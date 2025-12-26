package buffer

import (
	"database/sql"
	"encoding/json"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/probe-system/agent/pkg/protocol"
)

type Buffer struct {
	db *sql.DB
	mu sync.Mutex
}

func NewBuffer(path string) (*Buffer, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	// Create table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS buffered_metrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			data TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		db.Close()
		return nil, err
	}

	return &Buffer{db: db}, nil
}

func (b *Buffer) Store(metrics *protocol.MetricsPayload) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	data, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	_, err = b.db.Exec(`INSERT INTO buffered_metrics (data) VALUES (?)`, string(data))
	return err
}

func (b *Buffer) GetAll() ([]*protocol.MetricsPayload, []int64, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	rows, err := b.db.Query(`SELECT id, data FROM buffered_metrics ORDER BY timestamp ASC LIMIT 100`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var metrics []*protocol.MetricsPayload
	var ids []int64

	for rows.Next() {
		var id int64
		var data string
		if err := rows.Scan(&id, &data); err != nil {
			continue
		}

		var m protocol.MetricsPayload
		if err := json.Unmarshal([]byte(data), &m); err != nil {
			continue
		}

		metrics = append(metrics, &m)
		ids = append(ids, id)
	}

	return metrics, ids, nil
}

func (b *Buffer) Delete(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	for _, id := range ids {
		b.db.Exec(`DELETE FROM buffered_metrics WHERE id = ?`, id)
	}

	return nil
}

func (b *Buffer) Count() (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	var count int
	err := b.db.QueryRow(`SELECT COUNT(*) FROM buffered_metrics`).Scan(&count)
	return count, err
}

func (b *Buffer) Cleanup(maxAge time.Duration) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	_, err := b.db.Exec(`DELETE FROM buffered_metrics WHERE timestamp < ?`, cutoff)
	return err
}

func (b *Buffer) Close() error {
	return b.db.Close()
}

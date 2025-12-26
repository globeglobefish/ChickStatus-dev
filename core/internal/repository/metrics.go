package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/probe-system/core/internal/models"
)

type MetricsRepository struct {
	db *DB
}

func NewMetricsRepository(db *DB) *MetricsRepository {
	return &MetricsRepository{db: db}
}

func (r *MetricsRepository) Store(ctx context.Context, metrics *models.Metrics) error {
	memoryJSON, _ := json.Marshal(metrics.Memory)
	disksJSON, _ := json.Marshal(metrics.Disks)
	networkJSON, _ := json.Marshal(metrics.Network)

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO metrics (id, agent_id, cpu, memory, disks, network, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, metrics.ID, metrics.AgentID, metrics.CPU, string(memoryJSON),
		string(disksJSON), string(networkJSON), metrics.Timestamp)

	return err
}

func (r *MetricsRepository) GetLatest(ctx context.Context, agentID string) (*models.Metrics, error) {
	metrics := &models.Metrics{}
	var memoryJSON, disksJSON, networkJSON string

	err := r.db.QueryRowContext(ctx, `
		SELECT id, agent_id, cpu, memory, disks, network, timestamp
		FROM metrics WHERE agent_id = ? ORDER BY timestamp DESC LIMIT 1
	`, agentID).Scan(&metrics.ID, &metrics.AgentID, &metrics.CPU,
		&memoryJSON, &disksJSON, &networkJSON, &metrics.Timestamp)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(memoryJSON), &metrics.Memory)
	json.Unmarshal([]byte(disksJSON), &metrics.Disks)
	json.Unmarshal([]byte(networkJSON), &metrics.Network)

	return metrics, nil
}

func (r *MetricsRepository) GetHistory(ctx context.Context, agentID string, from, to time.Time) ([]*models.Metrics, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, agent_id, cpu, memory, disks, network, timestamp
		FROM metrics 
		WHERE agent_id = ? AND timestamp >= ? AND timestamp <= ?
		ORDER BY timestamp ASC
	`, agentID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []*models.Metrics{}
	for rows.Next() {
		metrics := &models.Metrics{}
		var memoryJSON, disksJSON, networkJSON string

		if err := rows.Scan(&metrics.ID, &metrics.AgentID, &metrics.CPU,
			&memoryJSON, &disksJSON, &networkJSON, &metrics.Timestamp); err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(memoryJSON), &metrics.Memory)
		json.Unmarshal([]byte(disksJSON), &metrics.Disks)
		json.Unmarshal([]byte(networkJSON), &metrics.Network)

		result = append(result, metrics)
	}

	return result, nil
}

func (r *MetricsRepository) Cleanup(ctx context.Context, retentionDays int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM metrics WHERE timestamp < ?
	`, cutoff)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *MetricsRepository) GetAggregated(ctx context.Context, agentID string, from, to time.Time, interval string) ([]*models.Metrics, error) {
	// For simplicity, return raw data - aggregation can be done in service layer
	return r.GetHistory(ctx, agentID, from, to)
}

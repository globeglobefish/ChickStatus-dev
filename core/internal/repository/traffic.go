package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/probe-system/core/internal/models"
)

type TrafficRepository struct {
	db *DB
}

func NewTrafficRepository(db *DB) *TrafficRepository {
	return &TrafficRepository{db: db}
}

func (r *TrafficRepository) CreateCycle(ctx context.Context, cycle *models.BillingCycle) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO billing_cycles (id, agent_id, start_date, duration, limit_bytes, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, cycle.ID, cycle.AgentID, cycle.StartDate, cycle.Duration, cycle.Limit, cycle.CreatedAt)
	return err
}

func (r *TrafficRepository) GetCycleByAgent(ctx context.Context, agentID string) (*models.BillingCycle, error) {
	cycle := &models.BillingCycle{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, agent_id, start_date, duration, limit_bytes, created_at
		FROM billing_cycles WHERE agent_id = ?
	`, agentID).Scan(&cycle.ID, &cycle.AgentID, &cycle.StartDate, 
		&cycle.Duration, &cycle.Limit, &cycle.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return cycle, nil
}

func (r *TrafficRepository) UpdateCycleStart(ctx context.Context, cycleID string, newStart time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE billing_cycles SET start_date = ? WHERE id = ?
	`, newStart, cycleID)
	return err
}

func (r *TrafficRepository) RecordTraffic(ctx context.Context, record *models.TrafficRecord) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO traffic_records (id, cycle_id, agent_id, bytes_sent, bytes_recv, timestamp)
		VALUES (?, ?, ?, ?, ?, ?)
	`, record.ID, record.CycleID, record.AgentID, record.BytesSent, record.BytesRecv, record.Timestamp)
	return err
}

func (r *TrafficRepository) GetCycleTraffic(ctx context.Context, cycleID string) (bytesSent, bytesRecv uint64, err error) {
	err = r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(bytes_sent), 0), COALESCE(SUM(bytes_recv), 0)
		FROM traffic_records WHERE cycle_id = ?
	`, cycleID).Scan(&bytesSent, &bytesRecv)
	return
}

func (r *TrafficRepository) GetTrafficStats(ctx context.Context, agentID string) (*models.TrafficStats, error) {
	cycle, err := r.GetCycleByAgent(ctx, agentID)
	if err != nil {
		return nil, err
	}
	if cycle == nil {
		return nil, nil
	}

	bytesSent, bytesRecv, err := r.GetCycleTraffic(ctx, cycle.ID)
	if err != nil {
		return nil, err
	}

	totalBytes := bytesSent + bytesRecv
	var percent float64
	if cycle.Limit > 0 {
		percent = float64(totalBytes) / float64(cycle.Limit) * 100
	}

	cycleEnd := cycle.StartDate.AddDate(0, 0, cycle.Duration)

	return &models.TrafficStats{
		CycleID:    cycle.ID,
		AgentID:    agentID,
		BytesSent:  bytesSent,
		BytesRecv:  bytesRecv,
		TotalBytes: totalBytes,
		Limit:      cycle.Limit,
		Percent:    percent,
		CycleStart: cycle.StartDate,
		CycleEnd:   cycleEnd,
	}, nil
}

func (r *TrafficRepository) DeleteCycleRecords(ctx context.Context, cycleID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM traffic_records WHERE cycle_id = ?`, cycleID)
	return err
}

func (r *TrafficRepository) GetExpiredCycles(ctx context.Context) ([]*models.BillingCycle, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, agent_id, start_date, duration, limit_bytes, created_at
		FROM billing_cycles
		WHERE date(start_date, '+' || duration || ' days') <= date('now')
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cycles := []*models.BillingCycle{}
	for rows.Next() {
		cycle := &models.BillingCycle{}
		if err := rows.Scan(&cycle.ID, &cycle.AgentID, &cycle.StartDate,
			&cycle.Duration, &cycle.Limit, &cycle.CreatedAt); err != nil {
			return nil, err
		}
		cycles = append(cycles, cycle)
	}
	return cycles, nil
}

func (r *TrafficRepository) ArchiveCycle(ctx context.Context, cycleID string, bytesSent, bytesRecv uint64) error {
	// For now, just delete old records - could store in archive table if needed
	return r.DeleteCycleRecords(ctx, cycleID)
}

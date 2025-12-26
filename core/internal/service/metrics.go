package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/probe-system/core/internal/models"
	"github.com/probe-system/core/internal/repository"
)

type MetricServiceImpl struct {
	repo *repository.MetricsRepository
}

func NewMetricService(repo *repository.MetricsRepository) *MetricServiceImpl {
	return &MetricServiceImpl{repo: repo}
}

func (s *MetricServiceImpl) Store(ctx context.Context, agentID string, metrics *models.Metrics) error {
	metrics.ID = uuid.New().String()
	metrics.AgentID = agentID
	metrics.Timestamp = time.Now()
	return s.repo.Store(ctx, metrics)
}

func (s *MetricServiceImpl) GetLatest(ctx context.Context, agentID string) (*models.Metrics, error) {
	return s.repo.GetLatest(ctx, agentID)
}

func (s *MetricServiceImpl) GetHistory(ctx context.Context, agentID string, from, to time.Time) ([]*models.Metrics, error) {
	return s.repo.GetHistory(ctx, agentID, from, to)
}

func (s *MetricServiceImpl) Cleanup(ctx context.Context, retentionDays int) (int64, error) {
	return s.repo.Cleanup(ctx, retentionDays)
}

// Traffic Service
type TrafficServiceImpl struct {
	repo *repository.TrafficRepository
}

func NewTrafficService(repo *repository.TrafficRepository) *TrafficServiceImpl {
	return &TrafficServiceImpl{repo: repo}
}

func (s *TrafficServiceImpl) RecordTraffic(ctx context.Context, agentID string, bytesSent, bytesRecv uint64) error {
	cycle, err := s.repo.GetCycleByAgent(ctx, agentID)
	if err != nil {
		return err
	}

	if cycle == nil {
		// Create default cycle (30 days, no limit)
		cycle = &models.BillingCycle{
			ID:        uuid.New().String(),
			AgentID:   agentID,
			StartDate: time.Now(),
			Duration:  30,
			Limit:     0,
			CreatedAt: time.Now(),
		}
		if err := s.repo.CreateCycle(ctx, cycle); err != nil {
			return err
		}
	}

	record := &models.TrafficRecord{
		ID:        uuid.New().String(),
		CycleID:   cycle.ID,
		AgentID:   agentID,
		BytesSent: bytesSent,
		BytesRecv: bytesRecv,
		Timestamp: time.Now(),
	}

	return s.repo.RecordTraffic(ctx, record)
}

func (s *TrafficServiceImpl) GetStats(ctx context.Context, agentID string) (*models.TrafficStats, error) {
	return s.repo.GetTrafficStats(ctx, agentID)
}

func (s *TrafficServiceImpl) ConfigureCycle(ctx context.Context, agentID string, startDate time.Time, durationDays int, limitBytes uint64) error {
	cycle := &models.BillingCycle{
		ID:        uuid.New().String(),
		AgentID:   agentID,
		StartDate: startDate,
		Duration:  durationDays,
		Limit:     limitBytes,
		CreatedAt: time.Now(),
	}
	return s.repo.CreateCycle(ctx, cycle)
}

func (s *TrafficServiceImpl) CheckAndResetCycles(ctx context.Context) error {
	cycles, err := s.repo.GetExpiredCycles(ctx)
	if err != nil {
		return err
	}

	for _, cycle := range cycles {
		// Get final stats before reset
		bytesSent, bytesRecv, err := s.repo.GetCycleTraffic(ctx, cycle.ID)
		if err != nil {
			continue
		}

		// Archive old data
		if err := s.repo.ArchiveCycle(ctx, cycle.ID, bytesSent, bytesRecv); err != nil {
			continue
		}

		// Reset cycle start date
		newStart := cycle.StartDate.AddDate(0, 0, cycle.Duration)
		if err := s.repo.UpdateCycleStart(ctx, cycle.ID, newStart); err != nil {
			continue
		}
	}

	return nil
}

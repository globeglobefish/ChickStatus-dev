package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/probe-system/core/internal/models"
	"github.com/probe-system/core/internal/repository"
)

type AlertServiceImpl struct {
	repo      *repository.AlertRepository
	notifiers []Notifier
}

func NewAlertService(repo *repository.AlertRepository) *AlertServiceImpl {
	return &AlertServiceImpl{
		repo:      repo,
		notifiers: []Notifier{},
	}
}

func (s *AlertServiceImpl) AddNotifier(n Notifier) {
	s.notifiers = append(s.notifiers, n)
}

func (s *AlertServiceImpl) CreateRule(ctx context.Context, rule *models.AlertRule) error {
	rule.ID = uuid.New().String()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	return s.repo.CreateRule(ctx, rule)
}

func (s *AlertServiceImpl) UpdateRule(ctx context.Context, rule *models.AlertRule) error {
	rule.UpdatedAt = time.Now()
	return s.repo.UpdateRule(ctx, rule)
}

func (s *AlertServiceImpl) DeleteRule(ctx context.Context, ruleID string) error {
	return s.repo.DeleteRule(ctx, ruleID)
}

func (s *AlertServiceImpl) GetRule(ctx context.Context, ruleID string) (*models.AlertRule, error) {
	return s.repo.GetRule(ctx, ruleID)
}

func (s *AlertServiceImpl) ListRules(ctx context.Context) ([]*models.AlertRule, error) {
	return s.repo.ListRules(ctx)
}

func (s *AlertServiceImpl) CheckAndTrigger(ctx context.Context, agentID string, metrics *models.Metrics) error {
	rules, err := s.repo.ListEnabledRules(ctx)
	if err != nil {
		return err
	}

	for _, rule := range rules {
		// Check if rule applies to this agent
		if len(rule.AgentIDs) > 0 && !contains(rule.AgentIDs, agentID) {
			continue
		}

		// Get metric value based on type
		value := s.getMetricValue(rule.MetricType, metrics)

		// Check threshold
		exceeded := rule.CheckThreshold(value)

		// Get existing firing alert
		existing, err := s.repo.GetFiringAlertByRuleAndAgent(ctx, rule.ID, agentID)
		if err != nil {
			continue
		}

		if exceeded {
			if existing == nil {
				// Check cooldown
				if s.isInCooldown(ctx, rule, agentID) {
					continue
				}

				// Create new alert
				alert := &models.Alert{
					ID:          uuid.New().String(),
					RuleID:      rule.ID,
					AgentID:     agentID,
					Status:      models.AlertStatusFiring,
					MetricType:  rule.MetricType,
					Value:       value,
					Threshold:   rule.Threshold,
					Message:     s.formatAlertMessage(rule, value),
					TriggeredAt: time.Now(),
				}

				if err := s.repo.CreateAlert(ctx, alert); err != nil {
					continue
				}

				// Send notifications
				s.notify(ctx, alert)
			}
		} else if existing != nil {
			// Resolve existing alert
			if err := s.repo.ResolveAlert(ctx, existing.ID); err != nil {
				continue
			}

			existing.Status = models.AlertStatusResolved
			now := time.Now()
			existing.ResolvedAt = &now

			// Send recovery notification
			s.notifyRecovery(ctx, existing)
		}
	}

	return nil
}

func (s *AlertServiceImpl) getMetricValue(metricType models.MetricType, metrics *models.Metrics) float64 {
	switch metricType {
	case models.MetricTypeCPU:
		return metrics.CPU
	case models.MetricTypeMemory:
		return metrics.Memory.Percent
	case models.MetricTypeDisk:
		if len(metrics.Disks) > 0 {
			// Return max disk usage
			var max float64
			for _, d := range metrics.Disks {
				if d.Percent > max {
					max = d.Percent
				}
			}
			return max
		}
		return 0
	default:
		return 0
	}
}

func (s *AlertServiceImpl) isInCooldown(ctx context.Context, rule *models.AlertRule, agentID string) bool {
	lastTime, err := s.repo.GetLastAlertTime(ctx, rule.ID, agentID)
	if err != nil || lastTime == nil {
		return false
	}

	cooldownEnd := lastTime.Add(time.Duration(rule.Cooldown) * time.Second)
	return time.Now().Before(cooldownEnd)
}

func (s *AlertServiceImpl) formatAlertMessage(rule *models.AlertRule, value float64) string {
	return fmt.Sprintf("[%s] %s: %.2f (threshold: %.2f)",
		rule.MetricType, rule.Name, value, rule.Threshold)
}

func (s *AlertServiceImpl) notify(ctx context.Context, alert *models.Alert) {
	for _, n := range s.notifiers {
		go n.Send(ctx, alert)
	}
}

func (s *AlertServiceImpl) notifyRecovery(ctx context.Context, alert *models.Alert) {
	for _, n := range s.notifiers {
		go n.SendRecovery(ctx, alert)
	}
}

func (s *AlertServiceImpl) ResolveAlert(ctx context.Context, alertID string) error {
	return s.repo.ResolveAlert(ctx, alertID)
}

func (s *AlertServiceImpl) GetActiveAlerts(ctx context.Context) ([]*models.Alert, error) {
	return s.repo.GetActiveAlerts(ctx)
}

func (s *AlertServiceImpl) GetAlertHistory(ctx context.Context, limit int) ([]*models.Alert, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.repo.GetAlertHistory(ctx, limit)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

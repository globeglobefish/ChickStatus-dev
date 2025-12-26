package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/probe-system/core/internal/models"
)

type AlertRepository struct {
	db *DB
}

func NewAlertRepository(db *DB) *AlertRepository {
	return &AlertRepository{db: db}
}

// Alert Rules
func (r *AlertRepository) CreateRule(ctx context.Context, rule *models.AlertRule) error {
	agentIDsJSON, _ := json.Marshal(rule.AgentIDs)

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO alert_rules (id, name, metric_type, operator, threshold, 
			duration_sec, cooldown_sec, agent_ids, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, rule.ID, rule.Name, rule.MetricType, rule.Operator, rule.Threshold,
		rule.Duration, rule.Cooldown, string(agentIDsJSON), rule.Enabled,
		rule.CreatedAt, rule.UpdatedAt)

	return err
}

func (r *AlertRepository) UpdateRule(ctx context.Context, rule *models.AlertRule) error {
	agentIDsJSON, _ := json.Marshal(rule.AgentIDs)

	_, err := r.db.ExecContext(ctx, `
		UPDATE alert_rules SET name=?, metric_type=?, operator=?, threshold=?, 
			duration_sec=?, cooldown_sec=?, agent_ids=?, enabled=?, updated_at=?
		WHERE id=?
	`, rule.Name, rule.MetricType, rule.Operator, rule.Threshold,
		rule.Duration, rule.Cooldown, string(agentIDsJSON), rule.Enabled,
		time.Now(), rule.ID)

	return err
}

func (r *AlertRepository) DeleteRule(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM alert_rules WHERE id = ?`, id)
	return err
}

func (r *AlertRepository) GetRule(ctx context.Context, id string) (*models.AlertRule, error) {
	rule := &models.AlertRule{}
	var agentIDsJSON string

	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, metric_type, operator, threshold, duration_sec, 
			cooldown_sec, agent_ids, enabled, created_at, updated_at
		FROM alert_rules WHERE id = ?
	`, id).Scan(&rule.ID, &rule.Name, &rule.MetricType, &rule.Operator, &rule.Threshold,
		&rule.Duration, &rule.Cooldown, &agentIDsJSON, &rule.Enabled,
		&rule.CreatedAt, &rule.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(agentIDsJSON), &rule.AgentIDs)
	return rule, nil
}

func (r *AlertRepository) ListRules(ctx context.Context) ([]*models.AlertRule, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, metric_type, operator, threshold, duration_sec, 
			cooldown_sec, agent_ids, enabled, created_at, updated_at
		FROM alert_rules ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := []*models.AlertRule{}
	for rows.Next() {
		rule := &models.AlertRule{}
		var agentIDsJSON string

		if err := rows.Scan(&rule.ID, &rule.Name, &rule.MetricType, &rule.Operator,
			&rule.Threshold, &rule.Duration, &rule.Cooldown, &agentIDsJSON,
			&rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(agentIDsJSON), &rule.AgentIDs)
		rules = append(rules, rule)
	}

	return rules, nil
}

func (r *AlertRepository) ListEnabledRules(ctx context.Context) ([]*models.AlertRule, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, metric_type, operator, threshold, duration_sec, 
			cooldown_sec, agent_ids, enabled, created_at, updated_at
		FROM alert_rules WHERE enabled = 1
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := []*models.AlertRule{}
	for rows.Next() {
		rule := &models.AlertRule{}
		var agentIDsJSON string

		if err := rows.Scan(&rule.ID, &rule.Name, &rule.MetricType, &rule.Operator,
			&rule.Threshold, &rule.Duration, &rule.Cooldown, &agentIDsJSON,
			&rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(agentIDsJSON), &rule.AgentIDs)
		rules = append(rules, rule)
	}

	return rules, nil
}

// Alerts
func (r *AlertRepository) CreateAlert(ctx context.Context, alert *models.Alert) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO alerts (id, rule_id, agent_id, status, metric_type, value, 
			threshold, message, triggered_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, alert.ID, alert.RuleID, alert.AgentID, alert.Status, alert.MetricType,
		alert.Value, alert.Threshold, alert.Message, alert.TriggeredAt)

	return err
}

func (r *AlertRepository) ResolveAlert(ctx context.Context, id string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx, `
		UPDATE alerts SET status = 'resolved', resolved_at = ? WHERE id = ?
	`, now, id)
	return err
}

func (r *AlertRepository) GetActiveAlerts(ctx context.Context) ([]*models.Alert, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT a.id, a.rule_id, r.name, a.agent_id, ag.custom_name, a.status, 
			a.metric_type, a.value, a.threshold, a.message, a.triggered_at, a.resolved_at
		FROM alerts a
		LEFT JOIN alert_rules r ON a.rule_id = r.id
		LEFT JOIN agents ag ON a.agent_id = ag.id
		WHERE a.status = 'firing'
		ORDER BY a.triggered_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAlerts(rows)
}

func (r *AlertRepository) GetAlertHistory(ctx context.Context, limit int) ([]*models.Alert, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT a.id, a.rule_id, r.name, a.agent_id, ag.custom_name, a.status, 
			a.metric_type, a.value, a.threshold, a.message, a.triggered_at, a.resolved_at
		FROM alerts a
		LEFT JOIN alert_rules r ON a.rule_id = r.id
		LEFT JOIN agents ag ON a.agent_id = ag.id
		ORDER BY a.triggered_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAlerts(rows)
}

func (r *AlertRepository) scanAlerts(rows *sql.Rows) ([]*models.Alert, error) {
	alerts := []*models.Alert{}
	for rows.Next() {
		alert := &models.Alert{}
		var ruleName, agentName sql.NullString
		var resolvedAt sql.NullTime

		if err := rows.Scan(&alert.ID, &alert.RuleID, &ruleName, &alert.AgentID,
			&agentName, &alert.Status, &alert.MetricType, &alert.Value,
			&alert.Threshold, &alert.Message, &alert.TriggeredAt, &resolvedAt); err != nil {
			return nil, err
		}

		if ruleName.Valid {
			alert.RuleName = ruleName.String
		}
		if agentName.Valid {
			alert.AgentName = agentName.String
		}
		if resolvedAt.Valid {
			alert.ResolvedAt = &resolvedAt.Time
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (r *AlertRepository) GetFiringAlertByRuleAndAgent(ctx context.Context, ruleID, agentID string) (*models.Alert, error) {
	alert := &models.Alert{}
	var resolvedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT id, rule_id, agent_id, status, metric_type, value, threshold, 
			message, triggered_at, resolved_at
		FROM alerts 
		WHERE rule_id = ? AND agent_id = ? AND status = 'firing'
		ORDER BY triggered_at DESC LIMIT 1
	`, ruleID, agentID).Scan(&alert.ID, &alert.RuleID, &alert.AgentID, &alert.Status,
		&alert.MetricType, &alert.Value, &alert.Threshold, &alert.Message,
		&alert.TriggeredAt, &resolvedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if resolvedAt.Valid {
		alert.ResolvedAt = &resolvedAt.Time
	}

	return alert, nil
}

func (r *AlertRepository) GetLastAlertTime(ctx context.Context, ruleID, agentID string) (*time.Time, error) {
	var triggeredAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		SELECT triggered_at FROM alerts 
		WHERE rule_id = ? AND agent_id = ?
		ORDER BY triggered_at DESC LIMIT 1
	`, ruleID, agentID).Scan(&triggeredAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if triggeredAt.Valid {
		return &triggeredAt.Time, nil
	}
	return nil, nil
}

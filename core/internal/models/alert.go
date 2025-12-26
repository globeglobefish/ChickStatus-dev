package models

import (
	"time"
)

type AlertStatus string

const (
	AlertStatusFiring   AlertStatus = "firing"
	AlertStatusResolved AlertStatus = "resolved"
)

type MetricType string

const (
	MetricTypeCPU     MetricType = "cpu"
	MetricTypeMemory  MetricType = "memory"
	MetricTypeDisk    MetricType = "disk"
	MetricTypeTraffic MetricType = "traffic"
)

type Operator string

const (
	OperatorGT Operator = "gt"
	OperatorLT Operator = "lt"
	OperatorEQ Operator = "eq"
)

type AlertRule struct {
	ID           string     `json:"id" db:"id"`
	Name         string     `json:"name" db:"name"`
	MetricType   MetricType `json:"metric_type" db:"metric_type"`
	Operator     Operator   `json:"operator" db:"operator"`
	Threshold    float64    `json:"threshold" db:"threshold"`
	Duration     int        `json:"duration" db:"duration_sec"`
	Cooldown     int        `json:"cooldown" db:"cooldown_sec"`
	AgentIDs     []string   `json:"agent_ids" db:"-"`
	AgentIDsJSON string     `json:"-" db:"agent_ids"`
	Enabled      bool       `json:"enabled" db:"enabled"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

type Alert struct {
	ID          string      `json:"id" db:"id"`
	RuleID      string      `json:"rule_id" db:"rule_id"`
	RuleName    string      `json:"rule_name" db:"-"`
	AgentID     string      `json:"agent_id" db:"agent_id"`
	AgentName   string      `json:"agent_name" db:"-"`
	Status      AlertStatus `json:"status" db:"status"`
	MetricType  MetricType  `json:"metric_type" db:"metric_type"`
	Value       float64     `json:"value" db:"value"`
	Threshold   float64     `json:"threshold" db:"threshold"`
	Message     string      `json:"message" db:"message"`
	TriggeredAt time.Time   `json:"triggered_at" db:"triggered_at"`
	ResolvedAt  *time.Time  `json:"resolved_at" db:"resolved_at"`
}

func (r *AlertRule) CheckThreshold(value float64) bool {
	switch r.Operator {
	case OperatorGT:
		return value > r.Threshold
	case OperatorLT:
		return value < r.Threshold
	case OperatorEQ:
		return value == r.Threshold
	default:
		return false
	}
}

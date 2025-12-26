package models

import (
	"time"
)

type TaskType string

const (
	TaskTypePing   TaskType = "ping"
	TaskTypeScript TaskType = "script"
)

type TaskStatus string

const (
	TaskStatusPending  TaskStatus = "pending"
	TaskStatusRunning  TaskStatus = "running"
	TaskStatusComplete TaskStatus = "complete"
	TaskStatusFailed   TaskStatus = "failed"
	TaskStatusCanceled TaskStatus = "canceled"
)

type Task struct {
	ID         string            `json:"id" db:"id"`
	Type       TaskType          `json:"type" db:"type"`
	Name       string            `json:"name" db:"name"`
	Target     string            `json:"target" db:"target"`
	ScriptID   string            `json:"script_id" db:"script_id"`
	Params     map[string]string `json:"params" db:"-"`
	ParamsJSON string            `json:"-" db:"params"`
	Interval   int               `json:"interval" db:"interval_sec"`
	Timeout    int               `json:"timeout" db:"timeout_sec"`
	Status     TaskStatus        `json:"status" db:"status"`
	AgentIDs   []string          `json:"agent_ids" db:"-"`
	AgentIDsJSON string          `json:"-" db:"agent_ids"`
	CreatedAt  time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at" db:"updated_at"`
}

type TaskResult struct {
	ID        string    `json:"id" db:"id"`
	TaskID    string    `json:"task_id" db:"task_id"`
	AgentID   string    `json:"agent_id" db:"agent_id"`
	Success   bool      `json:"success" db:"success"`
	Output    string    `json:"output" db:"output"`
	Error     string    `json:"error" db:"error"`
	Duration  int64     `json:"duration" db:"duration_ms"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}

type PingResult struct {
	Target     string  `json:"target"`
	Latency    float64 `json:"latency"`
	PacketLoss float64 `json:"packet_loss"`
	Success    bool    `json:"success"`
	Error      string  `json:"error,omitempty"`
}

func (p *PingResult) Validate() bool {
	if p.Latency < 0 && p.Success {
		return false
	}
	if p.PacketLoss < 0 || p.PacketLoss > 100 {
		return false
	}
	if p.PacketLoss >= 100 && p.Success {
		return false
	}
	return true
}

type Script struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Content     string    `json:"content" db:"content"`
	Checksum    string    `json:"checksum" db:"checksum"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

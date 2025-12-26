package protocol

import (
	"encoding/json"
	"time"
)

const (
	MsgTypeHeartbeat   = "heartbeat"
	MsgTypeRegister    = "register"
	MsgTypeRegisterAck = "register_ack"
	MsgTypeMetrics     = "metrics"
	MsgTypeMetricsAck  = "metrics_ack"
	MsgTypeTaskAssign  = "task_assign"
	MsgTypeTaskAck     = "task_ack"
	MsgTypeTaskResult  = "task_result"
	MsgTypeConfig      = "config"
	MsgTypeError       = "error"
)

type Message struct {
	Type      string          `json:"type"`
	ID        string          `json:"id"`
	Timestamp int64           `json:"ts"`
	Payload   json.RawMessage `json:"payload"`
}

func NewMessage(msgType string, id string, payload interface{}) (*Message, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return &Message{
		Type:      msgType,
		ID:        id,
		Timestamp: time.Now().UnixMilli(),
		Payload:   data,
	}, nil
}

type RegisterPayload struct {
	Hostname string `json:"hostname"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Version  string `json:"version"`
	Token    string `json:"token"`
}

type RegisterAckPayload struct {
	AgentID string `json:"agent_id"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type MetricsPayload struct {
	CPU     float64      `json:"cpu"`
	Memory  MemoryStats  `json:"memory"`
	Disks   []DiskStats  `json:"disks"`
	Network NetworkStats `json:"network"`
}

type MemoryStats struct {
	Total     uint64  `json:"total"`
	Used      uint64  `json:"used"`
	Available uint64  `json:"available"`
	Percent   float64 `json:"percent"`
}

type DiskStats struct {
	Path      string  `json:"path"`
	Total     uint64  `json:"total"`
	Used      uint64  `json:"used"`
	Available uint64  `json:"available"`
	Percent   float64 `json:"percent"`
}

type NetworkStats struct {
	BytesSent     uint64 `json:"bytes_sent"`
	BytesRecv     uint64 `json:"bytes_recv"`
	BytesSentRate uint64 `json:"bytes_sent_rate"`
	BytesRecvRate uint64 `json:"bytes_recv_rate"`
}

type TaskAssignPayload struct {
	TaskID   string            `json:"task_id"`
	Type     string            `json:"type"`
	Target   string            `json:"target,omitempty"`
	ScriptID string            `json:"script_id,omitempty"`
	Params   map[string]string `json:"params,omitempty"`
	Interval int               `json:"interval,omitempty"`
	Timeout  int               `json:"timeout,omitempty"`
}

type TaskAckPayload struct {
	TaskID  string `json:"task_id"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type TaskResultPayload struct {
	TaskID   string `json:"task_id"`
	Success  bool   `json:"success"`
	Output   string `json:"output"`
	Error    string `json:"error,omitempty"`
	Duration int64  `json:"duration"`
}

type ConfigPayload struct {
	MetricInterval int `json:"metric_interval"`
}

type ErrorPayload struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

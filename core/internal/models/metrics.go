package models

import (
	"time"
)

type MemoryStats struct {
	Total     uint64  `json:"total"`
	Used      uint64  `json:"used"`
	Available uint64  `json:"available"`
	Percent   float64 `json:"percent"`
}

func (m *MemoryStats) Validate() bool {
	if m.Total == 0 {
		return false
	}
	if m.Used > m.Total {
		return false
	}
	if m.Percent < 0 || m.Percent > 100 {
		return false
	}
	return true
}

type DiskStats struct {
	Path      string  `json:"path"`
	Total     uint64  `json:"total"`
	Used      uint64  `json:"used"`
	Available uint64  `json:"available"`
	Percent   float64 `json:"percent"`
}

func (d *DiskStats) Validate() bool {
	if d.Total == 0 {
		return false
	}
	if d.Used > d.Total {
		return false
	}
	if d.Percent < 0 || d.Percent > 100 {
		return false
	}
	return true
}

type NetworkStats struct {
	BytesSent     uint64 `json:"bytes_sent"`
	BytesRecv     uint64 `json:"bytes_recv"`
	BytesSentRate uint64 `json:"bytes_sent_rate"`
	BytesRecvRate uint64 `json:"bytes_recv_rate"`
}

type Metrics struct {
	ID        string       `json:"id" db:"id"`
	AgentID   string       `json:"agent_id" db:"agent_id"`
	CPU       float64      `json:"cpu" db:"cpu"`
	Memory    MemoryStats  `json:"memory" db:"-"`
	MemoryJSON string      `json:"-" db:"memory"`
	Disks     []DiskStats  `json:"disks" db:"-"`
	DisksJSON string       `json:"-" db:"disks"`
	Network   NetworkStats `json:"network" db:"-"`
	NetworkJSON string     `json:"-" db:"network"`
	Timestamp time.Time    `json:"timestamp" db:"timestamp"`
}

func (m *Metrics) ValidateCPU() bool {
	return m.CPU >= 0 && m.CPU <= 100
}

type BillingCycle struct {
	ID        string    `json:"id" db:"id"`
	AgentID   string    `json:"agent_id" db:"agent_id"`
	StartDate time.Time `json:"start_date" db:"start_date"`
	Duration  int       `json:"duration" db:"duration"`
	Limit     uint64    `json:"limit" db:"limit_bytes"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type TrafficStats struct {
	CycleID    string    `json:"cycle_id"`
	AgentID    string    `json:"agent_id"`
	BytesSent  uint64    `json:"bytes_sent"`
	BytesRecv  uint64    `json:"bytes_recv"`
	TotalBytes uint64    `json:"total_bytes"`
	Limit      uint64    `json:"limit"`
	Percent    float64   `json:"percent"`
	CycleStart time.Time `json:"cycle_start"`
	CycleEnd   time.Time `json:"cycle_end"`
}

type TrafficRecord struct {
	ID        string    `json:"id" db:"id"`
	CycleID   string    `json:"cycle_id" db:"cycle_id"`
	AgentID   string    `json:"agent_id" db:"agent_id"`
	BytesSent uint64    `json:"bytes_sent" db:"bytes_sent"`
	BytesRecv uint64    `json:"bytes_recv" db:"bytes_recv"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}

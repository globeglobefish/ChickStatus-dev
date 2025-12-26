package executor

import (
	"context"
	"fmt"
	"net"
	"time"
)

type PingResult struct {
	Target     string  `json:"target"`
	Latency    float64 `json:"latency"`
	PacketLoss float64 `json:"packet_loss"`
	Success    bool    `json:"success"`
	Error      string  `json:"error,omitempty"`
}

type PingExecutor struct {
	timeout time.Duration
	count   int
}

func NewPingExecutor(timeout time.Duration, count int) *PingExecutor {
	if count <= 0 {
		count = 4
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &PingExecutor{
		timeout: timeout,
		count:   count,
	}
}

func (e *PingExecutor) Execute(ctx context.Context, target string) (*PingResult, error) {
	result := &PingResult{
		Target: target,
	}

	// Use TCP connection as a simple ping alternative (works without root)
	var successCount int
	var totalLatency float64

	for i := 0; i < e.count; i++ {
		select {
		case <-ctx.Done():
			result.Error = "context cancelled"
			return result, ctx.Err()
		default:
		}

		latency, err := e.tcpPing(target)
		if err == nil {
			successCount++
			totalLatency += latency
		}

		if i < e.count-1 {
			time.Sleep(200 * time.Millisecond)
		}
	}

	if successCount > 0 {
		result.Success = true
		result.Latency = totalLatency / float64(successCount)
	} else {
		result.Success = false
		result.Latency = -1
		result.Error = "all pings failed"
	}

	result.PacketLoss = float64(e.count-successCount) / float64(e.count) * 100

	return result, nil
}

func (e *PingExecutor) tcpPing(target string) (float64, error) {
	// Add default port if not specified
	host, port, err := net.SplitHostPort(target)
	if err != nil {
		// No port specified, try common ports
		host = target
		port = "80"
	}

	addr := net.JoinHostPort(host, port)

	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr, e.timeout)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	latency := time.Since(start).Seconds() * 1000 // ms
	return latency, nil
}

func (e *PingExecutor) FormatOutput(result *PingResult) string {
	if result.Success {
		return fmt.Sprintf("PING %s: latency=%.2fms, packet_loss=%.1f%%",
			result.Target, result.Latency, result.PacketLoss)
	}
	return fmt.Sprintf("PING %s: FAILED - %s", result.Target, result.Error)
}

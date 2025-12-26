package collector

import (
	"time"

	"github.com/probe-system/agent/pkg/protocol"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type Collector struct {
	interval         time.Duration
	lastNetStats     *net.IOCountersStat
	lastNetStatsTime time.Time
}

func NewCollector(interval time.Duration) *Collector {
	return &Collector{
		interval: interval,
	}
}

func (c *Collector) Collect() (*protocol.MetricsPayload, error) {
	metrics := &protocol.MetricsPayload{}

	// CPU
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		metrics.CPU = cpuPercent[0]
	}

	// Memory
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		metrics.Memory = protocol.MemoryStats{
			Total:     memInfo.Total,
			Used:      memInfo.Used,
			Available: memInfo.Available,
			Percent:   memInfo.UsedPercent,
		}
	}

	// Disk
	partitions, err := disk.Partitions(false)
	if err == nil {
		for _, p := range partitions {
			usage, err := disk.Usage(p.Mountpoint)
			if err != nil {
				continue
			}

			metrics.Disks = append(metrics.Disks, protocol.DiskStats{
				Path:      p.Mountpoint,
				Total:     usage.Total,
				Used:      usage.Used,
				Available: usage.Free,
				Percent:   usage.UsedPercent,
			})
		}
	}

	// Network
	netStats, err := net.IOCounters(false)
	if err == nil && len(netStats) > 0 {
		current := &netStats[0]
		now := time.Now()

		metrics.Network = protocol.NetworkStats{
			BytesSent: current.BytesSent,
			BytesRecv: current.BytesRecv,
		}

		// Calculate rate
		if c.lastNetStats != nil {
			duration := now.Sub(c.lastNetStatsTime).Seconds()
			if duration > 0 {
				metrics.Network.BytesSentRate = uint64(float64(current.BytesSent-c.lastNetStats.BytesSent) / duration)
				metrics.Network.BytesRecvRate = uint64(float64(current.BytesRecv-c.lastNetStats.BytesRecv) / duration)
			}
		}

		c.lastNetStats = current
		c.lastNetStatsTime = now
	}

	return metrics, nil
}

func (c *Collector) GetInterval() time.Duration {
	return c.interval
}

func (c *Collector) SetInterval(interval time.Duration) {
	c.interval = interval
}

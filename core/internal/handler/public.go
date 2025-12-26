package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/probe-system/core/internal/service"
)

type PublicHandler struct {
	agentSvc   service.AgentService
	metricSvc  service.MetricService
	trafficSvc service.TrafficService
}

func NewPublicHandler(
	agentSvc service.AgentService,
	metricSvc service.MetricService,
	trafficSvc service.TrafficService,
) *PublicHandler {
	return &PublicHandler{
		agentSvc:   agentSvc,
		metricSvc:  metricSvc,
		trafficSvc: trafficSvc,
	}
}

type PublicAgentResponse struct {
	ID          string                 `json:"id"`
	CustomName  string                 `json:"name"`
	Status      string                 `json:"status"`
	GroupID     *string                `json:"group_id"`
	Location    *PublicLocationResponse `json:"location"`
	Metrics     *PublicMetricsResponse `json:"metrics,omitempty"`
	Traffic     *PublicTrafficResponse `json:"traffic,omitempty"`
}

type PublicLocationResponse struct {
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
	Region      string `json:"region"`
}

type PublicMetricsResponse struct {
	CPU           float64 `json:"cpu"`
	MemoryPercent float64 `json:"memory_percent"`
	DiskPercent   float64 `json:"disk_percent"`
}

type PublicTrafficResponse struct {
	Used    uint64  `json:"used"`
	Limit   uint64  `json:"limit"`
	Percent float64 `json:"percent"`
}

func (h *PublicHandler) ListAgents(c *gin.Context) {
	ctx := c.Request.Context()

	agents, err := h.agentSvc.ListPublic(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]PublicAgentResponse, 0, len(agents))

	for _, agent := range agents {
		resp := PublicAgentResponse{
			ID:         agent.ID,
			CustomName: agent.CustomName,
			Status:     string(agent.Status),
			GroupID:    agent.GroupID,
		}

		// Add location (without IP)
		if agent.Location != nil {
			resp.Location = &PublicLocationResponse{
				Country:     agent.Location.Country,
				CountryCode: agent.Location.CountryCode,
				Region:      agent.Location.Region,
			}
		}

		// Get latest metrics
		if metrics, err := h.metricSvc.GetLatest(ctx, agent.ID); err == nil && metrics != nil {
			var maxDisk float64
			for _, d := range metrics.Disks {
				if d.Percent > maxDisk {
					maxDisk = d.Percent
				}
			}

			resp.Metrics = &PublicMetricsResponse{
				CPU:           metrics.CPU,
				MemoryPercent: metrics.Memory.Percent,
				DiskPercent:   maxDisk,
			}
		}

		// Get traffic stats
		if traffic, err := h.trafficSvc.GetStats(ctx, agent.ID); err == nil && traffic != nil {
			resp.Traffic = &PublicTrafficResponse{
				Used:    traffic.TotalBytes,
				Limit:   traffic.Limit,
				Percent: traffic.Percent,
			}
		}

		response = append(response, resp)
	}

	c.JSON(http.StatusOK, response)
}

func (h *PublicHandler) GetAgent(c *gin.Context) {
	ctx := c.Request.Context()
	agentID := c.Param("id")

	agent, err := h.agentSvc.GetByID(ctx, agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if agent == nil || !agent.PublicVisible {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	resp := PublicAgentResponse{
		ID:         agent.ID,
		CustomName: agent.CustomName,
		Status:     string(agent.Status),
		GroupID:    agent.GroupID,
	}

	if agent.Location != nil {
		resp.Location = &PublicLocationResponse{
			Country:     agent.Location.Country,
			CountryCode: agent.Location.CountryCode,
			Region:      agent.Location.Region,
		}
	}

	if metrics, err := h.metricSvc.GetLatest(ctx, agent.ID); err == nil && metrics != nil {
		var maxDisk float64
		for _, d := range metrics.Disks {
			if d.Percent > maxDisk {
				maxDisk = d.Percent
			}
		}

		resp.Metrics = &PublicMetricsResponse{
			CPU:           metrics.CPU,
			MemoryPercent: metrics.Memory.Percent,
			DiskPercent:   maxDisk,
		}
	}

	if traffic, err := h.trafficSvc.GetStats(ctx, agent.ID); err == nil && traffic != nil {
		resp.Traffic = &PublicTrafficResponse{
			Used:    traffic.TotalBytes,
			Limit:   traffic.Limit,
			Percent: traffic.Percent,
		}
	}

	c.JSON(http.StatusOK, resp)
}

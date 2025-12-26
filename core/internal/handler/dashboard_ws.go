package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/probe-system/core/internal/service"
)

type DashboardWSHandler struct {
	upgrader   websocket.Upgrader
	clients    sync.Map
	agentSvc   service.AgentService
	metricSvc  service.MetricService
	trafficSvc service.TrafficService
}

func NewDashboardWSHandler(
	agentSvc service.AgentService,
	metricSvc service.MetricService,
	trafficSvc service.TrafficService,
) *DashboardWSHandler {
	h := &DashboardWSHandler{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		agentSvc:   agentSvc,
		metricSvc:  metricSvc,
		trafficSvc: trafficSvc,
	}

	// Start broadcast loop
	go h.broadcastLoop()

	return h
}

func (h *DashboardWSHandler) ServeWS(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Dashboard WebSocket upgrade failed: %v", err)
		return
	}

	clientID := time.Now().UnixNano()
	h.clients.Store(clientID, conn)

	defer func() {
		h.clients.Delete(clientID)
		conn.Close()
	}()

	// Send initial data
	h.sendInitialData(conn)

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (h *DashboardWSHandler) sendInitialData(conn *websocket.Conn) {
	ctx := conn.LocalAddr()
	_ = ctx

	agents, err := h.agentSvc.ListPublic(nil)
	if err != nil {
		return
	}

	data := make([]map[string]interface{}, 0, len(agents))
	for _, agent := range agents {
		item := map[string]interface{}{
			"id":     agent.ID,
			"name":   agent.CustomName,
			"status": agent.Status,
		}

		if agent.Location != nil {
			item["location"] = map[string]string{
				"country":      agent.Location.Country,
				"country_code": agent.Location.CountryCode,
			}
		}

		if metrics, err := h.metricSvc.GetLatest(nil, agent.ID); err == nil && metrics != nil {
			item["metrics"] = map[string]interface{}{
				"cpu":    metrics.CPU,
				"memory": metrics.Memory.Percent,
			}
		}

		if traffic, err := h.trafficSvc.GetStats(nil, agent.ID); err == nil && traffic != nil {
			item["traffic"] = map[string]interface{}{
				"used":    traffic.TotalBytes,
				"limit":   traffic.Limit,
				"percent": traffic.Percent,
			}
		}

		data = append(data, item)
	}

	msg := map[string]interface{}{
		"type": "init",
		"data": data,
	}

	jsonData, _ := json.Marshal(msg)
	conn.WriteMessage(websocket.TextMessage, jsonData)
}

func (h *DashboardWSHandler) broadcastLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		h.broadcastUpdate()
	}
}

func (h *DashboardWSHandler) broadcastUpdate() {
	agents, err := h.agentSvc.ListPublic(nil)
	if err != nil {
		return
	}

	updates := make([]map[string]interface{}, 0)

	for _, agent := range agents {
		update := map[string]interface{}{
			"id":     agent.ID,
			"status": agent.Status,
		}

		if metrics, err := h.metricSvc.GetLatest(nil, agent.ID); err == nil && metrics != nil {
			update["metrics"] = map[string]interface{}{
				"cpu":    metrics.CPU,
				"memory": metrics.Memory.Percent,
			}
		}

		if traffic, err := h.trafficSvc.GetStats(nil, agent.ID); err == nil && traffic != nil {
			update["traffic"] = map[string]interface{}{
				"used":    traffic.TotalBytes,
				"limit":   traffic.Limit,
				"percent": traffic.Percent,
			}
		}

		updates = append(updates, update)
	}

	msg := map[string]interface{}{
		"type": "update",
		"data": updates,
	}

	jsonData, _ := json.Marshal(msg)

	h.clients.Range(func(key, value interface{}) bool {
		conn := value.(*websocket.Conn)
		conn.WriteMessage(websocket.TextMessage, jsonData)
		return true
	})
}

func (h *DashboardWSHandler) BroadcastAlert(alert map[string]interface{}) {
	msg := map[string]interface{}{
		"type": "alert",
		"data": alert,
	}

	jsonData, _ := json.Marshal(msg)

	h.clients.Range(func(key, value interface{}) bool {
		conn := value.(*websocket.Conn)
		conn.WriteMessage(websocket.TextMessage, jsonData)
		return true
	})
}

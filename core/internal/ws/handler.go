package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/probe-system/core/internal/models"
	"github.com/probe-system/core/internal/service"
	"github.com/probe-system/core/pkg/protocol"
)

type Handler struct {
	hub          *Hub
	upgrader     websocket.Upgrader
	agentToken   string
	agentSvc     service.AgentService
	metricSvc    service.MetricService
	trafficSvc   service.TrafficService
	taskSvc      service.TaskService
	alertSvc     service.AlertService
}

func NewHandler(hub *Hub, agentToken string) *Handler {
	h := &Handler{
		hub:        hub,
		agentToken: agentToken,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}

	hub.SetMessageHandler(h.handleMessage)
	hub.SetDisconnectHandler(h.handleDisconnect)

	return h
}

func (h *Handler) SetServices(
	agentSvc service.AgentService,
	metricSvc service.MetricService,
	trafficSvc service.TrafficService,
	taskSvc service.TaskService,
	alertSvc service.AlertService,
) {
	h.agentSvc = agentSvc
	h.metricSvc = metricSvc
	h.trafficSvc = trafficSvc
	h.taskSvc = taskSvc
	h.alertSvc = alertSvc
}

func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Get client IP
	clientIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = strings.Split(forwarded, ",")[0]
	}
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		clientIP = realIP
	}

	// Extract IP without port
	if idx := strings.LastIndex(clientIP, ":"); idx != -1 {
		clientIP = clientIP[:idx]
	}

	// Wait for register message
	_, message, err := conn.ReadMessage()
	if err != nil {
		conn.Close()
		return
	}

	var msg protocol.Message
	if err := json.Unmarshal(message, &msg); err != nil {
		conn.Close()
		return
	}

	if msg.Type != protocol.MsgTypeRegister {
		conn.Close()
		return
	}

	var payload protocol.RegisterPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		conn.Close()
		return
	}

	// Validate token
	if h.agentToken != "" && payload.Token != h.agentToken {
		log.Printf("Invalid token from %s", clientIP)
		h.sendError(conn, "invalid token")
		conn.Close()
		return
	}

	// Register agent
	ctx := context.Background()
	agent, err := h.agentSvc.Register(ctx, &service.RegisterRequest{
		Hostname: payload.Hostname,
		IP:       clientIP,
		OS:       payload.OS,
		Arch:     payload.Arch,
		Version:  payload.Version,
	})
	if err != nil {
		log.Printf("Agent registration failed: %v", err)
		h.sendError(conn, "registration failed")
		conn.Close()
		return
	}

	// Send ack
	ack, _ := protocol.NewMessage(protocol.MsgTypeRegisterAck, uuid.New().String(), protocol.RegisterAckPayload{
		AgentID: agent.ID,
		Success: true,
	})
	data, _ := json.Marshal(ack)
	conn.WriteMessage(websocket.TextMessage, data)

	// Create agent connection
	agentConn := &AgentConn{
		ID:       agent.ID,
		Conn:     conn,
		SendChan: make(chan []byte, 256),
		hub:      h.hub,
	}

	h.hub.register <- agentConn

	go agentConn.WritePump()
	go agentConn.ReadPump()

	// Send pending tasks
	h.sendPendingTasks(agent.ID)
}

func (h *Handler) handleMessage(agentID string, msg *protocol.Message) {
	ctx := context.Background()

	switch msg.Type {
	case protocol.MsgTypeHeartbeat:
		h.agentSvc.UpdateLastSeen(ctx, agentID)

	case protocol.MsgTypeMetrics:
		var payload protocol.MetricsPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}

		metrics := &models.Metrics{
			CPU: payload.CPU,
			Memory: models.MemoryStats{
				Total:     payload.Memory.Total,
				Used:      payload.Memory.Used,
				Available: payload.Memory.Available,
				Percent:   payload.Memory.Percent,
			},
			Network: models.NetworkStats{
				BytesSent:     payload.Network.BytesSent,
				BytesRecv:     payload.Network.BytesRecv,
				BytesSentRate: payload.Network.BytesSentRate,
				BytesRecvRate: payload.Network.BytesRecvRate,
			},
		}

		for _, d := range payload.Disks {
			metrics.Disks = append(metrics.Disks, models.DiskStats{
				Path:      d.Path,
				Total:     d.Total,
				Used:      d.Used,
				Available: d.Available,
				Percent:   d.Percent,
			})
		}

		h.metricSvc.Store(ctx, agentID, metrics)
		h.agentSvc.UpdateLastSeen(ctx, agentID)

		// Record traffic
		h.trafficSvc.RecordTraffic(ctx, agentID, payload.Network.BytesSentRate, payload.Network.BytesRecvRate)

		// Check alerts
		h.alertSvc.CheckAndTrigger(ctx, agentID, metrics)

	case protocol.MsgTypeTaskResult:
		var payload protocol.TaskResultPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}

		result := &models.TaskResult{
			TaskID:   payload.TaskID,
			AgentID:  agentID,
			Success:  payload.Success,
			Output:   payload.Output,
			Error:    payload.Error,
			Duration: payload.Duration,
		}

		h.taskSvc.RecordResult(ctx, result)
	}
}

func (h *Handler) handleDisconnect(agentID string) {
	ctx := context.Background()
	h.agentSvc.UpdateStatus(ctx, agentID, models.AgentStatusOffline)
}

func (h *Handler) sendPendingTasks(agentID string) {
	ctx := context.Background()
	tasks, err := h.taskSvc.ListByAgent(ctx, agentID)
	if err != nil {
		return
	}

	for _, task := range tasks {
		payload := protocol.TaskAssignPayload{
			TaskID:   task.ID,
			Type:     string(task.Type),
			Target:   task.Target,
			ScriptID: task.ScriptID,
			Params:   task.Params,
			Interval: task.Interval,
			Timeout:  task.Timeout,
		}

		msg, _ := protocol.NewMessage(protocol.MsgTypeTaskAssign, uuid.New().String(), payload)
		h.hub.SendToAgent(agentID, msg)
	}
}

func (h *Handler) sendError(conn *websocket.Conn, message string) {
	msg, _ := protocol.NewMessage(protocol.MsgTypeError, uuid.New().String(), protocol.ErrorPayload{
		Code:    400,
		Message: message,
	})
	data, _ := json.Marshal(msg)
	conn.WriteMessage(websocket.TextMessage, data)
}

func (h *Handler) AssignTask(agentID string, task *models.Task) error {
	payload := protocol.TaskAssignPayload{
		TaskID:   task.ID,
		Type:     string(task.Type),
		Target:   task.Target,
		ScriptID: task.ScriptID,
		Params:   task.Params,
		Interval: task.Interval,
		Timeout:  task.Timeout,
	}

	msg, err := protocol.NewMessage(protocol.MsgTypeTaskAssign, uuid.New().String(), payload)
	if err != nil {
		return err
	}

	return h.hub.SendToAgent(agentID, msg)
}

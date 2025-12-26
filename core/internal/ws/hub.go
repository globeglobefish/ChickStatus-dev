package ws

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/probe-system/core/pkg/protocol"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024
)

type Hub struct {
	agents       sync.Map
	broadcast    chan []byte
	register     chan *AgentConn
	unregister   chan *AgentConn
	onMessage    func(agentID string, msg *protocol.Message)
	onConnect    func(agentID string)
	onDisconnect func(agentID string)
}

type AgentConn struct {
	ID       string
	Conn     *websocket.Conn
	SendChan chan []byte
	hub      *Hub
	mu       sync.Mutex
	lastSeen time.Time
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte, 256),
		register:   make(chan *AgentConn),
		unregister: make(chan *AgentConn),
	}
}

func (h *Hub) SetMessageHandler(handler func(agentID string, msg *protocol.Message)) {
	h.onMessage = handler
}

func (h *Hub) SetConnectHandler(handler func(agentID string)) {
	h.onConnect = handler
}

func (h *Hub) SetDisconnectHandler(handler func(agentID string)) {
	h.onDisconnect = handler
}

func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.agents.Store(conn.ID, conn)
			log.Printf("Agent connected: %s", conn.ID)
			if h.onConnect != nil {
				go h.onConnect(conn.ID)
			}

		case conn := <-h.unregister:
			if _, ok := h.agents.Load(conn.ID); ok {
				h.agents.Delete(conn.ID)
				close(conn.SendChan)
				log.Printf("Agent disconnected: %s", conn.ID)
				if h.onDisconnect != nil {
					go h.onDisconnect(conn.ID)
				}
			}

		case message := <-h.broadcast:
			h.agents.Range(func(key, value interface{}) bool {
				conn := value.(*AgentConn)
				select {
				case conn.SendChan <- message:
				default:
					h.agents.Delete(key)
					close(conn.SendChan)
				}
				return true
			})
		}
	}
}

func (h *Hub) GetAgent(agentID string) *AgentConn {
	if conn, ok := h.agents.Load(agentID); ok {
		return conn.(*AgentConn)
	}
	return nil
}

func (h *Hub) SendToAgent(agentID string, msg *protocol.Message) error {
	conn := h.GetAgent(agentID)
	if conn == nil {
		return nil
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case conn.SendChan <- data:
		return nil
	default:
		return nil
	}
}

func (h *Hub) Broadcast(msg *protocol.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	h.broadcast <- data
	return nil
}

func (h *Hub) GetOnlineAgentIDs() []string {
	var ids []string
	h.agents.Range(func(key, value interface{}) bool {
		ids = append(ids, key.(string))
		return true
	})
	return ids
}

func (h *Hub) GetAgentCount() int {
	count := 0
	h.agents.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

func (c *AgentConn) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		c.lastSeen = time.Now()

		var msg protocol.Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Invalid message format: %v", err)
			continue
		}

		if c.hub.onMessage != nil {
			c.hub.onMessage(c.ID, &msg)
		}
	}
}

func (c *AgentConn) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.SendChan:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *AgentConn) Send(msg *protocol.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case c.SendChan <- data:
		return nil
	default:
		return nil
	}
}

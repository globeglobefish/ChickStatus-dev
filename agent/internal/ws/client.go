package ws

import (
	"encoding/json"
	"log"
	"math"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/probe-system/agent/pkg/protocol"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024
	maxBackoff     = 60 * time.Second
)

type Client struct {
	serverURL string
	token     string
	version   string
	agentID   string
	conn      *websocket.Conn
	sendChan  chan []byte
	stopChan  chan struct{}
	mu        sync.Mutex
	connected bool
	onMessage func(*protocol.Message)
}

func NewClient(serverURL, token, version string) *Client {
	return &Client{
		serverURL: serverURL,
		token:     token,
		version:   version,
		sendChan:  make(chan []byte, 256),
		stopChan:  make(chan struct{}),
	}
}

func (c *Client) SetMessageHandler(handler func(*protocol.Message)) {
	c.onMessage = handler
}

func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, _, err := websocket.DefaultDialer.Dial(c.serverURL, nil)
	if err != nil {
		return err
	}

	c.conn = conn

	// Send register message
	hostname, _ := os.Hostname()
	payload := protocol.RegisterPayload{
		Hostname: hostname,
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Version:  c.version,
		Token:    c.token,
	}

	msg, _ := protocol.NewMessage(protocol.MsgTypeRegister, uuid.New().String(), payload)
	data, _ := json.Marshal(msg)

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		conn.Close()
		return err
	}

	// Wait for ack
	_, message, err := conn.ReadMessage()
	if err != nil {
		conn.Close()
		return err
	}

	var ackMsg protocol.Message
	if err := json.Unmarshal(message, &ackMsg); err != nil {
		conn.Close()
		return err
	}

	if ackMsg.Type == protocol.MsgTypeError {
		var errPayload protocol.ErrorPayload
		json.Unmarshal(ackMsg.Payload, &errPayload)
		conn.Close()
		return &ConnectionError{Message: errPayload.Message}
	}

	var ackPayload protocol.RegisterAckPayload
	if err := json.Unmarshal(ackMsg.Payload, &ackPayload); err != nil {
		conn.Close()
		return err
	}

	if !ackPayload.Success {
		conn.Close()
		return &ConnectionError{Message: ackPayload.Error}
	}

	c.agentID = ackPayload.AgentID
	c.connected = true

	log.Printf("Connected to server, agent ID: %s", c.agentID)

	return nil
}

func (c *Client) Run() {
	go c.readPump()
	go c.writePump()
	go c.heartbeatLoop()
}

func (c *Client) readPump() {
	defer func() {
		c.mu.Lock()
		c.connected = false
		if c.conn != nil {
			c.conn.Close()
		}
		c.mu.Unlock()
		c.reconnect()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		select {
		case <-c.stopChan:
			return
		default:
		}

		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			return
		}

		var msg protocol.Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Invalid message: %v", err)
			continue
		}

		if c.onMessage != nil {
			c.onMessage(&msg)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return

		case message := <-c.sendChan:
			c.mu.Lock()
			if !c.connected || c.conn == nil {
				c.mu.Unlock()
				continue
			}

			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.mu.Unlock()
				return
			}
			c.mu.Unlock()

		case <-ticker.C:
			c.mu.Lock()
			if !c.connected || c.conn == nil {
				c.mu.Unlock()
				continue
			}

			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.mu.Unlock()
				return
			}
			c.mu.Unlock()
		}
	}
}

func (c *Client) heartbeatLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			c.SendHeartbeat()
		}
	}
}

func (c *Client) reconnect() {
	attempt := 0
	for {
		select {
		case <-c.stopChan:
			return
		default:
		}

		backoff := c.calculateBackoff(attempt)
		log.Printf("Reconnecting in %v (attempt %d)", backoff, attempt+1)
		time.Sleep(backoff)

		if err := c.Connect(); err != nil {
			log.Printf("Reconnection failed: %v", err)
			attempt++
			continue
		}

		c.Run()
		return
	}
}

func (c *Client) calculateBackoff(attempt int) time.Duration {
	backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
	if backoff > maxBackoff {
		backoff = maxBackoff
	}
	return backoff
}

func (c *Client) Send(msg *protocol.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case c.sendChan <- data:
		return nil
	default:
		return &SendError{Message: "send buffer full"}
	}
}

func (c *Client) SendHeartbeat() error {
	msg, err := protocol.NewMessage(protocol.MsgTypeHeartbeat, uuid.New().String(), nil)
	if err != nil {
		return err
	}
	return c.Send(msg)
}

func (c *Client) SendMetrics(metrics *protocol.MetricsPayload) error {
	msg, err := protocol.NewMessage(protocol.MsgTypeMetrics, uuid.New().String(), metrics)
	if err != nil {
		return err
	}
	return c.Send(msg)
}

func (c *Client) SendTaskResult(result *protocol.TaskResultPayload) error {
	msg, err := protocol.NewMessage(protocol.MsgTypeTaskResult, uuid.New().String(), result)
	if err != nil {
		return err
	}
	return c.Send(msg)
}

func (c *Client) Stop() {
	close(c.stopChan)
	c.mu.Lock()
	if c.conn != nil {
		c.conn.Close()
	}
	c.mu.Unlock()
}

func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

func (c *Client) GetAgentID() string {
	return c.agentID
}

type ConnectionError struct {
	Message string
}

func (e *ConnectionError) Error() string {
	return e.Message
}

type SendError struct {
	Message string
}

func (e *SendError) Error() string {
	return e.Message
}

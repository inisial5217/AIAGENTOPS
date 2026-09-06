package ws

import (
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 30 * time.Second
	maxMessageSize = 65536
	sendBufferSize = 1000
)

// Client represents ws connection
type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan WSMessage
	topics    map[string]bool
	topicsMu  sync.RWMutex
	logger    *slog.Logger
	userID    string
	username  string
	closeOnce sync.Once
}

// NewClient creates client instance
func NewClient(hub *Hub, conn *websocket.Conn, userID string, username string, logger *slog.Logger) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan WSMessage, sendBufferSize),
		topics:   make(map[string]bool),
		logger:   logger,
		userID:   userID,
		username: username,
	}
}

// Subscribed checks topic membership
func (c *Client) Subscribed(topic string) bool {
	c.topicsMu.RLock()
	defer c.topicsMu.RUnlock()
	return c.topics[topic]
}

// AddTopic adds subscribed topic
func (c *Client) AddTopic(topic string) {
	c.topicsMu.Lock()
	c.topics[topic] = true
	c.topicsMu.Unlock()
}

// RemoveTopic removes subscribed topic
func (c *Client) RemoveTopic(topic string) {
	c.topicsMu.Lock()
	delete(c.topics, topic)
	c.topicsMu.Unlock()
}

// GetTopics returns subscribed topics
func (c *Client) GetTopics() []string {
	c.topicsMu.RLock()
	defer c.topicsMu.RUnlock()
	res := make([]string, 0, len(c.topics))
	for t := range c.topics {
		res = append(res, t)
	}
	return res
}

// Close terminates client connection
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		if c.conn != nil {
			_ = c.conn.Close()
		}
		close(c.send)
	})
}

// ReadPump handles incoming messages
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Debug("ws read error", slog.String("error", err.Error()))
			}
			break
		}

		var clientMsg ClientMessage
		if err := json.Unmarshal(messageBytes, &clientMsg); err != nil {
			c.logger.Debug("ws decode error", slog.String("error", err.Error()))
			continue
		}

		switch clientMsg.Action {
		case ActionSubscribe:
			if clientMsg.Topic != "" {
				c.hub.Subscribe(c, clientMsg.Topic)
				c.Enqueue(NewWSMessage(TypeAck, clientMsg.Topic, map[string]string{"action": "subscribed"}))
			}
		case ActionUnsubscribe:
			if clientMsg.Topic != "" {
				c.hub.Unsubscribe(c, clientMsg.Topic)
				c.Enqueue(NewWSMessage(TypeAck, clientMsg.Topic, map[string]string{"action": "unsubscribed"}))
			}
		case ActionPing:
			c.Enqueue(NewWSMessage(TypePong, "", map[string]string{"reply": "pong"}))
		}
	}
}

// WritePump pushes outgoing messages
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(msg); err != nil {
				c.logger.Debug("ws write error", slog.String("error", err.Error()))
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Enqueue sends message non-blocking
func (c *Client) Enqueue(msg WSMessage) bool {
	select {
	case c.send <- msg:
		return true
	default:
		// slow consumer drop message
		c.logger.Warn("ws buffer full, dropped", slog.String("topic", msg.Topic))
		return false
	}
}

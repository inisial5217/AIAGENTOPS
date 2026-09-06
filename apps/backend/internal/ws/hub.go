package ws

import (
	"context"
	"log/slog"
	"sync"
)

// Hub manages active connections
type Hub struct {
	mu           sync.RWMutex
	clients      map[*Client]bool
	topics       map[string]map[*Client]bool
	onTopicSub   func(topic string)
	onTopicEmpty func(topic string)
	logger       *slog.Logger
}

// NewHub creates hub instance
func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		clients: make(map[*Client]bool),
		topics:  make(map[string]map[*Client]bool),
		logger:  logger,
	}
}

// SetCallbacks sets topic hooks
func (h *Hub) SetCallbacks(onSub func(string), onEmpty func(string)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onTopicSub = onSub
	h.onTopicEmpty = onEmpty
}

// Register adds client connection
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = true
	h.logger.Debug("ws client connected", slog.String("user", c.username))
}

// Unregister removes client connection
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	if !h.clients[c] {
		h.mu.Unlock()
		return
	}
	delete(h.clients, c)

	// clean up client topics
	topicsToNotify := make([]string, 0)
	for _, topic := range c.GetTopics() {
		if subs, exists := h.topics[topic]; exists {
			delete(subs, c)
			if len(subs) == 0 {
				delete(h.topics, topic)
				topicsToNotify = append(topicsToNotify, topic)
			}
		}
	}
	emptyCb := h.onTopicEmpty
	h.mu.Unlock()

	c.Close()
	h.logger.Debug("ws client disconnected", slog.String("user", c.username))

	if emptyCb != nil {
		for _, t := range topicsToNotify {
			emptyCb(t)
		}
	}
}

// Subscribe binds client topic
func (h *Hub) Subscribe(c *Client, topic string) {
	h.mu.Lock()
	if !h.clients[c] {
		h.mu.Unlock()
		return
	}

	firstSub := false
	if _, exists := h.topics[topic]; !exists {
		h.topics[topic] = make(map[*Client]bool)
		firstSub = true
	}
	h.topics[topic][c] = true
	c.AddTopic(topic)
	subCb := h.onTopicSub
	h.mu.Unlock()

	h.logger.Debug("client subscribed topic", slog.String("user", c.username), slog.String("topic", topic))
	if firstSub && subCb != nil {
		subCb(topic)
	}
}

// Unsubscribe unbinds client topic
func (h *Hub) Unsubscribe(c *Client, topic string) {
	h.mu.Lock()
	emptyTopic := false
	if subs, exists := h.topics[topic]; exists {
		delete(subs, c)
		if len(subs) == 0 {
			delete(h.topics, topic)
			emptyTopic = true
		}
	}
	c.RemoveTopic(topic)
	emptyCb := h.onTopicEmpty
	h.mu.Unlock()

	h.logger.Debug("client unsubscribed topic", slog.String("user", c.username), slog.String("topic", topic))
	if emptyTopic && emptyCb != nil {
		emptyCb(topic)
	}
}

// Broadcast sends to topic
func (h *Hub) Broadcast(topic string, msg WSMessage) {
	h.mu.RLock()
	subs, exists := h.topics[topic]
	if !exists || len(subs) == 0 {
		h.mu.RUnlock()
		return
	}

	targets := make([]*Client, 0, len(subs))
	for c := range subs {
		targets = append(targets, c)
	}
	h.mu.RUnlock()

	for _, c := range targets {
		c.Enqueue(msg)
	}
}

// BroadcastAll sends to all
func (h *Hub) BroadcastAll(msg WSMessage) {
	h.mu.RLock()
	targets := make([]*Client, 0, len(h.clients))
	for c := range h.clients {
		targets = append(targets, c)
	}
	h.mu.RUnlock()

	for _, c := range targets {
		c.Enqueue(msg)
	}
}

// GetSubscriberCount counts topic clients
func (h *Hub) GetSubscriberCount(topic string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if subs, exists := h.topics[topic]; exists {
		return len(subs)
	}
	return 0
}

// GetTotalClients counts connected clients
func (h *Hub) GetTotalClients() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// CloseAll closes all connections
func (h *Hub) CloseAll() {
	h.mu.Lock()
	targets := make([]*Client, 0, len(h.clients))
	for c := range h.clients {
		targets = append(targets, c)
	}
	h.clients = make(map[*Client]bool)
	h.topics = make(map[string]map[*Client]bool)
	h.mu.Unlock()

	for _, c := range targets {
		c.Close()
	}
}

// Run executes hub lifecycle
func (h *Hub) Run(ctx context.Context) {
	<-ctx.Done()
	h.CloseAll()
}

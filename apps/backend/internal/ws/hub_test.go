package ws

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHubLifecycle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	hub := NewHub(logger)

	client := &Client{
		hub:      hub,
		send:     make(chan WSMessage, 10),
		topics:   make(map[string]bool),
		logger:   logger,
		userID:   "user-1",
		username: "admin",
	}

	// register
	hub.Register(client)
	assert.Equal(t, 1, hub.GetTotalClients())

	// subscribe topic
	subHookCalled := false
	hub.SetCallbacks(func(topic string) {
		if topic == "notifications" {
			subHookCalled = true
		}
	}, nil)

	hub.Subscribe(client, "notifications")
	assert.Equal(t, 1, hub.GetSubscriberCount("notifications"))
	assert.True(t, client.Subscribed("notifications"))
	assert.True(t, subHookCalled)

	// broadcast to topic
	msg := NewWSMessage(TypeNotification, "notifications", map[string]string{"alert": "high_cpu"})
	hub.Broadcast("notifications", msg)

	select {
	case received := <-client.send:
		assert.Equal(t, TypeNotification, received.Type)
		assert.Equal(t, "notifications", received.Topic)
	case <-time.After(1 * time.Second):
		t.Fatal("expected message not received")
	}

	// unsubscribe topic
	emptyHookCalled := false
	hub.SetCallbacks(nil, func(topic string) {
		if topic == "notifications" {
			emptyHookCalled = true
		}
	})

	hub.Unsubscribe(client, "notifications")
	assert.Equal(t, 0, hub.GetSubscriberCount("notifications"))
	assert.False(t, client.Subscribed("notifications"))
	assert.True(t, emptyHookCalled)

	// unregister client
	hub.Unregister(client)
	assert.Equal(t, 0, hub.GetTotalClients())
}

func TestClientSlowConsumerDrop(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	client := &Client{
		send:   make(chan WSMessage, 1),
		topics: make(map[string]bool),
		logger: logger,
	}

	msg1 := NewWSMessage(TypeSystemEvent, "system_events", "event-1")
	msg2 := NewWSMessage(TypeSystemEvent, "system_events", "event-2")

	// first message succeeds
	ok1 := client.Enqueue(msg1)
	assert.True(t, ok1)

	// second message should drop gracefully without blocking
	ok2 := client.Enqueue(msg2)
	assert.False(t, ok2)
}

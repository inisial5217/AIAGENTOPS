package ws

import "time"

// MessageType defines ws type
type MessageType string

const (
	TypeLogEntry       MessageType = "log_entry"
	TypeNotification   MessageType = "notification"
	TypeMetricUpdate   MessageType = "metric_update"
	TypeContainerEvent MessageType = "container_event"
	TypeK8sEvent       MessageType = "k8s_event"
	TypeSystemEvent    MessageType = "system_event"
	TypePing           MessageType = "ping"
	TypePong           MessageType = "pong"
	TypeAck            MessageType = "ack"
	TypeError          MessageType = "error"
)

// ClientAction defines client request
type ClientAction string

const (
	ActionSubscribe   ClientAction = "subscribe"
	ActionUnsubscribe ClientAction = "unsubscribe"
	ActionPing        ClientAction = "ping"
)

// WSMessage outgoing envelope
type WSMessage struct {
	Type      MessageType `json:"type"`
	Topic     string      `json:"topic,omitempty"`
	Data      any         `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// ClientMessage incoming envelope
type ClientMessage struct {
	Action  ClientAction `json:"action"`
	Topic   string       `json:"topic,omitempty"`
	Payload any          `json:"payload,omitempty"`
}

// LogPayload represents container log
type LogPayload struct {
	Source    string `json:"source"`
	ID        string `json:"id,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
	Stream    string `json:"stream,omitempty"`
	Log       string `json:"log"`
	Timestamp string `json:"timestamp,omitempty"`
}

// NotificationPayload represents system alert
type NotificationPayload struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	Severity  string `json:"severity"`
	Timestamp string `json:"timestamp"`
	Source    string `json:"source,omitempty"`
}

// EventPayload represents cluster event
type EventPayload struct {
	Type      string `json:"type"`
	Resource  string `json:"resource"`
	Action    string `json:"action"`
	Reason    string `json:"reason,omitempty"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// NewWSMessage helper constructor
func NewWSMessage(msgType MessageType, topic string, data any) WSMessage {
	return WSMessage{
		Type:      msgType,
		Topic:     topic,
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	}
}

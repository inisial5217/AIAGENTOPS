package ws

import "sync"

// Hub manages ws connections
type Hub struct {
	mu sync.RWMutex
}

// NewHub creates hub instance
func NewHub() *Hub {
	return &Hub{}
}

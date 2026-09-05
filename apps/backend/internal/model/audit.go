package model

import (
	"encoding/json"
	"time"
)

// AuditLog represents general security and system audit log
type AuditLog struct {
	ID           string          `json:"id" db:"id"`
	Timestamp    time.Time       `json:"timestamp" db:"timestamp"`
	ActorType    string          `json:"actor_type" db:"actor_type" validate:"required,oneof=user system agent"`
	ActorID      string          `json:"actor_id" db:"actor_id" validate:"required"`
	Action       string          `json:"action" db:"action" validate:"required"`
	ResourceType string          `json:"resource_type" db:"resource_type" validate:"required"`
	ResourceID   *string         `json:"resource_id,omitempty" db:"resource_id"`
	Details      json.RawMessage `json:"details,omitempty" db:"details"`
	IPAddress    *string         `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    *string         `json:"user_agent,omitempty" db:"user_agent"`
	Result       string          `json:"result" db:"result" validate:"required,oneof=success failure"`
}

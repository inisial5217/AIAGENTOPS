package model

import (
	"time"
)

// User represents authenticated user
type User struct {
	ID         string    `json:"id" db:"id"`
	Email      string    `json:"email" db:"email" validate:"required,email"`
	Name       string    `json:"name" db:"name" validate:"required"`
	Role       string    `json:"role" db:"role" validate:"required,oneof=admin devops viewer"`
	KeycloakID string    `json:"keycloak_id,omitempty" db:"keycloak_id"`
	IsActive   bool      `json:"is_active" db:"is_active"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

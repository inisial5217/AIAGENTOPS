package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AISession conversation session
type AISession struct {
	ID              uuid.UUID `json:"id" db:"id"`
	UserID          uuid.UUID `json:"user_id" db:"user_id"`
	Status          string    `json:"status" db:"status"`
	ModelPreference *string   `json:"model_preference,omitempty" db:"model_preference"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	LastActivityAt  time.Time `json:"last_activity_at" db:"last_activity_at"`
}

// AIMessage conversation message
type AIMessage struct {
	ID           uuid.UUID `json:"id" db:"id"`
	SessionID    uuid.UUID `json:"session_id" db:"session_id"`
	Role         string    `json:"role" db:"role"`
	Content      string    `json:"content" db:"content"`
	ModelUsed    *string   `json:"model_used,omitempty" db:"model_used"`
	InputTokens  int       `json:"input_tokens" db:"input_tokens"`
	OutputTokens int       `json:"output_tokens" db:"output_tokens"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// AIUsageTracking usage and cost
type AIUsageTracking struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	UserID           uuid.UUID  `json:"user_id" db:"user_id"`
	SessionID        *uuid.UUID `json:"session_id,omitempty" db:"session_id"`
	ModelProvider    string     `json:"model_provider" db:"model_provider"`
	ModelName        string     `json:"model_name" db:"model_name"`
	InputTokens      int        `json:"input_tokens" db:"input_tokens"`
	OutputTokens     int        `json:"output_tokens" db:"output_tokens"`
	EstimatedCostUSD float64    `json:"estimated_cost_usd" db:"estimated_cost_usd"`
	Timestamp        time.Time  `json:"timestamp" db:"timestamp"`
}

// AIActionAuditLog action audit
type AIActionAuditLog struct {
	ID               uuid.UUID       `json:"id" db:"id"`
	UserID           uuid.UUID       `json:"user_id" db:"user_id"`
	SessionID        *uuid.UUID      `json:"session_id,omitempty" db:"session_id"`
	PromptInputHash  string          `json:"prompt_input_hash" db:"prompt_input_hash"`
	AIOutputSummary  *string         `json:"ai_output_summary,omitempty" db:"ai_output_summary"`
	ToolName         *string         `json:"tool_name,omitempty" db:"tool_name"`
	ToolParameters   json.RawMessage `json:"tool_parameters,omitempty" db:"tool_parameters"`
	ApprovalStatus   string          `json:"approval_status" db:"approval_status"`
	ExecutionResult  *string         `json:"execution_result,omitempty" db:"execution_result"`
	ModelUsed        *string         `json:"model_used,omitempty" db:"model_used"`
	Timestamp        time.Time       `json:"timestamp" db:"timestamp"`
}

// AIToolCall parsed tool call
type AIToolCall struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Parameters       map[string]interface{} `json:"parameters"`
	RequiresApproval bool                   `json:"requires_approval"`
	RequiredRole     string                 `json:"required_role"`
	ApprovalID       *uuid.UUID             `json:"approval_id,omitempty"`
	Status           string                 `json:"status"`
	Result           string                 `json:"result,omitempty"`
}

// AIChatRequest incoming chat
type AIChatRequest struct {
	SessionID       *uuid.UUID `json:"session_id,omitempty"`
	Message         string     `json:"message" validate:"required"`
	ModelPreference *string    `json:"model_preference,omitempty"`
}

// AIChatResponse chat response
type AIChatResponse struct {
	SessionID        uuid.UUID    `json:"session_id"`
	MessageID        uuid.UUID    `json:"message_id"`
	Role             string       `json:"role"`
	Content          string       `json:"content"`
	ModelUsed        string       `json:"model_used"`
	ProviderName     string       `json:"provider_name"`
	InputTokens      int          `json:"input_tokens"`
	OutputTokens     int          `json:"output_tokens"`
	EstimatedCostUSD float64      `json:"estimated_cost_usd"`
	ToolCalls        []AIToolCall `json:"tool_calls"`
	SecurityFlag     *string      `json:"security_flag,omitempty"`
}

// ToolApprovalRequest approve action
type ToolApprovalRequest struct {
	Action string `json:"action" validate:"required,oneof=approve reject"`
}

// AIUsageStats usage summary
type AIUsageStats struct {
	TotalTokens      int               `json:"total_tokens"`
	TotalCostUSD     float64           `json:"total_cost_usd"`
	RequestCount     int               `json:"request_count"`
	ProviderBreakdown map[string]int    `json:"provider_breakdown"`
	ModelBreakdown   map[string]int    `json:"model_breakdown"`
}

// RCARequest rca query
type RCARequest struct {
	IncidentID uuid.UUID `json:"incident_id"`
}

// RCAResponse rca result
type RCAResponse struct {
	IncidentID       uuid.UUID `json:"incident_id"`
	RCASummary       string    `json:"rca_summary"`
	ModelUsed        string    `json:"model_used"`
	ProviderName     string    `json:"provider_name"`
	EstimatedCostUSD float64   `json:"estimated_cost_usd"`
}

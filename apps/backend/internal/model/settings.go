package model

import "time"

// SystemSettings system-wide configuration
type SystemSettings struct {
	ID                       string    `json:"id" db:"id"`
	AppName                  string    `json:"app_name" db:"app_name"`
	DefaultTheme             string    `json:"default_theme" db:"default_theme"`
	Language                 string    `json:"language" db:"language"`
	Timezone                 string    `json:"timezone" db:"timezone"`
	AIDefaultModel           string    `json:"ai_default_model" db:"ai_default_model"`
	AIDefaultProvider        string    `json:"ai_default_provider" db:"ai_default_provider"`
	AIMonthlyBudgetUSD       float64   `json:"ai_monthly_budget_usd" db:"ai_monthly_budget_usd"`
	AIMaxTokensPerRequest    int       `json:"ai_max_tokens_per_request" db:"ai_max_tokens_per_request"`
	AIModelPreferenceOrder   []string  `json:"ai_model_preference_order" db:"ai_model_preference_order"`
	SessionTimeoutMinutes    int       `json:"session_timeout_minutes" db:"session_timeout_minutes"`
	MaxLoginAttempts         int       `json:"max_login_attempts" db:"max_login_attempts"`
	RequireMFA               bool      `json:"require_mfa" db:"require_mfa"`
	MaintenanceMode          bool      `json:"maintenance_mode" db:"maintenance_mode"`
	CreatedAt                time.Time `json:"created_at" db:"created_at"`
	UpdatedAt                time.Time `json:"updated_at" db:"updated_at"`
}

// NotificationSettings telegram and in-app alert config
type NotificationSettings struct {
	ID                         string    `json:"id" db:"id"`
	TelegramBotTokenRef        string    `json:"telegram_bot_token_ref" db:"telegram_bot_token_ref"`
	TelegramChatID             string    `json:"telegram_chat_id" db:"telegram_chat_id"`
	TelegramEnabled            bool      `json:"telegram_enabled" db:"telegram_enabled"`
	InAppEnabled               bool      `json:"inapp_enabled" db:"inapp_enabled"`
	AlertBatchingWindowSeconds int       `json:"alert_batching_window_seconds" db:"alert_batching_window_seconds"`
	CreatedAt                  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt                  time.Time `json:"updated_at" db:"updated_at"`
}

// CombinedSettings holds all settings categories
type CombinedSettings struct {
	System       *SystemSettings       `json:"system"`
	Notification *NotificationSettings `json:"notification"`
}

// SystemSettingsPartial sub-object
type SystemSettingsPartial struct {
	AppName                *string  `json:"app_name,omitempty"`
	DefaultTheme           *string  `json:"default_theme,omitempty"`
	Language               *string  `json:"language,omitempty"`
	Timezone               *string  `json:"timezone,omitempty"`
	AIDefaultModel         *string  `json:"ai_default_model,omitempty"`
	AIDefaultProvider      *string  `json:"ai_default_provider,omitempty"`
	AIMonthlyBudgetUSD     *float64 `json:"ai_monthly_budget_usd,omitempty"`
	AIMaxTokensPerRequest  *int     `json:"ai_max_tokens_per_request,omitempty"`
	AIModelPreferenceOrder []string `json:"ai_model_preference_order,omitempty"`
	SessionTimeoutMinutes  *int     `json:"session_timeout_minutes,omitempty"`
	MaxLoginAttempts       *int     `json:"max_login_attempts,omitempty"`
	RequireMFA             *bool    `json:"require_mfa,omitempty"`
	MaintenanceMode        *bool    `json:"maintenance_mode,omitempty"`
}

// NotificationSettingsPartial sub-object
type NotificationSettingsPartial struct {
	TelegramBotTokenRef        *string `json:"telegram_bot_token_ref,omitempty"`
	TelegramChatID             *string `json:"telegram_chat_id,omitempty"`
	TelegramEnabled            *bool   `json:"telegram_enabled,omitempty"`
	InAppEnabled               *bool   `json:"inapp_enabled,omitempty"`
	AlertBatchingWindowSeconds *int    `json:"alert_batching_window_seconds,omitempty"`
}

// UpdateSettingsRequest payload for updating settings
type UpdateSettingsRequest struct {
	System                     *SystemSettingsPartial       `json:"system,omitempty"`
	Notification               *NotificationSettingsPartial `json:"notification,omitempty"`
	AppName                    *string                      `json:"app_name,omitempty"`
	DefaultTheme               *string                      `json:"default_theme,omitempty"`
	Language                   *string                      `json:"language,omitempty"`
	Timezone                   *string                      `json:"timezone,omitempty"`
	AIDefaultModel             *string                      `json:"ai_default_model,omitempty"`
	AIDefaultProvider          *string                      `json:"ai_default_provider,omitempty"`
	AIMonthlyBudgetUSD         *float64                     `json:"ai_monthly_budget_usd,omitempty"`
	AIMaxTokensPerRequest      *int                         `json:"ai_max_tokens_per_request,omitempty"`
	AIModelPreferenceOrder     []string                     `json:"ai_model_preference_order,omitempty"`
	SessionTimeoutMinutes      *int                         `json:"session_timeout_minutes,omitempty"`
	MaxLoginAttempts           *int                         `json:"max_login_attempts,omitempty"`
	RequireMFA                 *bool                        `json:"require_mfa,omitempty"`
	MaintenanceMode            *bool                        `json:"maintenance_mode,omitempty"`
	TelegramBotTokenRef        *string                      `json:"telegram_bot_token_ref,omitempty"`
	TelegramChatID             *string                      `json:"telegram_chat_id,omitempty"`
	TelegramEnabled            *bool                        `json:"telegram_enabled,omitempty"`
	InAppEnabled               *bool                        `json:"inapp_enabled,omitempty"`
	AlertBatchingWindowSeconds *int                         `json:"alert_batching_window_seconds,omitempty"`
}

// UpdateUserRoleRequest payload for role alteration
type UpdateUserRoleRequest struct {
	Role string `json:"role"`
}

// UpdateUserStatusRequest payload for activating/deactivating
type UpdateUserStatusRequest struct {
	IsActive bool `json:"is_active"`
}

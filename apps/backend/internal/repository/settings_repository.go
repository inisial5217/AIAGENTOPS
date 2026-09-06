package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SettingsRepository interface
type SettingsRepository interface {
	GetSystemSettings(ctx context.Context) (*model.SystemSettings, error)
	UpdateSystemSettings(ctx context.Context, s *model.SystemSettings) (*model.SystemSettings, error)
	GetNotificationSettings(ctx context.Context) (*model.NotificationSettings, error)
	UpdateNotificationSettings(ctx context.Context, n *model.NotificationSettings) (*model.NotificationSettings, error)
}

// PostgresSettingsRepository pgxpool implementation
type PostgresSettingsRepository struct {
	pool *pgxpool.Pool
}

// NewSettingsRepository constructor
func NewSettingsRepository(pool *pgxpool.Pool) *PostgresSettingsRepository {
	return &PostgresSettingsRepository{pool: pool}
}

// GetSystemSettings retrieves system settings
func (r *PostgresSettingsRepository) GetSystemSettings(ctx context.Context) (*model.SystemSettings, error) {
	query := `
		SELECT id, app_name, default_theme, language, timezone,
		       ai_default_model, ai_default_provider, ai_monthly_budget_usd, ai_max_tokens_per_request,
		       ai_model_preference_order, session_timeout_minutes, max_login_attempts,
		       require_mfa, maintenance_mode, created_at, updated_at
		FROM system_settings
		ORDER BY created_at ASC
		LIMIT 1`

	var s model.SystemSettings
	err := r.pool.QueryRow(ctx, query).Scan(
		&s.ID, &s.AppName, &s.DefaultTheme, &s.Language, &s.Timezone,
		&s.AIDefaultModel, &s.AIDefaultProvider, &s.AIMonthlyBudgetUSD, &s.AIMaxTokensPerRequest,
		&s.AIModelPreferenceOrder, &s.SessionTimeoutMinutes, &s.MaxLoginAttempts,
		&s.RequireMFA, &s.MaintenanceMode, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// default fallback
			return &model.SystemSettings{
				AppName:                "CIFO Platform",
				DefaultTheme:           "dark",
				Language:               "id",
				Timezone:               "Asia/Jakarta",
				AIDefaultModel:         "gemini-2.0-flash",
				AIDefaultProvider:      "google",
				AIMonthlyBudgetUSD:     50.00,
				AIMaxTokensPerRequest:  4096,
				AIModelPreferenceOrder: []string{"google", "openai", "anthropic", "ollama", "mock"},
				SessionTimeoutMinutes:  60,
				MaxLoginAttempts:       5,
			}, nil
		}
		return nil, fmt.Errorf("get system settings: %w", err)
	}

	return &s, nil
}

// UpdateSystemSettings updates system settings
func (r *PostgresSettingsRepository) UpdateSystemSettings(ctx context.Context, s *model.SystemSettings) (*model.SystemSettings, error) {
	query := `
		UPDATE system_settings
		SET app_name = $1, default_theme = $2, language = $3, timezone = $4,
		    ai_default_model = $5, ai_default_provider = $6, ai_monthly_budget_usd = $7,
		    ai_max_tokens_per_request = $8, ai_model_preference_order = $9,
		    session_timeout_minutes = $10, max_login_attempts = $11,
		    require_mfa = $12, maintenance_mode = $13, updated_at = NOW()
		WHERE id = $14
		RETURNING id, app_name, default_theme, language, timezone,
		          ai_default_model, ai_default_provider, ai_monthly_budget_usd, ai_max_tokens_per_request,
		          ai_model_preference_order, session_timeout_minutes, max_login_attempts,
		          require_mfa, maintenance_mode, created_at, updated_at`

	var updated model.SystemSettings
	err := r.pool.QueryRow(ctx, query,
		s.AppName, s.DefaultTheme, s.Language, s.Timezone,
		s.AIDefaultModel, s.AIDefaultProvider, s.AIMonthlyBudgetUSD,
		s.AIMaxTokensPerRequest, s.AIModelPreferenceOrder,
		s.SessionTimeoutMinutes, s.MaxLoginAttempts,
		s.RequireMFA, s.MaintenanceMode, s.ID,
	).Scan(
		&updated.ID, &updated.AppName, &updated.DefaultTheme, &updated.Language, &updated.Timezone,
		&updated.AIDefaultModel, &updated.AIDefaultProvider, &updated.AIMonthlyBudgetUSD, &updated.AIMaxTokensPerRequest,
		&updated.AIModelPreferenceOrder, &updated.SessionTimeoutMinutes, &updated.MaxLoginAttempts,
		&updated.RequireMFA, &updated.MaintenanceMode, &updated.CreatedAt, &updated.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update system settings: %w", err)
	}

	return &updated, nil
}

// GetNotificationSettings retrieves notification settings
func (r *PostgresSettingsRepository) GetNotificationSettings(ctx context.Context) (*model.NotificationSettings, error) {
	query := `
		SELECT id, COALESCE(telegram_bot_token_ref, ''), COALESCE(telegram_chat_id, ''),
		       telegram_enabled, inapp_enabled, alert_batching_window_seconds,
		       created_at, updated_at
		FROM notification_settings
		ORDER BY created_at ASC
		LIMIT 1`

	var n model.NotificationSettings
	err := r.pool.QueryRow(ctx, query).Scan(
		&n.ID, &n.TelegramBotTokenRef, &n.TelegramChatID,
		&n.TelegramEnabled, &n.InAppEnabled, &n.AlertBatchingWindowSeconds,
		&n.CreatedAt, &n.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &model.NotificationSettings{
				TelegramBotTokenRef:        "vault:secret/telegram#token",
				TelegramChatID:             "-100123456789",
				TelegramEnabled:            true,
				InAppEnabled:               true,
				AlertBatchingWindowSeconds: 120,
			}, nil
		}
		return nil, fmt.Errorf("get notification settings: %w", err)
	}

	return &n, nil
}

// UpdateNotificationSettings updates notification settings
func (r *PostgresSettingsRepository) UpdateNotificationSettings(ctx context.Context, n *model.NotificationSettings) (*model.NotificationSettings, error) {
	query := `
		UPDATE notification_settings
		SET telegram_bot_token_ref = $1, telegram_chat_id = $2,
		    telegram_enabled = $3, inapp_enabled = $4,
		    alert_batching_window_seconds = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING id, COALESCE(telegram_bot_token_ref, ''), COALESCE(telegram_chat_id, ''),
		          telegram_enabled, inapp_enabled, alert_batching_window_seconds,
		          created_at, updated_at`

	var updated model.NotificationSettings
	err := r.pool.QueryRow(ctx, query,
		n.TelegramBotTokenRef, n.TelegramChatID,
		n.TelegramEnabled, n.InAppEnabled,
		n.AlertBatchingWindowSeconds, n.ID,
	).Scan(
		&updated.ID, &updated.TelegramBotTokenRef, &updated.TelegramChatID,
		&updated.TelegramEnabled, &updated.InAppEnabled, &updated.AlertBatchingWindowSeconds,
		&updated.CreatedAt, &updated.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update notification settings: %w", err)
	}

	return &updated, nil
}

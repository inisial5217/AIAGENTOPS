-- Migration 008: Create system_settings table and seed defaults
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'system_settings' AND column_name = 'key'
    ) THEN
        ALTER TABLE system_settings RENAME TO system_settings_kv;
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS system_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_name VARCHAR(100) NOT NULL DEFAULT 'CIFO Platform',
    default_theme VARCHAR(20) NOT NULL DEFAULT 'dark',
    language VARCHAR(10) NOT NULL DEFAULT 'id',
    timezone VARCHAR(50) NOT NULL DEFAULT 'Asia/Jakarta',
    ai_default_model VARCHAR(50) NOT NULL DEFAULT 'gemini-2.0-flash',
    ai_default_provider VARCHAR(50) NOT NULL DEFAULT 'google',
    ai_monthly_budget_usd NUMERIC(10,2) NOT NULL DEFAULT 50.00,
    ai_max_tokens_per_request INTEGER NOT NULL DEFAULT 4096,
    ai_model_preference_order TEXT[] NOT NULL DEFAULT ARRAY['google', 'openai', 'anthropic', 'ollama', 'mock'],
    session_timeout_minutes INTEGER NOT NULL DEFAULT 60,
    max_login_attempts INTEGER NOT NULL DEFAULT 5,
    require_mfa BOOLEAN NOT NULL DEFAULT false,
    maintenance_mode BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed initial row in system_settings if table is empty
INSERT INTO system_settings (
    app_name, default_theme, language, timezone,
    ai_default_model, ai_default_provider, ai_monthly_budget_usd, ai_max_tokens_per_request,
    ai_model_preference_order, session_timeout_minutes, max_login_attempts, require_mfa, maintenance_mode
)
SELECT
    'CIFO Platform', 'dark', 'id', 'Asia/Jakarta',
    'gemini-2.0-flash', 'google', 50.00, 4096,
    ARRAY['google', 'openai', 'anthropic', 'ollama', 'mock'], 60, 5, false, false
WHERE NOT EXISTS (SELECT 1 FROM system_settings);

-- Seed initial row in notification_settings if empty
INSERT INTO notification_settings (
    telegram_bot_token_ref, telegram_chat_id, telegram_enabled, inapp_enabled, alert_batching_window_seconds
)
SELECT
    'vault:secret/telegram#token', '-100123456789', true, true, 120
WHERE NOT EXISTS (SELECT 1 FROM notification_settings);

CREATE TABLE IF NOT EXISTS notification_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    telegram_bot_token_ref VARCHAR(255),
    telegram_chat_id VARCHAR(100),
    telegram_enabled BOOLEAN NOT NULL DEFAULT false,
    inapp_enabled BOOLEAN NOT NULL DEFAULT true,
    alert_batching_window_seconds INTEGER NOT NULL DEFAULT 120,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

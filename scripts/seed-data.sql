-- System config seed data (no mock/dummy user data)

-- Default system notification settings
INSERT INTO notification_settings (
    id,
    telegram_bot_token_ref,
    telegram_chat_id,
    telegram_enabled,
    inapp_enabled,
    alert_batching_window_seconds,
    created_at,
    updated_at
) VALUES (
    '00000000-0000-0000-0000-000000000001',
    'vault:secret/telegram#token',
    '-1001234567890',
    false,
    true,
    120,
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

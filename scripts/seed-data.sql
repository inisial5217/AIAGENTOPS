-- System config seed data (no dummy/fake user data)

-- 1. Create table structure if running directly before migrations
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'viewer',
    keycloak_id VARCHAR(255) UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

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

-- 2. Insert initial system notification settings
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

-- 3. Insert default administrative operator placeholder
INSERT INTO users (
    id,
    email,
    name,
    role,
    is_active,
    created_at,
    updated_at
) VALUES (
    '00000000-0000-0000-0000-000000000099',
    'admin@cifo.id',
    'CIFO Admin Operator',
    'admin',
    true,
    NOW(),
    NOW()
) ON CONFLICT (email) DO NOTHING;

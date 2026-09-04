CREATE TABLE IF NOT EXISTS audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    actor_type VARCHAR(20) NOT NULL,
    actor_id VARCHAR(255) NOT NULL,
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id VARCHAR(255),
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    result VARCHAR(20) NOT NULL DEFAULT 'success'
);

CREATE INDEX IF NOT EXISTS idx_audit_log_timestamp ON audit_log(timestamp);
CREATE INDEX IF NOT EXISTS idx_audit_log_actor_id ON audit_log(actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_action ON audit_log(action);

CREATE TABLE IF NOT EXISTS ai_action_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id UUID REFERENCES ai_sessions(id) ON DELETE SET NULL,
    prompt_input_hash VARCHAR(64) NOT NULL,
    ai_output_summary TEXT,
    tool_name VARCHAR(100),
    tool_parameters JSONB,
    approval_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    execution_result TEXT,
    model_used VARCHAR(50),
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_action_audit_user ON ai_action_audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_action_audit_timestamp ON ai_action_audit_log(timestamp);

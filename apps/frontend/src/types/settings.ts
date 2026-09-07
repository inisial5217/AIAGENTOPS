// Settings domain models for CIFO Monitoring Platform

export interface SystemSettings {
  id: string;
  app_name: string;
  default_theme?: string;
  language?: string;
  timezone?: string;
  retention_days: number;
  refresh_interval: number;
  ai_auto_remediation: boolean;
  ai_analysis_threshold: number;
  ai_default_model?: string;
  ai_default_provider?: string;
  ai_monthly_budget_usd?: number;
  ai_max_tokens_per_request?: number;
  ai_model_preference_order?: string[];
  session_timeout_minutes: number;
  session_timeout_mins?: number; // alias for backwards compatibility
  max_login_attempts?: number;
  require_mfa: boolean;
  mfa_enforced?: boolean; // alias for backwards compatibility
  maintenance_mode?: boolean;
  created_at?: string;
  updated_at: string;
}

export interface NotificationSettings {
  id: string;
  telegram_enabled: boolean;
  telegram_bot_token?: string;
  telegram_bot_token_ref?: string;
  telegram_chat_id?: string;
  inapp_enabled?: boolean;
  alert_batching_window_seconds?: number;
  email_enabled: boolean;
  email_recipients?: string;
  critical_alert: boolean;
  warning_alert: boolean;
  info_alert: boolean;
  auto_resolve_alert: boolean;
  quiet_hours_enabled: boolean;
  quiet_hours_start?: string;
  quiet_hours_end?: string;
  created_at?: string;
  updated_at: string;
}

export interface CombinedSettings {
  system: SystemSettings;
  notification: NotificationSettings;
}

export interface UpdateSettingsRequest {
  system?: Partial<SystemSettings>;
  notification?: Partial<NotificationSettings>;
  // Flattened aliases supported by backend
  app_name?: string;
  session_timeout_minutes?: number;
  require_mfa?: boolean;
  ai_monthly_budget_usd?: number;
  ai_model_preference_order?: string[];
  telegram_bot_token_ref?: string;
  telegram_chat_id?: string;
  telegram_enabled?: boolean;
}

export interface UserAdmin {
  id: string;
  keycloak_id?: string;
  email: string;
  name: string;
  role: "admin" | "devops" | "viewer";
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface ActiveSession {
  id: string;
  device: string;
  ip: string;
  last_active: string;
  is_current: boolean;
}

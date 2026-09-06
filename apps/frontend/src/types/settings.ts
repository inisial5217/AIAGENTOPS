// Settings domain models for CIFO Monitoring Platform

export interface SystemSettings {
  id: string;
  app_name: string;
  retention_days: number;
  refresh_interval: number;
  ai_auto_remediation: boolean;
  ai_analysis_threshold: number;
  session_timeout_mins: number;
  mfa_enforced: boolean;
  updated_at: string;
}

export interface NotificationSettings {
  id: string;
  telegram_enabled: boolean;
  telegram_bot_token?: string;
  telegram_chat_id?: string;
  email_enabled: boolean;
  email_recipients?: string;
  critical_alert: boolean;
  warning_alert: boolean;
  info_alert: boolean;
  auto_resolve_alert: boolean;
  quiet_hours_enabled: boolean;
  quiet_hours_start?: string;
  quiet_hours_end?: string;
  updated_at: string;
}

export interface CombinedSettings {
  system: SystemSettings;
  notification: NotificationSettings;
}

export interface UpdateSettingsRequest {
  system?: Partial<SystemSettings>;
  notification?: Partial<NotificationSettings>;
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

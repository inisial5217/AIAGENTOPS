export type IncidentStatus =
  | "open"
  | "acknowledged"
  | "investigating"
  | "resolved"
  | "closed";

export type IncidentSeverity = "critical" | "warning" | "info";

export type IncidentSource = "alertmanager" | "docker" | "kubernetes" | "manual";

export interface NotificationRecord {
  id: string;
  incident_id?: string;
  channel: string;
  recipient: string;
  title: string;
  message: string;
  severity: string;
  status: string;
  created_at: string;
  error_message?: string;
}

export interface IncidentSummary {
  id: string;
  title: string;
  description: string;
  severity: IncidentSeverity;
  status: IncidentStatus;
  source: string;
  alert_name?: string;
  resource_type?: string;
  resource_id?: string;
  namespace?: string;
  acknowledged_by?: string;
  acknowledged_by_name?: string;
  acknowledged_at?: string;
  resolved_by?: string;
  resolved_by_name?: string;
  resolved_at?: string;
  closed_by?: string;
  closed_by_name?: string;
  closed_at?: string;
  created_at: string;
  updated_at: string;
}

export interface IncidentDetail extends IncidentSummary {
  rca_summary?: string;
  labels?: Record<string, string>;
  annotations?: Record<string, string>;
  notifications?: NotificationRecord[];
}

export interface IncidentStats {
  total: number;
  open: number;
  acknowledged: number;
  investigating: number;
  resolved: number;
  closed: number;
  critical_count: number;
}

export interface IncidentFilter {
  status?: string;
  severity?: string;
  source?: string;
  search?: string;
  page?: number;
  limit?: number;
}

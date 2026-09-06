export type WebSocketStatus = "connecting" | "connected" | "disconnected" | "error";

export type WSMessageType =
  | "log_entry"
  | "notification"
  | "metric_update"
  | "container_event"
  | "k8s_event"
  | "system_event"
  | "ping"
  | "pong"
  | "ack"
  | "error";

export interface WSMessage<T = unknown> {
  type: WSMessageType;
  topic?: string;
  data: T;
  timestamp: number;
}

export interface ClientActionMessage {
  action: "subscribe" | "unsubscribe" | "ping";
  topic?: string;
  payload?: unknown;
}

export interface LogPayload {
  source: string;
  id?: string;
  namespace?: string;
  name?: string;
  stream?: string;
  log: string;
  timestamp?: string;
}

export interface NotificationPayload {
  id: string;
  title: string;
  message: string;
  severity: "info" | "warning" | "error" | "critical";
  timestamp: string;
  source?: string;
}

export interface EventPayload {
  type: string;
  resource: string;
  action: string;
  reason?: string;
  message: string;
  timestamp: string;
}

import { WSMessage, ClientActionMessage, WebSocketStatus } from "../types/websocket";

type MessageHandler = (message: WSMessage) => void;
type StatusHandler = (status: WebSocketStatus) => void;

export class WebSocketClient {
  private socket: WebSocket | null = null;
  private status: WebSocketStatus = "disconnected";
  private subscribedTopics: Set<string> = new Set();
  private messageHandlers: Map<string, Set<MessageHandler>> = new Map();
  private globalHandlers: Set<MessageHandler> = new Set();
  private statusHandlers: Set<StatusHandler> = new Set();
  private reconnectAttempt = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private shouldReconnect = true;

  constructor() {}

  public getStatus(): WebSocketStatus {
    return this.status;
  }

  public connect(): void {
    if (typeof window === "undefined") return;
    if (this.socket && (this.socket.readyState === WebSocket.CONNECTING || this.socket.readyState === WebSocket.OPEN)) {
      return;
    }

    this.shouldReconnect = true;
    this.setStatus("connecting");

    const token = localStorage.getItem("cifo_access_token") || "dev-token-admin";
    const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://127.0.0.1:8080";
    const wsBaseUrl = apiUrl.replace(/^http/, "ws");
    const wsUrl = `${wsBaseUrl}/ws?token=${encodeURIComponent(token)}`;

    try {
      this.socket = new WebSocket(wsUrl);

      this.socket.onopen = () => {
        this.setStatus("connected");
        this.reconnectAttempt = 0;
        // resubscribe topics on connect
        this.subscribedTopics.forEach((topic) => {
          this.sendAction("subscribe", topic);
        });
      };

      this.socket.onmessage = (event: MessageEvent) => {
        try {
          const data: WSMessage = JSON.parse(event.data);
          this.dispatchMessage(data);
        } catch {
          // ignore malformed frame
        }
      };

      this.socket.onclose = () => {
        this.setStatus("disconnected");
        this.socket = null;
        if (this.shouldReconnect) {
          this.scheduleReconnect();
        }
      };

      this.socket.onerror = () => {
        this.setStatus("error");
      };
    } catch {
      this.setStatus("error");
      if (this.shouldReconnect) {
        this.scheduleReconnect();
      }
    }
  }

  public disconnect(): void {
    this.shouldReconnect = false;
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
    this.setStatus("disconnected");
  }

  public subscribe(topic: string, handler?: MessageHandler): void {
    this.subscribedTopics.add(topic);
    if (handler) {
      this.on(topic, handler);
    }
    if (this.status === "connected") {
      this.sendAction("subscribe", topic);
    } else {
      this.connect();
    }
  }

  public unsubscribe(topic: string, handler?: MessageHandler): void {
    if (handler) {
      this.off(topic, handler);
    }
    // only fully unsubscribe if no more handlers
    const handlers = this.messageHandlers.get(topic);
    if (!handlers || handlers.size === 0) {
      this.subscribedTopics.delete(topic);
      if (this.status === "connected") {
        this.sendAction("unsubscribe", topic);
      }
    }
  }

  public on(topic: string, handler: MessageHandler): void {
    if (!this.messageHandlers.has(topic)) {
      this.messageHandlers.set(topic, new Set());
    }
    this.messageHandlers.get(topic)!.add(handler);
  }

  public off(topic: string, handler: MessageHandler): void {
    const handlers = this.messageHandlers.get(topic);
    if (handlers) {
      handlers.delete(handler);
      if (handlers.size === 0) {
        this.messageHandlers.delete(topic);
      }
    }
  }

  public onAny(handler: MessageHandler): () => void {
    this.globalHandlers.add(handler);
    return () => {
      this.globalHandlers.delete(handler);
    };
  }

  public onStatusChange(handler: StatusHandler): () => void {
    this.statusHandlers.add(handler);
    handler(this.status);
    return () => {
      this.statusHandlers.delete(handler);
    };
  }

  public send(action: "subscribe" | "unsubscribe" | "ping", topic?: string, payload?: unknown): void {
    this.sendAction(action, topic, payload);
  }

  private sendAction(action: "subscribe" | "unsubscribe" | "ping", topic?: string, payload?: unknown): void {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      const msg: ClientActionMessage = { action, topic, payload };
      this.socket.send(JSON.stringify(msg));
    }
  }

  private setStatus(newStatus: WebSocketStatus): void {
    if (this.status !== newStatus) {
      this.status = newStatus;
      this.statusHandlers.forEach((handler) => handler(newStatus));
    }
  }

  private dispatchMessage(message: WSMessage): void {
    this.globalHandlers.forEach((handler) => handler(message));
    if (message.topic) {
      const handlers = this.messageHandlers.get(message.topic);
      if (handlers) {
        handlers.forEach((handler) => handler(message));
      }
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer) return;
    this.reconnectAttempt++;
    // exponential backoff: 1s, 2s, 4s, 8s, max 30s
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempt - 1), 30000);
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      if (this.shouldReconnect) {
        this.connect();
      }
    }, delay);
  }
}

// singleton instance
let instance: WebSocketClient | null = null;

export function getWebSocketClient(): WebSocketClient {
  if (!instance) {
    instance = new WebSocketClient();
  }
  return instance;
}

export const wsClient = getWebSocketClient();

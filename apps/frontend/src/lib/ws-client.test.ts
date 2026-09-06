import { describe, it, expect, vi, beforeEach } from "vitest";
import { WebSocketClient } from "./ws-client";
import { WSMessage } from "../types/websocket";

describe("WebSocketClient", () => {
  let client: WebSocketClient;

  beforeEach(() => {
    client = new WebSocketClient();
  });

  it("initializes with disconnected status", () => {
    expect(client.getStatus()).toBe("disconnected");
  });

  it("registers and triggers status handlers", () => {
    const statusHandler = vi.fn();
    client.onStatusChange(statusHandler);

    expect(statusHandler).toHaveBeenCalledWith("disconnected");
  });

  it("subscribes to topics and registers message handler", () => {
    const handler = vi.fn();
    client.subscribe("notifications", handler);

    // dispatch test message
    const msg: WSMessage = {
      type: "notification",
      topic: "notifications",
      data: { title: "Test Alert", severity: "warning" },
      timestamp: Date.now(),
    };

    // simulate dispatch
    (client as any).dispatchMessage(msg);

    expect(handler).toHaveBeenCalledWith(msg);
  });

  it("unsubscribes from topics and unregisters handler", () => {
    const handler = vi.fn();
    client.subscribe("system_events", handler);
    client.unsubscribe("system_events", handler);

    const msg: WSMessage = {
      type: "system_event",
      topic: "system_events",
      data: { message: "Test Event" },
      timestamp: Date.now(),
    };

    (client as any).dispatchMessage(msg);

    expect(handler).not.toHaveBeenCalled();
  });

  it("registers global onAny message handler", () => {
    const globalHandler = vi.fn();
    const unsub = client.onAny(globalHandler);

    const msg: WSMessage = {
      type: "log_entry",
      topic: "docker_logs:c-1",
      data: { log: "test log" },
      timestamp: Date.now(),
    };

    (client as any).dispatchMessage(msg);
    expect(globalHandler).toHaveBeenCalledWith(msg);

    unsub();
    (client as any).dispatchMessage(msg);
    expect(globalHandler).toHaveBeenCalledTimes(1);
  });
});

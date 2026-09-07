import { renderHook, act } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { useWebSocket } from "./use-websocket";
import * as wsLib from "../lib/ws-client";
import { WebSocketStatus, WSMessage } from "../types/websocket";

describe("useWebSocket hook", () => {
  let statusCallback: (status: WebSocketStatus) => void;
  let topicHandlers: Map<string, (msg: WSMessage) => void>;
  let mockClient: any;

  beforeEach(() => {
    topicHandlers = new Map();
    mockClient = {
      getStatus: vi.fn().mockReturnValue("disconnected"),
      connect: vi.fn(),
      disconnect: vi.fn(),
      onStatusChange: vi.fn().mockImplementation((cb) => {
        statusCallback = cb;
        return vi.fn();
      }),
      subscribe: vi.fn().mockImplementation((topic, handler) => {
        if (handler) topicHandlers.set(topic, handler);
      }),
      unsubscribe: vi.fn().mockImplementation((topic) => {
        topicHandlers.delete(topic);
      }),
      send: vi.fn(),
    };

    vi.spyOn(wsLib, "getWebSocketClient").mockReturnValue(mockClient);
  });

  it("subscribes to topics and updates status on connection", () => {
    const { result } = renderHook(() => useWebSocket(["telemetry:cpu", "incidents:live"]));

    expect(mockClient.connect).toHaveBeenCalled();
    expect(mockClient.subscribe).toHaveBeenCalledWith("telemetry:cpu", expect.any(Function));
    expect(mockClient.subscribe).toHaveBeenCalledWith("incidents:live", expect.any(Function));
    expect(result.current.isConnected).toBe(false);
    expect(result.current.status).toBe("disconnected");

    // Simulate connection established
    act(() => {
      statusCallback("connected");
    });

    expect(result.current.isConnected).toBe(true);
    expect(result.current.status).toBe("connected");
  });

  it("receives and stores message when topic handler fires", () => {
    const { result } = renderHook(() => useWebSocket(["telemetry:cpu"]));

    const msg: WSMessage = {
      type: "telemetry",
      topic: "telemetry:cpu",
      data: { usage_percent: 45.8 },
      timestamp: new Date().toISOString(),
    };

    act(() => {
      const handler = topicHandlers.get("telemetry:cpu");
      if (handler) {
        handler(msg);
      }
    });

    expect(result.current.lastMessage).toEqual(msg);
  });

  it("forwards sendMessage calls to wsClient", () => {
    const { result } = renderHook(() => useWebSocket([]));

    act(() => {
      result.current.sendMessage("subscribe", "docker:containers", { filter: "running" });
    });

    expect(mockClient.send).toHaveBeenCalledWith(
      "subscribe",
      "docker:containers",
      { filter: "running" }
    );
  });

  it("does not connect if enabled is false", () => {
    renderHook(() => useWebSocket(["logs"], { enabled: false }));

    expect(mockClient.connect).not.toHaveBeenCalled();
    expect(mockClient.subscribe).not.toHaveBeenCalled();
  });
});

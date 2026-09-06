"use client";

import * as React from "react";
import { WebSocketStatus, WSMessage } from "../types/websocket";
import { getWebSocketClient } from "../lib/ws-client";

interface UseWebSocketOptions {
  enabled?: boolean;
}

export function useWebSocket(topics: string[] = [], options: UseWebSocketOptions = {}) {
  const { enabled = true } = options;
  const client = React.useMemo(() => getWebSocketClient(), []);
  const [status, setStatus] = React.useState<WebSocketStatus>("disconnected");
  const [lastMessage, setLastMessage] = React.useState<WSMessage | null>(null);

  React.useEffect(() => {
    if (!enabled) return;

    // listen to connection status changes
    const unsubStatus = client.onStatusChange((newStatus) => {
      setStatus(newStatus);
    });

    client.connect();

    // message handler for requested topics
    const handleMsg = (msg: WSMessage) => {
      setLastMessage(msg);
    };

    // subscribe to all requested topics
    topics.forEach((topic) => {
      client.subscribe(topic, handleMsg);
    });

    return () => {
      unsubStatus();
      topics.forEach((topic) => {
        client.unsubscribe(topic, handleMsg);
      });
    };
  }, [client, enabled, topics.join(",")]);

  const subscribe = React.useCallback(
    (topic: string, handler?: (msg: WSMessage) => void) => {
      client.subscribe(topic, handler);
    },
    [client]
  );

  const unsubscribe = React.useCallback(
    (topic: string, handler?: (msg: WSMessage) => void) => {
      client.unsubscribe(topic, handler);
    },
    [client]
  );

  const sendMessage = React.useCallback(
    (action: "subscribe" | "unsubscribe" | "ping", topic?: string, payload?: unknown) => {
      client.send(action, topic, payload);
    },
    [client]
  );

  return {
    isConnected: status === "connected",
    status,
    lastMessage,
    subscribe,
    unsubscribe,
    sendMessage,
  };
}

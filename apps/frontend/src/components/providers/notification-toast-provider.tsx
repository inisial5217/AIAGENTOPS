"use client";

import * as React from "react";
import {
  ToastProvider,
  ToastViewport,
  Toast,
  ToastTitle,
  ToastDescription,
} from "../ui/toast";
import { useWebSocket } from "../../hooks/use-websocket";
import { useNotificationStore } from "../../store/notification-store";
import { NotificationPayload, WSMessage } from "../../types/websocket";
import { AlertCircle, AlertTriangle, CheckCircle2, Info } from "lucide-react";

interface ActiveToast extends NotificationPayload {
  open: boolean;
}

export function NotificationToastProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  const [toasts, setToasts] = React.useState<ActiveToast[]>([]);
  const { addNotification } = useNotificationStore();

  // subscribe to notifications topic
  const { isConnected } = useWebSocket(["notifications"]);

  React.useEffect(() => {
    const { wsClient } = require("../../lib/ws-client");

    const handleNotification = (msg: WSMessage) => {
      if (msg.type === "notification" && msg.data) {
        const payload = msg.data as NotificationPayload;

        // add to global notification store for bell center
        addNotification({
          title: payload.title,
          message: payload.message,
          severity: payload.severity,
        });

        // add to active toast queue (max 3 on screen)
        const newToast: ActiveToast = {
          ...payload,
          open: true,
        };

        setToasts((prev) => {
          // ignore duplicate messages if already showing
          if (prev.some((t) => t.title === newToast.title && t.message === newToast.message && t.open)) {
            return prev;
          }
          return [newToast, ...prev.slice(0, 2)];
        });
      }
    };

    wsClient.on("notifications", handleNotification);
    return () => {
      wsClient.off("notifications", handleNotification);
    };
  }, [addNotification]);

  const handleOpenChange = (id: string, open: boolean) => {
    if (!open) {
      setToasts((prev) => prev.filter((t) => t.id !== id));
    } else {
      setToasts((prev) =>
        prev.map((t) => (t.id === id ? { ...t, open } : t))
      );
    }
  };

  const getVariant = (severity: string): "default" | "success" | "warning" | "error" => {
    switch (severity) {
      case "critical":
      case "error":
        return "error";
      case "warning":
        return "warning";
      case "success":
        return "success";
      default:
        return "default";
    }
  };

  const getIcon = (severity: string) => {
    switch (severity) {
      case "critical":
      case "error":
        return <AlertCircle className="w-4 h-4 text-rose-400 shrink-0 mt-0.5" />;
      case "warning":
        return <AlertTriangle className="w-4 h-4 text-amber-400 shrink-0 mt-0.5" />;
      case "success":
        return <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0 mt-0.5" />;
      default:
        return <Info className="w-4 h-4 text-cyan-400 shrink-0 mt-0.5" />;
    }
  };

  // determine duration: CRITICAL persistent (1000000ms), WARNING 10s, INFO 5s
  const getDuration = (severity: string) => {
    if (severity === "critical") {
      return 10000000; // persistent
    }
    if (severity === "warning") {
      return 10000; // 10s
    }
    return 5000; // 5s
  };

  return (
    <ToastProvider swipeDirection="right">
      {children}

      {toasts.map((toast) => (
        <Toast
          key={toast.id}
          open={toast.open}
          onOpenChange={(open) => handleOpenChange(toast.id, open)}
          duration={getDuration(toast.severity)}
          variant={getVariant(toast.severity)}
        >
          <div className="flex items-start gap-2.5">
            {getIcon(toast.severity)}
            <div className="grid gap-1">
              <ToastTitle className="flex items-center gap-2">
                <span>{toast.title}</span>
                <span className="text-[10px] text-[var(--text-muted)] font-mono">
                  {toast.timestamp}
                </span>
              </ToastTitle>
              <ToastDescription>{toast.message}</ToastDescription>
            </div>
          </div>
        </Toast>
      ))}

      <ToastViewport />
    </ToastProvider>
  );
}

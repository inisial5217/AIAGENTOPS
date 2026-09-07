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
import { wsClient } from "../../lib/ws-client";
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
  const recentAlertsRef = React.useRef<Map<string, number>>(new Map());

  // subscribe to notifications topic
  useWebSocket(["notifications"]);

  React.useEffect(() => {
    const handleNotification = (msg: WSMessage) => {
      if (msg.type === "notification" && msg.data) {
        const payload = msg.data as NotificationPayload;
        const alertKey = `${payload.title}:${payload.message}`;
        const now = Date.now();

        // 1. Synchronous deduplication: ignore identical alert if received within last 10 seconds
        const lastSeen = recentAlertsRef.current.get(alertKey);
        if (lastSeen && now - lastSeen < 10000) {
          return;
        }
        recentAlertsRef.current.set(alertKey, now);

        // Prune old entries from ref
        if (recentAlertsRef.current.size > 200) {
          for (const [k, ts] of recentAlertsRef.current.entries()) {
            if (now - ts > 30000) {
              recentAlertsRef.current.delete(k);
            }
          }
        }

        // 2. Add to global notification store for bell center
        addNotification({
          title: payload.title,
          message: payload.message,
          severity: payload.severity,
        });

        // 3. Add to active toast queue with guaranteed unique ID
        const uniqueId = payload.id || `toast-${now}-${Math.random().toString(36).slice(2, 7)}`;
        const newToast: ActiveToast = {
          ...payload,
          id: uniqueId,
          open: true,
        };

        // 4. Hard cap of at most 3 toasts on screen simultaneously
        setToasts((prev) => {
          const filtered = prev.filter(
            (t) => t.open && t.id !== newToast.id && !(t.title === newToast.title && t.message === newToast.message)
          );
          return [newToast, ...filtered.slice(0, 2)];
        });
      }
    };

    wsClient.on("notifications", handleNotification);
    return () => {
      wsClient.off("notifications", handleNotification);
    };
  }, [addNotification]);

  const handleOpenChange = (id: string, open: boolean) => {
    setToasts((prev) => {
      if (!open) {
        return prev.filter((t) => t.id !== id);
      }
      return prev.map((t) => (t.id === id ? { ...t, open } : t));
    });
  };

  const handleDismissAll = () => {
    setToasts([]);
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

  // determine duration: CRITICAL 12s, WARNING 5s, INFO 4s
  const getDuration = (severity: string) => {
    if (severity === "critical") {
      return 12000;
    }
    if (severity === "warning") {
      return 5000;
    }
    return 4000;
  };

  return (
    <ToastProvider swipeDirection="right">
      {children}

      {toasts.map((toast, index) => (
        <Toast
          key={`${toast.id}-${index}`}
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

      {toasts.length > 1 && (
        <div className="fixed bottom-3 right-4 z-50 pointer-events-auto">
          <button
            type="button"
            onClick={handleDismissAll}
            className="px-2.5 py-1 text-[11px] font-mono rounded bg-zinc-900/90 hover:bg-zinc-800 text-amber-300 border border-amber-500/40 shadow-xl cursor-pointer transition-colors backdrop-blur-sm flex items-center gap-1.5"
          >
            <span>Tutup Semua ({toasts.length})</span>
          </button>
        </div>
      )}

      <ToastViewport />
    </ToastProvider>
  );
}

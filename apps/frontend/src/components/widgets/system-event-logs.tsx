"use client";

import * as React from "react";
import { Download, Terminal, Radio } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { Skeleton } from "../ui/skeleton";
import { Badge } from "../ui/badge";
import apiClient from "../../lib/api-client";
import { useWebSocket } from "../../hooks/use-websocket";
import { EventPayload, WSMessage } from "../../types/websocket";

export interface LogEntry {
  id: string;
  time: string;
  tag: "INFO" | "CRITICAL" | "AI-OPS" | "WAITING" | "WARN";
  message: string;
}

interface AuditLogItem {
  id: string;
  user_id: string;
  username: string;
  action: string;
  resource: string;
  ip_address: string;
  status: string;
  details?: string;
  created_at: string;
}

const defaultFallbackLogs: LogEntry[] = [
  {
    id: "l-1",
    time: "10:14:02",
    tag: "INFO",
    message: "Docker daemon sync connected (//./pipe/docker_engine).",
  },
  {
    id: "l-2",
    time: "10:14:15",
    tag: "INFO",
    message: "Container prober engine initialized across namespaces.",
  },
  {
    id: "l-3",
    time: "10:14:18",
    tag: "AI-OPS",
    message: "Autonomous telemetry collector started for Docker host.",
  },
  {
    id: "l-4",
    time: "10:14:25",
    tag: "INFO",
    message: "Docker images and volumes cache pre-warmed via Redis.",
  },
  {
    id: "l-5",
    time: "10:14:30",
    tag: "INFO",
    message: "System ready & listening for operational commands.",
  },
];

export function SystemEventLogs({ isLoading = false }: { isLoading?: boolean }) {
  const [autoScroll, setAutoScroll] = React.useState(true);
  const [liveEvents, setLiveEvents] = React.useState<LogEntry[]>([]);
  const logsContainerRef = React.useRef<HTMLDivElement>(null);

  // connect to system_events via websocket
  const { isConnected } = useWebSocket(["system_events"]);

  // listen to real-time events from websocket
  React.useEffect(() => {
    const { wsClient } = require("../../lib/ws-client");

    const handleSystemEvent = (msg: WSMessage) => {
      if ((msg.type === "system_event" || msg.type === "container_event" || msg.type === "k8s_event") && msg.data) {
        const payload = msg.data as EventPayload;
        let tag: LogEntry["tag"] = "INFO";
        const act = (payload.action || "").toLowerCase();

        if (act === "die" || act === "oom" || act.includes("fail") || act.includes("error")) {
          tag = "CRITICAL";
        } else if (act === "stop" || act === "pause" || act.includes("warn") || act.includes("scale")) {
          tag = "WARN";
        } else if (payload.type === "kubernetes") {
          tag = "AI-OPS";
        }

        const newEntry: LogEntry = {
          id: `ws-${Date.now()}-${Math.random().toString(36).substring(2, 7)}`,
          time: payload.timestamp || new Date().toLocaleTimeString(),
          tag,
          message: payload.message || `[${payload.type}] ${payload.resource}: ${payload.action}`,
        };

        setLiveEvents((prev) => [...prev.slice(-100), newEntry]);
      }
    };

    wsClient.on("system_events", handleSystemEvent);
    return () => {
      wsClient.off("system_events", handleSystemEvent);
    };
  }, []);

  // Fetch initial audit logs from backend database as baseline
  const { data: auditLogs = [], isLoading: isAuditLoading } = useQuery<AuditLogItem[]>({
    queryKey: ["admin", "audit-logs"],
    queryFn: async () => {
      try {
        const res = await apiClient.get<{ status: string; data: AuditLogItem[] }>("/api/v1/admin/audit-logs");
        return res.data?.data || [];
      } catch {
        return [];
      }
    },
    refetchInterval: isConnected ? false : 15000,
  });

  const formattedLogs: LogEntry[] = React.useMemo(() => {
    const base: LogEntry[] =
      auditLogs && auditLogs.length > 0
        ? auditLogs.map((log) => {
            const date = new Date(log.created_at);
            const time = isNaN(date.getTime())
              ? "00:00:00"
              : date.toTimeString().split(" ")[0];

            let tag: LogEntry["tag"] = "INFO";
            if (log.status === "failed") tag = "CRITICAL";
            else if (log.action.includes("STOP") || log.action.includes("DELETE")) tag = "WARN";
            else if (log.action.includes("AI_") || log.action.includes("AUTO_")) tag = "AI-OPS";

            return {
              id: log.id,
              time,
              tag,
              message: `${log.username || "system"} performed ${log.action} on ${log.resource} (${log.status})`,
            };
          })
        : defaultFallbackLogs;

    // combine baseline audit logs with incoming live websocket events
    return [...base, ...liveEvents];
  }, [auditLogs, liveEvents]);

  React.useEffect(() => {
    if (autoScroll && logsContainerRef.current) {
      logsContainerRef.current.scrollTop = logsContainerRef.current.scrollHeight;
    }
  }, [autoScroll, formattedLogs]);

  const tagColors: Record<string, string> = {
    INFO: "text-cyan-400 font-semibold",
    CRITICAL: "text-rose-500 font-bold",
    "AI-OPS": "text-sky-300 font-semibold",
    WAITING: "text-amber-400 font-semibold",
    WARN: "text-yellow-400 font-semibold",
  };

  const handleExport = () => {
    const content = formattedLogs
      .map((l) => `[${l.time}] [${l.tag}] ${l.message}`)
      .join("\n");
    const blob = new Blob([content], { type: "text/plain" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `cifo-event-logs-${Date.now()}.txt`;
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div className="rounded-[var(--radius-xl)] bg-[var(--bg-card)] border border-[var(--border-default)] p-5 shadow-lg flex flex-col h-[320px] text-left">
      <div className="flex items-center justify-between pb-3 border-b border-[var(--border-subtle)] shrink-0">
        <div className="flex items-center gap-2">
          <Terminal className="w-4 h-4 text-[var(--accent-default)]" />
          <h3 className="text-sm font-semibold text-[var(--text-primary)]">
            System Event Logs
          </h3>
          <div className="flex items-center gap-1 ml-1.5">
            <Radio
              className={`w-3 h-3 ${
                isConnected ? "text-emerald-400 animate-pulse" : "text-slate-400"
              }`}
            />
            <Badge variant={isConnected ? "success" : "neutral"} size="sm">
              {isConnected ? "LIVE WS" : "POLLING"}
            </Badge>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setAutoScroll(!autoScroll)}
            className={`px-2 py-0.5 rounded text-[10px] font-mono border transition-colors cursor-pointer ${
              autoScroll
                ? "bg-cyan-500/10 text-cyan-400 border-cyan-500/30"
                : "bg-gray-800 text-gray-400 border-gray-700"
            }`}
          >
            Auto-scroll: {autoScroll ? "ON" : "OFF"}
          </button>
          <button
            onClick={handleExport}
            className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-mono bg-[var(--bg-secondary)] hover:bg-[var(--bg-hover)] border border-[var(--border-default)] text-[var(--text-secondary)] hover:text-[var(--text-primary)] transition-colors cursor-pointer"
          >
            <Download className="w-3 h-3" />
            Export
          </button>
        </div>
      </div>

      <div
        ref={logsContainerRef}
        className="mt-3 flex-1 overflow-y-auto font-mono text-[11px] leading-relaxed space-y-2 pr-1"
      >
        {isLoading || isAuditLoading ? (
          <div className="space-y-2 pt-1">
            <Skeleton variant="text" className="h-4 w-5/6" />
            <Skeleton variant="text" className="h-4 w-full" />
            <Skeleton variant="text" className="h-4 w-4/5" />
            <Skeleton variant="text" className="h-4 w-11/12" />
            <Skeleton variant="text" className="h-4 w-3/4" />
          </div>
        ) : (
          formattedLogs.map((log) => (
            <div
              key={log.id}
              className="flex items-start gap-2 hover:bg-[var(--bg-secondary)]/50 p-1 rounded transition-colors"
            >
              <span className="text-[var(--text-muted)] shrink-0">
                [{log.time}]
              </span>
              <span className={tagColors[log.tag] || "text-gray-400"}>
                [{log.tag}]
              </span>
              <span className="text-[var(--text-primary)] break-all">
                {log.message}
              </span>
            </div>
          ))
        )}
      </div>
    </div>
  );
}

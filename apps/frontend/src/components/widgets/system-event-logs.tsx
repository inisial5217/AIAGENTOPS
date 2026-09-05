"use client";

import * as React from "react";
import { Download, Terminal } from "lucide-react";
import { Skeleton } from "../ui/skeleton";

export interface LogEntry {
  id: string;
  time: string;
  tag: "INFO" | "CRITICAL" | "AI-OPS" | "WAITING" | "WARN";
  message: string;
}

const defaultLogs: LogEntry[] = [
  {
    id: "l-1",
    time: "10:14:02",
    tag: "INFO",
    message: "Docker daemon sync completed successfully.",
  },
  {
    id: "l-2",
    time: "10:14:15",
    tag: "INFO",
    message: "Prober engine checking 145 targets...",
  },
  {
    id: "l-3",
    time: "10:14:18",
    tag: "CRITICAL",
    message: "target 'payment-gateway' CONNECTION TIMEOUT!",
  },
  {
    id: "l-4",
    time: "10:14:19",
    tag: "CRITICAL",
    message: "node-worker-1 CPU spiked to 99%.",
  },
  {
    id: "l-5",
    time: "10:14:20",
    tag: "AI-OPS",
    message: "Agent initialized Root Cause Analysis...",
  },
  {
    id: "l-6",
    time: "10:14:25",
    tag: "AI-OPS",
    message: "Generating incident report #8842 into Database.",
  },
  {
    id: "l-7",
    time: "10:14:30",
    tag: "WAITING",
    message: "Waiting for user approval on Auto-Remediation.",
  },
];

export function SystemEventLogs({ isLoading = false }: { isLoading?: boolean }) {
  const [autoScroll, setAutoScroll] = React.useState(true);
  const logsContainerRef = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    if (autoScroll && logsContainerRef.current) {
      logsContainerRef.current.scrollTop = logsContainerRef.current.scrollHeight;
    }
  }, [autoScroll]);

  const tagColors: Record<string, string> = {
    INFO: "text-cyan-400 font-semibold",
    CRITICAL: "text-rose-500 font-bold",
    "AI-OPS": "text-sky-300 font-semibold",
    WAITING: "text-amber-400 font-semibold",
    WARN: "text-yellow-400 font-semibold",
  };

  const handleExport = () => {
    const content = defaultLogs
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
        {isLoading ? (
          <div className="space-y-2 pt-1">
            <Skeleton variant="text" className="h-4 w-5/6" />
            <Skeleton variant="text" className="h-4 w-full" />
            <Skeleton variant="text" className="h-4 w-4/5" />
            <Skeleton variant="text" className="h-4 w-11/12" />
            <Skeleton variant="text" className="h-4 w-3/4" />
          </div>
        ) : (
          defaultLogs.map((log) => (
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

"use client";

import * as React from "react";
import {
  Search,
  Trash2,
  Download,
  ArrowDownCircle,
  Radio,
  RefreshCw,
} from "lucide-react";
import { Button } from "../ui/button";
import { Badge } from "../ui/badge";
import { useWebSocket } from "../../hooks/use-websocket";
import { LogPayload, WSMessage } from "../../types/websocket";

export interface LogTerminalProps {
  topic?: string;
  initialLogs?: string[];
  title?: string;
  height?: string;
  maxLines?: number;
  onRefresh?: () => void;
}

export function LogTerminal({
  topic,
  initialLogs = [],
  title = "Live Container Output",
  height = "h-[450px]",
  maxLines = 5000,
  onRefresh,
}: LogTerminalProps) {
  const [logs, setLogs] = React.useState<string[]>(initialLogs);
  const [filterText, setFilterText] = React.useState("");
  const [autoScroll, setAutoScroll] = React.useState(true);
  const terminalEndRef = React.useRef<HTMLDivElement>(null);
  const containerRef = React.useRef<HTMLDivElement>(null);

  // connect to ws topic if provided
  const topics = React.useMemo(() => (topic ? [topic] : []), [topic]);
  const { isConnected, status } = useWebSocket(topics, { enabled: !!topic });

  // sync initial logs if updated
  React.useEffect(() => {
    if (initialLogs && initialLogs.length > 0) {
      setLogs((prev) => {
        if (prev.length === 0) return initialLogs;
        return prev;
      });
    }
  }, [initialLogs]);

  // listen to ws messages for this topic
  React.useEffect(() => {
    if (!topic) return;

    const { wsClient } = require("../../lib/ws-client");
    const handleLog = (msg: WSMessage) => {
      if (msg.type === "log_entry" && msg.data) {
        const payload = msg.data as LogPayload;
        if (payload.log) {
          setLogs((prev) => {
            const next = [...prev, payload.log];
            if (next.length > maxLines) {
              return next.slice(next.length - maxLines);
            }
            return next;
          });
        }
      }
    };

    wsClient.on(topic, handleLog);
    return () => {
      wsClient.off(topic, handleLog);
    };
  }, [topic, maxLines]);

  // handle auto scroll
  React.useEffect(() => {
    if (autoScroll && containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  }, [logs, autoScroll]);

  // manual scroll detection to toggle auto-scroll
  const handleScroll = () => {
    if (!containerRef.current) return;
    const { scrollTop, scrollHeight, clientHeight } = containerRef.current;
    const isAtBottom = scrollHeight - scrollTop - clientHeight < 40;
    if (isAtBottom !== autoScroll) {
      setAutoScroll(isAtBottom);
    }
  };

  // filter logs
  const filteredLogs = React.useMemo(() => {
    if (!filterText.trim()) return logs;
    const query = filterText.toLowerCase();
    return logs.filter((line) => line.toLowerCase().includes(query));
  }, [logs, filterText]);

  // clear logs
  const handleClear = () => {
    setLogs([]);
  };

  // export logs to .log file
  const handleExport = () => {
    const content = logs.join("\n");
    const blob = new Blob([content], { type: "text/plain;charset=utf-8" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `terminal-logs-${Date.now()}.log`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  };

  // format syntax highlighting for log line
  const formatLogLine = (line: string) => {
    const lower = line.toLowerCase();
    let textColor = "text-slate-300";

    if (
      lower.includes("error") ||
      lower.includes("fatal") ||
      lower.includes("critical") ||
      lower.includes("failed") ||
      lower.includes("panic")
    ) {
      textColor = "text-rose-400 font-medium";
    } else if (lower.includes("warn") || lower.includes("warning")) {
      textColor = "text-amber-400";
    } else if (
      lower.includes("info") ||
      lower.includes("success") ||
      lower.includes("started") ||
      lower.includes("ready") ||
      lower.includes("listening")
    ) {
      textColor = "text-emerald-400";
    } else if (lower.includes("debug") || lower.includes("trace")) {
      textColor = "text-cyan-400/80";
    }

    if (!filterText.trim()) {
      return <span className={textColor}>{line}</span>;
    }

    // highlight search match
    const parts = line.split(new RegExp(`(${filterText})`, "gi"));
    return (
      <span className={textColor}>
        {parts.map((part, i) =>
          part.toLowerCase() === filterText.toLowerCase() ? (
            <mark key={i} className="bg-yellow-500/30 text-yellow-200 px-0.5 rounded">
              {part}
            </mark>
          ) : (
            part
          )
        )}
      </span>
    );
  };

  return (
    <div className="flex flex-col rounded-xl border border-[var(--border-default)] bg-[#070a12] shadow-2xl overflow-hidden font-mono text-xs">
      {/* Terminal Header Bar */}
      <div className="flex flex-wrap items-center justify-between gap-3 px-4 py-2.5 bg-[#0d121f] border-b border-[var(--border-default)]">
        <div className="flex items-center gap-2.5">
          {/* Mac window dots */}
          <div className="flex items-center gap-1.5 mr-1">
            <span className="w-2.5 h-2.5 rounded-full bg-rose-500/80 inline-block" />
            <span className="w-2.5 h-2.5 rounded-full bg-amber-500/80 inline-block" />
            <span className="w-2.5 h-2.5 rounded-full bg-emerald-500/80 inline-block" />
          </div>

          <span className="font-semibold text-slate-200">{title}</span>

          {topic && (
            <div className="flex items-center gap-1.5 ml-2">
              <Radio
                className={`w-3.5 h-3.5 ${
                  isConnected ? "text-emerald-400 animate-pulse" : "text-amber-400"
                }`}
              />
              <Badge variant={isConnected ? "success" : "warning"} size="sm">
                {isConnected ? "LIVE STREAM" : status.toUpperCase()}
              </Badge>
            </div>
          )}

          <span className="text-[11px] text-slate-400 font-sans ml-1">
            ({logs.length.toLocaleString()} lines)
          </span>
        </div>

        {/* Action Controls */}
        <div className="flex items-center gap-2">
          {/* Search Filter Bar */}
          <div className="relative w-40 sm:w-48">
            <Search className="w-3.5 h-3.5 absolute left-2.5 top-1/2 -translate-y-1/2 text-slate-500" />
            <input
              type="text"
              value={filterText}
              onChange={(e) => setFilterText(e.target.value)}
              placeholder="Search logs..."
              className="w-full h-7 pl-8 pr-2 rounded-md bg-[#131929] border border-slate-700/60 text-[11px] text-slate-200 placeholder:text-slate-500 focus:outline-none focus:border-cyan-500/50"
            />
            {filterText && (
              <span className="absolute right-2 top-1/2 -translate-y-1/2 text-[10px] text-cyan-400 font-sans">
                {filteredLogs.length}
              </span>
            )}
          </div>

          {/* Auto Scroll Toggle */}
          <button
            type="button"
            onClick={() => setAutoScroll(!autoScroll)}
            className={`flex items-center gap-1 px-2.5 h-7 rounded-md border text-[11px] transition-colors cursor-pointer ${
              autoScroll
                ? "bg-cyan-950/50 border-cyan-500/40 text-cyan-300 font-medium"
                : "bg-[#131929] border-slate-700/60 text-slate-400 hover:text-slate-200"
            }`}
            title="Auto-scroll to latest log"
          >
            <ArrowDownCircle className="w-3 h-3" />
            <span className="hidden sm:inline">Auto-scroll</span>
          </button>

          {/* Refresh Button */}
          {onRefresh && (
            <Button
              variant="outline"
              size="sm"
              onClick={onRefresh}
              className="h-7 px-2 border-slate-700/60 text-slate-300 hover:text-white"
              title="Refresh logs from server"
            >
              <RefreshCw className="w-3 h-3" />
            </Button>
          )}

          {/* Clear Button */}
          <Button
            variant="outline"
            size="sm"
            onClick={handleClear}
            className="h-7 px-2 border-slate-700/60 text-slate-300 hover:text-rose-300"
            title="Clear terminal display"
          >
            <Trash2 className="w-3 h-3" />
          </Button>

          {/* Export Button */}
          <Button
            variant="outline"
            size="sm"
            onClick={handleExport}
            className="h-7 px-2 border-slate-700/60 text-slate-300 hover:text-emerald-300"
            title="Export log output to file"
          >
            <Download className="w-3 h-3" />
          </Button>
        </div>
      </div>

      {/* Terminal Log Viewport */}
      <div
        ref={containerRef}
        onScroll={handleScroll}
        className={`${height} overflow-y-auto p-4 space-y-0.5 select-text selection:bg-cyan-500/30 selection:text-cyan-200`}
        style={{ scrollBehavior: autoScroll ? "smooth" : "auto" }}
      >
        {filteredLogs.length === 0 ? (
          <div className="h-full flex flex-col items-center justify-center text-slate-500 py-12">
            <Radio className="w-6 h-6 mb-2 opacity-40 animate-pulse text-cyan-400" />
            <p className="text-sm font-sans">
              {logs.length === 0
                ? "Waiting for stream output..."
                : `No logs matching "${filterText}"`}
            </p>
          </div>
        ) : (
          filteredLogs.map((line, idx) => (
            <div
              key={idx}
              className="leading-relaxed hover:bg-slate-900/50 px-1 rounded flex items-start gap-2 group"
            >
              <span className="text-slate-600 select-none text-[10px] w-8 shrink-0 text-right opacity-60 group-hover:opacity-100">
                {idx + 1}
              </span>
              <div className="break-all whitespace-pre-wrap flex-1">
                {formatLogLine(line)}
              </div>
            </div>
          ))
        )}
        <div ref={terminalEndRef} />
      </div>
    </div>
  );
}

"use client";

import * as React from "react";
import { Cpu, HardDrive, Wifi, Sparkles, RefreshCw } from "lucide-react";

export function HostResourceUsage() {
  const [activeTab, setActiveTab] = React.useState<"cpu" | "network" | "ai">("cpu");
  const [isRefreshing, setIsRefreshing] = React.useState(false);

  // Simulated live telemetry series
  const [dataPoints, setDataPoints] = React.useState([
    { time: "10:00", cpu: 32, ram: 62 },
    { time: "10:05", cpu: 45, ram: 63 },
    { time: "10:10", cpu: 38, ram: 64 },
    { time: "10:15", cpu: 52, ram: 63 },
    { time: "10:20", cpu: 78, ram: 65 },
    { time: "10:25", cpu: 42, ram: 64 },
    { time: "10:30", cpu: 48, ram: 64 },
    { time: "10:35", cpu: 55, ram: 64 },
  ]);

  const handleRefresh = () => {
    setIsRefreshing(true);
    setTimeout(() => {
      setDataPoints((prev) =>
        prev.map((pt) => ({
          ...pt,
          cpu: Math.floor(30 + Math.random() * 40),
          ram: Math.floor(62 + Math.random() * 4),
        }))
      );
      setIsRefreshing(false);
    }, 600);
  };

  // Generate SVG path for multi-point area chart
  const getPolyline = (key: "cpu" | "ram") => {
    const width = 500;
    const height = 140;
    const stepX = width / (dataPoints.length - 1);

    return dataPoints
      .map((d, i) => {
        const x = i * stepX;
        const val = d[key];
        const y = height - (val / 100) * height;
        return `${x},${y}`;
      })
      .join(" ");
  };

  const getAreaPolygon = (key: "cpu" | "ram") => {
    const polyline = getPolyline(key);
    const width = 500;
    const height = 140;
    return `0,${height} ${polyline} ${width},${height}`;
  };

  return (
    <div className="bg-[var(--bg-card)] border border-[var(--border-default)] rounded-xl p-5 shadow-sm flex flex-col h-[320px]">
      {/* Header */}
      <div className="flex items-center justify-between pb-3 border-b border-[var(--border-subtle)]">
        <div className="flex items-center gap-2.5 text-left">
          <div className="p-2 rounded-lg bg-cyan-500/10 border border-cyan-500/30 text-cyan-400">
            <Cpu className="w-4 h-4" />
          </div>
          <div>
            <h3 className="text-sm font-bold text-[var(--text-primary)]">
              Host Resource Usage
            </h3>
            <span className="text-[11px] font-mono text-[var(--text-muted)]">
              production-cifo-1 (Ubuntu 24.04 LTS)
            </span>
          </div>
        </div>

        {/* Tab Switcher & Refresh */}
        <div className="flex items-center gap-2">
          <div className="flex bg-[var(--bg-secondary)] p-0.5 rounded-lg border border-[var(--border-subtle)] text-[11px] font-mono">
            <button
              onClick={() => setActiveTab("cpu")}
              className={`px-2.5 py-1 rounded-md transition-all cursor-pointer ${
                activeTab === "cpu"
                  ? "bg-[var(--bg-card)] text-cyan-400 font-bold shadow-sm"
                  : "text-[var(--text-muted)] hover:text-[var(--text-primary)]"
              }`}
            >
              CPU & RAM
            </button>
            <button
              onClick={() => setActiveTab("network")}
              className={`px-2.5 py-1 rounded-md transition-all cursor-pointer ${
                activeTab === "network"
                  ? "bg-[var(--bg-card)] text-cyan-400 font-bold shadow-sm"
                  : "text-[var(--text-muted)] hover:text-[var(--text-primary)]"
              }`}
            >
              Network I/O
            </button>
            <button
              onClick={() => setActiveTab("ai")}
              className={`px-2.5 py-1 rounded-md transition-all cursor-pointer ${
                activeTab === "ai"
                  ? "bg-[var(--bg-card)] text-cyan-400 font-bold shadow-sm"
                  : "text-[var(--text-muted)] hover:text-[var(--text-primary)]"
              }`}
            >
              AI Tokens
            </button>
          </div>

          <button
            onClick={handleRefresh}
            title="Refresh metrics"
            className="p-1.5 rounded-lg border border-[var(--border-subtle)] bg-[var(--bg-secondary)] text-[var(--text-muted)] hover:text-[var(--text-primary)] cursor-pointer"
          >
            <RefreshCw
              className={`w-3.5 h-3.5 ${isRefreshing ? "animate-spin text-cyan-400" : ""}`}
            />
          </button>
        </div>
      </div>

      {/* Chart Canvas */}
      <div className="flex-1 pt-3 flex flex-col justify-between">
        <div className="relative w-full h-[140px] overflow-hidden rounded-lg bg-[var(--bg-secondary)]/40 p-2">
          {activeTab === "cpu" && (
            <svg
              viewBox="0 0 500 140"
              preserveAspectRatio="none"
              className="w-full h-full overflow-visible"
            >
              <defs>
                <linearGradient id="cyanGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="#00f2fe" stopOpacity="0.35" />
                  <stop offset="100%" stopColor="#00f2fe" stopOpacity="0.0" />
                </linearGradient>
                <linearGradient id="pinkGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="#ff007a" stopOpacity="0.3" />
                  <stop offset="100%" stopColor="#ff007a" stopOpacity="0.0" />
                </linearGradient>
              </defs>

              {/* Grid Lines */}
              <line x1="0" y1="35" x2="500" y2="35" stroke="var(--border-subtle)" strokeDasharray="3 3" />
              <line x1="0" y1="70" x2="500" y2="70" stroke="var(--border-subtle)" strokeDasharray="3 3" />
              <line x1="0" y1="105" x2="500" y2="105" stroke="var(--border-subtle)" strokeDasharray="3 3" />

              {/* RAM Area & Line */}
              <polygon points={getAreaPolygon("ram")} fill="url(#pinkGrad)" />
              <polyline
                points={getPolyline("ram")}
                fill="none"
                stroke="#ff007a"
                strokeWidth="2"
              />

              {/* CPU Area & Line */}
              <polygon points={getAreaPolygon("cpu")} fill="url(#cyanGrad)" />
              <polyline
                points={getPolyline("cpu")}
                fill="none"
                stroke="#00f2fe"
                strokeWidth="2.5"
              />
            </svg>
          )}

          {activeTab === "network" && (
            <div className="w-full h-full flex flex-col justify-center items-center gap-2 text-xs font-mono text-[var(--text-secondary)]">
              <Wifi className="w-6 h-6 text-emerald-400 animate-pulse" />
              <span>Ingress: 12.4 MB/s &bull; Egress: 8.9 MB/s</span>
              <span className="text-[10px] text-[var(--text-muted)]">Packets Dropped: 0.00%</span>
            </div>
          )}

          {activeTab === "ai" && (
            <div className="w-full h-full flex flex-col justify-center items-center gap-2 text-xs font-mono text-[var(--text-secondary)]">
              <Sparkles className="w-6 h-6 text-cyan-400 animate-pulse" />
              <span>Prompt Tokens: 8,420 &bull; Completion Tokens: 5,780</span>
              <span className="text-[10px] text-[var(--text-muted)]">Avg Inference: 42ms (Gemini 1.5)</span>
            </div>
          )}
        </div>

        {/* Legend / Metrics Footer */}
        <div className="flex items-center justify-between pt-2 px-1 text-xs font-mono border-t border-[var(--border-subtle)]">
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-1.5">
              <span className="w-2.5 h-2.5 rounded-full bg-[#00f2fe]" />
              <span className="text-[var(--text-secondary)]">
                CPU: <strong className="text-[var(--text-primary)]">48%</strong>
              </span>
            </div>
            <div className="flex items-center gap-1.5">
              <span className="w-2.5 h-2.5 rounded-full bg-[#ff007a]" />
              <span className="text-[var(--text-secondary)]">
                RAM: <strong className="text-[var(--text-primary)]">64%</strong> (41.2 GB)
              </span>
            </div>
          </div>

          <div className="flex items-center gap-1 text-[10px] text-[var(--text-muted)]">
            <HardDrive className="w-3 h-3" />
            <span>NVMe Root: 182 GB / 500 GB (36%)</span>
          </div>
        </div>
      </div>
    </div>
  );
}

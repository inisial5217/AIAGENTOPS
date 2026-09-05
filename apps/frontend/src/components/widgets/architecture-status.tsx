"use client";

import * as React from "react";
import { Server, Activity, GitBranch, Cpu, CheckCircle2, ShieldAlert } from "lucide-react";
import { Badge } from "../ui/badge";

export function ArchitectureStatus() {
  return (
    <div className="bg-[var(--bg-card)] border border-[var(--border-default)] rounded-xl p-5 shadow-sm flex flex-col h-[320px]">
      {/* Header */}
      <div className="flex items-center justify-between pb-3 border-b border-[var(--border-subtle)]">
        <div className="flex items-center gap-2.5 text-left">
          <div className="p-2 rounded-lg bg-pink-500/10 border border-pink-500/30 text-[var(--accent-pink)]">
            <Server className="w-4 h-4" />
          </div>
          <div>
            <h3 className="text-sm font-bold text-[var(--text-primary)]">
              Agent Architecture & Model Usage
            </h3>
            <span className="text-[11px] font-mono text-[var(--text-muted)]">
              Sub-system Telemetry & AI Model Engine
            </span>
          </div>
        </div>

        <Badge variant="cyan" size="sm" pulse>
          System Nominal
        </Badge>
      </div>

      {/* Grid of Microservice Engines */}
      <div className="grid grid-cols-2 gap-3 py-3 flex-1">
        {/* Prober Engine */}
        <div className="p-3 rounded-lg bg-[var(--bg-secondary)]/70 border border-[var(--border-subtle)] flex flex-col justify-between text-left">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Activity className="w-4 h-4 text-emerald-400" />
              <span className="text-xs font-bold text-[var(--text-primary)]">
                Prober Daemon
              </span>
            </div>
            <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              8ms latency
            </span>
          </div>
          <div className="mt-2 text-[11px] text-[var(--text-secondary)] font-mono">
            TCP/HTTP Syn Probing active across 12 namespaces.
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-[10px] font-mono text-emerald-400">
            <CheckCircle2 className="w-3 h-3" />
            <span>State: HEALTHY &bull; 1,248 pings/min</span>
          </div>
        </div>

        {/* Docker Daemon */}
        <div className="p-3 rounded-lg bg-[var(--bg-secondary)]/70 border border-[var(--border-subtle)] flex flex-col justify-between text-left">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Server className="w-4 h-4 text-cyan-400" />
              <span className="text-xs font-bold text-[var(--text-primary)]">
                Docker Daemon
              </span>
            </div>
            <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-cyan-500/10 text-cyan-400 border border-cyan-500/20">
              v26.1 API
            </span>
          </div>
          <div className="mt-2 text-[11px] text-[var(--text-secondary)] font-mono">
            Unix Socket streaming events to Redis pipeline.
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-[10px] font-mono text-cyan-400">
            <CheckCircle2 className="w-3 h-3" />
            <span>145 Containers &bull; 22 Images cached</span>
          </div>
        </div>

        {/* ArgoCD Sync */}
        <div className="p-3 rounded-lg bg-[var(--bg-secondary)]/70 border border-[var(--border-subtle)] flex flex-col justify-between text-left">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <GitBranch className="w-4 h-4 text-amber-400" />
              <span className="text-xs font-bold text-[var(--text-primary)]">
                ArgoCD Controller
              </span>
            </div>
            <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-amber-500/10 text-amber-400 border border-amber-500/20">
              GitOps
            </span>
          </div>
          <div className="mt-2 text-[11px] text-[var(--text-secondary)] font-mono">
            Target commit: <span className="text-cyan-400">#f389a1c</span> (main branch)
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-[10px] font-mono text-amber-400">
            <CheckCircle2 className="w-3 h-3" />
            <span>Synced &bull; 0 Configuration Drifts</span>
          </div>
        </div>

        {/* Model AI Engine */}
        <div className="p-3 rounded-lg bg-[var(--bg-secondary)]/70 border border-[var(--border-subtle)] flex flex-col justify-between text-left">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Cpu className="w-4 h-4 text-[var(--accent-pink)]" />
              <span className="text-xs font-bold text-[var(--text-primary)]">
                Gemini 1.5 Flash
              </span>
            </div>
            <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-pink-500/10 text-[var(--accent-pink)] border border-pink-500/20">
              99.4% Acc
            </span>
          </div>
          <div className="mt-2 text-[11px] text-[var(--text-secondary)] font-mono">
            Autonomous Level 2 &bull; Context Window: 1M
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-[10px] font-mono text-[var(--accent-pink)]">
            <CheckCircle2 className="w-3 h-3" />
            <span>Active &bull; Cost: Free Tier / Dev</span>
          </div>
        </div>
      </div>
    </div>
  );
}

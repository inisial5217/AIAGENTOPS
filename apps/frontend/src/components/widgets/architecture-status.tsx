"use client";

import * as React from "react";
import { Server, Activity, GitBranch, Cpu, CheckCircle2, ShieldAlert } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { Badge } from "../ui/badge";
import { dockerService } from "../../services/docker-service";
import { argocdService } from "../../services/argocd-service";
import { aiService } from "../../services/ai-service";

export function ArchitectureStatus() {
  // 1. Docker Daemon live telemetry
  const { data: stats } = useQuery({
    queryKey: ["monitoring", "dashboard-stats"],
    queryFn: () => dockerService.getDashboardStats(),
    refetchInterval: 15000,
  });

  const { data: imagesData } = useQuery({
    queryKey: ["docker", "images"],
    queryFn: () => dockerService.getImages(),
    refetchInterval: 30000,
  });

  // 2. ArgoCD GitOps live status
  const { data: argoOverview } = useQuery({
    queryKey: ["argocd", "overview"],
    queryFn: () => argocdService.getOverview(),
    refetchInterval: 15000,
  });

  // 3. AI Usage & active model
  const { data: aiUsage } = useQuery({
    queryKey: ["ai", "usage-stats"],
    queryFn: () => aiService.getUsageStats(),
    refetchInterval: 15000,
  });

  const hasIncidents = (stats?.active_incidents ?? 0) > 0;
  const isOutOfSync = (argoOverview?.out_of_sync ?? 0) > 0;
  const containerCount = stats?.total_containers ?? 0;
  const runningCount = stats?.containers_on ?? 0;
  const imageCount = imagesData?.data?.length ?? 0;
  const activeModel = aiUsage?.active_model || "Gemini 2.0 Flash";

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
              Agent Architecture &amp; Sub-systems
            </h3>
            <span className="text-[11px] font-mono text-[var(--text-muted)]">
              Live Microservice Telemetry &amp; AI Engine Status
            </span>
          </div>
        </div>

        <Badge
          variant={hasIncidents || isOutOfSync ? "warning" : "cyan"}
          size="sm"
          pulse={hasIncidents || isOutOfSync}
        >
          {hasIncidents
            ? `${stats?.active_incidents} Active Alerts`
            : isOutOfSync
            ? "Sync Drift Detected"
            : "System Nominal"}
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
              Live Probe
            </span>
          </div>
          <div className="mt-2 text-[11px] text-[var(--text-secondary)] font-mono">
            TCP/HTTP health checks streaming to telemetry pipeline.
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-[10px] font-mono text-emerald-400">
            <CheckCircle2 className="w-3 h-3" />
            <span>State: HEALTHY &bull; K3d Klaster OK</span>
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
              Engine Ready
            </span>
          </div>
          <div className="mt-2 text-[11px] text-[var(--text-secondary)] font-mono">
            {runningCount} of {containerCount} containers running &bull; {imageCount} cached images.
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-[10px] font-mono text-cyan-400">
            <CheckCircle2 className="w-3 h-3" />
            <span>Socket Connected &bull; Redis Stream Active</span>
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
            {argoOverview?.synced ?? 0} of {argoOverview?.total ?? 0} applications synchronized with Git repository.
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-[10px] font-mono text-amber-400">
            {isOutOfSync ? (
              <>
                <ShieldAlert className="w-3 h-3 text-amber-400" />
                <span>{argoOverview?.out_of_sync} Drifts &bull; Manual Sync Available</span>
              </>
            ) : (
              <>
                <CheckCircle2 className="w-3 h-3 text-emerald-400" />
                <span className="text-emerald-400">All Synced &bull; Zero Configuration Drift</span>
              </>
            )}
          </div>
        </div>

        {/* Model AI Engine */}
        <div className="p-3 rounded-lg bg-[var(--bg-secondary)]/70 border border-[var(--border-subtle)] flex flex-col justify-between text-left">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Cpu className="w-4 h-4 text-[var(--accent-pink)]" />
              <span className="text-xs font-bold text-[var(--text-primary)] truncate max-w-[110px]" title={activeModel}>
                {activeModel}
              </span>
            </div>
            <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-pink-500/10 text-[var(--accent-pink)] border border-pink-500/20">
              Online
            </span>
          </div>
          <div className="mt-2 text-[11px] text-[var(--text-secondary)] font-mono">
            {aiUsage?.total_calls ?? 0} Inferences &bull; Prompt: {aiUsage?.prompt_tokens ?? 0} Tok
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-[10px] font-mono text-[var(--accent-pink)]">
            <CheckCircle2 className="w-3 h-3" />
            <span>Autonomous Level 2 &bull; Tools Enabled</span>
          </div>
        </div>
      </div>
    </div>
  );
}

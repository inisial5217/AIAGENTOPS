"use client";

import * as React from "react";
import Link from "next/link";
import { GitPullRequest, CheckCircle2, AlertCircle, ArrowUpRight, RefreshCw } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { argocdService } from "../../services/argocd-service";
import { Badge } from "../ui/badge";

export function ArgoCDStatusWidget() {
  const { data: overview, isLoading, refetch, isFetching } = useQuery({
    queryKey: ["argocd", "overview"],
    queryFn: () => argocdService.getOverview(),
    refetchInterval: 15000,
  });

  const total = overview?.total ?? 0;
  const synced = overview?.synced ?? 0;
  const outOfSync = overview?.out_of_sync ?? 0;
  const healthy = overview?.healthy ?? 0;

  return (
    <div className="bg-[var(--bg-card)] border border-[var(--border-default)] rounded-xl p-5 shadow-sm flex flex-col justify-between">
      {/* Header */}
      <div className="flex items-center justify-between pb-3 border-b border-[var(--border-subtle)]">
        <div className="flex items-center gap-2.5 text-left">
          <div className="p-2 rounded-lg bg-cyan-500/10 border border-cyan-500/30 text-cyan-400">
            <GitPullRequest className="w-4 h-4" />
          </div>
          <div>
            <h3 className="text-sm font-bold text-[var(--text-primary)]">
              ArgoCD GitOps Sync Status
            </h3>
            <span className="text-[11px] font-mono text-[var(--text-muted)]">
              Continuous Delivery &amp; Desired State
            </span>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => refetch()}
            disabled={isFetching}
            className="p-1.5 rounded text-[var(--text-muted)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-hover)] transition-colors cursor-pointer"
            title="Refresh GitOps Status"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${isFetching ? "animate-spin text-cyan-400" : ""}`} />
          </button>
          <Badge variant={outOfSync > 0 ? "warning" : "success"} size="sm" pulse={outOfSync > 0}>
            {outOfSync > 0 ? `${outOfSync} Out of Sync` : "All Synced"}
          </Badge>
        </div>
      </div>

      {/* Stats Body */}
      <div className="grid grid-cols-3 gap-3 py-4 text-left">
        <div className="p-3 rounded-lg bg-[var(--bg-secondary)]/60 border border-[var(--border-subtle)]">
          <span className="text-[11px] font-mono text-[var(--text-muted)]">Applications</span>
          <div className="text-xl font-bold font-mono text-[var(--text-primary)] mt-0.5">
            {isLoading ? "--" : total}
          </div>
          <span className="text-[10px] text-cyan-400 font-mono">Managed CRDs</span>
        </div>

        <div className="p-3 rounded-lg bg-[var(--bg-secondary)]/60 border border-[var(--border-subtle)]">
          <div className="flex items-center justify-between">
            <span className="text-[11px] font-mono text-[var(--text-muted)]">Synced</span>
            <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" />
          </div>
          <div className="text-xl font-bold font-mono text-emerald-400 mt-0.5">
            {isLoading ? "--" : synced}
          </div>
          <span className="text-[10px] text-[var(--text-muted)] font-mono">Match Target</span>
        </div>

        <div className="p-3 rounded-lg bg-[var(--bg-secondary)]/60 border border-[var(--border-subtle)]">
          <div className="flex items-center justify-between">
            <span className="text-[11px] font-mono text-[var(--text-muted)]">Health</span>
            <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
          </div>
          <div className="text-xl font-bold font-mono text-cyan-400 mt-0.5">
            {isLoading ? "--" : `${healthy}/${total}`}
          </div>
          <span className="text-[10px] text-[var(--text-muted)] font-mono">Healthy Workloads</span>
        </div>
      </div>

      {/* Footer Link */}
      <div className="pt-2 border-t border-[var(--border-subtle)] flex items-center justify-between">
        <span className="text-[11px] font-mono text-[var(--text-muted)]">
          Sync Engine: Active on K3d
        </span>
        <Link
          href="/argocd"
          className="inline-flex items-center gap-1 text-xs font-mono text-cyan-400 hover:text-cyan-300 transition-colors"
        >
          <span>Open GitOps Manager</span>
          <ArrowUpRight className="w-3.5 h-3.5" />
        </Link>
      </div>
    </div>
  );
}

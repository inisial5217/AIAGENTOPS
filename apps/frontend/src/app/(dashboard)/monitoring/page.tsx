"use client";

import * as React from "react";
import { useQuery } from "@tanstack/react-query";
import { StatCard } from "../../../components/widgets/stat-card";
import { HostResourceUsage } from "../../../components/widgets/host-resource-usage";
import { SystemEventLogs } from "../../../components/widgets/system-event-logs";
import { ArchitectureStatus } from "../../../components/widgets/architecture-status";
import { AIAssistantWidget } from "../../../components/widgets/ai-assistant-widget";
import { ArgoCDStatusWidget } from "../../../components/widgets/argocd-status-widget";
import { dockerService } from "../../../services/docker-service";

export default function MonitoringPage() {
  // Query live metrics from Docker daemon via Go backend
  const { data: stats, isLoading } = useQuery({
    queryKey: ["monitoring", "dashboard-stats"],
    queryFn: () => dockerService.getDashboardStats(),
    refetchInterval: 15000,
  });

  // Calculate RAM formatted strings
  const ramUsedGB = stats ? (stats.used_ram_bytes / (1024 * 1024 * 1024)).toFixed(1) : "0.0";
  const ramTotalGB = stats ? (stats.total_ram_bytes / (1024 * 1024 * 1024)).toFixed(1) : "0.0";
  const ramPercentStr = stats ? `${stats.overall_ram_percent.toFixed(1)}%` : "0%";
  const uptimePercent = stats && stats.total_containers > 0
    ? ((stats.containers_on / stats.total_containers) * 100).toFixed(1)
    : "100";

  return (
    <div className="space-y-6 max-w-7xl mx-auto">
      {/* Tier 1: 6 KPI Stat Cards - Live Telemetry from Docker Engine */}
      <section className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
        <StatCard
          value={stats ? stats.total_containers : 0}
          label="Total Kontainer"
          sublabel={stats ? `${stats.containers_on} Up &bull; ${stats.containers_off} Down` : "Loading..."}
          color="default"
          accentTop="none"
          isLoading={isLoading}
        />
        <StatCard
          value={stats ? stats.total_replicas : 0}
          label="Total Replika"
          sublabel="K3d &amp; Docker Swarm"
          color="cyan"
          accentTop="cyan"
          isLoading={isLoading}
        />
        <StatCard
          value={ramPercentStr}
          label="Overall RAM"
          sublabel={`${ramUsedGB} GB / ${ramTotalGB} GB`}
          color={stats && stats.overall_ram_percent > 85 ? "danger" : "warning"}
          accentTop="amber"
          isLoading={isLoading}
        />
        <StatCard
          value={stats ? stats.containers_on : 0}
          label="Kontainer ON"
          sublabel={`${uptimePercent}% Operational`}
          color="success"
          accentTop="emerald"
          isLoading={isLoading}
        />
        <StatCard
          value={stats ? stats.containers_off : 0}
          label="Kontainer OFF"
          sublabel={stats && stats.containers_off > 0 ? "Exited / Stopped" : "All Running"}
          color={stats && stats.containers_off > 0 ? "danger" : "default"}
          accentTop="pink"
          isLoading={isLoading}
        />
        <StatCard
          value={stats ? stats.active_incidents : 0}
          label="Active Incidents"
          sublabel={stats && stats.active_incidents > 0 ? "Requires Remediation" : "Zero Alerts"}
          color={stats && stats.active_incidents > 0 ? "critical" : "default"}
          accentTop="red"
          isLoading={isLoading}
        />
      </section>

      {/* Tier 2: Host Resource Usage & System Event Logs */}
      <section className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <HostResourceUsage />
        <SystemEventLogs isLoading={isLoading} />
      </section>

      {/* Tier 3: Agent Architecture Status & AI Autonomous Assistant */}
      <section className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <ArchitectureStatus />
        <AIAssistantWidget isLoading={isLoading} />
      </section>

      {/* Tier 4: GitOps & ArgoCD Continuous Delivery */}
      <section className="grid grid-cols-1 gap-6">
        <ArgoCDStatusWidget />
      </section>
    </div>
  );
}

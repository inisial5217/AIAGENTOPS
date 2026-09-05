"use client";

import * as React from "react";
import { StatCard } from "../../../components/widgets/stat-card";
import { HostResourceUsage } from "../../../components/widgets/host-resource-usage";
import { SystemEventLogs } from "../../../components/widgets/system-event-logs";
import { ArchitectureStatus } from "../../../components/widgets/architecture-status";
import { AIAssistantWidget } from "../../../components/widgets/ai-assistant-widget";

export default function MonitoringPage() {
  const [isLoading, setIsLoading] = React.useState(true);

  // Initial load simulation for skeleton demonstration
  React.useEffect(() => {
    const timer = setTimeout(() => {
      setIsLoading(false);
    }, 400);
    return () => clearTimeout(timer);
  }, []);

  return (
    <div className="space-y-6 max-w-7xl mx-auto">
      {/* Tier 1: 6 KPI Stat Cards */}
      <section className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
        <StatCard
          value={145}
          label="Total Kontainer"
          sublabel="12 Namespaces"
          color="default"
          accentTop="none"
          isLoading={isLoading}
        />
        <StatCard
          value={312}
          label="Total Replika"
          sublabel="4 K3d Clusters"
          color="cyan"
          accentTop="cyan"
          isLoading={isLoading}
        />
        <StatCard
          value="64%"
          label="Overall RAM"
          sublabel="41.2 GB / 64 GB"
          color="warning"
          accentTop="amber"
          isLoading={isLoading}
        />
        <StatCard
          value={142}
          label="Kontainer ON"
          sublabel="97.9% Operational"
          color="success"
          accentTop="emerald"
          isLoading={isLoading}
        />
        <StatCard
          value={3}
          label="Kontainer OFF"
          sublabel="CrashLoopBackOff"
          color="danger"
          accentTop="pink"
          isLoading={isLoading}
        />
        <StatCard
          value={2}
          label="Active Incidents"
          sublabel="Requires Remediation"
          color="critical"
          accentTop="red"
          isLoading={isLoading}
        />
      </section>

      {/* Tier 2: Host Resource Usage & System Event Logs */}
      <section className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <HostResourceUsage />
        <SystemEventLogs />
      </section>

      {/* Tier 3: Agent Architecture Status & AI Autonomous Assistant */}
      <section className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <ArchitectureStatus />
        <AIAssistantWidget isLoading={isLoading} />
      </section>
    </div>
  );
}

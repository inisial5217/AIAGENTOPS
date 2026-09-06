"use client";

import * as React from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { dockerService } from "../../../services/docker-service";
import { Badge } from "../../../components/ui/badge";
import { Button } from "../../../components/ui/button";
import { ContainerDetailModal } from "../../../components/widgets/container-detail-modal";
import { DockerNavTabs } from "../../../components/layout/docker-nav-tabs";
import { useAuth } from "../../../hooks/use-auth";
import { useNotificationStore } from "../../../store/notification-store";
import {
  Container,
  RotateCw,
  Square,
  Search,
  RefreshCw,
  ExternalLink,
  ShieldCheck,
} from "lucide-react";

export default function DockerContainersPage() {
  const queryClient = useQueryClient();
  const { isDevOps, isAdmin } = useAuth();
  const { addNotification } = useNotificationStore();
  const [statusFilter, setStatusFilter] = React.useState<string>("all");
  const [searchQuery, setSearchQuery] = React.useState<string>("");
  const [selectedContainerId, setSelectedContainerId] = React.useState<string | null>(null);

  // fetch live containers list
  const { data, isLoading, isRefetching, refetch } = useQuery({
    queryKey: ["docker", "containers", statusFilter],
    queryFn: () => dockerService.getContainers(statusFilter),
    refetchInterval: 10000,
  });

  // restart container mutation
  const restartMutation = useMutation({
    mutationFn: (id: string) => dockerService.restartContainer(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: ["docker"] });
      addNotification({
        title: "Container Restarted",
        message: `Container ${id.slice(0, 12)} successfully restarted.`,
        severity: "info",
      });
    },
    onError: (err: any, id) => {
      addNotification({
        title: "Restart Failed",
        message: `Failed to restart container ${id.slice(0, 12)}: ${err.message}`,
        severity: "error",
      });
    },
  });

  // stop container mutation
  const stopMutation = useMutation({
    mutationFn: (id: string) => dockerService.stopContainer(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: ["docker"] });
      addNotification({
        title: "Container Stopped",
        message: `Container ${id.slice(0, 12)} stopped.`,
        severity: "warning",
      });
    },
    onError: (err: any, id) => {
      addNotification({
        title: "Stop Failed",
        message: `Failed to stop container ${id.slice(0, 12)}: ${err.message}`,
        severity: "error",
      });
    },
  });

  const containers = data?.data || [];
  const filtered = containers.filter((c) => {
    if (!searchQuery) return true;
    const nameMatch = c.names.some((n) =>
      n.toLowerCase().includes(searchQuery.toLowerCase())
    );
    const imageMatch = c.image.toLowerCase().includes(searchQuery.toLowerCase());
    return nameMatch || imageMatch;
  });

  const runningCount = containers.filter((c) => c.state.toLowerCase() === "running").length;
  const stoppedCount = containers.length - runningCount;

  return (
    <div className="space-y-6 max-w-7xl mx-auto">
      {/* Top Header & Overview */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-[var(--border-subtle)]">
        <div className="text-left">
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold tracking-tight text-[var(--text-primary)]">
              Docker Engine Containers
            </h1>
            <Badge variant="cyan" size="sm">
              Live Socket
            </Badge>
          </div>
          <p className="text-xs text-[var(--text-muted)] font-mono mt-1">
            Realtime container inventory &bull; {containers.length} Total ({runningCount} Running, {stoppedCount} Stopped)
          </p>
        </div>

        {/* Action controls */}
        <div className="flex items-center gap-3">
          <div className="relative w-64">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-[var(--text-muted)]" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Filter by name or image..."
              className="w-full h-8 pl-8 pr-3 rounded-lg bg-[var(--bg-secondary)] border border-[var(--border-default)] text-xs text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:outline-none focus:border-cyan-500"
            />
          </div>

          <Button
            size="sm"
            variant="outline"
            onClick={() => refetch()}
            isLoading={isRefetching}
            title="Refresh list"
          >
            <RefreshCw className="w-3.5 h-3.5" />
          </Button>
        </div>
      </div>

      {/* Navigation Sub-Tabs */}
      <DockerNavTabs />

      {/* Filter Tabs */}
      <div className="flex items-center gap-2 text-xs font-mono">
        <button
          onClick={() => setStatusFilter("all")}
          className={`px-3 py-1 rounded-lg border transition-all cursor-pointer ${
            statusFilter === "all"
              ? "bg-cyan-500/10 border-cyan-500/40 text-cyan-400 font-bold"
              : "border-[var(--border-subtle)] bg-[var(--bg-secondary)] text-[var(--text-muted)] hover:text-[var(--text-primary)]"
          }`}
        >
          All ({containers.length})
        </button>
        <button
          onClick={() => setStatusFilter("running")}
          className={`px-3 py-1 rounded-lg border transition-all cursor-pointer ${
            statusFilter === "running"
              ? "bg-emerald-500/10 border-emerald-500/40 text-emerald-400 font-bold"
              : "border-[var(--border-subtle)] bg-[var(--bg-secondary)] text-[var(--text-muted)] hover:text-[var(--text-primary)]"
          }`}
        >
          Running ({runningCount})
        </button>
        <button
          onClick={() => setStatusFilter("stopped")}
          className={`px-3 py-1 rounded-lg border transition-all cursor-pointer ${
            statusFilter === "stopped"
              ? "bg-rose-500/10 border-rose-500/40 text-rose-400 font-bold"
              : "border-[var(--border-subtle)] bg-[var(--bg-secondary)] text-[var(--text-muted)] hover:text-[var(--text-primary)]"
          }`}
        >
          Stopped ({stoppedCount})
        </button>
      </div>

      {/* Containers Table */}
      <div className="bg-[var(--bg-card)] border border-[var(--border-default)] rounded-xl overflow-hidden shadow-sm">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs font-mono">
            <thead className="bg-[var(--bg-secondary)] border-b border-[var(--border-subtle)] text-[var(--text-muted)]">
              <tr>
                <th className="py-3 px-4">Container Name</th>
                <th className="py-3 px-4">Image</th>
                <th className="py-3 px-4">State</th>
                <th className="py-3 px-4">Status</th>
                <th className="py-3 px-4">Ports</th>
                <th className="py-3 px-4 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[var(--border-subtle)]">
              {isLoading ? (
                <tr>
                  <td colSpan={6} className="py-8 text-center text-[var(--text-muted)]">
                    Loading containers from Docker Engine...
                  </td>
                </tr>
              ) : filtered.length === 0 ? (
                <tr>
                  <td colSpan={6} className="py-8 text-center text-[var(--text-muted)]">
                    No containers found matching the criteria.
                  </td>
                </tr>
              ) : (
                filtered.map((c) => {
                  const name = c.names[0] ? c.names[0].replace(/^\//, "") : c.id.substring(0, 12);
                  const isRunning = c.state.toLowerCase() === "running";

                  return (
                    <tr
                      key={c.id}
                      className="hover:bg-[var(--bg-hover)] transition-colors group"
                    >
                      {/* Name */}
                      <td className="py-3 px-4 font-semibold text-[var(--text-primary)]">
                        <div className="flex items-center gap-2">
                          <Container className="w-3.5 h-3.5 text-cyan-400 shrink-0" />
                          <button
                            onClick={() => setSelectedContainerId(c.id)}
                            className="hover:text-cyan-400 hover:underline transition-colors text-left truncate max-w-xs cursor-pointer"
                          >
                            {name}
                          </button>
                        </div>
                      </td>

                      {/* Image */}
                      <td className="py-3 px-4 text-[var(--text-secondary)] truncate max-w-[200px]">
                        {c.image}
                      </td>

                      {/* State Badge */}
                      <td className="py-3 px-4">
                        <Badge
                          variant={isRunning ? "success" : "critical"}
                          pulse={isRunning}
                          size="sm"
                        >
                          {c.state}
                        </Badge>
                      </td>

                      {/* Status */}
                      <td className="py-3 px-4 text-[var(--text-muted)] text-[11px]">
                        {c.status}
                      </td>

                      {/* Ports */}
                      <td className="py-3 px-4 text-[var(--text-secondary)]">
                        {c.ports && c.ports.length > 0 ? (
                          <div className="flex flex-wrap gap-1">
                            {c.ports.slice(0, 2).map((p, idx) => (
                              <span
                                key={idx}
                                className="px-1.5 py-0.5 rounded bg-[var(--bg-secondary)] text-[10px] text-cyan-300 border border-[var(--border-subtle)]"
                              >
                                {p.public_port ? `${p.public_port}->` : ""}{p.private_port}/{p.type}
                              </span>
                            ))}
                          </div>
                        ) : (
                          <span className="text-[var(--text-muted)]">-</span>
                        )}
                      </td>

                      {/* Action buttons */}
                      <td className="py-3 px-4 text-right">
                        <div className="flex items-center justify-end gap-1.5">
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => setSelectedContainerId(c.id)}
                            title="Inspect Details"
                          >
                            <ExternalLink className="w-3 h-3" />
                          </Button>

                          {isDevOps && (
                            <Button
                              size="sm"
                              variant="outline"
                              onClick={() => restartMutation.mutate(c.id)}
                              isLoading={restartMutation.isPending && restartMutation.variables === c.id}
                              title="Restart Container"
                            >
                              <RotateCw className="w-3 h-3" />
                            </Button>
                          )}

                          {isAdmin && isRunning && (
                            <Button
                              size="sm"
                              variant="danger"
                              onClick={() => stopMutation.mutate(c.id)}
                              isLoading={stopMutation.isPending && stopMutation.variables === c.id}
                              title="Stop Container"
                            >
                              <Square className="w-3 h-3" />
                            </Button>
                          )}
                        </div>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Container Details Modal */}
      <ContainerDetailModal
        containerId={selectedContainerId}
        onClose={() => setSelectedContainerId(null)}
      />
    </div>
  );
}

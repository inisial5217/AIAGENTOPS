"use client";

import * as React from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { dockerService } from "../../services/docker-service";
import { Modal } from "../ui/modal";
import { Badge } from "../ui/badge";
import { Button } from "../ui/button";
import { useAuth } from "../../hooks/use-auth";
import { useNotificationStore } from "../../store/notification-store";
import { formatBytes } from "../../lib/format";
import { LogTerminal } from "../terminal/log-terminal";
import {
  RotateCw,
  Square,
  FileText,
  Activity,
  Info,
  Download,
  Terminal,
  Cpu,
  HardDrive,
  Network,
} from "lucide-react";

interface ContainerDetailModalProps {
  containerId: string | null;
  onClose: () => void;
}

export function ContainerDetailModal({
  containerId,
  onClose,
}: ContainerDetailModalProps) {
  const queryClient = useQueryClient();
  const { isAdmin, isDevOps } = useAuth();
  const { addNotification } = useNotificationStore();
  const [activeTab, setActiveTab] = React.useState<"info" | "stats" | "logs">("info");
  const [logTail, setLogTail] = React.useState(200);

  // fetch container detail
  const { data: detail, isLoading: isDetailLoading } = useQuery({
    queryKey: ["docker", "container", containerId],
    queryFn: () => dockerService.getContainer(containerId!),
    enabled: !!containerId,
  });

  // fetch container stats
  const { data: stats, isLoading: isStatsLoading } = useQuery({
    queryKey: ["docker", "stats", containerId],
    queryFn: () => dockerService.getContainerStats(containerId!),
    enabled: !!containerId && activeTab === "stats",
    refetchInterval: activeTab === "stats" ? 3000 : false,
  });

  // fetch container logs
  const { data: logsData, isLoading: isLogsLoading } = useQuery({
    queryKey: ["docker", "logs", containerId, logTail],
    queryFn: () => dockerService.getContainerLogs(containerId!, logTail),
    enabled: !!containerId && activeTab === "logs",
  });

  // restart mutation
  const restartMutation = useMutation({
    mutationFn: () => dockerService.restartContainer(containerId!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["docker"] });
      addNotification({
        title: "Container Restarted",
        message: `Container ${detail?.name || containerId?.slice(0, 12)} successfully restarted.`,
        severity: "info",
      });
    },
    onError: (err: any) => {
      addNotification({
        title: "Restart Failed",
        message: `Failed to restart ${detail?.name || "container"}: ${err.message}`,
        severity: "error",
      });
    },
  });

  // stop mutation
  const stopMutation = useMutation({
    mutationFn: () => dockerService.stopContainer(containerId!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["docker"] });
      addNotification({
        title: "Container Stopped",
        message: `Container ${detail?.name || containerId?.slice(0, 12)} stopped.`,
        severity: "warning",
      });
    },
    onError: (err: any) => {
      addNotification({
        title: "Stop Failed",
        message: `Failed to stop ${detail?.name || "container"}: ${err.message}`,
        severity: "error",
      });
    },
  });

  const handleDownloadLogs = () => {
    if (!logsData?.logs) return;
    const blob = new Blob([logsData.logs], { type: "text/plain" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `${detail?.name || "container"}-logs.txt`;
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <Modal
      isOpen={!!containerId}
      onClose={onClose}
      className="max-w-2xl max-h-[85vh] flex flex-col"
    >
      {/* Modal Header */}
      <div className="flex items-center justify-between pb-3 border-b border-[var(--border-subtle)]">
        <div className="flex items-center gap-2.5">
          <div className="p-2 rounded-lg bg-cyan-500/10 border border-cyan-500/30 text-cyan-400">
            <Terminal className="w-4 h-4" />
          </div>
          <div className="text-left">
            <h3 className="text-sm font-bold text-[var(--text-primary)]">
              {detail?.name || "Container Details"}
            </h3>
            <span className="text-[10px] font-mono text-[var(--text-muted)] truncate max-w-xs block">
              ID: {containerId}
            </span>
          </div>
        </div>

        {detail && (
          <Badge
            variant={detail.state.running ? "success" : "critical"}
            pulse={detail.state.running}
            size="sm"
          >
            {detail.state.status}
          </Badge>
        )}
      </div>

      {/* Tabs */}
      <div className="flex items-center justify-between border-b border-[var(--border-subtle)] my-3">
        <div className="flex gap-1 text-xs font-mono">
          <button
            onClick={() => setActiveTab("info")}
            className={`px-3 py-1.5 border-b-2 transition-colors cursor-pointer ${
              activeTab === "info"
                ? "border-cyan-400 text-cyan-400 font-bold"
                : "border-transparent text-[var(--text-muted)] hover:text-[var(--text-primary)]"
            }`}
          >
            <Info className="w-3.5 h-3.5 inline mr-1" />
            Info
          </button>
          <button
            onClick={() => setActiveTab("stats")}
            className={`px-3 py-1.5 border-b-2 transition-colors cursor-pointer ${
              activeTab === "stats"
                ? "border-cyan-400 text-cyan-400 font-bold"
                : "border-transparent text-[var(--text-muted)] hover:text-[var(--text-primary)]"
            }`}
          >
            <Activity className="w-3.5 h-3.5 inline mr-1" />
            Live Stats
          </button>
          <button
            onClick={() => setActiveTab("logs")}
            className={`px-3 py-1.5 border-b-2 transition-colors cursor-pointer ${
              activeTab === "logs"
                ? "border-cyan-400 text-cyan-400 font-bold"
                : "border-transparent text-[var(--text-muted)] hover:text-[var(--text-primary)]"
            }`}
          >
            <FileText className="w-3.5 h-3.5 inline mr-1" />
            Console Logs
          </button>
        </div>

        {/* Action Buttons */}
        <div className="flex items-center gap-1.5">
          {isDevOps && (
            <Button
              size="sm"
              variant="outline"
              onClick={() => restartMutation.mutate()}
              isLoading={restartMutation.isPending}
              title="Restart Container"
            >
              <RotateCw className="w-3 h-3" />
              Restart
            </Button>
          )}
          {isAdmin && (
            <Button
              size="sm"
              variant="danger"
              onClick={() => stopMutation.mutate()}
              isLoading={stopMutation.isPending}
              title="Stop Container"
            >
              <Square className="w-3 h-3" />
              Stop
            </Button>
          )}
        </div>
      </div>

      {/* Tab Content */}
      <div className="flex-1 overflow-y-auto text-left text-xs font-mono py-2">
        {activeTab === "info" && detail && (
          <div className="space-y-3">
            <div className="grid grid-cols-2 gap-2 p-3 rounded-lg bg-[var(--bg-secondary)] border border-[var(--border-subtle)]">
              <div>
                <span className="text-[10px] text-[var(--text-muted)] block">Image</span>
                <span className="text-[var(--text-primary)] font-bold truncate block">{detail.image}</span>
              </div>
              <div>
                <span className="text-[10px] text-[var(--text-muted)] block">IP Address</span>
                <span className="text-cyan-400">{detail.ip_address || "host network"}</span>
              </div>
              <div>
                <span className="text-[10px] text-[var(--text-muted)] block">Started At</span>
                <span className="text-[var(--text-secondary)]">{detail.state.started_at}</span>
              </div>
              <div>
                <span className="text-[10px] text-[var(--text-muted)] block">Platform</span>
                <span className="text-[var(--text-secondary)]">{detail.platform}</span>
              </div>
            </div>

            {/* Mounts */}
            <div>
              <span className="text-[11px] font-bold text-[var(--text-primary)] block mb-1">
                Volume Mounts ({detail.mounts?.length || 0})
              </span>
              <div className="space-y-1">
                {detail.mounts?.map((m, i) => (
                  <div key={i} className="p-2 rounded bg-[var(--bg-secondary)] text-[11px] flex justify-between border border-[var(--border-subtle)]">
                    <span className="text-[var(--text-secondary)]">{m.destination}</span>
                    <span className="text-[10px] text-cyan-400">{m.type} ({m.rw ? "rw" : "ro"})</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}

        {activeTab === "stats" && stats && (
          <div className="space-y-4">
            <div className="grid grid-cols-3 gap-3">
              <div className="p-3 rounded-lg bg-[var(--bg-secondary)] border border-[var(--border-subtle)] text-center">
                <Cpu className="w-5 h-5 text-cyan-400 mx-auto mb-1" />
                <span className="text-[10px] text-[var(--text-muted)] block">CPU Usage</span>
                <span className="text-base font-bold text-cyan-400">{stats.cpu_percentage.toFixed(2)}%</span>
              </div>
              <div className="p-3 rounded-lg bg-[var(--bg-secondary)] border border-[var(--border-subtle)] text-center">
                <HardDrive className="w-5 h-5 text-pink-400 mx-auto mb-1" />
                <span className="text-[10px] text-[var(--text-muted)] block">RAM Usage</span>
                <span className="text-base font-bold text-pink-400">{stats.memory_percentage.toFixed(2)}%</span>
                <span className="text-[9px] text-[var(--text-muted)] block">{formatBytes(stats.memory_usage_bytes)}</span>
              </div>
              <div className="p-3 rounded-lg bg-[var(--bg-secondary)] border border-[var(--border-subtle)] text-center">
                <Network className="w-5 h-5 text-emerald-400 mx-auto mb-1" />
                <span className="text-[10px] text-[var(--text-muted)] block">Network I/O</span>
                <span className="text-xs font-bold text-emerald-400">{formatBytes(stats.network_rx_bytes)}</span>
                <span className="text-[9px] text-[var(--text-muted)] block">Tx: {formatBytes(stats.network_tx_bytes)}</span>
              </div>
            </div>
          </div>
        )}

        {activeTab === "logs" && (
          <div className="flex flex-col pt-1">
            <LogTerminal
              topic={`docker_logs:${containerId}`}
              initialLogs={logsData?.logs ? logsData.logs.trim().split("\n") : []}
              title={`Logs: ${detail?.name || containerId?.slice(0, 12)}`}
              height="h-[340px]"
            />
          </div>
        )}
      </div>
    </Modal>
  );
}

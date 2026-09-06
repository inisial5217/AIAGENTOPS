"use client";

import * as React from "react";
import Link from "next/link";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  AlertTriangle,
  AlertOctagon,
  CheckCircle2,
  Clock,
  ShieldAlert,
  RefreshCw,
  Search,
  Filter,
  ExternalLink,
  ArrowRight,
  X,
  Bot,
  Sparkles,
  Bell,
  Check,
  CheckCircle,
  XCircle,
} from "lucide-react";
import { useAuth } from "../../../hooks/use-auth";
import { useWebSocket } from "../../../hooks/use-websocket";
import { incidentService } from "../../../services/incident-service";
import { Modal } from "../../../components/ui/modal";
import {
  IncidentSummary,
  IncidentDetail,
  IncidentSeverity,
  IncidentStatus,
} from "../../../types/incident";
import { WSMessage, NotificationPayload } from "../../../types/websocket";

export default function IncidentsPage() {
  const queryClient = useQueryClient();
  const { user } = useAuth();

  // state
  const [searchQuery, setSearchQuery] = React.useState("");
  const [statusFilter, setStatusFilter] = React.useState<string>("all");
  const [severityFilter, setSeverityFilter] = React.useState<string>("all");
  const [sourceFilter, setSourceFilter] = React.useState<string>("all");
  const [selectedIncidentId, setSelectedIncidentId] = React.useState<string | null>(null);
  const [actionError, setActionError] = React.useState<string | null>(null);

  // user role check
  const isDevOpsOrAdmin = user?.role === "devops" || user?.role === "admin";
  const isAdmin = user?.role === "admin";

  // query incidents list
  const {
    data: incidentListData,
    isLoading: isLoadingList,
    isRefetching,
    refetch: refetchList,
  } = useQuery({
    queryKey: ["incidents", statusFilter, severityFilter, sourceFilter, searchQuery],
    queryFn: () =>
      incidentService.listIncidents({
        status: statusFilter,
        severity: severityFilter,
        source: sourceFilter,
        search: searchQuery,
      }),
    refetchInterval: 15000,
  });

  // query incident stats
  const { data: statsData, refetch: refetchStats } = useQuery({
    queryKey: ["incident-stats"],
    queryFn: () => incidentService.getIncidentStats(),
    refetchInterval: 15000,
  });

  // query incident detail
  const {
    data: selectedIncident,
    isLoading: isLoadingDetail,
    refetch: refetchDetail,
  } = useQuery({
    queryKey: ["incident-detail", selectedIncidentId],
    queryFn: () => (selectedIncidentId ? incidentService.getIncident(selectedIncidentId) : null),
    enabled: Boolean(selectedIncidentId),
  });

  // mutations
  const ackMutation = useMutation({
    mutationFn: (id: string) => incidentService.acknowledgeIncident(id),
    onSuccess: () => {
      setActionError(null);
      queryClient.invalidateQueries({ queryKey: ["incidents"] });
      queryClient.invalidateQueries({ queryKey: ["incident-stats"] });
      if (selectedIncidentId) {
        queryClient.invalidateQueries({ queryKey: ["incident-detail", selectedIncidentId] });
      }
    },
    onError: (err: any) => {
      setActionError(err.response?.data?.message || err.message || "Failed to acknowledge incident");
    },
  });

  const resolveMutation = useMutation({
    mutationFn: (id: string) => incidentService.resolveIncident(id),
    onSuccess: () => {
      setActionError(null);
      queryClient.invalidateQueries({ queryKey: ["incidents"] });
      queryClient.invalidateQueries({ queryKey: ["incident-stats"] });
      if (selectedIncidentId) {
        queryClient.invalidateQueries({ queryKey: ["incident-detail", selectedIncidentId] });
      }
    },
    onError: (err: any) => {
      setActionError(err.response?.data?.message || err.message || "Failed to resolve incident");
    },
  });

  const closeMutation = useMutation({
    mutationFn: (id: string) => incidentService.closeIncident(id),
    onSuccess: () => {
      setActionError(null);
      queryClient.invalidateQueries({ queryKey: ["incidents"] });
      queryClient.invalidateQueries({ queryKey: ["incident-stats"] });
      if (selectedIncidentId) {
        queryClient.invalidateQueries({ queryKey: ["incident-detail", selectedIncidentId] });
      }
    },
    onError: (err: any) => {
      setActionError(err.response?.data?.message || err.message || "Failed to close incident");
    },
  });

  // websocket real-time hook
  const { status: wsStatus, lastMessage } = useWebSocket(["notifications"]);

  React.useEffect(() => {
    if (lastMessage && lastMessage.topic === "notifications") {
      queryClient.invalidateQueries({ queryKey: ["incidents"] });
      queryClient.invalidateQueries({ queryKey: ["incident-stats"] });
    }
  }, [lastMessage, queryClient]);

  // format helpers
  const formatTime = (ts?: string | null) => {
    if (!ts) return "-";
    try {
      const d = new Date(ts);
      return d.toLocaleString("en-US", {
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
        hour12: false,
      });
    } catch {
      return ts;
    }
  };

  const formatRelativeTime = (ts?: string | null) => {
    if (!ts) return "-";
    try {
      const diff = Math.floor((Date.now() - new Date(ts).getTime()) / 1000);
      if (diff < 60) return `${diff}s ago`;
      if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
      if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
      return `${Math.floor(diff / 86400)}d ago`;
    } catch {
      return ts;
    }
  };

  const getSeverityBadge = (severity: IncidentSeverity) => {
    switch (severity?.toLowerCase()) {
      case "critical":
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-red-500/10 text-red-400 border border-red-500/30">
            <span className="w-1.5 h-1.5 rounded-full bg-red-400 animate-pulse" />
            CRITICAL
          </span>
        );
      case "warning":
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-amber-500/10 text-amber-400 border border-amber-500/30">
            <span className="w-1.5 h-1.5 rounded-full bg-amber-400" />
            WARNING
          </span>
        );
      default:
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-cyan-500/10 text-cyan-400 border border-cyan-500/30">
            <span className="w-1.5 h-1.5 rounded-full bg-cyan-400" />
            INFO
          </span>
        );
    }
  };

  const getStatusBadge = (status: IncidentStatus) => {
    switch (status) {
      case "open":
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-500/15 text-red-300 border border-red-500/40 font-mono">
            <span className="w-1.5 h-1.5 rounded-full bg-red-400 animate-ping" />
            OPEN
          </span>
        );
      case "acknowledged":
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-amber-500/15 text-amber-300 border border-amber-500/40 font-mono">
            <Clock className="w-3 h-3 text-amber-400" />
            ACKNOWLEDGED
          </span>
        );
      case "investigating":
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-500/15 text-blue-300 border border-blue-500/40 font-mono">
            <Sparkles className="w-3 h-3 text-blue-400" />
            INVESTIGATING
          </span>
        );
      case "resolved":
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-500/15 text-emerald-300 border border-emerald-500/40 font-mono">
            <CheckCircle2 className="w-3 h-3 text-emerald-400" />
            RESOLVED
          </span>
        );
      case "closed":
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-zinc-800 text-zinc-400 border border-zinc-700 font-mono">
            CLOSED
          </span>
        );
      default:
        return <span>{status}</span>;
    }
  };

  const incidents = incidentListData?.data || [];
  const stats = statsData || {
    total: 0,
    open: 0,
    acknowledged: 0,
    investigating: 0,
    resolved: 0,
    closed: 0,
    critical_count: 0,
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold tracking-tight text-[var(--text-primary)]">
              Incident Management
            </h1>
            <span
              className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded text-[11px] font-mono uppercase tracking-wider font-semibold ${
                wsStatus === "connected"
                  ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/30"
                  : "bg-zinc-800 text-zinc-400 border border-zinc-700"
              }`}
            >
              <span
                className={`w-1.5 h-1.5 rounded-full ${
                  wsStatus === "connected" ? "bg-emerald-400 animate-pulse" : "bg-zinc-500"
                }`}
              />
              {wsStatus === "connected" ? "LIVE ALERTS" : "WS CONNECTING"}
            </span>
          </div>
          <p className="text-sm text-[var(--text-secondary)] mt-1">
            Real-time infrastructure incident triage, automated Alertmanager webhook receiver & audit lifecycle
          </p>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => {
              refetchList();
              refetchStats();
            }}
            disabled={isRefetching}
            className="flex items-center gap-2 px-3 py-2 text-xs font-medium rounded-[var(--radius-md)] bg-[var(--bg-secondary)] hover:bg-[var(--bg-hover)] border border-[var(--border-default)] text-[var(--text-primary)] transition-colors cursor-pointer"
            title="Refresh incidents"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${isRefetching ? "animate-spin" : ""}`} />
            Refresh
          </button>
        </div>
      </div>

      {/* KPI Stats Cards */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3.5">
        <div className="p-4 rounded-[var(--radius-lg)] bg-[var(--bg-secondary)] border border-[var(--border-default)] flex flex-col justify-between">
          <div className="flex items-center justify-between text-xs text-[var(--text-secondary)]">
            <span>Total Incidents</span>
            <ShieldAlert className="w-4 h-4 text-[var(--accent-default)]" />
          </div>
          <div className="text-2xl font-bold text-[var(--text-primary)] mt-2 font-mono">
            {stats.total}
          </div>
        </div>

        <div className="p-4 rounded-[var(--radius-lg)] bg-red-950/20 border border-red-500/30 flex flex-col justify-between relative overflow-hidden">
          <div className="flex items-center justify-between text-xs text-red-400">
            <span>Open (Action Needed)</span>
            <span className="w-2 h-2 rounded-full bg-red-500 animate-ping" />
          </div>
          <div className="text-2xl font-bold text-red-400 mt-2 font-mono">
            {stats.open}
          </div>
        </div>

        <div className="p-4 rounded-[var(--radius-lg)] bg-amber-950/20 border border-amber-500/30 flex flex-col justify-between">
          <div className="flex items-center justify-between text-xs text-amber-400">
            <span>Acknowledged</span>
            <Clock className="w-4 h-4 text-amber-400" />
          </div>
          <div className="text-2xl font-bold text-amber-400 mt-2 font-mono">
            {stats.acknowledged}
          </div>
        </div>

        <div className="p-4 rounded-[var(--radius-lg)] bg-emerald-950/20 border border-emerald-500/30 flex flex-col justify-between">
          <div className="flex items-center justify-between text-xs text-emerald-400">
            <span>Resolved</span>
            <CheckCircle2 className="w-4 h-4 text-emerald-400" />
          </div>
          <div className="text-2xl font-bold text-emerald-400 mt-2 font-mono">
            {stats.resolved}
          </div>
        </div>

        <div className="p-4 rounded-[var(--radius-lg)] bg-[var(--bg-secondary)] border border-[var(--border-default)] flex flex-col justify-between">
          <div className="flex items-center justify-between text-xs text-red-400">
            <span>Critical Severity</span>
            <AlertOctagon className="w-4 h-4 text-red-400" />
          </div>
          <div className="text-2xl font-bold text-red-400 mt-2 font-mono">
            {stats.critical_count}
          </div>
        </div>
      </div>

      {/* Filter and Search Bar */}
      <div className="p-4 rounded-[var(--radius-lg)] bg-[var(--bg-secondary)] border border-[var(--border-default)] flex flex-col md:flex-row gap-4 justify-between items-stretch md:items-center">
        {/* Status Tabs */}
        <div className="flex items-center gap-1.5 overflow-x-auto pb-1 md:pb-0">
          {[
            { id: "all", label: "All Status" },
            { id: "open", label: "Open" },
            { id: "acknowledged", label: "Acknowledged" },
            { id: "resolved", label: "Resolved" },
            { id: "closed", label: "Closed" },
          ].map((tab) => {
            const active = statusFilter === tab.id;
            return (
              <button
                key={tab.id}
                onClick={() => setStatusFilter(tab.id)}
                className={`px-3 py-1.5 rounded-[var(--radius-md)] text-xs font-medium transition-colors whitespace-nowrap cursor-pointer ${
                  active
                    ? "bg-[var(--accent-default)] text-black font-semibold shadow"
                    : "text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-hover)]"
                }`}
              >
                {tab.label}
              </button>
            );
          })}
        </div>

        {/* Severity, Source & Search */}
        <div className="flex items-center gap-2.5 flex-wrap md:flex-nowrap">
          {/* Severity Dropdown */}
          <div className="flex items-center gap-1.5">
            <select
              value={severityFilter}
              onChange={(e) => setSeverityFilter(e.target.value)}
              className="px-2.5 py-1.5 rounded-[var(--radius-md)] bg-[var(--bg-primary)] border border-[var(--border-default)] text-xs text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent-default)] cursor-pointer"
            >
              <option value="all">All Severities</option>
              <option value="critical">Critical</option>
              <option value="warning">Warning</option>
              <option value="info">Info</option>
            </select>
          </div>

          {/* Search Box */}
          <div className="relative flex-1 sm:w-60">
            <Search className="w-3.5 h-3.5 absolute left-2.5 top-2.5 text-[var(--text-muted)]" />
            <input
              type="text"
              placeholder="Search title, resource..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-8 pr-3 py-1.5 rounded-[var(--radius-md)] bg-[var(--bg-primary)] border border-[var(--border-default)] text-xs text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:outline-none focus:border-[var(--accent-default)]"
            />
          </div>
        </div>
      </div>

      {/* Incidents Table */}
      <div className="rounded-[var(--radius-lg)] bg-[var(--bg-secondary)] border border-[var(--border-default)] overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs border-collapse">
            <thead>
              <tr className="border-b border-[var(--border-subtle)] bg-[var(--bg-primary)]/50 text-[var(--text-secondary)] font-mono uppercase text-[11px] tracking-wider">
                <th className="py-3 px-4">Severity</th>
                <th className="py-3 px-4">Incident Title & Details</th>
                <th className="py-3 px-4">Source & Resource</th>
                <th className="py-3 px-4">Status</th>
                <th className="py-3 px-4">Created</th>
                <th className="py-3 px-4">Actor</th>
                <th className="py-3 px-4 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[var(--border-subtle)]">
              {isLoadingList ? (
                <tr>
                  <td colSpan={7} className="py-12 text-center text-[var(--text-muted)]">
                    <RefreshCw className="w-5 h-5 animate-spin mx-auto mb-2 text-[var(--accent-default)]" />
                    Loading incidents...
                  </td>
                </tr>
              ) : incidents.length === 0 ? (
                <tr>
                  <td colSpan={7} className="py-12 text-center text-[var(--text-muted)]">
                    <ShieldAlert className="w-8 h-8 mx-auto mb-2 text-zinc-600" />
                    No incidents match the active filters.
                  </td>
                </tr>
              ) : (
                incidents.map((inc) => {
                  return (
                    <tr
                      key={inc.id}
                      className="hover:bg-[var(--bg-hover)] transition-colors group cursor-pointer"
                      onClick={() => setSelectedIncidentId(inc.id)}
                    >
                      <td className="py-3 px-4">{getSeverityBadge(inc.severity)}</td>
                      <td className="py-3 px-4 max-w-xs">
                        <div className="font-semibold text-[var(--text-primary)] truncate group-hover:text-[var(--accent-default)] transition-colors">
                          {inc.title}
                        </div>
                        <div className="text-[11px] font-mono text-[var(--text-muted)] truncate mt-0.5">
                          ID: {inc.id.slice(0, 8)}...
                          {inc.alert_name && ` • Alert: ${inc.alert_name}`}
                        </div>
                      </td>
                      <td className="py-3 px-4">
                        <div className="font-medium text-[var(--text-primary)]">
                          {inc.resource_id || "System"}
                        </div>
                        <div className="text-[11px] text-[var(--text-muted)] capitalize">
                          {inc.source} {inc.resource_type ? `(${inc.resource_type})` : ""}
                        </div>
                      </td>
                      <td className="py-3 px-4">{getStatusBadge(inc.status)}</td>
                      <td className="py-3 px-4 font-mono text-[11px]" title={inc.created_at}>
                        <span className="text-[var(--text-primary)]">
                          {formatRelativeTime(inc.created_at)}
                        </span>
                        <div className="text-[10px] text-[var(--text-muted)]">
                          {formatTime(inc.created_at)}
                        </div>
                      </td>
                      <td className="py-3 px-4 text-[11px]">
                        {inc.acknowledged_by_name ? (
                          <div>
                            <span className="text-amber-400 font-medium">Ack:</span>{" "}
                            <span className="text-[var(--text-secondary)]">
                              {inc.acknowledged_by_name}
                            </span>
                          </div>
                        ) : inc.resolved_by_name ? (
                          <div>
                            <span className="text-emerald-400 font-medium">Resolved:</span>{" "}
                            <span className="text-[var(--text-secondary)]">
                              {inc.resolved_by_name}
                            </span>
                          </div>
                        ) : (
                          <span className="text-[var(--text-muted)] italic">Unassigned</span>
                        )}
                      </td>
                      <td className="py-3 px-4 text-right" onClick={(e) => e.stopPropagation()}>
                        <div className="flex items-center justify-end gap-1.5">
                          {inc.status === "open" && isDevOpsOrAdmin && (
                            <button
                              onClick={() => ackMutation.mutate(inc.id)}
                              disabled={ackMutation.isPending}
                              className="px-2.5 py-1 text-[11px] font-medium rounded bg-amber-500/10 text-amber-400 hover:bg-amber-500/20 border border-amber-500/30 transition-colors cursor-pointer"
                              title="Acknowledge incident"
                            >
                              Ack
                            </button>
                          )}
                          {(inc.status === "open" || inc.status === "acknowledged") && isDevOpsOrAdmin && (
                            <button
                              onClick={() => resolveMutation.mutate(inc.id)}
                              disabled={resolveMutation.isPending}
                              className="px-2.5 py-1 text-[11px] font-medium rounded bg-emerald-500/10 text-emerald-400 hover:bg-emerald-500/20 border border-emerald-500/30 transition-colors cursor-pointer"
                              title="Resolve incident"
                            >
                              Resolve
                            </button>
                          )}
                          {inc.status !== "closed" && isAdmin && (
                            <button
                              onClick={() => closeMutation.mutate(inc.id)}
                              disabled={closeMutation.isPending}
                              className="px-2.5 py-1 text-[11px] font-medium rounded bg-zinc-800 text-zinc-300 hover:bg-zinc-700 border border-zinc-700 transition-colors cursor-pointer"
                              title="Close incident permanently"
                            >
                              Close
                            </button>
                          )}
                          <button
                            onClick={() => setSelectedIncidentId(inc.id)}
                            className="px-2 py-1 text-[11px] font-medium rounded bg-[var(--bg-primary)] text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-hover)] border border-[var(--border-default)] transition-colors cursor-pointer"
                            title="Inspect details"
                          >
                            Inspect
                          </button>
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

      {/* Incident Detail Modal */}
      {selectedIncidentId && (
        <Modal
          isOpen={Boolean(selectedIncidentId)}
          onClose={() => {
            setSelectedIncidentId(null);
            setActionError(null);
          }}
          title="Incident Lifecycle & Details"
          description={selectedIncident ? `UUID: ${selectedIncident.id}` : "Loading incident..."}
          className="max-w-3xl"
        >
          {isLoadingDetail || !selectedIncident ? (
            <div className="py-16 text-center text-[var(--text-muted)]">
              <RefreshCw className="w-6 h-6 animate-spin mx-auto mb-2 text-[var(--accent-default)]" />
              Loading incident details...
            </div>
          ) : (
            <div className="space-y-6">
              {/* Error banner if action fails */}
              {actionError && (
                <div className="p-3 rounded-[var(--radius-md)] bg-red-500/10 border border-red-500/30 text-red-400 text-xs flex items-center justify-between">
                  <span>{actionError}</span>
                  <button
                    onClick={() => setActionError(null)}
                    className="text-red-400 hover:text-red-300 ml-2"
                  >
                    <X className="w-3.5 h-3.5" />
                  </button>
                </div>
              )}

              {/* Title & Badges */}
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 p-4 rounded-[var(--radius-md)] bg-[var(--bg-primary)] border border-[var(--border-subtle)]">
                <div>
                  <div className="flex items-center gap-2 mb-1.5">
                    {getSeverityBadge(selectedIncident.severity)}
                    {getStatusBadge(selectedIncident.status)}
                    <span className="text-xs text-[var(--text-muted)] uppercase font-mono tracking-wider">
                      Source: {selectedIncident.source}
                    </span>
                  </div>
                  <h3 className="text-base font-bold text-[var(--text-primary)]">
                    {selectedIncident.title}
                  </h3>
                </div>

                {/* Resource link button */}
                {selectedIncident.resource_id && (
                  <div className="shrink-0">
                    {selectedIncident.resource_type === "container" || selectedIncident.source === "docker" ? (
                      <Link
                        href={`/docker/containers`}
                        className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-[var(--radius-md)] bg-cyan-500/10 text-cyan-400 hover:bg-cyan-500/20 border border-cyan-500/30 text-xs font-medium transition-colors"
                      >
                        Inspect Container <ExternalLink className="w-3.5 h-3.5" />
                      </Link>
                    ) : selectedIncident.resource_type === "pod" || selectedIncident.source === "kubernetes" ? (
                      <Link
                        href={`/kubernetes`}
                        className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-[var(--radius-md)] bg-cyan-500/10 text-cyan-400 hover:bg-cyan-500/20 border border-cyan-500/30 text-xs font-medium transition-colors"
                      >
                        Inspect Pod <ExternalLink className="w-3.5 h-3.5" />
                      </Link>
                    ) : (
                      <span className="text-xs text-[var(--text-muted)] font-mono">
                        Resource: {selectedIncident.resource_id}
                      </span>
                    )}
                  </div>
                )}
              </div>

              {/* Lifecycle Visual Timeline */}
              <div className="p-4 rounded-[var(--radius-md)] bg-[var(--bg-primary)] border border-[var(--border-subtle)]">
                <h4 className="text-xs font-mono uppercase text-[var(--text-secondary)] tracking-wider mb-4">
                  Incident Lifecycle Progress
                </h4>

                <div className="relative">
                  <div className="grid grid-cols-4 gap-2 text-center relative z-10">
                    {/* 1. Created */}
                    <div className="flex flex-col items-center">
                      <div className="w-7 h-7 rounded-full bg-emerald-500/20 text-emerald-400 border border-emerald-500/50 flex items-center justify-center text-xs font-bold mb-1.5">
                        <Check className="w-4 h-4" />
                      </div>
                      <span className="text-xs font-semibold text-[var(--text-primary)]">Created</span>
                      <span className="text-[10px] text-[var(--text-muted)] mt-0.5">
                        {formatTime(selectedIncident.created_at)}
                      </span>
                      <span className="text-[10px] text-[var(--text-secondary)] capitalize mt-0.5">
                        By {selectedIncident.source}
                      </span>
                    </div>

                    {/* 2. Acknowledged */}
                    <div className="flex flex-col items-center">
                      <div
                        className={`w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold mb-1.5 ${
                          selectedIncident.acknowledged_at
                            ? "bg-emerald-500/20 text-emerald-400 border border-emerald-500/50"
                            : selectedIncident.status === "open"
                            ? "bg-amber-500/20 text-amber-400 border border-amber-500/50 animate-pulse"
                            : "bg-zinc-800 text-zinc-500 border border-zinc-700"
                        }`}
                      >
                        {selectedIncident.acknowledged_at ? (
                          <Check className="w-4 h-4" />
                        ) : (
                          <Clock className="w-3.5 h-3.5" />
                        )}
                      </div>
                      <span className="text-xs font-semibold text-[var(--text-primary)]">Acknowledged</span>
                      <span className="text-[10px] text-[var(--text-muted)] mt-0.5">
                        {formatTime(selectedIncident.acknowledged_at)}
                      </span>
                      <span className="text-[10px] text-[var(--text-secondary)] mt-0.5">
                        {selectedIncident.acknowledged_by_name || "-"}
                      </span>
                    </div>

                    {/* 3. Resolved */}
                    <div className="flex flex-col items-center">
                      <div
                        className={`w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold mb-1.5 ${
                          selectedIncident.resolved_at
                            ? "bg-emerald-500/20 text-emerald-400 border border-emerald-500/50"
                            : selectedIncident.status === "acknowledged"
                            ? "bg-blue-500/20 text-blue-400 border border-blue-500/50 animate-pulse"
                            : "bg-zinc-800 text-zinc-500 border border-zinc-700"
                        }`}
                      >
                        {selectedIncident.resolved_at ? (
                          <Check className="w-4 h-4" />
                        ) : (
                          <CheckCircle className="w-3.5 h-3.5" />
                        )}
                      </div>
                      <span className="text-xs font-semibold text-[var(--text-primary)]">Resolved</span>
                      <span className="text-[10px] text-[var(--text-muted)] mt-0.5">
                        {formatTime(selectedIncident.resolved_at)}
                      </span>
                      <span className="text-[10px] text-[var(--text-secondary)] mt-0.5">
                        {selectedIncident.resolved_by_name || "-"}
                      </span>
                    </div>

                    {/* 4. Closed */}
                    <div className="flex flex-col items-center">
                      <div
                        className={`w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold mb-1.5 ${
                          selectedIncident.closed_at
                            ? "bg-emerald-500/20 text-emerald-400 border border-emerald-500/50"
                            : selectedIncident.status === "resolved"
                            ? "bg-purple-500/20 text-purple-400 border border-purple-500/50 animate-pulse"
                            : "bg-zinc-800 text-zinc-500 border border-zinc-700"
                        }`}
                      >
                        {selectedIncident.closed_at ? (
                          <Check className="w-4 h-4" />
                        ) : (
                          <XCircle className="w-3.5 h-3.5" />
                        )}
                      </div>
                      <span className="text-xs font-semibold text-[var(--text-primary)]">Closed</span>
                      <span className="text-[10px] text-[var(--text-muted)] mt-0.5">
                        {formatTime(selectedIncident.closed_at)}
                      </span>
                      <span className="text-[10px] text-[var(--text-secondary)] mt-0.5">
                        {selectedIncident.closed_by_name || "-"}
                      </span>
                    </div>
                  </div>
                </div>
              </div>

              {/* Incident Details & Description */}
              <div className="space-y-3">
                <h4 className="text-xs font-mono uppercase text-[var(--text-secondary)] tracking-wider">
                  Description & Anomaly Details
                </h4>
                <div className="p-3.5 rounded-[var(--radius-md)] bg-[var(--bg-primary)] border border-[var(--border-subtle)] font-mono text-xs text-[var(--text-primary)] leading-relaxed whitespace-pre-wrap">
                  {selectedIncident.description || "No description provided."}
                </div>
              </div>

              {/* Metadata Key-Value */}
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs">
                <div className="p-3 rounded bg-[var(--bg-primary)] border border-[var(--border-subtle)]">
                  <span className="text-[var(--text-muted)] block text-[10px] uppercase font-mono">Alert Name</span>
                  <span className="font-mono text-[var(--text-primary)] font-semibold mt-0.5 block truncate">
                    {selectedIncident.alert_name || "N/A"}
                  </span>
                </div>
                <div className="p-3 rounded bg-[var(--bg-primary)] border border-[var(--border-subtle)]">
                  <span className="text-[var(--text-muted)] block text-[10px] uppercase font-mono">Resource Type</span>
                  <span className="font-mono text-[var(--text-primary)] font-semibold mt-0.5 block truncate">
                    {selectedIncident.resource_type || "N/A"}
                  </span>
                </div>
                <div className="p-3 rounded bg-[var(--bg-primary)] border border-[var(--border-subtle)]">
                  <span className="text-[var(--text-muted)] block text-[10px] uppercase font-mono">Namespace</span>
                  <span className="font-mono text-[var(--text-primary)] font-semibold mt-0.5 block truncate">
                    {selectedIncident.namespace || "default"}
                  </span>
                </div>
                <div className="p-3 rounded bg-[var(--bg-primary)] border border-[var(--border-subtle)]">
                  <span className="text-[var(--text-muted)] block text-[10px] uppercase font-mono">Last Updated</span>
                  <span className="font-mono text-[var(--text-primary)] font-semibold mt-0.5 block truncate">
                    {formatTime(selectedIncident.updated_at)}
                  </span>
                </div>
              </div>

              {/* AI RCA Assistant Card (Phase 9 Integration Ready) */}
              <div className="p-4 rounded-[var(--radius-lg)] bg-gradient-to-br from-pink-500/10 via-purple-500/5 to-cyan-500/10 border border-purple-500/30">
                <div className="flex items-center gap-2 mb-2 text-purple-300">
                  <Bot className="w-4 h-4 text-pink-400" />
                  <span className="text-xs font-bold uppercase tracking-wider font-mono">
                    AI Root Cause Analysis (AIOps Agent)
                  </span>
                  <span className="ml-auto px-2 py-0.5 rounded text-[10px] font-mono bg-purple-500/20 text-purple-300 border border-purple-500/30">
                    Phase 9 Ready
                  </span>
                </div>
                <p className="text-xs text-[var(--text-secondary)] leading-relaxed">
                  {selectedIncident.rca_summary ||
                    "Automated AI log diagnosis, remediation recommendation, and blast radius correlation will be active upon Phase 9 enablement. The incident payload is already structured with proper resource pointers."}
                </p>
              </div>

              {/* Notification Audit History */}
              {selectedIncident.notifications && selectedIncident.notifications.length > 0 && (
                <div className="space-y-2">
                  <h4 className="text-xs font-mono uppercase text-[var(--text-secondary)] tracking-wider flex items-center gap-1.5">
                    <Bell className="w-3.5 h-3.5 text-[var(--accent-default)]" /> Notification History
                  </h4>
                  <div className="divide-y divide-[var(--border-subtle)] rounded bg-[var(--bg-primary)] border border-[var(--border-subtle)] overflow-hidden">
                    {selectedIncident.notifications.map((n) => (
                      <div key={n.id} className="p-2.5 text-xs flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <span className="uppercase text-[10px] font-mono px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-300">
                            {n.channel}
                          </span>
                          <span className="text-[var(--text-primary)] font-medium">{n.title}</span>
                        </div>
                        <div className="flex items-center gap-2 text-[10px] text-[var(--text-muted)] font-mono">
                          <span className="capitalize">{n.status}</span>
                          <span>•</span>
                          <span>{formatTime(n.created_at)}</span>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Action Buttons Footer */}
              <div className="pt-4 border-t border-[var(--border-subtle)] flex items-center justify-between">
                <div className="text-xs text-[var(--text-muted)] font-mono">
                  Role: <span className="text-[var(--text-primary)] capitalize">{user?.role || "Viewer"}</span>
                </div>

                <div className="flex items-center gap-2">
                  {selectedIncident.status === "open" && isDevOpsOrAdmin && (
                    <button
                      onClick={() => ackMutation.mutate(selectedIncident.id)}
                      disabled={ackMutation.isPending}
                      className="px-4 py-2 text-xs font-medium rounded-[var(--radius-md)] bg-amber-500/20 text-amber-300 hover:bg-amber-500/30 border border-amber-500/40 transition-colors cursor-pointer"
                    >
                      {ackMutation.isPending ? "Acknowledging..." : "Acknowledge Incident"}
                    </button>
                  )}

                  {(selectedIncident.status === "open" || selectedIncident.status === "acknowledged") &&
                    isDevOpsOrAdmin && (
                      <button
                        onClick={() => resolveMutation.mutate(selectedIncident.id)}
                        disabled={resolveMutation.isPending}
                        className="px-4 py-2 text-xs font-medium rounded-[var(--radius-md)] bg-emerald-500/20 text-emerald-300 hover:bg-emerald-500/30 border border-emerald-500/40 transition-colors cursor-pointer"
                      >
                        {resolveMutation.isPending ? "Resolving..." : "Mark as Resolved"}
                      </button>
                    )}

                  {selectedIncident.status !== "closed" && isAdmin && (
                    <button
                      onClick={() => closeMutation.mutate(selectedIncident.id)}
                      disabled={closeMutation.isPending}
                      className="px-4 py-2 text-xs font-medium rounded-[var(--radius-md)] bg-zinc-800 text-zinc-300 hover:bg-zinc-700 border border-zinc-700 transition-colors cursor-pointer"
                    >
                      {closeMutation.isPending ? "Closing..." : "Close Incident"}
                    </button>
                  )}

                  <button
                    onClick={() => {
                      setSelectedIncidentId(null);
                      setActionError(null);
                    }}
                    className="px-4 py-2 text-xs font-medium rounded-[var(--radius-md)] bg-[var(--bg-hover)] text-[var(--text-secondary)] hover:text-[var(--text-primary)] transition-colors cursor-pointer"
                  >
                    Dismiss
                  </button>
                </div>
              </div>
            </div>
          )}
        </Modal>
      )}
    </div>
  );
}

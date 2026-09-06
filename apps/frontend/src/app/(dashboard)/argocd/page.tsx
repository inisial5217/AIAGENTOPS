"use client";

import * as React from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  GitPullRequest,
  RefreshCw,
  Search,
  CheckCircle2,
  AlertCircle,
  Clock,
  GitBranch,
  FolderGit2,
  Server,
  Layers,
  ExternalLink,
  Sliders,
  Play,
  History,
  ShieldCheck,
} from "lucide-react";
import { argocdService } from "../../../services/argocd-service";
import {
  ArgoApplicationSummary,
  ArgoApplicationDetail,
  ArgoSyncRequest,
} from "../../../types/argocd";
import { Badge } from "../../../components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "../../../components/ui/modal";
import { useNotificationStore } from "../../../store/notification-store";

export default function ArgoCDDashboardPage() {
  const queryClient = useQueryClient();
  const addNotification = useNotificationStore((s) => s.addNotification);

  // Filter States
  const [searchQuery, setSearchQuery] = React.useState<string>("");
  const [syncFilter, setSyncFilter] = React.useState<string>("all");
  const [healthFilter, setHealthFilter] = React.useState<string>("all");

  // Modal States
  const [syncModalApp, setSyncModalApp] = React.useState<ArgoApplicationSummary | null>(null);
  const [syncPrune, setSyncPrune] = React.useState<boolean>(true);
  const [syncDryRun, setSyncDryRun] = React.useState<boolean>(false);

  const [inspectApp, setInspectApp] = React.useState<ArgoApplicationSummary | null>(null);

  // Queries
  const {
    data: overview,
    isLoading: isOverviewLoading,
    refetch: refetchOverview,
  } = useQuery({
    queryKey: ["argocd", "overview"],
    queryFn: () => argocdService.getOverview(),
    refetchInterval: 15000,
  });

  const {
    data: appsData,
    isLoading: isAppsLoading,
    refetch: refetchApps,
  } = useQuery({
    queryKey: ["argocd", "applications"],
    queryFn: () => argocdService.getApplications(),
    refetchInterval: 10000,
  });

  // Query Inspect Detail
  const { data: appDetail, isLoading: isDetailLoading } = useQuery({
    queryKey: ["argocd", "detail", inspectApp?.name],
    queryFn: () => (inspectApp ? argocdService.getApplication(inspectApp.name) : null),
    enabled: !!inspectApp,
  });

  // Sync Mutation
  const syncMutation = useMutation({
    mutationFn: ({
      name,
      req,
    }: {
      name: string;
      req: ArgoSyncRequest;
    }) => argocdService.syncApplication(name, req),
    onSuccess: (_, vars) => {
      addNotification({
        title: "ArgoCD Sync Triggered",
        message: `Synchronization initiated for application ${vars.name}.`,
        severity: "info",
      });
      setSyncModalApp(null);
      queryClient.invalidateQueries({ queryKey: ["argocd", "applications"] });
      queryClient.invalidateQueries({ queryKey: ["argocd", "overview"] });
    },
    onError: (err: any) => {
      addNotification({
        title: "Sync Failed",
        message: err?.response?.data?.detail || err.message || "Failed to trigger sync.",
        severity: "error",
      });
    },
  });

  const handleRefreshAll = () => {
    refetchOverview();
    refetchApps();
  };

  // Filtered applications
  const applications = React.useMemo(() => {
    if (!appsData?.data) return [];
    return appsData.data.filter((app) => {
      const matchSearch =
        searchQuery === "" ||
        app.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        app.path.toLowerCase().includes(searchQuery.toLowerCase()) ||
        app.repo_url.toLowerCase().includes(searchQuery.toLowerCase());

      const matchSync =
        syncFilter === "all" || app.sync_status.toLowerCase() === syncFilter.toLowerCase();

      const matchHealth =
        healthFilter === "all" || app.health_status.toLowerCase() === healthFilter.toLowerCase();

      return matchSearch && matchSync && matchHealth;
    });
  }, [appsData, searchQuery, syncFilter, healthFilter]);

  return (
    <div className="space-y-6 max-w-7xl mx-auto pb-12">
      {/* Top Header */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 border-b border-[var(--border-subtle)] pb-4">
        <div className="flex items-center gap-3 text-left">
          <div className="p-2.5 rounded-xl bg-gradient-to-br from-cyan-500/20 via-pink-500/10 to-transparent border border-cyan-500/30 text-cyan-400">
            <GitPullRequest className="w-6 h-6" />
          </div>
          <div>
            <h1 className="text-xl font-bold tracking-tight text-[var(--text-primary)]">
              ArgoCD GitOps Management
            </h1>
            <p className="text-xs font-mono text-[var(--text-muted)]">
              Continuous Delivery &bull; Desired Git State Reconciliation &bull; Zero Mock Data
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={handleRefreshAll}
            className="flex items-center gap-2 px-3 py-2 text-xs font-mono rounded-lg bg-[var(--bg-secondary)] border border-[var(--border-default)] hover:border-cyan-500/50 text-[var(--text-primary)] transition-all cursor-pointer shadow-xs"
          >
            <RefreshCw className="w-3.5 h-3.5" />
            <span>Refresh GitOps</span>
          </button>
        </div>
      </div>

      {/* KPI Cards Grid */}
      <section className="grid grid-cols-2 md:grid-cols-5 gap-4 text-left">
        <div className="p-4 rounded-xl bg-[var(--bg-card)] border border-[var(--border-default)] shadow-xs">
          <span className="text-[11px] font-mono text-[var(--text-muted)]">Total Applications</span>
          <div className="text-2xl font-bold font-mono text-[var(--text-primary)] mt-0.5">
            {isOverviewLoading ? "--" : overview?.total ?? 0}
          </div>
          <span className="text-[10px] text-cyan-400 font-mono">Managed CRDs</span>
        </div>

        <div className="p-4 rounded-xl bg-[var(--bg-card)] border border-[var(--border-default)] shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-[11px] font-mono text-[var(--text-muted)]">Synced</span>
            <CheckCircle2 className="w-4 h-4 text-emerald-400" />
          </div>
          <div className="text-2xl font-bold font-mono text-emerald-400 mt-0.5">
            {isOverviewLoading ? "--" : overview?.synced ?? 0}
          </div>
          <span className="text-[10px] text-[var(--text-muted)] font-mono">Match Git Head</span>
        </div>

        <div className="p-4 rounded-xl bg-[var(--bg-card)] border border-[var(--border-default)] shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-[11px] font-mono text-[var(--text-muted)]">Out of Sync</span>
            <AlertCircle className="w-4 h-4 text-amber-400" />
          </div>
          <div className="text-2xl font-bold font-mono text-amber-400 mt-0.5">
            {isOverviewLoading ? "--" : overview?.out_of_sync ?? 0}
          </div>
          <span className="text-[10px] text-amber-400 font-mono">Requires Sync</span>
        </div>

        <div className="p-4 rounded-xl bg-[var(--bg-card)] border border-[var(--border-default)] shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-[11px] font-mono text-[var(--text-muted)]">Healthy</span>
            <span className="w-2.5 h-2.5 rounded-full bg-emerald-400 animate-pulse" />
          </div>
          <div className="text-2xl font-bold font-mono text-emerald-400 mt-0.5">
            {isOverviewLoading ? "--" : overview?.healthy ?? 0}
          </div>
          <span className="text-[10px] text-[var(--text-muted)] font-mono">Passing Probes</span>
        </div>

        <div className="p-4 rounded-xl bg-[var(--bg-card)] border border-[var(--border-default)] shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-[11px] font-mono text-[var(--text-muted)]">Degraded</span>
            <span className="w-2.5 h-2.5 rounded-full bg-red-400" />
          </div>
          <div className="text-2xl font-bold font-mono text-red-400 mt-0.5">
            {isOverviewLoading ? "--" : overview?.degraded ?? 0}
          </div>
          <span className="text-[10px] text-[var(--text-muted)] font-mono">Incidents</span>
        </div>
      </section>

      {/* Toolbar: Filters & Search */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-3 bg-[var(--bg-card)] border border-[var(--border-default)] rounded-xl p-3 shadow-xs">
        <div className="flex flex-wrap items-center gap-2 text-xs font-mono">
          {/* Sync Filter Pills */}
          <div className="flex items-center gap-1 bg-[var(--bg-secondary)] p-1 rounded-lg border border-[var(--border-subtle)]">
            <span className="text-[11px] text-[var(--text-muted)] px-1.5">Sync:</span>
            {["all", "Synced", "OutOfSync"].map((st) => (
              <button
                key={st}
                onClick={() => setSyncFilter(st)}
                className={`px-2 py-0.5 rounded text-[11px] transition-colors cursor-pointer ${
                  syncFilter === st
                    ? "bg-cyan-500/20 text-cyan-300 font-bold"
                    : "text-[var(--text-muted)] hover:text-[var(--text-primary)]"
                }`}
              >
                {st}
              </button>
            ))}
          </div>

          {/* Health Filter Pills */}
          <div className="flex items-center gap-1 bg-[var(--bg-secondary)] p-1 rounded-lg border border-[var(--border-subtle)]">
            <span className="text-[11px] text-[var(--text-muted)] px-1.5">Health:</span>
            {["all", "Healthy", "Degraded"].map((h) => (
              <button
                key={h}
                onClick={() => setHealthFilter(h)}
                className={`px-2 py-0.5 rounded text-[11px] transition-colors cursor-pointer ${
                  healthFilter === h
                    ? "bg-cyan-500/20 text-cyan-300 font-bold"
                    : "text-[var(--text-muted)] hover:text-[var(--text-primary)]"
                }`}
              >
                {h}
              </button>
            ))}
          </div>
        </div>

        {/* Search */}
        <div className="relative w-full md:w-72">
          <Search className="w-3.5 h-3.5 absolute left-3 top-1/2 -translate-y-1/2 text-[var(--text-muted)]" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search applications..."
            className="w-full pl-9 pr-3 py-1.5 text-xs font-mono bg-[var(--bg-secondary)] border border-[var(--border-default)] rounded-lg text-[var(--text-primary)] placeholder:text-[var(--text-muted)] focus:outline-none focus:border-cyan-500/50"
          />
        </div>
      </div>

      {/* Applications Grid */}
      {isAppsLoading ? (
        <div className="py-16 text-center text-[var(--text-muted)] font-mono text-xs">
          Loading ArgoCD applications...
        </div>
      ) : applications.length === 0 ? (
        <div className="py-16 text-center text-[var(--text-muted)] font-mono text-xs bg-[var(--bg-card)] border border-[var(--border-default)] rounded-xl">
          No applications matching selected filters.
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {applications.map((app) => {
            const isSynced = app.sync_status === "Synced";
            const isHealthy = app.health_status === "Healthy";

            return (
              <div
                key={app.name}
                className="bg-[var(--bg-card)] border border-[var(--border-default)] hover:border-cyan-500/40 rounded-xl p-5 shadow-xs flex flex-col justify-between text-left transition-all"
              >
                <div>
                  {/* Card Header */}
                  <div className="flex items-start justify-between pb-3 border-b border-[var(--border-subtle)]">
                    <div className="flex items-center gap-2.5">
                      <div className="p-2 rounded-lg bg-cyan-500/10 border border-cyan-500/30 text-cyan-400">
                        <FolderGit2 className="w-4 h-4" />
                      </div>
                      <div>
                        <h3 className="text-sm font-bold text-[var(--text-primary)] font-mono">
                          {app.name}
                        </h3>
                        <span className="text-[11px] font-mono text-[var(--text-muted)]">
                          Project: {app.project || "default"}
                        </span>
                      </div>
                    </div>

                    <div className="flex items-center gap-1.5">
                      <Badge variant={isSynced ? "success" : "warning"} size="sm">
                        {app.sync_status}
                      </Badge>
                      <Badge variant={isHealthy ? "success" : "error"} size="sm">
                        {app.health_status}
                      </Badge>
                    </div>
                  </div>

                  {/* Card Details */}
                  <div className="py-3 space-y-2 text-xs font-mono text-[var(--text-muted)]">
                    <div className="flex items-center gap-2">
                      <GitBranch className="w-3.5 h-3.5 text-pink-400 shrink-0" />
                      <span className="truncate">
                        <strong className="text-[var(--text-secondary)]">Repo:</strong>{" "}
                        {app.repo_url}
                      </span>
                    </div>

                    <div className="flex items-center gap-2">
                      <Layers className="w-3.5 h-3.5 text-cyan-400 shrink-0" />
                      <span>
                        <strong className="text-[var(--text-secondary)]">Path:</strong>{" "}
                        {app.path || "/"}
                      </span>
                      <span className="text-[var(--border-default)]">&bull;</span>
                      <span>
                        <strong className="text-[var(--text-secondary)]">Rev:</strong>{" "}
                        {app.target_revision || "HEAD"}
                      </span>
                    </div>

                    <div className="flex items-center gap-2">
                      <Server className="w-3.5 h-3.5 text-amber-400 shrink-0" />
                      <span>
                        <strong className="text-[var(--text-secondary)]">Dest:</strong>{" "}
                        {app.destination_namespace || "default"}
                      </span>
                    </div>

                    {app.images && app.images.length > 0 && (
                      <div className="pt-1">
                        <span className="text-[10px] text-[var(--text-muted)] block">Images:</span>
                        <div className="text-[11px] text-cyan-300 truncate">
                          {app.images.join(", ")}
                        </div>
                      </div>
                    )}
                  </div>
                </div>

                {/* Card Actions */}
                <div className="pt-3 border-t border-[var(--border-subtle)] flex items-center justify-between">
                  <button
                    onClick={() => setInspectApp(app)}
                    className="inline-flex items-center gap-1 text-xs font-mono text-cyan-400 hover:text-cyan-300 transition-colors cursor-pointer"
                  >
                    <History className="w-3.5 h-3.5" />
                    <span>Inspect Resources</span>
                  </button>

                  <button
                    onClick={() => {
                      setSyncModalApp(app);
                      setSyncPrune(true);
                      setSyncDryRun(false);
                    }}
                    className="inline-flex items-center gap-1.5 px-3 py-1 text-xs font-mono font-semibold rounded-lg bg-cyan-500/20 hover:bg-cyan-500/30 text-cyan-300 border border-cyan-500/40 transition-colors cursor-pointer"
                  >
                    <Play className="w-3 h-3 fill-current" />
                    <span>Sync Application</span>
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Modal 1: Sync Application Modal */}
      <Dialog
        open={!!syncModalApp}
        onOpenChange={(open) => {
          if (!open) setSyncModalApp(null);
        }}
      >
        <DialogContent className="max-w-md">
          <DialogHeader>
            <div className="flex items-center gap-2 text-cyan-400">
              <Play className="w-5 h-5 fill-current" />
              <DialogTitle>Sync ArgoCD Application</DialogTitle>
            </div>
            <DialogDescription className="font-mono text-xs text-[var(--text-muted)]">
              Application: {syncModalApp?.name}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-3 text-left font-mono text-xs">
            <p className="text-[var(--text-secondary)] leading-relaxed">
              Trigger synchronization of live Kubernetes cluster state with Git repository desired
              state.
            </p>

            <div className="p-3 rounded-lg bg-[var(--bg-secondary)] border border-[var(--border-subtle)] space-y-2.5">
              <label className="flex items-center gap-2.5 cursor-pointer select-none">
                <input
                  type="checkbox"
                  checked={syncPrune}
                  onChange={(e) => setSyncPrune(e.target.checked)}
                  className="rounded border-[var(--border-default)] accent-cyan-400 cursor-pointer"
                />
                <span className="text-[var(--text-primary)] font-semibold">Prune Resources</span>
                <span className="text-[10px] text-[var(--text-muted)]">
                  (Delete resources no longer in Git)
                </span>
              </label>

              <label className="flex items-center gap-2.5 cursor-pointer select-none">
                <input
                  type="checkbox"
                  checked={syncDryRun}
                  onChange={(e) => setSyncDryRun(e.target.checked)}
                  className="rounded border-[var(--border-default)] accent-cyan-400 cursor-pointer"
                />
                <span className="text-[var(--text-primary)] font-semibold">Dry Run Mode</span>
                <span className="text-[10px] text-[var(--text-muted)]">
                  (Simulate sync without mutating cluster)
                </span>
              </label>
            </div>
          </div>

          <DialogFooter className="flex items-center justify-end gap-2">
            <button
              onClick={() => setSyncModalApp(null)}
              className="px-3 py-1.5 text-xs font-mono rounded-lg bg-[var(--bg-secondary)] hover:bg-[var(--bg-hover)] text-[var(--text-secondary)] transition-colors cursor-pointer"
            >
              Cancel
            </button>
            <button
              disabled={syncMutation.isPending}
              onClick={() => {
                if (syncModalApp) {
                  syncMutation.mutate({
                    name: syncModalApp.name,
                    req: { prune: syncPrune, dry_run: syncDryRun },
                  });
                }
              }}
              className="px-4 py-1.5 text-xs font-mono font-bold rounded-lg bg-cyan-500 hover:bg-cyan-400 text-black transition-colors cursor-pointer"
            >
              {syncMutation.isPending ? "Syncing..." : "Confirm & Sync"}
            </button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Modal 2: Inspect Resources & History Modal */}
      <Dialog
        open={!!inspectApp}
        onOpenChange={(open) => {
          if (!open) setInspectApp(null);
        }}
      >
        <DialogContent className="max-w-3xl max-h-[85vh] flex flex-col">
          <DialogHeader>
            <div className="flex items-center gap-2 text-cyan-400">
              <FolderGit2 className="w-5 h-5" />
              <DialogTitle className="font-mono">
                {inspectApp?.name} &bull; Resources &amp; Revisions
              </DialogTitle>
            </div>
            <DialogDescription className="font-mono text-xs text-[var(--text-muted)]">
              Managed Kubernetes Custom Resources and Git Deployment History
            </DialogDescription>
          </DialogHeader>

          <div className="flex-1 overflow-y-auto space-y-5 text-left font-mono text-xs pr-1">
            {/* Managed Resources Table */}
            <div>
              <h4 className="text-xs font-bold text-[var(--text-primary)] mb-2 flex items-center gap-1.5">
                <Layers className="w-4 h-4 text-cyan-400" />
                <span>Managed Cluster Resources</span>
              </h4>

              <div className="border border-[var(--border-subtle)] rounded-lg overflow-hidden">
                <table className="w-full text-left border-collapse">
                  <thead>
                    <tr className="bg-[var(--bg-secondary)] text-[var(--text-muted)] text-[11px] border-b border-[var(--border-subtle)]">
                      <th className="py-2 px-3">Kind</th>
                      <th className="py-2 px-3">Name</th>
                      <th className="py-2 px-3">Namespace</th>
                      <th className="py-2 px-3">Sync Status</th>
                      <th className="py-2 px-3">Health</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-[var(--border-subtle)]">
                    {isDetailLoading ? (
                      <tr>
                        <td colSpan={5} className="py-4 text-center text-[var(--text-muted)]">
                          Loading resources...
                        </td>
                      </tr>
                    ) : appDetail?.resources && appDetail.resources.length > 0 ? (
                      appDetail.resources.map((res, i) => (
                        <tr key={i} className="hover:bg-[var(--bg-secondary)]/40">
                          <td className="py-2 px-3 text-cyan-300 font-semibold">{res.kind}</td>
                          <td className="py-2 px-3 text-[var(--text-primary)]">{res.name}</td>
                          <td className="py-2 px-3 text-[var(--text-muted)]">{res.namespace}</td>
                          <td className="py-2 px-3">
                            <Badge
                              variant={res.status === "Synced" ? "success" : "warning"}
                              size="sm"
                            >
                              {res.status}
                            </Badge>
                          </td>
                          <td className="py-2 px-3">
                            {res.health ? (
                              <Badge
                                variant={res.health === "Healthy" ? "success" : "error"}
                                size="sm"
                              >
                                {res.health}
                              </Badge>
                            ) : (
                              <span className="text-[var(--text-muted)]">--</span>
                            )}
                          </td>
                        </tr>
                      ))
                    ) : (
                      <tr>
                        <td colSpan={5} className="py-4 text-center text-[var(--text-muted)]">
                          No resources recorded for this application.
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            </div>

            {/* Deployment History */}
            <div>
              <h4 className="text-xs font-bold text-[var(--text-primary)] mb-2 flex items-center gap-1.5">
                <History className="w-4 h-4 text-pink-400" />
                <span>Deployment Revision History</span>
              </h4>

              <div className="space-y-2">
                {isDetailLoading ? (
                  <div className="text-[var(--text-muted)]">Loading history...</div>
                ) : appDetail?.history && appDetail.history.length > 0 ? (
                  appDetail.history.map((hist, idx) => (
                    <div
                      key={idx}
                      className="p-3 rounded-lg bg-[var(--bg-secondary)] border border-[var(--border-subtle)] flex items-center justify-between"
                    >
                      <div>
                        <div className="flex items-center gap-2">
                          <span className="font-bold text-pink-400">
                            #{hist.id} &bull; {hist.revision.slice(0, 8)}
                          </span>
                          <span className="text-[10px] text-[var(--text-muted)]">
                            {hist.source || "repo source"}
                          </span>
                        </div>
                        <span className="text-[10px] text-[var(--text-muted)]">
                          Full SHA: {hist.revision}
                        </span>
                      </div>
                      <span className="text-[11px] text-[var(--text-muted)]">
                        {hist.deployed_at || "Recent"}
                      </span>
                    </div>
                  ))
                ) : (
                  <div className="text-[var(--text-muted)]">No revision history found.</div>
                )}
              </div>
            </div>
          </div>

          <DialogFooter className="pt-2 border-t border-[var(--border-subtle)]">
            <button
              onClick={() => setInspectApp(null)}
              className="px-4 py-1.5 text-xs font-mono rounded-lg bg-[var(--bg-secondary)] hover:bg-[var(--bg-hover)] text-[var(--text-primary)] transition-colors cursor-pointer"
            >
              Close
            </button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

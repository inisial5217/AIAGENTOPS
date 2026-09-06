"use client";

import * as React from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Layers,
  Box,
  Server,
  Network,
  RefreshCw,
  Search,
  Terminal,
  RotateCw,
  Sliders,
  CheckCircle2,
  AlertCircle,
  Clock,
  Cpu,
  HardDrive,
  Copy,
  Check,
  X,
  ExternalLink,
} from "lucide-react";
import { LogTerminal } from "../../../components/terminal/log-terminal";
import { kubernetesService } from "../../../services/kubernetes-service";
import {
  PodSummary,
  PodDetail,
  DeploymentSummary,
  NodeSummary,
  ServiceSummary,
} from "../../../types/kubernetes";
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

type TabKey = "pods" | "deployments" | "nodes" | "services";

export default function KubernetesDashboardPage() {
  const queryClient = useQueryClient();
  const addNotification = useNotificationStore((s) => s.addNotification);

  // States
  const [activeTab, setActiveTab] = React.useState<TabKey>("pods");
  const [selectedNamespace, setSelectedNamespace] = React.useState<string>("all");
  const [searchQuery, setSearchQuery] = React.useState<string>("");

  // Modal states
  const [selectedPodForLogs, setSelectedPodForLogs] = React.useState<PodSummary | null>(null);
  const [selectedContainer, setSelectedContainer] = React.useState<string>("");
  const [logTail, setLogTail] = React.useState<number>(200);
  const [autoRefreshLogs, setAutoRefreshLogs] = React.useState<boolean>(false);
  const [copiedLogs, setCopiedLogs] = React.useState<boolean>(false);

  const [scaleDeploymentModal, setScaleDeploymentModal] = React.useState<DeploymentSummary | null>(null);
  const [targetReplicas, setTargetReplicas] = React.useState<number>(1);

  const [restartDeploymentModal, setRestartDeploymentModal] = React.useState<DeploymentSummary | null>(null);

  // Queries
  const {
    data: overview,
    isLoading: isOverviewLoading,
    refetch: refetchOverview,
  } = useQuery({
    queryKey: ["k8s", "overview"],
    queryFn: () => kubernetesService.getOverview(),
    refetchInterval: 15000,
  });

  const {
    data: podsData,
    isLoading: isPodsLoading,
    refetch: refetchPods,
  } = useQuery({
    queryKey: ["k8s", "pods", selectedNamespace],
    queryFn: () => kubernetesService.getPods(selectedNamespace),
    refetchInterval: 10000,
  });

  const {
    data: deploymentsData,
    isLoading: isDeploymentsLoading,
    refetch: refetchDeployments,
  } = useQuery({
    queryKey: ["k8s", "deployments", selectedNamespace],
    queryFn: () => kubernetesService.getDeployments(selectedNamespace),
    refetchInterval: 10000,
  });

  const {
    data: nodesData,
    isLoading: isNodesLoading,
    refetch: refetchNodes,
  } = useQuery({
    queryKey: ["k8s", "nodes"],
    queryFn: () => kubernetesService.getNodes(),
    refetchInterval: 15000,
  });

  const {
    data: servicesData,
    isLoading: isServicesLoading,
    refetch: refetchServices,
  } = useQuery({
    queryKey: ["k8s", "services", selectedNamespace],
    queryFn: () => kubernetesService.getServices(selectedNamespace),
    refetchInterval: 15000,
  });

  // Query Pod Logs
  const { data: logsData, isFetching: isLogsFetching, refetch: refetchLogs } = useQuery({
    queryKey: [
      "k8s",
      "pod-logs",
      selectedPodForLogs?.namespace,
      selectedPodForLogs?.name,
      selectedContainer,
      logTail,
    ],
    queryFn: () =>
      selectedPodForLogs
        ? kubernetesService.getPodLogs(
            selectedPodForLogs.namespace,
            selectedPodForLogs.name,
            selectedContainer,
            logTail
          )
        : null,
    enabled: !!selectedPodForLogs,
    refetchInterval: autoRefreshLogs ? 5000 : false,
  });

  // Mutations
  const restartMutation = useMutation({
    mutationFn: ({ namespace, name }: { namespace: string; name: string }) =>
      kubernetesService.restartDeployment(namespace, name),
    onSuccess: (_, vars) => {
      addNotification({
        title: "Rollout Restart Triggered",
        message: `Deployment ${vars.namespace}/${vars.name} rollout restart initiated successfully.`,
        severity: "info",
      });
      setRestartDeploymentModal(null);
      queryClient.invalidateQueries({ queryKey: ["k8s", "deployments"] });
      queryClient.invalidateQueries({ queryKey: ["k8s", "pods"] });
    },
    onError: (err: any) => {
      addNotification({
        title: "Restart Failed",
        message: err?.response?.data?.detail || err.message || "Failed to restart deployment.",
        severity: "error",
      });
    },
  });

  const scaleMutation = useMutation({
    mutationFn: ({
      namespace,
      name,
      replicas,
    }: {
      namespace: string;
      name: string;
      replicas: number;
    }) => kubernetesService.scaleDeployment(namespace, name, replicas),
    onSuccess: (data, vars) => {
      addNotification({
        title: "Deployment Scaled",
        message: `Deployment ${vars.namespace}/${vars.name} target replicas updated to ${vars.replicas}.`,
        severity: "info",
      });
      setScaleDeploymentModal(null);
      queryClient.invalidateQueries({ queryKey: ["k8s", "deployments"] });
      queryClient.invalidateQueries({ queryKey: ["k8s", "pods"] });
      queryClient.invalidateQueries({ queryKey: ["k8s", "overview"] });
    },
    onError: (err: any) => {
      addNotification({
        title: "Scale Failed",
        message: err?.response?.data?.detail || err.message || "Failed to scale deployment.",
        severity: "error",
      });
    },
  });

  const handleRefreshAll = () => {
    refetchOverview();
    refetchPods();
    refetchDeployments();
    refetchNodes();
    refetchServices();
  };

  // Filtered datasets
  const pods = React.useMemo(() => {
    if (!podsData?.data) return [];
    return podsData.data.filter((p) => {
      const matchSearch =
        searchQuery === "" ||
        p.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        p.node.toLowerCase().includes(searchQuery.toLowerCase()) ||
        p.status.toLowerCase().includes(searchQuery.toLowerCase());
      return matchSearch;
    });
  }, [podsData, searchQuery]);

  const deployments = React.useMemo(() => {
    if (!deploymentsData?.data) return [];
    return deploymentsData.data.filter((d) => {
      const matchSearch =
        searchQuery === "" ||
        d.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        d.namespace.toLowerCase().includes(searchQuery.toLowerCase());
      return matchSearch;
    });
  }, [deploymentsData, searchQuery]);

  const nodes = React.useMemo(() => {
    if (!nodesData?.data) return [];
    return nodesData.data.filter((n) => {
      return (
        searchQuery === "" ||
        n.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        n.roles.toLowerCase().includes(searchQuery.toLowerCase()) ||
        n.internal_ip.includes(searchQuery)
      );
    });
  }, [nodesData, searchQuery]);

  const services = React.useMemo(() => {
    if (!servicesData?.data) return [];
    return servicesData.data.filter((s) => {
      return (
        searchQuery === "" ||
        s.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        s.cluster_ip.includes(searchQuery)
      );
    });
  }, [servicesData, searchQuery]);

  // Copy logs helper
  const handleCopyLogs = () => {
    if (!logsData?.logs) return;
    navigator.clipboard.writeText(logsData.logs);
    setCopiedLogs(true);
    setTimeout(() => setCopiedLogs(false), 2000);
  };

  return (
    <div className="space-y-6 max-w-7xl mx-auto pb-12">
      {/* Top Header */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 border-b border-[var(--border-subtle)] pb-4">
        <div className="flex items-center gap-3 text-left">
          <div className="p-2.5 rounded-xl bg-gradient-to-br from-cyan-500/20 via-pink-500/10 to-transparent border border-cyan-500/30 text-cyan-400">
            <Layers className="w-6 h-6" />
          </div>
          <div>
            <h1 className="text-xl font-bold tracking-tight text-[var(--text-primary)]">
              Kubernetes Cluster Management
            </h1>
            <p className="text-xs font-mono text-[var(--text-muted)]">
              Live Workload Telemetry &bull; K3d Local Cluster &bull; Desired Replicas
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={handleRefreshAll}
            className="flex items-center gap-2 px-3 py-2 text-xs font-mono rounded-lg bg-[var(--bg-secondary)] border border-[var(--border-default)] hover:border-cyan-500/50 text-[var(--text-primary)] transition-all cursor-pointer shadow-xs"
          >
            <RefreshCw className="w-3.5 h-3.5" />
            <span>Refresh Cluster</span>
          </button>
        </div>
      </div>

      {/* KPI Cards Grid */}
      <section className="grid grid-cols-2 md:grid-cols-4 gap-4 text-left">
        <div className="p-4 rounded-xl bg-[var(--bg-card)] border border-[var(--border-default)] shadow-xs flex items-center justify-between">
          <div>
            <span className="text-[11px] font-mono text-[var(--text-muted)]">Active Nodes</span>
            <div className="text-2xl font-bold font-mono text-[var(--text-primary)] mt-0.5">
              {isOverviewLoading ? "--" : `${overview?.ready_nodes || 0} / ${overview?.total_nodes || 0}`}
            </div>
            <span className="text-[10px] text-emerald-400 font-mono">100% Ready Status</span>
          </div>
          <div className="p-2.5 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400">
            <Server className="w-5 h-5" />
          </div>
        </div>

        <div className="p-4 rounded-xl bg-[var(--bg-card)] border border-[var(--border-default)] shadow-xs flex items-center justify-between">
          <div>
            <span className="text-[11px] font-mono text-[var(--text-muted)]">Running Pods</span>
            <div className="text-2xl font-bold font-mono text-cyan-400 mt-0.5">
              {isOverviewLoading ? "--" : `${overview?.running_pods || 0} / ${overview?.total_pods || 0}`}
            </div>
            <span className="text-[10px] text-[var(--text-muted)] font-mono">Managed Workloads</span>
          </div>
          <div className="p-2.5 rounded-lg bg-cyan-500/10 border border-cyan-500/20 text-cyan-400">
            <Box className="w-5 h-5" />
          </div>
        </div>

        <div className="p-4 rounded-xl bg-[var(--bg-card)] border border-[var(--border-default)] shadow-xs flex items-center justify-between">
          <div>
            <span className="text-[11px] font-mono text-[var(--text-muted)]">Deployments</span>
            <div className="text-2xl font-bold font-mono text-pink-400 mt-0.5">
              {isOverviewLoading ? "--" : `${overview?.ready_deployments || 0} / ${overview?.total_deployments || 0}`}
            </div>
            <span className="text-[10px] text-pink-400 font-mono">Replicas Healthy</span>
          </div>
          <div className="p-2.5 rounded-lg bg-pink-500/10 border border-pink-500/20 text-pink-400">
            <Layers className="w-5 h-5" />
          </div>
        </div>

        <div className="p-4 rounded-xl bg-[var(--bg-card)] border border-[var(--border-default)] shadow-xs flex items-center justify-between">
          <div>
            <span className="text-[11px] font-mono text-[var(--text-muted)]">Services</span>
            <div className="text-2xl font-bold font-mono text-amber-400 mt-0.5">
              {isOverviewLoading ? "--" : overview?.total_services || 0}
            </div>
            <span className="text-[10px] text-[var(--text-muted)] font-mono">Cluster Endpoints</span>
          </div>
          <div className="p-2.5 rounded-lg bg-amber-500/10 border border-amber-500/20 text-amber-400">
            <Network className="w-5 h-5" />
          </div>
        </div>
      </section>

      {/* Toolbar: Namespace selector & Search */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-3 bg-[var(--bg-card)] border border-[var(--border-default)] rounded-xl p-3 shadow-xs">
        {/* Namespace Pills */}
        <div className="flex items-center gap-1.5 overflow-x-auto text-xs font-mono">
          <span className="text-[11px] text-[var(--text-muted)] mr-1">Namespace:</span>
          {["all", "default", "argocd", "kube-system"].map((ns) => (
            <button
              key={ns}
              onClick={() => setSelectedNamespace(ns)}
              className={`px-2.5 py-1 rounded-md transition-all cursor-pointer ${
                selectedNamespace === ns
                  ? "bg-cyan-500/20 text-cyan-400 border border-cyan-500/40 font-bold"
                  : "bg-[var(--bg-secondary)] text-[var(--text-secondary)] hover:text-[var(--text-primary)] border border-transparent"
              }`}
            >
              {ns}
            </button>
          ))}
        </div>

        {/* Search */}
        <div className="relative w-full md:w-72">
          <Search className="w-3.5 h-3.5 absolute left-3 top-1/2 -translate-y-1/2 text-[var(--text-muted)]" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search resources..."
            className="w-full pl-9 pr-3 py-1.5 text-xs font-mono bg-[var(--bg-secondary)] border border-[var(--border-default)] rounded-lg text-[var(--text-primary)] placeholder:text-[var(--text-muted)] focus:outline-none focus:border-cyan-500/50"
          />
        </div>
      </div>

      {/* Tab Navigation */}
      <div className="flex items-center gap-2 border-b border-[var(--border-default)] pb-px">
        {[
          { key: "pods", label: "Workloads & Pods", count: pods.length },
          { key: "deployments", label: "Deployments", count: deployments.length },
          { key: "nodes", label: "Cluster Nodes", count: nodes.length },
          { key: "services", label: "Services & Routing", count: services.length },
        ].map((tab) => (
          <button
            key={tab.key}
            onClick={() => setActiveTab(tab.key as TabKey)}
            className={`flex items-center gap-2 px-4 py-2.5 text-xs font-mono border-b-2 transition-all cursor-pointer ${
              activeTab === tab.key
                ? "border-cyan-400 text-cyan-400 font-bold"
                : "border-transparent text-[var(--text-muted)] hover:text-[var(--text-secondary)]"
            }`}
          >
            <span>{tab.label}</span>
            <span
              className={`px-1.5 py-0.2 rounded text-[10px] ${
                activeTab === tab.key
                  ? "bg-cyan-500/20 text-cyan-300"
                  : "bg-[var(--bg-secondary)] text-[var(--text-muted)]"
              }`}
            >
              {tab.count}
            </span>
          </button>
        ))}
      </div>

      {/* Tab 1: Pods */}
      {activeTab === "pods" && (
        <div className="bg-[var(--bg-card)] border border-[var(--border-default)] rounded-xl overflow-hidden shadow-xs">
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs font-mono border-collapse">
              <thead>
                <tr className="bg-[var(--bg-secondary)]/80 text-[var(--text-muted)] border-b border-[var(--border-subtle)]">
                  <th className="py-3 px-4">Pod Name</th>
                  <th className="py-3 px-4">Namespace</th>
                  <th className="py-3 px-4">Status</th>
                  <th className="py-3 px-4">Restarts</th>
                  <th className="py-3 px-4">Node</th>
                  <th className="py-3 px-4">IP</th>
                  <th className="py-3 px-4">Requests (CPU/RAM)</th>
                  <th className="py-3 px-4">Age</th>
                  <th className="py-3 px-4 text-right">Action</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[var(--border-subtle)]">
                {isPodsLoading ? (
                  <tr>
                    <td colSpan={9} className="py-8 text-center text-[var(--text-muted)]">
                      Loading cluster pods...
                    </td>
                  </tr>
                ) : pods.length === 0 ? (
                  <tr>
                    <td colSpan={9} className="py-8 text-center text-[var(--text-muted)]">
                      No pods found in namespace &quot;{selectedNamespace}&quot;
                    </td>
                  </tr>
                ) : (
                  pods.map((pod) => {
                    const isRunning = pod.status === "Running";
                    const isPending = pod.status === "Pending";
                    const isCrash = pod.status === "CrashLoopBackOff" || pod.status === "Error";

                    return (
                      <tr
                        key={pod.name}
                        className="hover:bg-[var(--bg-secondary)]/40 transition-colors"
                      >
                        <td className="py-3 px-4 font-semibold text-[var(--text-primary)]">
                          <div className="flex items-center gap-2">
                            <Box className="w-3.5 h-3.5 text-cyan-400 shrink-0" />
                            <span className="truncate max-w-xs">{pod.name}</span>
                          </div>
                        </td>
                        <td className="py-3 px-4 text-[var(--text-muted)]">{pod.namespace}</td>
                        <td className="py-3 px-4">
                          <Badge
                            variant={isRunning ? "success" : isPending ? "warning" : isCrash ? "error" : "neutral"}
                            size="sm"
                            pulse={isRunning}
                          >
                            {pod.status}
                          </Badge>
                        </td>
                        <td className="py-3 px-4">
                          <span
                            className={
                              pod.restarts > 0 ? "text-amber-400 font-bold" : "text-[var(--text-muted)]"
                            }
                          >
                            {pod.restarts}
                          </span>
                        </td>
                        <td className="py-3 px-4 text-[var(--text-muted)] truncate max-w-[120px]">
                          {pod.node || "Pending"}
                        </td>
                        <td className="py-3 px-4 text-[var(--text-muted)]">{pod.ip || "--"}</td>
                        <td className="py-3 px-4 text-[var(--text-muted)] text-[11px]">
                          {pod.cpu_request || "--"} / {pod.memory_request || "--"}
                        </td>
                        <td className="py-3 px-4 text-[var(--text-muted)]">{pod.age}</td>
                        <td className="py-3 px-4 text-right">
                          <button
                            onClick={() => {
                              setSelectedPodForLogs(pod);
                              setSelectedContainer("");
                            }}
                            className="inline-flex items-center gap-1 px-2.5 py-1 text-[11px] rounded bg-cyan-500/10 hover:bg-cyan-500/20 text-cyan-400 border border-cyan-500/30 transition-colors cursor-pointer"
                          >
                            <Terminal className="w-3 h-3" />
                            <span>Logs</span>
                          </button>
                        </td>
                      </tr>
                    );
                  })
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Tab 2: Deployments */}
      {activeTab === "deployments" && (
        <div className="bg-[var(--bg-card)] border border-[var(--border-default)] rounded-xl overflow-hidden shadow-xs">
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs font-mono border-collapse">
              <thead>
                <tr className="bg-[var(--bg-secondary)]/80 text-[var(--text-muted)] border-b border-[var(--border-subtle)]">
                  <th className="py-3 px-4">Deployment</th>
                  <th className="py-3 px-4">Namespace</th>
                  <th className="py-3 px-4">Replicas (Ready / Desired)</th>
                  <th className="py-3 px-4">Image</th>
                  <th className="py-3 px-4">Age</th>
                  <th className="py-3 px-4 text-right">Orchestration</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[var(--border-subtle)]">
                {isDeploymentsLoading ? (
                  <tr>
                    <td colSpan={6} className="py-8 text-center text-[var(--text-muted)]">
                      Loading deployments...
                    </td>
                  </tr>
                ) : deployments.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="py-8 text-center text-[var(--text-muted)]">
                      No deployments found
                    </td>
                  </tr>
                ) : (
                  deployments.map((d) => {
                    const isHealthy = d.ready_replicas === d.replicas && d.replicas > 0;

                    return (
                      <tr key={d.name} className="hover:bg-[var(--bg-secondary)]/40 transition-colors">
                        <td className="py-3 px-4 font-semibold text-[var(--text-primary)]">
                          <div className="flex items-center gap-2">
                            <Layers className="w-3.5 h-3.5 text-pink-400 shrink-0" />
                            <span>{d.name}</span>
                          </div>
                        </td>
                        <td className="py-3 px-4 text-[var(--text-muted)]">{d.namespace}</td>
                        <td className="py-3 px-4">
                          <div className="flex items-center gap-2">
                            <span
                              className={`font-bold ${
                                isHealthy ? "text-emerald-400" : "text-amber-400"
                              }`}
                            >
                              {d.ready_replicas} / {d.replicas}
                            </span>
                            <span className="text-[10px] text-[var(--text-muted)]">
                              ({d.available_replicas} available)
                            </span>
                          </div>
                        </td>
                        <td className="py-3 px-4 text-[var(--text-muted)] truncate max-w-xs text-[11px]">
                          {d.images?.join(", ") || "--"}
                        </td>
                        <td className="py-3 px-4 text-[var(--text-muted)]">{d.age}</td>
                        <td className="py-3 px-4 text-right">
                          <div className="flex items-center justify-end gap-1.5">
                            <button
                              onClick={() => {
                                setScaleDeploymentModal(d);
                                setTargetReplicas(d.replicas);
                              }}
                              className="inline-flex items-center gap-1 px-2.5 py-1 text-[11px] rounded bg-[var(--bg-secondary)] hover:bg-[var(--bg-hover)] text-[var(--text-primary)] border border-[var(--border-default)] transition-colors cursor-pointer"
                              title="Scale Replicas"
                            >
                              <Sliders className="w-3 h-3 text-cyan-400" />
                              <span>Scale</span>
                            </button>

                            <button
                              onClick={() => setRestartDeploymentModal(d)}
                              className="inline-flex items-center gap-1 px-2.5 py-1 text-[11px] rounded bg-pink-500/10 hover:bg-pink-500/20 text-pink-400 border border-pink-500/30 transition-colors cursor-pointer"
                              title="Rollout Restart"
                            >
                              <RotateCw className="w-3 h-3" />
                              <span>Restart</span>
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
      )}

      {/* Tab 3: Nodes */}
      {activeTab === "nodes" && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {isNodesLoading ? (
            <div className="col-span-3 py-12 text-center text-[var(--text-muted)] font-mono text-xs">
              Loading cluster nodes...
            </div>
          ) : (
            nodes.map((node) => {
              const memGB = (node.memory_capacity_bytes / (1024 * 1024 * 1024)).toFixed(1);

              return (
                <div
                  key={node.name}
                  className="bg-[var(--bg-card)] border border-[var(--border-default)] rounded-xl p-5 shadow-xs flex flex-col justify-between text-left"
                >
                  <div className="flex items-start justify-between pb-3 border-b border-[var(--border-subtle)]">
                    <div className="flex items-center gap-2.5">
                      <div className="p-2 rounded-lg bg-emerald-500/10 border border-emerald-500/30 text-emerald-400">
                        <Server className="w-4 h-4" />
                      </div>
                      <div>
                        <h4 className="text-sm font-bold text-[var(--text-primary)] font-mono">
                          {node.name}
                        </h4>
                        <span className="text-[11px] font-mono text-cyan-400">
                          Role: {node.roles}
                        </span>
                      </div>
                    </div>
                    <Badge variant={node.status === "Ready" ? "success" : "error"} size="sm">
                      {node.status}
                    </Badge>
                  </div>

                  <div className="grid grid-cols-2 gap-3 py-4 text-xs font-mono text-[var(--text-muted)]">
                    <div>
                      <span className="text-[10px] block">Kubelet Version</span>
                      <span className="text-[var(--text-primary)] font-bold">{node.version}</span>
                    </div>
                    <div>
                      <span className="text-[10px] block">Internal IP</span>
                      <span className="text-[var(--text-primary)] font-bold">{node.internal_ip}</span>
                    </div>
                    <div>
                      <span className="text-[10px] block">CPU Capacity</span>
                      <span className="text-[var(--text-primary)] font-bold">
                        {node.cpu_capacity} cores
                      </span>
                    </div>
                    <div>
                      <span className="text-[10px] block">RAM Capacity</span>
                      <span className="text-[var(--text-primary)] font-bold">{memGB} GB</span>
                    </div>
                  </div>

                  <div className="pt-3 border-t border-[var(--border-subtle)] flex items-center justify-between text-xs font-mono">
                    <span className="text-[var(--text-muted)]">Workloads Hosted</span>
                    <span className="text-cyan-400 font-bold">{node.pod_count} Active Pods</span>
                  </div>
                </div>
              );
            })
          )}
        </div>
      )}

      {/* Tab 4: Services */}
      {activeTab === "services" && (
        <div className="bg-[var(--bg-card)] border border-[var(--border-default)] rounded-xl overflow-hidden shadow-xs">
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs font-mono border-collapse">
              <thead>
                <tr className="bg-[var(--bg-secondary)]/80 text-[var(--text-muted)] border-b border-[var(--border-subtle)]">
                  <th className="py-3 px-4">Service Name</th>
                  <th className="py-3 px-4">Namespace</th>
                  <th className="py-3 px-4">Type</th>
                  <th className="py-3 px-4">Cluster IP</th>
                  <th className="py-3 px-4">External IP</th>
                  <th className="py-3 px-4">Port Mappings</th>
                  <th className="py-3 px-4">Age</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[var(--border-subtle)]">
                {isServicesLoading ? (
                  <tr>
                    <td colSpan={7} className="py-8 text-center text-[var(--text-muted)]">
                      Loading services...
                    </td>
                  </tr>
                ) : services.length === 0 ? (
                  <tr>
                    <td colSpan={7} className="py-8 text-center text-[var(--text-muted)]">
                      No services found
                    </td>
                  </tr>
                ) : (
                  services.map((svc) => (
                    <tr key={svc.name} className="hover:bg-[var(--bg-secondary)]/40 transition-colors">
                      <td className="py-3 px-4 font-semibold text-[var(--text-primary)]">
                        <div className="flex items-center gap-2">
                          <Network className="w-3.5 h-3.5 text-amber-400 shrink-0" />
                          <span>{svc.name}</span>
                        </div>
                      </td>
                      <td className="py-3 px-4 text-[var(--text-muted)]">{svc.namespace}</td>
                      <td className="py-3 px-4">
                        <Badge variant="cyan" size="sm">
                          {svc.type}
                        </Badge>
                      </td>
                      <td className="py-3 px-4 text-[var(--text-muted)]">{svc.cluster_ip}</td>
                      <td className="py-3 px-4 text-[var(--text-muted)]">{svc.external_ip || "--"}</td>
                      <td className="py-3 px-4 text-[11px] text-[var(--text-muted)]">
                        {svc.ports
                          ?.map((p) => `${p.port}${p.node_port ? `:${p.node_port}` : ""}/${p.protocol}`)
                          .join(", ") || "--"}
                      </td>
                      <td className="py-3 px-4 text-[var(--text-muted)]">{svc.age}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Modal 1: Live Pod Logs Viewer */}
      <Dialog
        open={!!selectedPodForLogs}
        onOpenChange={(open) => {
          if (!open) {
            setSelectedPodForLogs(null);
            setAutoRefreshLogs(false);
          }
        }}
      >
        <DialogContent className="max-w-4xl h-[640px] flex flex-col">
          <DialogHeader>
            <div className="flex items-center justify-between pr-6">
              <div className="flex items-center gap-2.5">
                <Terminal className="w-5 h-5 text-cyan-400" />
                <DialogTitle className="font-mono text-sm">
                  Logs: {selectedPodForLogs?.namespace}/{selectedPodForLogs?.name}
                </DialogTitle>
              </div>

              <div className="flex items-center gap-2 text-xs font-mono">
                {/* Tail Selector */}
                <select
                  value={logTail}
                  onChange={(e) => setLogTail(Number(e.target.value))}
                  className="bg-[var(--bg-secondary)] border border-[var(--border-default)] rounded px-2 py-1 text-xs text-[var(--text-primary)] focus:outline-none"
                >
                  <option value={100}>100 lines</option>
                  <option value={200}>200 lines</option>
                  <option value={500}>500 lines</option>
                  <option value={1000}>1000 lines</option>
                </select>

                {/* Auto Refresh Toggle */}
                <button
                  onClick={() => setAutoRefreshLogs(!autoRefreshLogs)}
                  className={`px-2 py-1 rounded border transition-colors cursor-pointer ${
                    autoRefreshLogs
                      ? "bg-emerald-500/20 text-emerald-300 border-emerald-500/40 font-bold"
                      : "bg-[var(--bg-secondary)] text-[var(--text-muted)] border-[var(--border-default)]"
                  }`}
                >
                  {autoRefreshLogs ? "Live: 5s" : "Auto-refresh: Off"}
                </button>

                {/* Manual Refresh */}
                <button
                  onClick={() => refetchLogs()}
                  className="p-1.5 rounded bg-[var(--bg-secondary)] text-[var(--text-primary)] hover:bg-[var(--bg-hover)] transition-colors cursor-pointer"
                  title="Refresh Logs"
                >
                  <RefreshCw
                    className={`w-3.5 h-3.5 ${isLogsFetching ? "animate-spin text-cyan-400" : ""}`}
                  />
                </button>

                {/* Copy */}
                <button
                  onClick={handleCopyLogs}
                  className="p-1.5 rounded bg-[var(--bg-secondary)] text-[var(--text-primary)] hover:bg-[var(--bg-hover)] transition-colors cursor-pointer"
                  title="Copy Logs to Clipboard"
                >
                  {copiedLogs ? (
                    <Check className="w-3.5 h-3.5 text-emerald-400" />
                  ) : (
                    <Copy className="w-3.5 h-3.5" />
                  )}
                </button>
              </div>
            </div>
            <DialogDescription className="font-mono text-[11px] text-[var(--text-muted)]">
              Container Output Stream &bull; Live Kubernetes Pod Logs
            </DialogDescription>
          </DialogHeader>

          {/* Terminal View */}
          <div className="flex-1 pt-1">
            <LogTerminal
              initialLogs={logsData?.logs ? logsData.logs.trim().split("\n") : []}
              title={`${selectedPodForLogs?.name} (${selectedContainer || "main"})`}
              height="h-[460px]"
              onRefresh={() => refetchLogs()}
            />
          </div>
        </DialogContent>
      </Dialog>

      {/* Modal 2: Scale Deployment Modal */}
      <Dialog
        open={!!scaleDeploymentModal}
        onOpenChange={(open) => {
          if (!open) setScaleDeploymentModal(null);
        }}
      >
        <DialogContent className="max-w-md">
          <DialogHeader>
            <div className="flex items-center gap-2 text-cyan-400">
              <Sliders className="w-5 h-5" />
              <DialogTitle>Scale Deployment Replicas</DialogTitle>
            </div>
            <DialogDescription className="font-mono text-xs text-[var(--text-muted)]">
              {scaleDeploymentModal?.namespace} / {scaleDeploymentModal?.name}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4 text-left font-mono text-xs">
            <div className="p-3 rounded-lg bg-[var(--bg-secondary)] border border-[var(--border-subtle)] flex items-center justify-between">
              <span className="text-[var(--text-muted)]">Current Replicas:</span>
              <span className="text-base font-bold text-cyan-400">
                {scaleDeploymentModal?.replicas}
              </span>
            </div>

            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <label className="text-[var(--text-secondary)]">Desired Target Replicas:</label>
                <span className="text-base font-bold text-pink-400">{targetReplicas}</span>
              </div>
              <input
                type="range"
                min="0"
                max="10"
                value={targetReplicas}
                onChange={(e) => setTargetReplicas(Number(e.target.value))}
                className="w-full accent-cyan-400 cursor-pointer"
              />
              <div className="flex justify-between text-[10px] text-[var(--text-muted)]">
                <span>0 (Stopped)</span>
                <span>5 Replicas</span>
                <span>10 Max</span>
              </div>
            </div>
          </div>

          <DialogFooter className="flex items-center justify-end gap-2">
            <button
              onClick={() => setScaleDeploymentModal(null)}
              className="px-3 py-1.5 text-xs font-mono rounded-lg bg-[var(--bg-secondary)] hover:bg-[var(--bg-hover)] text-[var(--text-secondary)] transition-colors cursor-pointer"
            >
              Cancel
            </button>
            <button
              disabled={scaleMutation.isPending}
              onClick={() => {
                if (scaleDeploymentModal) {
                  scaleMutation.mutate({
                    namespace: scaleDeploymentModal.namespace,
                    name: scaleDeploymentModal.name,
                    replicas: targetReplicas,
                  });
                }
              }}
              className="px-4 py-1.5 text-xs font-mono font-bold rounded-lg bg-cyan-500 hover:bg-cyan-400 text-black transition-colors cursor-pointer"
            >
              {scaleMutation.isPending ? "Scaling..." : "Confirm Scale"}
            </button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Modal 3: Restart Deployment Modal */}
      <Dialog
        open={!!restartDeploymentModal}
        onOpenChange={(open) => {
          if (!open) setRestartDeploymentModal(null);
        }}
      >
        <DialogContent className="max-w-md">
          <DialogHeader>
            <div className="flex items-center gap-2 text-pink-400">
              <RotateCw className="w-5 h-5" />
              <DialogTitle>Rollout Restart Deployment</DialogTitle>
            </div>
            <DialogDescription className="font-mono text-xs text-[var(--text-muted)]">
              {restartDeploymentModal?.namespace} / {restartDeploymentModal?.name}
            </DialogDescription>
          </DialogHeader>

          <div className="py-3 text-left font-mono text-xs text-[var(--text-secondary)] space-y-2">
            <p>
              This will perform a zero-downtime rolling restart by updating the pod template
              annotation. All current replica pods will be terminated sequentially once new pods
              reach ready state.
            </p>
            <div className="p-2.5 rounded bg-pink-500/10 border border-pink-500/20 text-pink-300 text-[11px]">
              Requires DevOps or Admin authorization. Action is recorded in audit logs.
            </div>
          </div>

          <DialogFooter className="flex items-center justify-end gap-2">
            <button
              onClick={() => setRestartDeploymentModal(null)}
              className="px-3 py-1.5 text-xs font-mono rounded-lg bg-[var(--bg-secondary)] hover:bg-[var(--bg-hover)] text-[var(--text-secondary)] transition-colors cursor-pointer"
            >
              Cancel
            </button>
            <button
              disabled={restartMutation.isPending}
              onClick={() => {
                if (restartDeploymentModal) {
                  restartMutation.mutate({
                    namespace: restartDeploymentModal.namespace,
                    name: restartDeploymentModal.name,
                  });
                }
              }}
              className="px-4 py-1.5 text-xs font-mono font-bold rounded-lg bg-pink-500 hover:bg-pink-400 text-black transition-colors cursor-pointer"
            >
              {restartMutation.isPending ? "Triggering..." : "Rollout Restart"}
            </button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

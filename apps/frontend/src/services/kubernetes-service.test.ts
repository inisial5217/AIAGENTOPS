import { describe, it, expect, vi, beforeEach } from "vitest";
import { kubernetesService } from "./kubernetes-service";
import { apiClient } from "../lib/api-client";

vi.mock("../lib/api-client", () => {
  const mockClient = {
    get: vi.fn(),
    post: vi.fn(),
  };
  return {
    default: mockClient,
    apiClient: mockClient,
  };
});

describe("kubernetesService", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls getPods with optional namespace filter", async () => {
    const mockPods = [{ name: "nginx-123", namespace: "default", status: "Running" }];
    vi.mocked(apiClient.get).mockResolvedValueOnce({
      data: { data: mockPods, total: 1 },
    });

    const result = await kubernetesService.getPods("default");
    expect(apiClient.get).toHaveBeenCalledWith("/api/v1/kubernetes/pods", {
      params: { namespace: "default" },
    });
    expect(result.data).toEqual(mockPods);
  });

  it("calls getPod with namespace and name", async () => {
    const mockPod = { name: "nginx-123", namespace: "default", qos_class: "BestEffort", containers: [] };
    vi.mocked(apiClient.get).mockResolvedValueOnce({ data: mockPod });

    const result = await kubernetesService.getPod("default", "nginx-123");
    expect(apiClient.get).toHaveBeenCalledWith("/api/v1/kubernetes/pods/default/nginx-123");
    expect(result.name).toBe("nginx-123");
  });

  it("calls getPodLogs with container and tail", async () => {
    const mockLogs = { namespace: "default", pod: "p1", container: "c1", tail: 200, logs: "test logs" };
    vi.mocked(apiClient.get).mockResolvedValueOnce({ data: mockLogs });

    const result = await kubernetesService.getPodLogs("default", "p1", "c1", 200);
    expect(apiClient.get).toHaveBeenCalledWith("/api/v1/kubernetes/pods/default/p1/logs", {
      params: { container: "c1", tail: 200 },
    });
    expect(result.logs).toBe("test logs");
  });

  it("calls getDeployments with optional namespace", async () => {
    const mockDeps = [{ name: "api", namespace: "default", replicas: 3, ready_replicas: 3 }];
    vi.mocked(apiClient.get).mockResolvedValueOnce({ data: { data: mockDeps, total: 1 } });

    const result = await kubernetesService.getDeployments("default");
    expect(apiClient.get).toHaveBeenCalledWith("/api/v1/kubernetes/deployments", {
      params: { namespace: "default" },
    });
    expect(result.total).toBe(1);
  });

  it("calls restartDeployment with POST", async () => {
    vi.mocked(apiClient.post).mockResolvedValueOnce({
      data: { status: "success", message: "rollout restarted" },
    });

    const result = await kubernetesService.restartDeployment("default", "api");
    expect(apiClient.post).toHaveBeenCalledWith("/api/v1/kubernetes/deployments/default/api/restart");
    expect(result.status).toBe("success");
  });

  it("calls scaleDeployment with POST body", async () => {
    vi.mocked(apiClient.post).mockResolvedValueOnce({
      data: { status: "success", message: "scaled", replicas: 5 },
    });

    const result = await kubernetesService.scaleDeployment("default", "api", 5);
    expect(apiClient.post).toHaveBeenCalledWith(
      "/api/v1/kubernetes/deployments/default/api/scale",
      { replicas: 5 }
    );
    expect(result.replicas).toBe(5);
  });

  it("calls getNodes", async () => {
    const mockNodes = [{ name: "node-1", status: "Ready", roles: "control-plane" }];
    vi.mocked(apiClient.get).mockResolvedValueOnce({ data: { data: mockNodes, total: 1 } });

    const result = await kubernetesService.getNodes();
    expect(apiClient.get).toHaveBeenCalledWith("/api/v1/kubernetes/nodes");
    expect(result.data).toEqual(mockNodes);
  });

  it("calls getServices", async () => {
    const mockServices = [{ name: "svc-1", namespace: "default", type: "ClusterIP" }];
    vi.mocked(apiClient.get).mockResolvedValueOnce({ data: { data: mockServices, total: 1 } });

    const result = await kubernetesService.getServices();
    expect(apiClient.get).toHaveBeenCalledWith("/api/v1/kubernetes/services", { params: {} });
    expect(result.data).toEqual(mockServices);
  });

  it("calls getOverview", async () => {
    const mockOverview = {
      total_nodes: 1,
      ready_nodes: 1,
      total_pods: 24,
      running_pods: 24,
      total_deployments: 15,
      ready_deployments: 15,
      total_services: 17,
    };
    vi.mocked(apiClient.get).mockResolvedValueOnce({ data: mockOverview });

    const result = await kubernetesService.getOverview();
    expect(apiClient.get).toHaveBeenCalledWith("/api/v1/kubernetes/overview");
    expect(result.total_pods).toBe(24);
  });
});

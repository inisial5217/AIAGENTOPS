import { describe, it, expect, vi, beforeEach } from "vitest";
import { dockerService } from "./docker-service";
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

describe("dockerService", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls getContainers with optional status filter", async () => {
    const mockContainers = [{ id: "c-123", names: ["/cifo-postgres"], state: "running" }];
    vi.mocked(apiClient.get).mockResolvedValueOnce({
      data: { data: mockContainers, total: 1 },
    });

    const result = await dockerService.getContainers("running");
    expect(apiClient.get).toHaveBeenCalledWith("/api/v1/docker/containers", {
      params: { status: "running" },
    });
    expect(result.data).toEqual(mockContainers);
  });

  it("calls getContainerStats with container id", async () => {
    const mockStats = {
      container_id: "c-123",
      container_name: "cifo-postgres",
      cpu_percentage: 4.5,
      memory_usage_bytes: 1000,
      memory_limit_bytes: 8000,
      memory_percentage: 12.2,
      network_rx_bytes: 100,
      network_tx_bytes: 200,
      block_read_bytes: 0,
      block_write_bytes: 0,
      pids_current: 5,
    };
    vi.mocked(apiClient.get).mockResolvedValueOnce({
      data: mockStats,
    });

    const result = await dockerService.getContainerStats("c-123");
    expect(apiClient.get).toHaveBeenCalledWith("/api/v1/docker/containers/c-123/stats");
    expect(result.cpu_percentage).toBe(4.5);
  });

  it("calls restartContainer endpoint with POST", async () => {
    vi.mocked(apiClient.post).mockResolvedValueOnce({
      data: { status: "success", message: "Container restarted" },
    });

    const result = await dockerService.restartContainer("c-123");
    expect(apiClient.post).toHaveBeenCalledWith("/api/v1/docker/containers/c-123/restart");
    expect(result.status).toBe("success");
  });

  it("calls getDashboardStats endpoint", async () => {
    const mockDashboard = {
      total_containers: 19,
      total_replicas: 4,
      overall_ram_percent: 62.5,
      used_ram_bytes: 1000,
      total_ram_bytes: 2000,
      containers_on: 13,
      containers_off: 6,
      active_incidents: 0,
    };
    vi.mocked(apiClient.get).mockResolvedValueOnce({
      data: mockDashboard,
    });

    const result = await dockerService.getDashboardStats();
    expect(apiClient.get).toHaveBeenCalledWith("/api/v1/monitoring/stats");
    expect(result.total_containers).toBe(19);
  });
});

import { apiClient } from "../lib/api";
import {
  ContainerSummary,
  ContainerDetail,
  ContainerStats,
  DockerSystemInfo,
  DockerImageInfo,
  DockerVolumeInfo,
  DockerNetworkInfo,
  DashboardStats,
  TimeSeriesPoint,
} from "../types/docker";

export const dockerService = {
  // getContainers retrieves container list
  async getContainers(status?: string): Promise<{ data: ContainerSummary[]; total: number }> {
    const params = status && status !== "all" ? { status } : {};
    const res = await apiClient.get("/api/v1/docker/containers", { params });
    return res.data;
  },

  // getContainer retrieves container detail
  async getContainer(id: string): Promise<ContainerDetail> {
    const res = await apiClient.get(`/api/v1/docker/containers/${id}`);
    return res.data;
  },

  // getContainerStats retrieves live resource stats
  async getContainerStats(id: string): Promise<ContainerStats> {
    const res = await apiClient.get(`/api/v1/docker/containers/${id}/stats`);
    return res.data;
  },

  // getContainerLogs retrieves container logs
  async getContainerLogs(
    id: string,
    tail: number = 200
  ): Promise<{ container_id: string; tail: number; logs: string }> {
    const res = await apiClient.get(`/api/v1/docker/containers/${id}/logs`, {
      params: { tail },
    });
    return res.data;
  },

  // restartContainer restarts a container
  async restartContainer(id: string): Promise<{ status: string; message: string }> {
    const res = await apiClient.post(`/api/v1/docker/containers/${id}/restart`);
    return res.data;
  },

  // stopContainer stops a container
  async stopContainer(id: string): Promise<{ status: string; message: string }> {
    const res = await apiClient.post(`/api/v1/docker/containers/${id}/stop`);
    return res.data;
  },

  // getImages retrieves docker images
  async getImages(): Promise<{ data: DockerImageInfo[]; total: number }> {
    const res = await apiClient.get("/api/v1/docker/images");
    return res.data;
  },

  // getVolumes retrieves docker volumes
  async getVolumes(): Promise<{ data: DockerVolumeInfo[]; total: number }> {
    const res = await apiClient.get("/api/v1/docker/volumes");
    return res.data;
  },

  // getNetworks retrieves docker networks
  async getNetworks(): Promise<{ data: DockerNetworkInfo[]; total: number }> {
    const res = await apiClient.get("/api/v1/docker/networks");
    return res.data;
  },

  // getSystemInfo retrieves docker system host info
  async getSystemInfo(): Promise<DockerSystemInfo> {
    const res = await apiClient.get("/api/v1/docker/system");
    return res.data;
  },

  // getDashboardStats retrieves aggregated KPI stats
  async getDashboardStats(): Promise<DashboardStats> {
    const res = await apiClient.get("/api/v1/monitoring/stats");
    return res.data;
  },

  // getCPUMetrics retrieves CPU time-series
  async getCPUMetrics(
    range: string = "1h"
  ): Promise<{ range: string; data: TimeSeriesPoint[] }> {
    const res = await apiClient.get("/api/v1/monitoring/metrics/cpu", {
      params: { range },
    });
    return res.data;
  },

  // getMemoryMetrics retrieves memory time-series
  async getMemoryMetrics(
    range: string = "1h"
  ): Promise<{ range: string; data: TimeSeriesPoint[] }> {
    const res = await apiClient.get("/api/v1/monitoring/metrics/memory", {
      params: { range },
    });
    return res.data;
  },

  // getNetworkMetrics retrieves network time-series
  async getNetworkMetrics(
    range: string = "1h"
  ): Promise<{ range: string; data: TimeSeriesPoint[] }> {
    const res = await apiClient.get("/api/v1/monitoring/metrics/network", {
      params: { range },
    });
    return res.data;
  },
};

import { apiClient } from "../lib/api-client";
import {
  PodSummary,
  PodDetail,
  DeploymentSummary,
  DeploymentDetail,
  NodeSummary,
  ServiceSummary,
  K8sClusterOverview,
} from "../types/kubernetes";

export const kubernetesService = {
  // getPods retrieves pod list
  async getPods(namespace?: string): Promise<{ data: PodSummary[]; total: number }> {
    const params = namespace && namespace !== "all" ? { namespace } : {};
    const res = await apiClient.get("/api/v1/kubernetes/pods", { params });
    return res.data;
  },

  // getPod retrieves pod detail
  async getPod(namespace: string, name: string): Promise<PodDetail> {
    const res = await apiClient.get(`/api/v1/kubernetes/pods/${namespace}/${name}`);
    return res.data;
  },

  // getPodLogs retrieves pod container logs
  async getPodLogs(
    namespace: string,
    name: string,
    container?: string,
    tail: number = 200
  ): Promise<{ namespace: string; pod: string; container: string; tail: number; logs: string }> {
    const res = await apiClient.get(`/api/v1/kubernetes/pods/${namespace}/${name}/logs`, {
      params: { container, tail },
    });
    return res.data;
  },

  // getDeployments retrieves deployment list
  async getDeployments(namespace?: string): Promise<{ data: DeploymentSummary[]; total: number }> {
    const params = namespace && namespace !== "all" ? { namespace } : {};
    const res = await apiClient.get("/api/v1/kubernetes/deployments", { params });
    return res.data;
  },

  // getDeployment retrieves deployment detail
  async getDeployment(namespace: string, name: string): Promise<DeploymentDetail> {
    const res = await apiClient.get(`/api/v1/kubernetes/deployments/${namespace}/${name}`);
    return res.data;
  },

  // restartDeployment triggers deployment restart
  async restartDeployment(namespace: string, name: string): Promise<{ status: string; message: string }> {
    const res = await apiClient.post(`/api/v1/kubernetes/deployments/${namespace}/${name}/restart`);
    return res.data;
  },

  // scaleDeployment scales deployment replicas
  async scaleDeployment(
    namespace: string,
    name: string,
    replicas: number
  ): Promise<{ status: string; message: string; replicas: number }> {
    const res = await apiClient.post(`/api/v1/kubernetes/deployments/${namespace}/${name}/scale`, {
      replicas,
    });
    return res.data;
  },

  // getNodes retrieves cluster nodes
  async getNodes(): Promise<{ data: NodeSummary[]; total: number }> {
    const res = await apiClient.get("/api/v1/kubernetes/nodes");
    return res.data;
  },

  // getServices retrieves services list
  async getServices(namespace?: string): Promise<{ data: ServiceSummary[]; total: number }> {
    const params = namespace && namespace !== "all" ? { namespace } : {};
    const res = await apiClient.get("/api/v1/kubernetes/services", { params });
    return res.data;
  },

  // getOverview retrieves cluster stats
  async getOverview(): Promise<K8sClusterOverview> {
    const res = await apiClient.get("/api/v1/kubernetes/overview");
    return res.data;
  },
};

import { apiClient } from "../lib/api-client";
import {
  ArgoApplicationSummary,
  ArgoApplicationDetail,
  ArgoSyncRequest,
  ArgoOverview,
} from "../types/argocd";

export const argocdService = {
  // getApplications retrieves argocd application list
  async getApplications(namespace?: string): Promise<{ data: ArgoApplicationSummary[]; total: number }> {
    const params = namespace && namespace !== "all" ? { namespace } : {};
    const res = await apiClient.get("/api/v1/argocd/applications", { params });
    return res.data;
  },

  // getApplication retrieves application detail
  async getApplication(name: string, namespace?: string): Promise<ArgoApplicationDetail> {
    const params = namespace && namespace !== "all" ? { namespace } : {};
    const res = await apiClient.get(`/api/v1/argocd/applications/${name}`, { params });
    return res.data;
  },

  // syncApplication triggers sync operation
  async syncApplication(
    name: string,
    req?: ArgoSyncRequest,
    namespace?: string
  ): Promise<{ status: string; message: string }> {
    const params = namespace && namespace !== "all" ? { namespace } : {};
    const res = await apiClient.post(`/api/v1/argocd/applications/${name}/sync`, req || {}, {
      params,
    });
    return res.data;
  },

  // getOverview retrieves application overview
  async getOverview(namespace?: string): Promise<ArgoOverview> {
    const params = namespace && namespace !== "all" ? { namespace } : {};
    const res = await apiClient.get("/api/v1/argocd/overview", { params });
    return res.data;
  },
};

import { apiClient } from "../lib/api-client";
import {
  IncidentSummary,
  IncidentDetail,
  IncidentStats,
  IncidentFilter,
} from "../types/incident";

export const incidentService = {
  // listIncidents fetches incidents with filter
  async listIncidents(
    filter?: IncidentFilter
  ): Promise<{ data: IncidentSummary[]; total: number; page?: number; limit?: number }> {
    const params: Record<string, string | number> = {};
    if (filter?.status && filter.status !== "all") {
      params.status = filter.status;
    }
    if (filter?.severity && filter.severity !== "all") {
      params.severity = filter.severity;
    }
    if (filter?.source && filter.source !== "all") {
      params.source = filter.source;
    }
    if (filter?.search) {
      params.search = filter.search;
    }
    if (filter?.page) {
      params.page = filter.page;
    }
    if (filter?.limit) {
      params.limit = filter.limit;
    }

    const res = await apiClient.get("/api/v1/incidents", { params });
    return res.data;
  },

  // getIncidentStats fetches aggregated metrics
  async getIncidentStats(): Promise<IncidentStats> {
    const res = await apiClient.get("/api/v1/incidents/stats");
    return res.data.data;
  },

  // getIncident fetches single incident detail
  async getIncident(id: string): Promise<IncidentDetail> {
    const res = await apiClient.get(`/api/v1/incidents/${id}`);
    return res.data.data;
  },

  // acknowledgeIncident acknowledges open incident
  async acknowledgeIncident(id: string): Promise<{ status: string; message: string }> {
    const res = await apiClient.post(`/api/v1/incidents/${id}/acknowledge`);
    return res.data;
  },

  // resolveIncident marks incident resolved
  async resolveIncident(id: string): Promise<{ status: string; message: string }> {
    const res = await apiClient.post(`/api/v1/incidents/${id}/resolve`);
    return res.data;
  },

  // closeIncident closes incident permanently
  async closeIncident(id: string): Promise<{ status: string; message: string }> {
    const res = await apiClient.post(`/api/v1/incidents/${id}/close`);
    return res.data;
  },
};

export default incidentService;

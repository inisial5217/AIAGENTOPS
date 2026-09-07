import { apiClient } from "../lib/api-client";
import {
  ModelInfo,
  AISession,
  AIMessage,
  AIChatRequest,
  AIChatResponse,
  ToolApprovalRequest,
  AIUsageStats,
  RCAResponse,
} from "../types/ai";

export const aiService = {
  // getModels fetches available AI providers and models
  async getModels(): Promise<ModelInfo[]> {
    const res = await apiClient.get("/api/v1/ai/models");
    const d = res.data;
    if (Array.isArray(d)) return d;
    if (d?.data && Array.isArray(d.data)) return d.data;
    if (d?.models && Array.isArray(d.models)) return d.models;
    if (d?.providers && Array.isArray(d.providers)) return d.providers;
    return [];
  },

  // listSessions fetches user AI chat sessions
  async listSessions(): Promise<AISession[]> {
    const res = await apiClient.get("/api/v1/ai/sessions");
    const d = res.data;
    if (Array.isArray(d)) return d;
    if (d?.sessions && Array.isArray(d.sessions)) return d.sessions;
    if (d?.data && Array.isArray(d.data)) return d.data;
    return [];
  },

  // createSession creates a new chat session
  async createSession(title?: string, provider?: string, model?: string): Promise<AISession> {
    const res = await apiClient.post("/api/v1/ai/sessions", { title, provider, model });
    return res.data?.data || res.data;
  },

  // getSessionMessages retrieves history for a session
  async getSessionMessages(sessionId: string): Promise<AIMessage[]> {
    const res = await apiClient.get(`/api/v1/ai/sessions/${sessionId}/messages`);
    const d = res.data;
    if (Array.isArray(d)) return d;
    if (d?.messages && Array.isArray(d.messages)) return d.messages;
    if (d?.data && Array.isArray(d.data)) return d.data;
    return [];
  },

  // sendMessage sends a prompt to AI assistant
  async sendMessage(req: AIChatRequest): Promise<AIChatResponse> {
    const res = await apiClient.post("/api/v1/ai/chat", req);
    return res.data?.data || res.data;
  },

  // approveTool approves or rejects a pending write tool
  async approveTool(req: ToolApprovalRequest): Promise<{ status: string; message: string; result?: string }> {
    const res = await apiClient.post(`/api/v1/ai/tools/${req.tool_call_id}/approve`, {
      session_id: req.session_id,
      approved: req.approved,
      reason: req.reason,
    });
    return res.data?.data || res.data;
  },

  // getUsageStats retrieves AI token and cost metrics
  async getUsageStats(): Promise<AIUsageStats> {
    const res = await apiClient.get("/api/v1/ai/usage");
    return res.data?.data || res.data;
  },

  // generateIncidentRCA triggers AI root cause analysis for an incident
  async generateIncidentRCA(incidentId: string, model?: string): Promise<RCAResponse> {
    const res = await apiClient.post(`/api/v1/incidents/${incidentId}/rca`, { model });
    return res.data?.data || res.data;
  },
};

export default aiService;

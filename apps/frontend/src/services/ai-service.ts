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
    return res.data.data;
  },

  // listSessions fetches user AI chat sessions
  async listSessions(): Promise<AISession[]> {
    const res = await apiClient.get("/api/v1/ai/sessions");
    return res.data.data;
  },

  // createSession creates a new chat session
  async createSession(title?: string, provider?: string, model?: string): Promise<AISession> {
    const res = await apiClient.post("/api/v1/ai/sessions", { title, provider, model });
    return res.data.data;
  },

  // getSessionMessages retrieves history for a session
  async getSessionMessages(sessionId: string): Promise<AIMessage[]> {
    const res = await apiClient.get(`/api/v1/ai/sessions/${sessionId}/messages`);
    return res.data.data;
  },

  // sendMessage sends a prompt to AI assistant
  async sendMessage(req: AIChatRequest): Promise<AIChatResponse> {
    const res = await apiClient.post("/api/v1/ai/chat", req);
    return res.data.data;
  },

  // approveTool approves or rejects a pending write tool
  async approveTool(req: ToolApprovalRequest): Promise<{ status: string; message: string; result?: string }> {
    const res = await apiClient.post(`/api/v1/ai/tools/${req.tool_call_id}/approve`, {
      session_id: req.session_id,
      approved: req.approved,
      reason: req.reason,
    });
    return res.data;
  },

  // getUsageStats retrieves AI token and cost metrics
  async getUsageStats(): Promise<AIUsageStats> {
    const res = await apiClient.get("/api/v1/ai/usage");
    return res.data.data;
  },

  // generateIncidentRCA triggers AI root cause analysis for an incident
  async generateIncidentRCA(incidentId: string, model?: string): Promise<RCAResponse> {
    const res = await apiClient.post(`/api/v1/incidents/${incidentId}/rca`, { model });
    return res.data.data;
  },
};

export default aiService;

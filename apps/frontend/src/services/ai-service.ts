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
  // getModels fetches available AI providers and models with guaranteed normalization
  async getModels(): Promise<ModelInfo[]> {
    try {
      const res = await apiClient.get("/api/v1/ai/models");
      const d = res.data;
      let rawList: any[] = [];
      if (Array.isArray(d)) rawList = d;
      else if (d?.data && Array.isArray(d.data)) rawList = d.data;
      else if (d?.models && Array.isArray(d.models)) rawList = d.models;
      else if (d?.providers && Array.isArray(d.providers)) rawList = d.providers;

      if (!rawList || rawList.length === 0) {
        return [
          {
            id: "gemini-2.0-flash",
            provider: "google",
            model_name: "Gemini 2.0 Flash",
            is_default: true,
            status: "available",
          },
          {
            id: "gpt-4o-mini",
            provider: "openai",
            model_name: "GPT-4o Mini",
            is_default: false,
            status: "available",
          },
          {
            id: "claude-3-5-sonnet",
            provider: "anthropic",
            model_name: "Claude 3.5 Sonnet",
            is_default: false,
            status: "available",
          },
          {
            id: "llama3",
            provider: "ollama",
            model_name: "Llama 3 (Local)",
            is_default: false,
            status: "available",
          },
        ];
      }

      return rawList.map((item: any, idx: number) => {
        const provider = item.provider || item.name || "google";
        const modelName = item.model_name || item.model || item.id || "Gemini 2.0 Flash";
        const id = item.id || item.model || `${provider}-${idx}`;
        const isDefault =
          item.is_default !== undefined
            ? Boolean(item.is_default)
            : Boolean(item.name === d?.active_provider || idx === 0);
        const status = item.circuit_state === "open" ? "degraded" : (item.status || "available");

        const modelObj: ModelInfo = {
          id: String(id),
          provider: String(provider),
          model_name: String(modelName),
          is_default: isDefault,
          status: status,
        };

        if (item.context_window !== undefined) {
          modelObj.context_window = item.context_window;
        }
        if (item.input_cost_per_1k !== undefined) {
          modelObj.input_cost_per_1k = item.input_cost_per_1k;
        }

        return modelObj;
      });
    } catch {
      return [
        {
          id: "gemini-2.0-flash",
          provider: "google",
          model_name: "Gemini 2.0 Flash",
          is_default: true,
          status: "available",
        },
      ];
    }
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

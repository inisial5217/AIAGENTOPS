import { describe, it, expect, vi, beforeEach } from "vitest";
import { aiService } from "./ai-service";
import { apiClient } from "../lib/api-client";

vi.mock("../lib/api-client", () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
  },
}));

describe("aiService", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("should fetch available AI models", async () => {
    const mockModels = [
      { id: "gemini-2.0-flash", provider: "google", model_name: "gemini-2.0-flash", is_default: true, status: "available" },
      { id: "gpt-4o", provider: "openai", model_name: "gpt-4o", is_default: false, status: "available" },
    ];
    vi.mocked(apiClient.get).mockResolvedValueOnce({ data: { data: mockModels } });

    const result = await aiService.getModels();
    expect(apiClient.get).toHaveBeenCalledWith("/api/v1/ai/models");
    expect(result).toEqual(mockModels);
  });

  it("should list user sessions", async () => {
    const mockSessions = [
      { id: "sess-1", user_id: "user-1", title: "K8s Debugging", provider: "google", model: "gemini-2.0-flash", created_at: "", updated_at: "" },
    ];
    vi.mocked(apiClient.get).mockResolvedValueOnce({ data: { data: mockSessions } });

    const result = await aiService.listSessions();
    expect(apiClient.get).toHaveBeenCalledWith("/api/v1/ai/sessions");
    expect(result).toEqual(mockSessions);
  });

  it("should send a chat message", async () => {
    const mockResponse = {
      session_id: "sess-1",
      message_id: "msg-1",
      reply: "All pods in default namespace are healthy.",
      provider: "google",
      model: "gemini-2.0-flash",
      latency_ms: 120,
      cost_usd: 0.0001,
    };
    vi.mocked(apiClient.post).mockResolvedValueOnce({ data: { data: mockResponse } });

    const result = await aiService.sendMessage({
      session_id: "sess-1",
      message: "Check cluster health",
    });

    expect(apiClient.post).toHaveBeenCalledWith("/api/v1/ai/chat", {
      session_id: "sess-1",
      message: "Check cluster health",
    });
    expect(result).toEqual(mockResponse);
  });

  it("should approve or reject tool execution", async () => {
    vi.mocked(apiClient.post).mockResolvedValueOnce({
      data: { status: "success", message: "tool approved and executed", result: "Restarted" },
    });

    const result = await aiService.approveTool({
      session_id: "sess-1",
      tool_call_id: "call-1",
      approved: true,
    });

    expect(apiClient.post).toHaveBeenCalledWith("/api/v1/ai/tools/call-1/approve", {
      session_id: "sess-1",
      approved: true,
      reason: undefined,
    });
    expect(result.status).toBe("success");
  });

  it("should trigger automated incident RCA", async () => {
    const mockRCA = {
      incident_id: "inc-99",
      rca_summary: "OOMKilled due to heap memory spike",
      root_cause: "Container exceeded memory limit of 512MiB",
      impact_analysis: "High latency on payment gateway",
      recommended_actions: ["Increase memory limit to 1GiB"],
      prevention_steps: ["Add memory threshold alert at 80%"],
      confidence_score: 0.95,
      generated_at: "2026-09-07T00:00:00Z",
    };
    vi.mocked(apiClient.post).mockResolvedValueOnce({ data: { data: mockRCA } });

    const result = await aiService.generateIncidentRCA("inc-99");
    expect(apiClient.post).toHaveBeenCalledWith("/api/v1/incidents/inc-99/rca", { model: undefined });
    expect(result).toEqual(mockRCA);
  });
});

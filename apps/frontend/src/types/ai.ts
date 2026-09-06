export type ModelProvider = "google" | "openai" | "anthropic" | "ollama" | "mock";

export interface ModelInfo {
  id: string;
  provider: ModelProvider | string;
  model_name: string;
  is_default: boolean;
  status: "available" | "unavailable" | "degraded";
  context_window?: number;
  input_cost_per_1k?: number;
  output_cost_per_1k?: number;
}

export type ToolApprovalStatus = "pending" | "approved" | "rejected" | "auto_approved" | "executed";

export interface ToolCallInfo {
  id: string;
  name: string;
  parameters: Record<string, any>;
  requires_approval: boolean;
  status: ToolApprovalStatus;
  result?: string;
  execution_time_ms?: number;
}

export interface AIMessage {
  id: string;
  session_id: string;
  role: "user" | "assistant" | "system" | "tool";
  content: string;
  tool_calls?: ToolCallInfo[];
  latency_ms?: number;
  tokens_prompt?: number;
  tokens_completion?: number;
  cost_usd?: number;
  created_at: string;
}

export interface AISession {
  id: string;
  user_id: string;
  title: string;
  provider: string;
  model: string;
  created_at: string;
  updated_at: string;
  messages?: AIMessage[];
}

export interface AIChatRequest {
  session_id?: string;
  message: string;
  provider?: string;
  model?: string;
  context?: Record<string, any>;
}

export interface AIChatResponse {
  session_id: string;
  message_id: string;
  reply: string;
  provider: string;
  model: string;
  tool_calls?: ToolCallInfo[];
  requires_approval?: boolean;
  pending_tool_call?: ToolCallInfo;
  latency_ms: number;
  cost_usd: number;
}

export interface ToolApprovalRequest {
  session_id: string;
  tool_call_id: string;
  approved: boolean;
  reason?: string;
}

export interface AIUsageStats {
  total_sessions: number;
  total_messages: number;
  total_tokens: number;
  total_cost_usd: number;
  provider_breakdown?: Record<string, number>;
}

export interface RCAResponse {
  incident_id: string;
  rca_summary: string;
  root_cause: string;
  impact_analysis: string;
  recommended_actions: string[];
  prevention_steps: string[];
  confidence_score: number;
  generated_at: string;
}

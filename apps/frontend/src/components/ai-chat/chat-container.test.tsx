import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { ChatContainer } from "./chat-container";
import { aiService } from "../../services/ai-service";

vi.mock("../../services/ai-service", () => ({
  aiService: {
    getModels: vi.fn().mockResolvedValue([
      { id: "gemini-2.0-flash", provider: "google", model_name: "Gemini 2.0 Flash", is_default: true, status: "available" }
    ]),
    listSessions: vi.fn().mockResolvedValue([]),
    createSession: vi.fn().mockResolvedValue({ id: "sess-1", title: "New Chat", provider: "google", model: "gemini-2.0-flash" }),
    getSessionMessages: vi.fn().mockResolvedValue([]),
    sendMessage: vi.fn().mockResolvedValue({
      session_id: "sess-1",
      message_id: "msg-1",
      reply: "Ready to assist",
      provider: "google",
      model: "gemini-2.0-flash",
      latency_ms: 100,
      cost_usd: 0.0001,
    }),
    approveTool: vi.fn().mockResolvedValue({ status: "success", message: "executed" }),
  },
}));

describe("ChatContainer Component", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders floating action button initially", () => {
    render(<ChatContainer />);
    const fab = screen.getByTitle("Open AI DevOps Assistant");
    expect(fab).toBeDefined();
  });

  it("opens chat window when FAB is clicked", async () => {
    render(<ChatContainer />);
    const fab = screen.getByTitle("Open AI DevOps Assistant");
    fireEvent.click(fab);

    expect(screen.getByText("CIFO AI Assistant")).toBeDefined();
    expect(screen.getByText("Autonomous DevOps Agent")).toBeDefined();
    expect(screen.getByPlaceholderText("Ask AI DevOps Assistant... (Shift+Enter for newline)")).toBeDefined();
  });

  it("closes chat window when close button is clicked", async () => {
    render(<ChatContainer />);
    const fab = screen.getByTitle("Open AI DevOps Assistant");
    fireEvent.click(fab);

    const closeBtn = screen.getByTitle("Close AI Assistant");
    fireEvent.click(closeBtn);

    expect(screen.queryByText("CIFO AI Assistant")).toBeNull();
    expect(screen.getByTitle("Open AI DevOps Assistant")).toBeDefined();
  });
});

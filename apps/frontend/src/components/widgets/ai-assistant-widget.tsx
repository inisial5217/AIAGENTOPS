"use client";

import * as React from "react";
import { Bot, Send, Sparkles, Loader2, CheckCircle, ShieldAlert, CornerDownLeft } from "lucide-react";
import { Skeleton } from "../ui/skeleton";
import { aiService } from "../../services/ai-service";
import { AIMessage, ToolCallInfo } from "../../types/ai";
import { useAuthStore } from "../../lib/auth";

export function AIAssistantWidget({ isLoading: isDashboardLoading = false }: { isLoading?: boolean }) {
  const user = useAuthStore((s) => s.user);
  const [messages, setMessages] = React.useState<AIMessage[]>([]);
  const [activeSessionId, setActiveSessionId] = React.useState<string | null>(null);
  const [input, setInput] = React.useState("");
  const [isSending, setIsSending] = React.useState(false);
  const [executingToolId, setExecutingToolId] = React.useState<string | null>(null);
  const messagesEndRef = React.useRef<HTMLDivElement>(null);

  // Initialize session and load real chat messages from database
  React.useEffect(() => {
    let isMounted = true;

    async function initSession() {
      try {
        const savedId =
          typeof window !== "undefined"
            ? localStorage.getItem("cifo_active_ai_session_id")
            : null;

        const sessions = await aiService.listSessions();
        if (!isMounted) return;

        let targetId = savedId && sessions?.some((s) => s.id === savedId) ? savedId : null;

        if (!targetId && sessions && sessions.length > 0) {
          targetId = sessions[0].id;
        }

        if (!targetId) {
          const newSession = await aiService.createSession("Monitoring Command Center Session");
          if (newSession && newSession.id) {
            targetId = newSession.id;
          }
        }

        if (targetId) {
          setActiveSessionId(targetId);
          if (typeof window !== "undefined") {
            localStorage.setItem("cifo_active_ai_session_id", targetId);
          }
          const msgs = await aiService.getSessionMessages(targetId);
          if (isMounted && msgs) {
            setMessages(msgs);
          }
        }
      } catch (err) {
        console.warn("Could not load AI messages for widget", err);
      }
    }

    initSession();

    return () => {
      isMounted = false;
    };
  }, [user?.id]);

  // Auto-scroll inside widget
  React.useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages, isSending]);

  const handleSend = async (textToSend?: string) => {
    const queryText = (textToSend || input).trim();
    if (!queryText || isSending) return;

    setInput("");
    setIsSending(true);

    // Optimistic user message
    const tempUserMsg: AIMessage = {
      id: `temp-${Date.now()}`,
      session_id: activeSessionId || "",
      role: "user",
      content: queryText,
      created_at: new Date().toISOString(),
    };
    setMessages((prev) => [...prev, tempUserMsg]);

    try {
      const response = await aiService.sendMessage({
        session_id: activeSessionId || undefined,
        message: queryText,
        provider: "google",
        model_preference: "gemini-2.0-flash",
      });

      if (response.session_id && response.session_id !== activeSessionId) {
        setActiveSessionId(response.session_id);
        if (typeof window !== "undefined") {
          localStorage.setItem("cifo_active_ai_session_id", response.session_id);
        }
      }

      const asstMsg: AIMessage = {
        id: response.message_id || `asst-${Date.now()}`,
        session_id: response.session_id,
        role: "assistant",
        content: response.content,
        tool_calls: response.tool_calls?.map((tc) => ({
          id: tc.id,
          name: tc.name,
          parameters: tc.parameters,
          requires_approval: tc.requires_approval,
          status: tc.status as any,
          result: tc.result,
        })),
        latency_ms: 0,
        tokens_prompt: response.input_tokens,
        tokens_completion: response.output_tokens,
        cost_usd: response.estimated_cost_usd,
        created_at: new Date().toISOString(),
      };

      setMessages((prev) => [...prev.filter((m) => m.id !== tempUserMsg.id), tempUserMsg, asstMsg]);
    } catch (err: any) {
      console.error("Failed sending AI message", err);
      const errMsg: AIMessage = {
        id: `err-${Date.now()}`,
        session_id: activeSessionId || "",
        role: "assistant",
        content: `Error: ${err?.message || "Gagal menghubungi AI service."}`,
        created_at: new Date().toISOString(),
      };
      setMessages((prev) => [...prev, errMsg]);
    } finally {
      setIsSending(false);
    }
  };

  const handleApproveTool = async (toolCallId: string) => {
    setExecutingToolId(toolCallId);
    try {
      const result = await aiService.approveToolCall(toolCallId);
      // update tool call state in messages
      setMessages((prev) =>
        prev.map((msg) => {
          if (!msg.tool_calls) return msg;
          return {
            ...msg,
            tool_calls: msg.tool_calls.map((tc) =>
              tc.id === toolCallId
                ? { ...tc, status: "executed", result: result.result || "Sukses dieksekusi" }
                : tc
            ),
          };
        })
      );
    } catch (err: any) {
      alert(`Gagal mengeksekusi aksi: ${err.message}`);
    } finally {
      setExecutingToolId(null);
    }
  };

  return (
    <div className="rounded-[var(--radius-xl)] bg-[var(--bg-card)] border border-[var(--border-default)] p-5 shadow-lg flex flex-col h-[320px] text-left">
      {/* Header */}
      <div className="flex items-center justify-between pb-3 border-b border-[var(--border-subtle)] shrink-0">
        <div className="flex items-center gap-2">
          <Bot className="w-4 h-4 text-[var(--accent-default)]" />
          <h3 className="text-sm font-semibold text-[var(--text-primary)]">
            AI Autonomous Assistant
          </h3>
        </div>
        <div className="flex items-center gap-1.5 px-2.5 py-0.5 rounded-full bg-emerald-500/10 border border-emerald-500/30 text-[10px] font-mono text-emerald-400">
          <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
          Real-time Engine Active
        </div>
      </div>

      {/* Messages Scroll Area */}
      <div className="mt-3 flex-1 overflow-y-auto space-y-3 pr-1 text-xs font-mono">
        {isDashboardLoading && messages.length === 0 ? (
          <div className="space-y-3 pt-2">
            <Skeleton variant="rectangular" className="h-14 w-4/5 rounded-lg" />
            <Skeleton variant="rectangular" className="h-10 w-3/5 rounded-lg ml-auto" />
            <Skeleton variant="rectangular" className="h-14 w-4/5 rounded-lg" />
          </div>
        ) : messages.length === 0 ? (
          <div className="h-full flex flex-col items-center justify-center text-center text-[var(--text-muted)] gap-2 py-6">
            <Bot className="w-7 h-7 text-cyan-400 animate-pulse" />
            <span className="text-xs font-bold text-[var(--text-primary)]">CIFO SRE AI Assistant Siap</span>
            <span className="text-[11px] text-[var(--text-secondary)]">
              Ketik &quot;hai&quot; atau &quot;deploy dari gitea&quot; untuk menjalankan operasi klaster.
            </span>
          </div>
        ) : (
          messages.map((m) => {
            const isUser = m.role === "user";
            return (
              <div
                key={m.id}
                className={`flex flex-col ${isUser ? "items-end" : "items-start"}`}
              >
                <div
                  className={`max-w-[88%] rounded-lg p-2.5 text-xs leading-relaxed ${
                    isUser
                      ? "bg-cyan-600/20 border border-cyan-500/40 text-cyan-100 shadow-sm"
                      : "bg-[var(--bg-secondary)] border border-[var(--border-subtle)] text-[var(--text-primary)]"
                  }`}
                >
                  {!isUser && (
                    <span className="font-bold text-[var(--accent-default)] block mb-1 text-[11px]">
                      Agent SRE:
                    </span>
                  )}
                  <div className="whitespace-pre-wrap">{m.content}</div>

                  {/* Render Tool Calling Cards if present */}
                  {m.tool_calls && m.tool_calls.length > 0 && (
                    <div className="mt-2.5 space-y-2 border-t border-[var(--border-subtle)] pt-2">
                      {m.tool_calls.map((tc) => {
                        const isPending = tc.status === "pending" || tc.requires_approval;
                        const isExecuted = tc.status === "executed" || tc.status === "approved";
                        return (
                          <div
                            key={tc.id}
                            className="p-2 rounded bg-[var(--bg-card)] border border-cyan-500/30 text-[11px]"
                          >
                            <div className="flex items-center justify-between gap-2">
                              <span className="font-bold text-cyan-400 flex items-center gap-1">
                                <ShieldAlert className="w-3.5 h-3.5 text-amber-400" />
                                {tc.name}
                              </span>
                              <span className="text-[10px] uppercase text-[var(--text-muted)]">
                                {tc.status}
                              </span>
                            </div>
                            <div className="text-[10px] text-[var(--text-secondary)] mt-1 font-mono">
                              Params: {JSON.stringify(tc.parameters)}
                            </div>
                            {isPending && !isExecuted && (
                              <button
                                onClick={() => handleApproveTool(tc.id)}
                                disabled={executingToolId === tc.id}
                                className="mt-2 w-full py-1 rounded bg-cyan-500 hover:bg-cyan-400 text-slate-950 font-bold text-[11px] transition-colors flex items-center justify-center gap-1.5 cursor-pointer"
                              >
                                {executingToolId === tc.id ? (
                                  <>
                                    <Loader2 className="w-3 h-3 animate-spin" />
                                    <span>Mengeksekusi...</span>
                                  </>
                                ) : (
                                  <>
                                    <CheckCircle className="w-3 h-3" />
                                    <span>Approve &amp; Eksekusi</span>
                                  </>
                                )}
                              </button>
                            )}
                            {isExecuted && (
                              <div className="mt-1.5 text-[10px] text-emerald-400 flex items-center gap-1">
                                <CheckCircle className="w-3 h-3" />
                                <span>{tc.result || "Aksi berhasil dieksekusi"}</span>
                              </div>
                            )}
                          </div>
                        );
                      })}
                    </div>
                  )}
                </div>
              </div>
            );
          })
        )}

        {isSending && (
          <div className="flex items-center gap-2 text-[11px] text-[var(--text-muted)] py-1">
            <Loader2 className="w-3.5 h-3.5 animate-spin text-cyan-400" />
            <span>Memproses instruksi dengan LLM orchestrator...</span>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Footer Controls & Quick Prompts */}
      <div className="pt-2 border-t border-[var(--border-subtle)] shrink-0 space-y-2">
        <div className="flex items-center gap-1.5 overflow-x-auto text-[10px] pb-1 no-scrollbar">
          <button
            type="button"
            onClick={() => handleSend("Deploy versi terbaru dari Gitea")}
            disabled={isSending}
            className="inline-flex items-center gap-1 px-2 py-0.5 rounded bg-[var(--bg-secondary)] hover:bg-[var(--bg-hover)] border border-[var(--border-default)] text-[var(--text-secondary)] hover:text-cyan-400 transition-colors whitespace-nowrap cursor-pointer"
          >
            <Sparkles className="w-2.5 h-2.5 text-cyan-400" />
            Deploy Gitea
          </button>
          <button
            type="button"
            onClick={() => handleSend("Cek status pod di namespace default")}
            disabled={isSending}
            className="px-2 py-0.5 rounded bg-[var(--bg-secondary)] hover:bg-[var(--bg-hover)] border border-[var(--border-default)] text-[var(--text-secondary)] hover:text-[var(--text-primary)] transition-colors whitespace-nowrap cursor-pointer"
          >
            Cek Pods K3d
          </button>
          <button
            type="button"
            onClick={() => handleSend("Analisis root cause insiden OOMKilled")}
            disabled={isSending}
            className="px-2 py-0.5 rounded bg-[var(--bg-secondary)] hover:bg-[var(--bg-hover)] border border-[var(--border-default)] text-[var(--text-secondary)] hover:text-[var(--text-primary)] transition-colors whitespace-nowrap cursor-pointer"
          >
            Analisis RCA
          </button>
        </div>

        <form
          onSubmit={(e) => {
            e.preventDefault();
            handleSend();
          }}
          className="flex items-center gap-2"
        >
          <input
            type="text"
            placeholder="Ketik 'hai' atau perintahkan deploy / diagnosa..."
            value={input}
            onChange={(e) => setInput(e.target.value)}
            disabled={isSending}
            className="flex-1 px-3 py-1.5 text-xs bg-[var(--bg-secondary)] border border-[var(--border-default)] rounded-[var(--radius-md)] text-[var(--text-primary)] placeholder:text-[var(--text-muted)] focus:outline-none focus:border-cyan-500 font-mono"
          />
          <button
            type="submit"
            disabled={!input.trim() || isSending}
            className="p-1.5 rounded-[var(--radius-md)] bg-cyan-500 hover:bg-cyan-400 text-slate-950 font-bold transition-colors cursor-pointer disabled:opacity-50"
          >
            <Send className="w-3.5 h-3.5" />
          </button>
        </form>
      </div>
    </div>
  );
}

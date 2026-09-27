"use client";

import React, { useState, useEffect, useRef } from "react";
import {
  Bot,
  Sparkles,
  X,
  Plus,
  MessageSquare,
  ChevronDown,
  Minimize2,
  Maximize2,
  AlertCircle,
  History,
  Trash2,
  Clock,
  ArrowLeft,
} from "lucide-react";
import { ModelInfo, AISession, AIMessage, ToolCallInfo } from "../../types/ai";
import { aiService } from "../../services/ai-service";
import { useAuthStore } from "../../lib/auth";
import { ModelIndicator } from "./model-indicator";
import { ChatBubble } from "./chat-bubble";
import { ChatInput } from "./chat-input";

function formatRelativeTime(dateStr?: string): string {
  if (!dateStr) return "";
  try {
    const d = new Date(dateStr);
    const now = new Date();
    const diffMs = now.getTime() - d.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffMins < 1) return "Just now";
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    return d.toLocaleDateString(undefined, { month: "short", day: "numeric" });
  } catch {
    return "";
  }
}

export const ChatContainer: React.FC = () => {
  const [isOpen, setIsOpen] = useState(false);
  const [isExpanded, setIsExpanded] = useState(false);
  const [showHistory, setShowHistory] = useState(false);
  const [models, setModels] = useState<ModelInfo[]>([]);
  const [selectedModel, setSelectedModel] = useState<string>("gemini-2.0-flash");
  const [selectedProvider, setSelectedProvider] = useState<string>("google");
  const [sessions, setSessions] = useState<AISession[]>([]);
  const [activeSessionId, setActiveSessionId] = useState<string | null>(null);
  const [messages, setMessages] = useState<AIMessage[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const user = useAuthStore((s) => s.user);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // Load models on mount
  useEffect(() => {
    async function loadModels() {
      try {
        const fetched = await aiService.getModels();
        if (fetched && fetched.length > 0) {
          setModels(fetched);
          const def = fetched.find((m) => m.is_default) || fetched[0];
          setSelectedModel(def.model_name || def.id);
          setSelectedProvider(def.provider);
        }
      } catch (err) {
        console.warn("Could not load AI models, using default provider", err);
      }
    }
    loadModels();
  }, []);

  // Reload sessions on mount or when user changes
  useEffect(() => {
    loadSessions();
  }, [user?.id]);

  // Load messages when active session changes and sync to localStorage
  useEffect(() => {
    if (activeSessionId) {
      if (typeof window !== "undefined") {
        localStorage.setItem("cifo_active_ai_session_id", activeSessionId);
      }
      loadSessionMessages(activeSessionId);
    } else {
      setMessages([]);
      if (typeof window !== "undefined") {
        localStorage.removeItem("cifo_active_ai_session_id");
      }
    }
  }, [activeSessionId]);

  // Auto-scroll on new messages
  useEffect(() => {
    if (!showHistory) {
      messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }
  }, [messages, isLoading, showHistory]);

  const loadSessions = async () => {
    try {
      const data = await aiService.listSessions();
      setSessions(data || []);
      const savedId =
        typeof window !== "undefined"
          ? localStorage.getItem("cifo_active_ai_session_id")
          : null;
      if (savedId && data && data.some((s) => s.id === savedId)) {
        setActiveSessionId(savedId);
      } else if (!activeSessionId && data && data.length > 0) {
        setActiveSessionId(data[0].id);
      }
    } catch (err) {
      console.warn("Failed to load AI sessions", err);
    }
  };

  const loadSessionMessages = async (sessionId: string) => {
    try {
      const data = await aiService.getSessionMessages(sessionId);
      setMessages(data || []);
    } catch (err) {
      console.warn("Failed to load messages", err);
    }
  };

  const handleStartNewChat = async () => {
    try {
      const newSession = await aiService.createSession(
        "New Conversation",
        selectedProvider,
        selectedModel
      );
      if (newSession && newSession.id) {
        setSessions((prev) => [newSession, ...prev.filter((s) => s.id !== newSession.id)]);
        setActiveSessionId(newSession.id);
      } else {
        setActiveSessionId(null);
      }
      setMessages([]);
      setShowHistory(false);
      setError(null);
    } catch (err) {
      console.error("Failed to create new session", err);
      setActiveSessionId(null);
      setMessages([]);
      setShowHistory(false);
    }
  };

  const handleDeleteSession = async (sessionId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await aiService.deleteSession(sessionId);
      setSessions((prev) => prev.filter((s) => s.id !== sessionId));
      if (activeSessionId === sessionId) {
        const remaining = sessions.filter((s) => s.id !== sessionId);
        if (remaining.length > 0) {
          setActiveSessionId(remaining[0].id);
        } else {
          setActiveSessionId(null);
          setMessages([]);
        }
      }
    } catch (err) {
      console.error("Failed to delete session", err);
      setError("Failed to delete chat session");
    }
  };

  const handleSelectSession = (sessionId: string) => {
    setActiveSessionId(sessionId);
    setShowHistory(false);
  };

  const handleSelectModel = (model: ModelInfo) => {
    setSelectedModel(model.model_name || model.id);
    setSelectedProvider(model.provider);
  };

  const handleSendMessage = async (text: string) => {
    if (!text.trim()) return;

    setError(null);
    setIsLoading(true);

    const tempUserMsg: AIMessage = {
      id: `temp-${Date.now()}`,
      session_id: activeSessionId || "",
      role: "user",
      content: text,
      created_at: new Date().toISOString(),
    };

    setMessages((prev) => [...prev, tempUserMsg]);

    try {
      const res = await aiService.sendMessage({
        session_id: activeSessionId || undefined,
        message: text,
        provider: selectedProvider,
        model: selectedModel,
      });

      if (res.session_id) {
        if (!activeSessionId) {
          setActiveSessionId(res.session_id);
        }
        loadSessions();
      }

      const assistantMsg: AIMessage = {
        id: res.message_id,
        session_id: res.session_id,
        role: "assistant",
        content: (res as any).reply || (res as any).content || "",
        tool_calls: res.tool_calls || (res.pending_tool_call ? [res.pending_tool_call] : undefined),
        latency_ms: res.latency_ms,
        cost_usd: res.cost_usd,
        created_at: new Date().toISOString(),
      };

      setMessages((prev) => [...prev, assistantMsg]);
    } catch (err: any) {
      const errMsg =
        err?.response?.data?.message || err?.message || "Failed to communicate with AI service";
      setError(errMsg);
    } finally {
      setIsLoading(false);
    }
  };

  const handleApproveTool = async (toolCallId: string) => {
    if (!activeSessionId) return;

    try {
      const res = await aiService.approveTool({
        session_id: activeSessionId,
        tool_call_id: toolCallId,
        approved: true,
      });

      // Update local message tool calls
      setMessages((prev) =>
        prev.map((msg) => {
          if (!msg.tool_calls) return msg;
          return {
            ...msg,
            tool_calls: msg.tool_calls.map((tc) =>
              tc.id === toolCallId
                ? { ...tc, status: "executed", result: res.result || res.message }
                : tc
            ),
          };
        })
      );
    } catch (err: any) {
      console.error("Tool approval failed", err);
      setError(err?.response?.data?.message || "Failed to execute approved action");
    }
  };

  const handleRejectTool = async (toolCallId: string, reason?: string) => {
    if (!activeSessionId) return;

    try {
      await aiService.approveTool({
        session_id: activeSessionId,
        tool_call_id: toolCallId,
        approved: false,
        reason: reason || "User rejected execution",
      });

      setMessages((prev) =>
        prev.map((msg) => {
          if (!msg.tool_calls) return msg;
          return {
            ...msg,
            tool_calls: msg.tool_calls.map((tc) =>
              tc.id === toolCallId
                ? { ...tc, status: "rejected", result: `Rejected: ${reason || "User rejected"}` }
                : tc
            ),
          };
        })
      );
    } catch (err: any) {
      console.error("Tool rejection failed", err);
    }
  };

  return (
    <>
      {/* Floating Action Button */}
      {!isOpen && (
        <button
          type="button"
          id="open-ai-chat-fab"
          onClick={() => setIsOpen(true)}
          className="fixed bottom-6 right-6 z-50 h-14 w-14 rounded-full bg-gradient-to-tr from-blue-600 to-indigo-500 hover:from-blue-500 hover:to-indigo-400 text-white shadow-2xl shadow-blue-500/40 flex items-center justify-center transition-transform hover:scale-105 active:scale-95 group focus:outline-none focus:ring-2 focus:ring-blue-400"
          title="Open AI DevOps Assistant"
        >
          <span className="absolute -top-1 -right-1 flex h-3.5 w-3.5">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
            <span className="relative inline-flex rounded-full h-3.5 w-3.5 bg-emerald-500"></span>
          </span>
          <Sparkles className="h-6 w-6 text-white group-hover:rotate-12 transition-transform duration-300" />
        </button>
      )}

      {/* Floating Chat Window */}
      {isOpen && (
        <div
          id="ai-chat-drawer"
          className={`fixed z-50 bg-slate-900/95 backdrop-blur-2xl border border-slate-700/80 rounded-2xl shadow-2xl flex flex-col overflow-hidden transition-all duration-300 ease-out ${
            isExpanded
              ? "inset-4 md:inset-10 w-auto h-auto"
              : "bottom-6 right-6 w-[440px] max-w-[calc(100vw-2rem)] h-[650px] max-h-[calc(100vh-4rem)]"
          }`}
        >
          {/* Header */}
          <div className="bg-slate-900 border-b border-slate-800 px-4 py-3 flex items-center justify-between gap-2 shrink-0">
            <div className="flex items-center gap-2.5">
              <div className="h-8 w-8 rounded-lg bg-gradient-to-tr from-blue-600 to-indigo-500 flex items-center justify-center text-white shadow-md">
                <Bot className="h-4 w-4" />
              </div>
              <div>
                <div className="flex items-center gap-1.5">
                  <h3 className="font-semibold text-sm text-white tracking-tight">
                    CIFO AI Assistant
                  </h3>
                  <span className="h-2 w-2 rounded-full bg-emerald-500"></span>
                </div>
                <p className="text-[10px] text-slate-400">Autonomous DevOps Agent</p>
              </div>
            </div>

            <div className="flex items-center gap-1.5">
              <ModelIndicator
                models={models}
                selectedModel={selectedModel}
                selectedProvider={selectedProvider}
                onSelectModel={handleSelectModel}
                isLoading={isLoading}
              />

              <button
                type="button"
                id="ai-chat-history-btn"
                onClick={() => setShowHistory(!showHistory)}
                title={showHistory ? "Back to Chat" : "View Chat History"}
                className={`p-1.5 rounded-lg transition-colors ${
                  showHistory
                    ? "text-blue-400 bg-blue-950/60 border border-blue-800/60"
                    : "text-slate-400 hover:text-white hover:bg-slate-800"
                }`}
              >
                <History className="h-4 w-4" />
              </button>

              <button
                type="button"
                id="ai-chat-new-btn"
                onClick={handleStartNewChat}
                title="New Chat Session"
                className="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors"
              >
                <Plus className="h-4 w-4" />
              </button>

              <button
                type="button"
                id="ai-chat-expand-btn"
                onClick={() => setIsExpanded(!isExpanded)}
                title={isExpanded ? "Collapse" : "Expand"}
                className="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors hidden sm:block"
              >
                {isExpanded ? <Minimize2 className="h-4 w-4" /> : <Maximize2 className="h-4 w-4" />}
              </button>

              <button
                type="button"
                id="close-ai-chat-btn"
                onClick={() => setIsOpen(false)}
                title="Close AI Assistant"
                className="p-1.5 rounded-lg text-slate-400 hover:text-rose-400 hover:bg-slate-800 transition-colors"
              >
                <X className="h-4 w-4" />
              </button>
            </div>
          </div>

          {/* Error Banner */}
          {error && (
            <div className="bg-rose-950/40 border-b border-rose-800/40 px-3.5 py-2 flex items-center justify-between text-xs text-rose-300">
              <div className="flex items-center gap-2">
                <AlertCircle className="h-4 w-4 text-rose-400 shrink-0" />
                <span>{error}</span>
              </div>
              <button
                type="button"
                onClick={() => setError(null)}
                className="text-rose-400 hover:text-rose-200"
              >
                <X className="h-3.5 w-3.5" />
              </button>
            </div>
          )}

          {/* Body: Either History Panel or Active Chat */}
          {showHistory ? (
            <div className="flex-1 flex flex-col bg-slate-900/90 overflow-hidden">
              {/* History Top Bar */}
              <div className="px-4 py-2.5 border-b border-slate-800/80 flex items-center justify-between bg-slate-950/40 shrink-0">
                <div className="flex items-center gap-2">
                  <button
                    type="button"
                    onClick={() => setShowHistory(false)}
                    className="p-1 rounded-md text-slate-400 hover:text-white hover:bg-slate-800"
                    title="Back to conversation"
                  >
                    <ArrowLeft className="h-3.5 w-3.5" />
                  </button>
                  <span className="text-xs font-semibold text-slate-200 uppercase tracking-wider">
                    Chat History
                  </span>
                  {user?.role && (
                    <span className="text-[10px] px-1.5 py-0.5 rounded bg-blue-950/40 border border-blue-800/50 text-blue-300 font-mono">
                      {user.role}
                    </span>
                  )}
                </div>
                <button
                  type="button"
                  onClick={handleStartNewChat}
                  className="flex items-center gap-1 px-2.5 py-1 rounded-lg text-xs font-medium bg-blue-600 hover:bg-blue-500 text-white shadow-sm transition-colors"
                >
                  <Plus className="h-3.5 w-3.5" />
                  <span>New Chat</span>
                </button>
              </div>

              {/* History Sessions List */}
              <div className="flex-1 overflow-y-auto p-3 space-y-2 scrollbar-thin">
                {sessions.length === 0 ? (
                  <div className="h-full flex flex-col items-center justify-center text-center p-6 text-slate-400">
                    <MessageSquare className="h-8 w-8 text-slate-600 mb-2" />
                    <p className="text-xs text-slate-300 font-medium">No previous chat sessions</p>
                    <p className="text-[11px] text-slate-500 mt-1 max-w-xs">
                      Conversations are stored per user account and retained across sessions.
                    </p>
                    <button
                      type="button"
                      onClick={handleStartNewChat}
                      className="mt-4 px-3 py-1.5 text-xs rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 flex items-center gap-1.5"
                    >
                      <Plus className="h-3.5 w-3.5" />
                      <span>Start First Conversation</span>
                    </button>
                  </div>
                ) : (
                  sessions.map((sess) => {
                    const isActive = sess.id === activeSessionId;
                    return (
                      <div
                        key={sess.id}
                        onClick={() => handleSelectSession(sess.id)}
                        className={`group relative flex items-center justify-between p-3 rounded-xl border cursor-pointer transition-all ${
                          isActive
                            ? "bg-blue-950/40 border-blue-600/60 shadow-md shadow-blue-950/30"
                            : "bg-slate-800/40 hover:bg-slate-800/80 border-slate-700/40 hover:border-slate-600"
                        }`}
                      >
                        <div className="flex-1 min-w-0 pr-3">
                          <div className="flex items-center gap-1.5">
                            <MessageSquare
                              className={`h-3.5 w-3.5 shrink-0 ${
                                isActive ? "text-blue-400" : "text-slate-400"
                              }`}
                            />
                            <h4
                              className={`text-xs font-medium truncate ${
                                isActive ? "text-blue-200 font-semibold" : "text-slate-200"
                              }`}
                              title={sess.title || "Conversation"}
                            >
                              {sess.title || "Conversation"}
                            </h4>
                          </div>
                          <div className="flex items-center gap-2 mt-1 text-[10px] text-slate-400">
                            <span className="flex items-center gap-1">
                              <Clock className="h-3 w-3" />
                              {formatRelativeTime(sess.updated_at || sess.created_at)}
                            </span>
                            {sess.model && (
                              <span className="px-1.5 py-0.5 rounded bg-slate-900/80 text-slate-400 border border-slate-800 font-mono text-[9px]">
                                {sess.model}
                              </span>
                            )}
                          </div>
                        </div>

                        <button
                          type="button"
                          onClick={(e) => handleDeleteSession(sess.id, e)}
                          title="Delete Conversation"
                          className="opacity-0 group-hover:opacity-100 p-1.5 rounded-lg text-slate-400 hover:text-rose-400 hover:bg-slate-700/80 transition-all"
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </button>
                      </div>
                    );
                  })
                )}
              </div>
            </div>
          ) : (
            <>
              {/* Messages Stream */}
              <div className="flex-1 overflow-y-auto p-4 space-y-2 scrollbar-thin">
                {messages.length === 0 ? (
                  <div className="h-full flex flex-col items-center justify-center text-center p-6 text-slate-400 space-y-3">
                    <div className="h-12 w-12 rounded-2xl bg-blue-500/10 border border-blue-500/20 flex items-center justify-center text-blue-400 mb-1">
                      <Sparkles className="h-6 w-6" />
                    </div>
                    <h4 className="font-semibold text-slate-200 text-sm">
                      How can I assist your infrastructure today?
                    </h4>
                    <p className="text-xs text-slate-400 max-w-xs leading-relaxed">
                      I can inspect Kubernetes pods, query Docker containers, diagnose incident alerts,
                      scale deployments, and trigger ArgoCD syncs with Human-in-the-Loop protection.
                    </p>
                  </div>
                ) : (
                  messages.map((msg) => (
                    <ChatBubble
                      key={msg.id}
                      message={msg}
                      onApproveTool={handleApproveTool}
                      onRejectTool={handleRejectTool}
                    />
                  ))
                )}

                {isLoading && (
                  <div className="flex items-center gap-2 text-xs text-slate-400 p-2 bg-slate-800/40 rounded-xl w-fit border border-slate-700/30">
                    <span className="flex h-2 w-2 relative">
                      <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75"></span>
                      <span className="relative inline-flex rounded-full h-2 w-2 bg-blue-500"></span>
                    </span>
                    <span>AI Agent is analyzing infrastructure...</span>
                  </div>
                )}

                <div ref={messagesEndRef} />
              </div>

              {/* Footer Input */}
              <ChatInput
                onSendMessage={handleSendMessage}
                isLoading={isLoading}
              />
            </>
          )}
        </div>
      )}
    </>
  );
};

export default ChatContainer;


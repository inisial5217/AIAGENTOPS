"use client";

import * as React from "react";
import { Bot, Send, Sparkles } from "lucide-react";
import { Skeleton } from "../ui/skeleton";

interface ChatMessage {
  id: string;
  sender: "agent" | "user";
  author: string;
  text: string;
}

const initialMessages: ChatMessage[] = [
  {
    id: "m-1",
    sender: "agent",
    author: "Agent SRE:",
    text: "Halo! Ada lonjakan CPU di auth-service dan payment-gateway terdeteksi DOWN. Ingin saya jalankan diagnosa mendalam?",
  },
  {
    id: "m-2",
    sender: "user",
    author: "You:",
    text: "Ya, tolong cek log dari payment-gateway.",
  },
  {
    id: "m-3",
    sender: "agent",
    author: "Agent SRE:",
    text: "Menganalisis log payment-gateway... Ditemukan 14 insiden CONNECTION TIMEOUT ke database PostgreSQL. Rekomendasi: lakukan restart pod dan periksa connection pool.",
  },
];

export function AIAssistantWidget({ isLoading = false }: { isLoading?: boolean }) {
  const [messages, setMessages] = React.useState<ChatMessage[]>(initialMessages);
  const [input, setInput] = React.useState("");

  const handleSend = (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    if (!input.trim()) return;

    const userMsg: ChatMessage = {
      id: `m-${Date.now()}`,
      sender: "user",
      author: "You:",
      text: input,
    };
    setMessages((prev) => [...prev, userMsg]);
    setInput("");

    // simulate autonomous response
    setTimeout(() => {
      const agentMsg: ChatMessage = {
        id: `m-${Date.now() + 1}`,
        sender: "agent",
        author: "Agent SRE:",
        text: `Memproses instruksi: "${userMsg.text}". Diagnosa telemetri sedang dijalankan pada cluster.`,
      };
      setMessages((prev) => [...prev, agentMsg]);
    }, 600);
  };

  const handleQuickPrompt = (prompt: string) => {
    setInput(prompt);
  };

  return (
    <div className="rounded-[var(--radius-xl)] bg-[var(--bg-card)] border border-[var(--border-default)] p-5 shadow-lg flex flex-col h-[320px] text-left">
      <div className="flex items-center justify-between pb-3 border-b border-[var(--border-subtle)] shrink-0">
        <div className="flex items-center gap-2">
          <Bot className="w-4 h-4 text-[var(--accent-default)]" />
          <h3 className="text-sm font-semibold text-[var(--text-primary)]">
            AI Autonomous Assistant
          </h3>
        </div>
        <div className="flex items-center gap-1.5 px-2.5 py-0.5 rounded-full bg-emerald-500/10 border border-emerald-500/30 text-[10px] font-mono text-emerald-400">
          <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
          Gemini 1.5 Active
        </div>
      </div>

      <div className="mt-3 flex-1 overflow-y-auto space-y-3 pr-1">
        {isLoading ? (
          <div className="space-y-3 pt-2">
            <Skeleton variant="rectangular" className="h-16 w-4/5 rounded-lg" />
            <Skeleton variant="rectangular" className="h-10 w-3/5 rounded-lg ml-auto" />
            <Skeleton variant="rectangular" className="h-14 w-4/5 rounded-lg" />
          </div>
        ) : (
          messages.map((m) => (
            <div
              key={m.id}
              className={`flex flex-col ${
                m.sender === "user" ? "items-end" : "items-start"
              }`}
            >
              <div
                className={`max-w-[85%] rounded-lg p-3 text-xs leading-relaxed ${
                  m.sender === "user"
                    ? "bg-[var(--accent-active)] text-white shadow-md"
                    : "bg-[var(--bg-secondary)] border border-[var(--border-subtle)] text-[var(--text-primary)]"
                }`}
              >
                {m.sender === "agent" && (
                  <span className="font-bold text-[var(--accent-default)] block mb-1">
                    {m.author}
                  </span>
                )}
                {m.text}
              </div>
            </div>
          ))
        )}
      </div>

      <div className="pt-2 border-t border-[var(--border-subtle)] shrink-0 space-y-2">
        <div className="flex items-center gap-1.5 overflow-x-auto text-[10px] pb-1 no-scrollbar">
          <button
            type="button"
            onClick={() => handleQuickPrompt("Jalankan Quick Fix pada payment-gateway")}
            className="inline-flex items-center gap-1 px-2 py-0.5 rounded bg-[var(--bg-secondary)] hover:bg-[var(--bg-hover)] border border-[var(--border-default)] text-[var(--text-secondary)] hover:text-[var(--accent-default)] transition-colors whitespace-nowrap cursor-pointer"
          >
            <Sparkles className="w-2.5 h-2.5 text-[var(--accent-default)]" />
            Quick Fix
          </button>
          <button
            type="button"
            onClick={() => handleQuickPrompt("Tampilkan pod crashlooping")}
            className="px-2 py-0.5 rounded bg-[var(--bg-secondary)] hover:bg-[var(--bg-hover)] border border-[var(--border-default)] text-[var(--text-secondary)] hover:text-[var(--text-primary)] transition-colors whitespace-nowrap cursor-pointer"
          >
            Cek Pods
          </button>
          <button
            type="button"
            onClick={() => handleQuickPrompt("Analisis CPU spike di node-worker-1")}
            className="px-2 py-0.5 rounded bg-[var(--bg-secondary)] hover:bg-[var(--bg-hover)] border border-[var(--border-default)] text-[var(--text-secondary)] hover:text-[var(--text-primary)] transition-colors whitespace-nowrap cursor-pointer"
          >
            Analisis Node
          </button>
        </div>

        <form onSubmit={handleSend} className="flex items-center gap-2">
          <input
            type="text"
            placeholder="Ask AI Assistant to diagnose or fix..."
            value={input}
            onChange={(e) => setInput(e.target.value)}
            className="flex-1 px-3 py-1.5 text-xs bg-[var(--bg-secondary)] border border-[var(--border-default)] rounded-[var(--radius-md)] text-[var(--text-primary)] placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent-default)]"
          />
          <button
            type="submit"
            className="p-1.5 rounded-[var(--radius-md)] bg-[var(--accent-default)] text-[var(--text-inverse)] hover:bg-[var(--accent-hover)] transition-colors cursor-pointer"
          >
            <Send className="w-3.5 h-3.5" />
          </button>
        </form>
      </div>
    </div>
  );
}

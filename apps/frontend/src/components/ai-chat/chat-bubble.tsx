"use client";

import React from "react";
import { Bot, User, Wrench, Clock, Coins } from "lucide-react";
import { AIMessage } from "../../types/ai";
import { ToolApproval } from "./tool-approval";

interface ChatBubbleProps {
  message: AIMessage;
  onApproveTool: (toolCallId: string) => Promise<void>;
  onRejectTool: (toolCallId: string, reason?: string) => Promise<void>;
}

export const ChatBubble: React.FC<ChatBubbleProps> = ({
  message,
  onApproveTool,
  onRejectTool,
}) => {
  const isUser = message.role === "user";
  const isTool = message.role === "tool";

  // Helper to render content with markdown-like codeblocks and bold
  const renderFormattedContent = (content: string) => {
    if (!content) return null;

    // Check for code blocks
    const codeBlockRegex = /```([a-zA-Z0-9_-]*)\n([\s\S]*?)```/g;
    const parts = [];
    let lastIdx = 0;
    let match;

    while ((match = codeBlockRegex.exec(content)) !== null) {
      if (match.index > lastIdx) {
        parts.push(
          <span key={lastIdx} className="whitespace-pre-wrap leading-relaxed">
            {content.substring(lastIdx, match.index)}
          </span>
        );
      }

      const lang = match[1] || "text";
      const code = match[2];
      parts.push(
        <div key={match.index} className="my-2.5 rounded-lg overflow-hidden border border-slate-700 bg-slate-950 font-mono text-xs shadow-md">
          <div className="bg-slate-800/80 px-3 py-1 text-[10px] text-slate-400 font-semibold uppercase flex justify-between">
            <span>{lang}</span>
            <span>Code</span>
          </div>
          <pre className="p-3 text-cyan-300 overflow-x-auto">
            <code>{code}</code>
          </pre>
        </div>
      );
      lastIdx = match.index + match[0].length;
    }

    if (lastIdx < content.length) {
      parts.push(
        <span key={lastIdx} className="whitespace-pre-wrap leading-relaxed">
          {content.substring(lastIdx)}
        </span>
      );
    }

    return parts.length > 0 ? parts : <span className="whitespace-pre-wrap">{content}</span>;
  };

  return (
    <div
      className={`flex gap-3 my-3 text-xs md:text-sm ${
        isUser ? "flex-row-reverse" : "flex-row"
      }`}
    >
      {/* Avatar */}
      <div
        className={`h-8 w-8 rounded-xl flex items-center justify-center shrink-0 shadow-md ${
          isUser
            ? "bg-blue-600 text-white"
            : isTool
            ? "bg-purple-600/30 text-purple-300 border border-purple-500/40"
            : "bg-gradient-to-br from-indigo-500 to-purple-600 text-white"
        }`}
      >
        {isUser ? <User className="h-4 w-4" /> : isTool ? <Wrench className="h-4 w-4" /> : <Bot className="h-4 w-4" />}
      </div>

      {/* Bubble Content */}
      <div
        className={`max-w-[85%] rounded-2xl p-3.5 shadow-sm transition-all ${
          isUser
            ? "bg-blue-600 text-white rounded-tr-none"
            : "bg-slate-800/90 text-slate-200 border border-slate-700/60 rounded-tl-none backdrop-blur-sm"
        }`}
      >
        {/* Author / Role Label */}
        <div className="flex items-center justify-between gap-3 mb-1.5 pb-1 border-b border-slate-700/40">
          <span className="font-semibold text-[11px] uppercase tracking-wider text-slate-300">
            {isUser ? "You" : isTool ? "Tool Result" : "CIFO Assistant"}
          </span>

          <div className="flex items-center gap-2 text-[10px] text-slate-400">
            {message.latency_ms ? (
              <span className="flex items-center gap-0.5">
                <Clock className="h-2.5 w-2.5" />
                {message.latency_ms}ms
              </span>
            ) : null}
            {message.cost_usd ? (
              <span className="flex items-center gap-0.5 text-emerald-400">
                <Coins className="h-2.5 w-2.5" />${message.cost_usd.toFixed(4)}
              </span>
            ) : null}
          </div>
        </div>

        {/* Message Body */}
        <div className="leading-relaxed break-words text-slate-100">
          {renderFormattedContent(message.content)}
        </div>

        {/* Tool Approval Cards if present */}
        {message.tool_calls && message.tool_calls.length > 0 && (
          <div className="mt-3 space-y-2">
            {message.tool_calls.map((toolCall) => (
              <ToolApproval
                key={toolCall.id}
                toolCall={toolCall}
                onApprove={onApproveTool}
                onReject={onRejectTool}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

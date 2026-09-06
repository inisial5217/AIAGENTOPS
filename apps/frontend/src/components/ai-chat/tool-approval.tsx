"use client";

import React, { useState } from "react";
import { AlertTriangle, CheckCircle, XCircle, ShieldAlert, Loader2, ArrowRight } from "lucide-react";
import { ToolCallInfo } from "../../types/ai";

interface ToolApprovalProps {
  toolCall: ToolCallInfo;
  onApprove: (toolCallId: string) => Promise<void>;
  onReject: (toolCallId: string, reason?: string) => Promise<void>;
}

export const ToolApproval: React.FC<ToolApprovalProps> = ({
  toolCall,
  onApprove,
  onReject,
}) => {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showRejectInput, setShowRejectInput] = useState(false);
  const [rejectReason, setRejectReason] = useState("");

  const handleApprove = async () => {
    setIsSubmitting(true);
    try {
      await onApprove(toolCall.id);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleReject = async () => {
    setIsSubmitting(true);
    try {
      await onReject(toolCall.id, rejectReason || "Rejected by user");
    } finally {
      setIsSubmitting(false);
      setShowRejectInput(false);
    }
  };

  const isPending = toolCall.status === "pending";

  return (
    <div
      id={`tool-approval-card-${toolCall.id}`}
      className={`rounded-xl border p-4 my-3 text-xs transition-all ${
        toolCall.status === "approved" || toolCall.status === "executed"
          ? "bg-emerald-950/20 border-emerald-800/40 text-slate-200"
          : toolCall.status === "rejected"
          ? "bg-rose-950/20 border-rose-800/40 text-slate-200"
          : "bg-amber-950/20 border-amber-600/40 text-slate-200 shadow-lg shadow-amber-950/10"
      }`}
    >
      <div className="flex items-center justify-between gap-2 mb-2.5">
        <div className="flex items-center gap-2">
          {toolCall.status === "pending" ? (
            <ShieldAlert className="h-4 w-4 text-amber-400" />
          ) : toolCall.status === "approved" || toolCall.status === "executed" ? (
            <CheckCircle className="h-4 w-4 text-emerald-400" />
          ) : (
            <XCircle className="h-4 w-4 text-rose-400" />
          )}
          <span className="font-semibold text-sm tracking-tight text-white">
            Action: <code className="text-amber-300 font-mono text-xs">{toolCall.name}</code>
          </span>
        </div>

        <span
          className={`px-2 py-0.5 rounded-full text-[10px] font-semibold uppercase tracking-wider ${
            toolCall.status === "pending"
              ? "bg-amber-500/20 text-amber-300 border border-amber-500/30"
              : toolCall.status === "approved" || toolCall.status === "executed"
              ? "bg-emerald-500/20 text-emerald-300 border border-emerald-500/30"
              : "bg-rose-500/20 text-rose-300 border border-rose-500/30"
          }`}
        >
          {toolCall.status}
        </span>
      </div>

      <div className="bg-slate-900/80 rounded-lg p-2.5 border border-slate-800 my-2">
        <div className="text-[10px] uppercase font-bold text-slate-400 mb-1 tracking-wider">
          Target Parameters
        </div>
        <div className="space-y-1 font-mono text-[11px] text-slate-300">
          {Object.entries(toolCall.parameters || {}).map(([key, val]) => (
            <div key={key} className="flex items-center justify-between">
              <span className="text-slate-400">{key}:</span>
              <span className="text-cyan-300 font-medium">{String(val)}</span>
            </div>
          ))}
        </div>
      </div>

      {toolCall.result && (
        <div className="bg-slate-900/90 rounded-lg p-2.5 border border-slate-700/60 my-2 font-mono text-[11px]">
          <div className="text-[10px] uppercase font-bold text-slate-400 mb-1 tracking-wider">
            Execution Result
          </div>
          <p className="text-emerald-300 break-words whitespace-pre-wrap">{toolCall.result}</p>
        </div>
      )}

      {isPending && (
        <div className="mt-3 pt-2 border-t border-amber-800/30">
          <div className="flex items-center gap-1.5 text-amber-300 text-[11px] mb-2.5">
            <AlertTriangle className="h-3.5 w-3.5 shrink-0" />
            <span>This write operation modifies cluster state. User confirmation required.</span>
          </div>

          {showRejectInput ? (
            <div className="space-y-2">
              <input
                type="text"
                value={rejectReason}
                onChange={(e) => setRejectReason(e.target.value)}
                placeholder="Reason for rejection (optional)..."
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-2.5 py-1.5 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-rose-500"
              />
              <div className="flex items-center justify-end gap-2">
                <button
                  type="button"
                  onClick={() => setShowRejectInput(false)}
                  className="px-2.5 py-1 rounded-lg text-slate-400 hover:text-white"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  onClick={handleReject}
                  disabled={isSubmitting}
                  className="px-3 py-1.5 bg-rose-600 hover:bg-rose-500 text-white rounded-lg font-medium flex items-center gap-1"
                >
                  {isSubmitting ? <Loader2 className="h-3 w-3 animate-spin" /> : null}
                  Confirm Reject
                </button>
              </div>
            </div>
          ) : (
            <div className="flex items-center justify-end gap-2">
              <button
                type="button"
                id={`reject-tool-btn-${toolCall.id}`}
                disabled={isSubmitting}
                onClick={() => setShowRejectInput(true)}
                className="px-3 py-1.5 rounded-lg border border-rose-700/50 hover:bg-rose-950/40 text-rose-300 font-medium transition-colors"
              >
                Reject Action
              </button>
              <button
                type="button"
                id={`approve-tool-btn-${toolCall.id}`}
                disabled={isSubmitting}
                onClick={handleApprove}
                className="px-3.5 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white font-medium shadow-md shadow-emerald-950/50 flex items-center gap-1.5 transition-all"
              >
                {isSubmitting ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <ArrowRight className="h-3.5 w-3.5" />}
                Approve & Execute
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

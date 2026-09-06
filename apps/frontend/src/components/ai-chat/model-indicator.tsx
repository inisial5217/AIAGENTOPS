"use client";

import React from "react";
import { Cpu, ChevronDown, Check, Sparkles } from "lucide-react";
import { ModelInfo } from "../../types/ai";

interface ModelIndicatorProps {
  models: ModelInfo[];
  selectedModel: string;
  selectedProvider: string;
  onSelectModel: (model: ModelInfo) => void;
  isLoading?: boolean;
}

export const ModelIndicator: React.FC<ModelIndicatorProps> = ({
  models,
  selectedModel,
  selectedProvider,
  onSelectModel,
  isLoading,
}) => {
  const [isOpen, setIsOpen] = React.useState(false);

  const currentModel = models.find(
    (m) => m.id === selectedModel || m.model_name === selectedModel
  ) || {
    id: selectedModel || "gemini-2.0-flash",
    provider: selectedProvider || "google",
    model_name: selectedModel || "Gemini 2.0 Flash",
    is_default: true,
    status: "available" as const,
  };

  const getProviderColor = (provider: string) => {
    switch (provider.toLowerCase()) {
      case "google":
        return "text-blue-400 bg-blue-500/10 border-blue-500/30";
      case "openai":
        return "text-emerald-400 bg-emerald-500/10 border-emerald-500/30";
      case "anthropic":
        return "text-amber-400 bg-amber-500/10 border-amber-500/30";
      case "ollama":
        return "text-purple-400 bg-purple-500/10 border-purple-500/30";
      default:
        return "text-cyan-400 bg-cyan-500/10 border-cyan-500/30";
    }
  };

  return (
    <div className="relative inline-block text-left" id="ai-model-selector-wrapper">
      <button
        type="button"
        id="ai-model-selector-btn"
        disabled={isLoading}
        onClick={() => setIsOpen(!isOpen)}
        className={`inline-flex items-center gap-2 px-2.5 py-1.5 rounded-lg border text-xs font-medium transition-all ${getProviderColor(
          currentModel.provider
        )} hover:bg-slate-800/60 focus:outline-none`}
      >
        <span className="flex h-2 w-2 relative">
          <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
          <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
        </span>
        <Sparkles className="h-3.5 w-3.5" />
        <span className="font-semibold uppercase tracking-wider text-[10px]">
          {currentModel.provider}
        </span>
        <span className="text-slate-200">{currentModel.model_name}</span>
        <ChevronDown className={`h-3 w-3 transition-transform ${isOpen ? "rotate-180" : ""}`} />
      </button>

      {isOpen && (
        <div
          id="ai-model-dropdown-menu"
          className="absolute left-0 mt-2 w-64 rounded-xl bg-slate-900 border border-slate-700/80 shadow-2xl z-50 p-1.5 focus:outline-none backdrop-blur-md"
        >
          <div className="px-2.5 py-1.5 text-[10px] font-semibold text-slate-400 uppercase tracking-wider border-b border-slate-800">
            Available AI Providers
          </div>
          <div className="py-1 space-y-0.5 max-h-56 overflow-y-auto">
            {models.length === 0 ? (
              <div className="px-3 py-2 text-xs text-slate-400">Loading models...</div>
            ) : (
              models.map((m) => {
                const isSelected =
                  m.id === currentModel.id || m.model_name === currentModel.model_name;
                return (
                  <button
                    key={m.id}
                    type="button"
                    onClick={() => {
                      onSelectModel(m);
                      setIsOpen(false);
                    }}
                    className={`w-full text-left px-2.5 py-2 rounded-lg text-xs flex items-center justify-between transition-colors ${
                      isSelected
                        ? "bg-blue-600/20 text-blue-300 font-medium"
                        : "text-slate-300 hover:bg-slate-800 hover:text-white"
                    }`}
                  >
                    <div className="flex flex-col">
                      <div className="flex items-center gap-1.5">
                        <span className="font-semibold text-[11px] text-slate-400 uppercase">
                          {m.provider}
                        </span>
                        <span className="text-slate-200">{m.model_name}</span>
                      </div>
                      {m.context_window && (
                        <span className="text-[10px] text-slate-500">
                          {Math.round(m.context_window / 1000)}k ctx • $
                          {m.input_cost_per_1k ? (m.input_cost_per_1k * 1000).toFixed(2) : "0"}/1M
                        </span>
                      )}
                    </div>
                    {isSelected && <Check className="h-4 w-4 text-blue-400" />}
                  </button>
                );
              })
            )}
          </div>
        </div>
      )}
    </div>
  );
};

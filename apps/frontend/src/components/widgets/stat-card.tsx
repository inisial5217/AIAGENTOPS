import * as React from "react";
import { clsx } from "clsx";
import { Skeleton } from "../ui/skeleton";

export interface StatCardProps {
  value: string | number;
  label: string;
  sublabel?: string;
  color?: "default" | "success" | "warning" | "danger" | "critical" | "cyan";
  accentTop?: "none" | "cyan" | "pink" | "emerald" | "amber" | "red";
  isLoading?: boolean;
}

export function StatCard({
  value,
  label,
  sublabel,
  color = "default",
  accentTop = "none",
  isLoading = false,
}: StatCardProps) {
  const colorStyles: Record<string, string> = {
    default: "text-[var(--text-primary)]",
    success: "text-[var(--status-success)]",
    warning: "text-[var(--status-warning)]",
    danger: "text-[var(--status-error)]",
    critical: "text-[var(--status-critical)]",
    cyan: "text-[var(--accent-default)]",
  };

  const accentBorders: Record<string, string> = {
    none: "",
    cyan: "border-t-2 border-t-[var(--accent-default)]",
    pink: "border-t-2 border-t-[var(--accent-pink)]",
    emerald: "border-t-2 border-t-[var(--status-success)]",
    amber: "border-t-2 border-t-[var(--status-warning)]",
    red: "border-t-2 border-t-[var(--status-critical)]",
  };

  if (isLoading) {
    return (
      <div className="rounded-[var(--radius-xl)] bg-[var(--bg-card)] border border-[var(--border-default)] p-4 shadow-lg flex flex-col justify-between h-28">
        <Skeleton variant="rectangular" className="h-8 w-16 mb-2" />
        <Skeleton variant="text" className="h-3 w-28 mb-1" />
        <Skeleton variant="text" className="h-2.5 w-20 opacity-70" />
      </div>
    );
  }

  return (
    <div
      className={clsx(
        "rounded-[var(--radius-xl)] bg-[var(--bg-card)] border border-[var(--border-default)] p-4 shadow-lg flex flex-col justify-between h-28 text-left transition-all hover:border-[var(--border-hover)] hover:bg-[var(--bg-card-hover)]",
        accentBorders[accentTop]
      )}
    >
      <div className={clsx("text-2xl font-bold tracking-tight font-sans", colorStyles[color])}>
        {value}
      </div>
      <div>
        <div className="text-[11px] font-semibold text-[var(--text-secondary)] uppercase tracking-wider">
          {label}
        </div>
        {sublabel && (
          <div className="text-[10px] text-[var(--text-muted)] font-mono mt-0.5 truncate">
            {sublabel}
          </div>
        )}
      </div>
    </div>
  );
}

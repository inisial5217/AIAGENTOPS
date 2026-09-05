import * as React from "react";
import { clsx } from "clsx";

export interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: "success" | "warning" | "error" | "critical" | "info" | "neutral" | "cyan";
  dot?: boolean;
  pulse?: boolean;
  size?: "sm" | "md";
}

export function Badge({
  className,
  variant = "neutral",
  dot = false,
  pulse = false,
  size = "md",
  children,
  ...props
}: BadgeProps) {
  const variantStyles: Record<string, { badge: string; dot: string }> = {
    success: {
      badge: "bg-emerald-500/10 text-emerald-400 border-emerald-500/30",
      dot: "bg-emerald-400",
    },
    warning: {
      badge: "bg-amber-500/10 text-amber-400 border-amber-500/30",
      dot: "bg-amber-400",
    },
    error: {
      badge: "bg-red-500/10 text-red-400 border-red-500/30",
      dot: "bg-red-400",
    },
    critical: {
      badge: "bg-rose-500/15 text-rose-400 border-rose-500/40 font-bold",
      dot: "bg-rose-500 animate-pulse",
    },
    info: {
      badge: "bg-blue-500/10 text-blue-400 border-blue-500/30",
      dot: "bg-blue-400",
    },
    cyan: {
      badge: "bg-cyan-500/10 text-cyan-400 border-cyan-500/30",
      dot: "bg-cyan-400 animate-pulse",
    },
    neutral: {
      badge: "bg-gray-500/10 text-gray-400 border-gray-500/20",
      dot: "bg-gray-400",
    },
  };

  const currentVariant = variantStyles[variant] || variantStyles.neutral;

  return (
    <span
      className={clsx(
        "inline-flex items-center gap-1.5 rounded-full font-mono border tracking-wide uppercase select-none",
        size === "sm" ? "px-2 py-0.2 text-[10px]" : "px-2.5 py-0.5 text-[11px]",
        currentVariant.badge,
        className
      )}
      {...props}
    >
      {(dot || pulse) && (
        <span
          className={clsx(
            "w-1.5 h-1.5 rounded-full shrink-0",
            currentVariant.dot,
            pulse && "animate-ping"
          )}
        />
      )}
      {children}
    </span>
  );
}

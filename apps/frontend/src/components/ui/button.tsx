import * as React from "react";
import { Loader2 } from "lucide-react";
import { clsx } from "clsx";

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "secondary" | "outline" | "ghost" | "danger" | "quickfix";
  size?: "sm" | "md" | "lg";
  isLoading?: boolean;
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  (
    {
      className,
      variant = "primary",
      size = "md",
      isLoading = false,
      disabled,
      children,
      ...props
    },
    ref
  ) => {
    const baseStyles =
      "inline-flex items-center justify-center font-medium transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-default)] disabled:pointer-events-none disabled:opacity-50 select-none cursor-pointer";

    const variantStyles: Record<string, string> = {
      primary:
        "bg-[var(--accent-default)] text-[var(--text-inverse)] hover:bg-[var(--accent-hover)] active:bg-[var(--accent-active)] shadow-md shadow-[var(--accent-glow)]",
      secondary:
        "bg-[var(--bg-secondary)] text-[var(--text-primary)] hover:bg-[var(--bg-hover)] border border-[var(--border-default)] hover:border-[var(--border-hover)]",
      outline:
        "border border-[var(--border-default)] text-[var(--text-primary)] hover:bg-[var(--bg-hover)] hover:border-[var(--border-accent)]",
      ghost:
        "text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-hover)]",
      danger:
        "bg-[var(--status-error)] text-white hover:bg-red-600 shadow-md shadow-[var(--status-error-glow)]",
      quickfix:
        "bg-[var(--accent-default)] text-[var(--text-inverse)] font-semibold hover:bg-[var(--accent-hover)] shadow-lg shadow-[var(--accent-glow)] border border-cyan-300/30",
    };

    const sizeStyles: Record<string, string> = {
      sm: "h-8 px-3 text-xs rounded-[var(--radius-md)] gap-1.5",
      md: "h-9 px-4 text-xs rounded-[var(--radius-md)] gap-2",
      lg: "h-11 px-6 text-sm rounded-[var(--radius-lg)] gap-2.5",
    };

    return (
      <button
        ref={ref}
        disabled={disabled || isLoading}
        className={clsx(
          baseStyles,
          variantStyles[variant],
          sizeStyles[size],
          className
        )}
        {...props}
      >
        {isLoading && <Loader2 className="w-4 h-4 animate-spin shrink-0" />}
        {children}
      </button>
    );
  }
);

Button.displayName = "Button";

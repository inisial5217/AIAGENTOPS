import * as React from "react";
import { clsx } from "clsx";

export interface InputProps
  extends Omit<React.InputHTMLAttributes<HTMLInputElement>, "prefix"> {
  label?: string;
  error?: string;
  icon?: React.ReactNode;
  prefix?: React.ReactNode;
}

export const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type = "text", label, error, icon, prefix, disabled, ...props }, ref) => {
    const leadingIcon = icon || prefix;
    return (
      <div className="w-full space-y-1.5 text-left">
        {label && (
          <label className="block text-xs font-medium text-[var(--text-secondary)]">
            {label}
          </label>
        )}
        <div className="relative flex items-center">
          {leadingIcon && (
            <div className="absolute left-3 flex items-center pointer-events-none text-[var(--text-muted)]">
              {leadingIcon}
            </div>
          )}
          <input
            type={type}
            ref={ref}
            disabled={disabled}
            className={clsx(
              "w-full h-9 px-3 text-xs bg-[var(--bg-secondary)] border rounded-[var(--radius-md)] text-[var(--text-primary)] placeholder:text-[var(--text-muted)] transition-colors focus:outline-none focus:ring-1 focus:ring-[var(--accent-default)] disabled:opacity-50 disabled:cursor-not-allowed",
              leadingIcon ? "pl-9" : "pl-3",
              error
                ? "border-[var(--status-error)] focus:border-[var(--status-error)]"
                : "border-[var(--border-default)] hover:border-[var(--border-hover)] focus:border-[var(--accent-default)]",
              className
            )}
            {...props}
          />
        </div>
        {error && (
          <span className="block text-[11px] text-[var(--status-error)]">
            {error}
          </span>
        )}
      </div>
    );
  }
);

Input.displayName = "Input";

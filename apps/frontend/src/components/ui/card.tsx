import * as React from "react";
import { clsx } from "clsx";

export interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
  accentTop?: "none" | "cyan" | "pink" | "emerald" | "amber" | "red";
}

export function Card({
  className,
  accentTop = "none",
  children,
  ...props
}: CardProps) {
  const accentBorders: Record<string, string> = {
    none: "",
    cyan: "border-t-2 border-t-[var(--accent-default)]",
    pink: "border-t-2 border-t-[var(--accent-pink)]",
    emerald: "border-t-2 border-t-[var(--status-success)]",
    amber: "border-t-2 border-t-[var(--status-warning)]",
    red: "border-t-2 border-t-[var(--status-critical)]",
  };

  return (
    <div
      className={clsx(
        "rounded-[var(--radius-xl)] bg-[var(--bg-card)] border border-[var(--border-default)] p-5 shadow-lg transition-all duration-200",
        accentBorders[accentTop],
        className
      )}
      {...props}
    >
      {children}
    </div>
  );
}

export function CardHeader({
  className,
  children,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={clsx(
        "flex items-center justify-between pb-3 border-b border-[var(--border-subtle)]",
        className
      )}
      {...props}
    >
      {children}
    </div>
  );
}

export function CardTitle({
  className,
  children,
  ...props
}: React.HTMLAttributes<HTMLHeadingElement>) {
  return (
    <h3
      className={clsx(
        "text-sm font-semibold tracking-tight text-[var(--text-primary)]",
        className
      )}
      {...props}
    >
      {children}
    </h3>
  );
}

export function CardDescription({
  className,
  children,
  ...props
}: React.HTMLAttributes<HTMLParagraphElement>) {
  return (
    <p
      className={clsx("text-xs text-[var(--text-secondary)] mt-0.5", className)}
      {...props}
    >
      {children}
    </p>
  );
}

export function CardContent({
  className,
  children,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div className={clsx("pt-3", className)} {...props}>
      {children}
    </div>
  );
}

export function CardFooter({
  className,
  children,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={clsx(
        "flex items-center justify-between pt-3 mt-3 border-t border-[var(--border-subtle)]",
        className
      )}
      {...props}
    >
      {children}
    </div>
  );
}

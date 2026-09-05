import * as React from "react";
import { clsx } from "clsx";

export interface SkeletonProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: "text" | "circular" | "rectangular";
}

export function Skeleton({
  className,
  variant = "rectangular",
  ...props
}: SkeletonProps) {
  const variantStyles: Record<string, string> = {
    text: "h-4 w-full rounded-[var(--radius-sm)]",
    circular: "rounded-full aspect-square",
    rectangular: "rounded-[var(--radius-md)]",
  };

  return (
    <div
      className={clsx(
        "animate-pulse bg-[var(--bg-hover)]/70 border border-[var(--border-subtle)]/40",
        variantStyles[variant],
        className
      )}
      {...props}
    />
  );
}

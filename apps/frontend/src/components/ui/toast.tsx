"use client";

import * as React from "react";
import * as ToastPrimitive from "@radix-ui/react-toast";
import { X } from "lucide-react";
import { clsx } from "clsx";

export const ToastProvider = ToastPrimitive.Provider;

export function ToastViewport({
  className,
  ...props
}: React.ComponentPropsWithoutRef<typeof ToastPrimitive.Viewport>) {
  return (
    <ToastPrimitive.Viewport
      className={clsx(
        "fixed bottom-0 right-0 z-50 flex max-h-screen w-full flex-col-reverse p-4 sm:bottom-0 sm:right-0 sm:top-auto sm:flex-col md:max-w-[420px]",
        className
      )}
      {...props}
    />
  );
}

export interface ToastProps
  extends React.ComponentPropsWithoutRef<typeof ToastPrimitive.Root> {
  variant?: "default" | "success" | "warning" | "error";
}

export function Toast({
  className,
  variant = "default",
  children,
  ...props
}: ToastProps) {
  const variantStyles: Record<string, string> = {
    default: "bg-[var(--bg-card)] border-[var(--border-default)] text-[var(--text-primary)]",
    success: "bg-emerald-950/80 border-emerald-500/40 text-emerald-300",
    warning: "bg-amber-950/80 border-amber-500/40 text-amber-300",
    error: "bg-red-950/80 border-red-500/40 text-red-300",
  };

  return (
    <ToastPrimitive.Root
      className={clsx(
        "group pointer-events-auto relative flex w-full items-center justify-between space-x-4 overflow-hidden rounded-[var(--radius-lg)] border p-4 shadow-xl transition-all data-[swipe=cancel]:translate-x-0 data-[swipe=end]:translate-x-[var(--radix-toast-swipe-end-x)] data-[swipe=move]:translate-x-[var(--radix-toast-swipe-move-x)] data-[swipe=move]:transition-none data-[state=open]:animate-in data-[state=closed]:animate-out data-[swipe=end]:animate-out data-[state=closed]:fade-out-80 data-[state=closed]:slide-out-to-right-full data-[state=open]:slide-in-from-top-full",
        variantStyles[variant],
        className
      )}
      {...props}
    >
      <div className="grid gap-1">{children}</div>
      <ToastPrimitive.Close className="rounded-md p-1 text-[var(--text-muted)] hover:text-[var(--text-primary)] focus:outline-none cursor-pointer">
        <X className="h-4 w-4" />
      </ToastPrimitive.Close>
    </ToastPrimitive.Root>
  );
}

export function ToastTitle({
  className,
  ...props
}: React.ComponentPropsWithoutRef<typeof ToastPrimitive.Title>) {
  return (
    <ToastPrimitive.Title
      className={clsx("text-xs font-semibold tracking-wide", className)}
      {...props}
    />
  );
}

export function ToastDescription({
  className,
  ...props
}: React.ComponentPropsWithoutRef<typeof ToastPrimitive.Description>) {
  return (
    <ToastPrimitive.Description
      className={clsx("text-xs opacity-90", className)}
      {...props}
    />
  );
}

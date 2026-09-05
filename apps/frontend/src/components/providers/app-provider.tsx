"use client";

import * as React from "react";
import { QueryProvider } from "./query-provider";
import { ToastProvider, ToastViewport } from "../ui/toast";
import { useThemeStore } from "../../store/theme-store";
import { useAuthStore } from "../../lib/auth";

export function AppProvider({ children }: { children: React.ReactNode }) {
  const initTheme = useThemeStore((state) => state.initTheme);
  const initAuth = useAuthStore((state) => state.initAuth);

  React.useEffect(() => {
    initTheme();
    initAuth();
  }, [initTheme, initAuth]);

  return (
    <QueryProvider>
      <ToastProvider>
        {children}
        <ToastViewport />
      </ToastProvider>
    </QueryProvider>
  );
}

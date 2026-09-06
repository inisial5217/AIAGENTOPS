"use client";

import * as React from "react";
import { QueryProvider } from "./query-provider";
import { NotificationToastProvider } from "./notification-toast-provider";
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
      <NotificationToastProvider>{children}</NotificationToastProvider>
    </QueryProvider>
  );
}

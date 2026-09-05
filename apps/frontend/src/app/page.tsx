"use client";

import * as React from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "../hooks/use-auth";

export default function HomePage() {
  const router = useRouter();
  const { isAuthenticated, token } = useAuth();

  React.useEffect(() => {
    // If token exists, go straight to monitoring dashboard, otherwise to login
    const storedToken =
      typeof window !== "undefined"
        ? localStorage.getItem("cifo_access_token")
        : null;

    if (isAuthenticated || storedToken || token) {
      router.replace("/monitoring");
    } else {
      router.replace("/login");
    }
  }, [isAuthenticated, token, router]);

  return (
    <div className="min-h-screen bg-[var(--bg-primary)] flex flex-col items-center justify-center">
      <div className="flex flex-col items-center gap-3">
        <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-pink-500/20 via-cyan-500/20 to-transparent border border-pink-500/30 flex items-center justify-center animate-pulse">
          <span className="text-sm font-black text-white">CF</span>
        </div>
        <span className="text-xs font-mono text-[var(--text-muted)]">
          Routing to Command Center...
        </span>
      </div>
    </div>
  );
}

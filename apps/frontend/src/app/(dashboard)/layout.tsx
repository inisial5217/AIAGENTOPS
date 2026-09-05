"use client";

import * as React from "react";
import { Sidebar } from "../../components/layout/sidebar";
import { Header } from "../../components/layout/header";
import { Breadcrumb } from "../../components/layout/breadcrumb";
import { useSidebarStore } from "../../store/sidebar-store";
import { ShieldCheck, Cpu } from "lucide-react";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { isCollapsed } = useSidebarStore();

  return (
    <div className="min-h-screen bg-[var(--bg-primary)] text-[var(--text-primary)] flex">
      {/* Fixed Sidebar */}
      <Sidebar />

      {/* Main Content Area */}
      <div
        className={`flex-1 flex flex-col min-w-0 transition-all duration-300 ${
          isCollapsed ? "ml-16" : "ml-64"
        }`}
      >
        {/* Sticky Header */}
        <Header />

        {/* Sub-header Bar (Breadcrumb & Ticker) */}
        <div className="h-10 px-6 border-b border-[var(--border-subtle)] bg-[var(--bg-secondary)]/60 flex items-center justify-between text-xs">
          <Breadcrumb />

          <div className="hidden sm:flex items-center gap-4 text-[11px] font-mono text-[var(--text-muted)]">
            <div className="flex items-center gap-1.5">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
              <span>Daemon Sync: OK</span>
            </div>
            <div className="flex items-center gap-1.5">
              <ShieldCheck className="w-3 h-3 text-cyan-400" />
              <span>RBAC Enforced</span>
            </div>
            <div className="flex items-center gap-1.5">
              <Cpu className="w-3 h-3 text-pink-400" />
              <span>Prober v1.2</span>
            </div>
          </div>
        </div>

        {/* Dynamic Page Content */}
        <main className="flex-1 p-6 overflow-y-auto">{children}</main>
      </div>
    </div>
  );
}

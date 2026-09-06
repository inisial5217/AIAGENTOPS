"use client";

import * as React from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import {
  Activity,
  Layers,
  Container,
  GitPullRequest,
  AlertTriangle,
  Settings,
  ChevronDown,
  ChevronRight,
  LogOut,
  PanelLeftClose,
  PanelLeftOpen,
} from "lucide-react";
import { useSidebarStore } from "../../store/sidebar-store";
import { useAuth } from "../../hooks/use-auth";

export function Sidebar() {
  const pathname = usePathname();
  const router = useRouter();
  const { isCollapsed, expandedSections, toggleSidebar, toggleSection } =
    useSidebarStore();
  const { logout } = useAuth();

  const handleExit = async () => {
    await logout();
    router.push("/login");
  };

  const isRouteActive = (href: string) => {
    if (href === "/monitoring") {
      return pathname === "/" || pathname === "/monitoring";
    }
    return pathname.startsWith(href);
  };

  return (
    <aside
      className={`fixed left-0 top-0 bottom-0 z-40 flex flex-col bg-[var(--bg-primary)] border-r border-[var(--border-default)] transition-all duration-300 ${
        isCollapsed ? "w-16" : "w-64"
      }`}
    >
      {/* Brand Header */}
      <div className="h-16 flex items-center justify-between px-4 border-b border-[var(--border-subtle)] shrink-0">
        {!isCollapsed && (
          <div className="flex items-center gap-2.5">
            <div className="relative flex items-center justify-center w-8 h-8 rounded-lg bg-gradient-to-br from-pink-500/20 via-cyan-500/20 to-transparent border border-pink-500/30">
              <span className="text-sm font-black tracking-tighter text-white">
                C<span className="text-[var(--accent-pink)]">I</span>F<span className="text-[var(--accent-default)]">O</span>
              </span>
              <span className="absolute -top-1 -right-1 w-2 h-2 rounded-full bg-[var(--accent-pink)] animate-ping" />
            </div>
            <div className="flex flex-col text-left">
              <span className="text-base font-bold tracking-tight text-[var(--text-primary)] leading-none">
                CIFO
              </span>
              <span className="text-[9px] text-[var(--text-muted)] font-mono tracking-wider lowercase">
                your unlimited it partner
              </span>
            </div>
          </div>
        )}

        {isCollapsed && (
          <div className="mx-auto flex items-center justify-center w-8 h-8 rounded-lg bg-gradient-to-br from-pink-500/20 to-cyan-500/20 border border-cyan-500/30">
            <span className="text-xs font-black text-white">CF</span>
          </div>
        )}

        <button
          onClick={toggleSidebar}
          className="p-1.5 rounded-md text-[var(--text-muted)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-hover)] transition-colors cursor-pointer"
          title={isCollapsed ? "Expand sidebar" : "Collapse sidebar"}
        >
          {isCollapsed ? (
            <PanelLeftOpen className="w-4 h-4" />
          ) : (
            <PanelLeftClose className="w-4 h-4" />
          )}
        </button>
      </div>

      {/* Navigation Links */}
      <nav className="flex-1 overflow-y-auto px-3 py-4 space-y-1 text-left font-mono text-xs">
        {/* Monitoring Root */}
        <Link
          href="/monitoring"
          className={`flex items-center gap-3 px-3 py-2.5 rounded-[var(--radius-md)] transition-all select-none ${
            isRouteActive("/monitoring")
              ? "bg-cyan-500/10 text-[var(--accent-default)] border border-cyan-500/30 font-semibold shadow-sm shadow-[var(--accent-glow)]"
              : "text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-hover)]"
          }`}
          title="Monitoring"
        >
          <Activity className="w-4 h-4 shrink-0" />
          {!isCollapsed && (
            <span className="flex-1 uppercase tracking-wider text-[11px]">
              {isRouteActive("/monitoring") && <span className="mr-1.5">•</span>}
              Monitoring
            </span>
          )}
        </Link>

        {/* Kubernetes Accordion */}
        <div className="pt-1">
          <button
            onClick={() => toggleSection("kubernetes")}
            className={`w-full flex items-center justify-between px-3 py-2.5 rounded-[var(--radius-md)] transition-colors cursor-pointer select-none ${
              isRouteActive("/kubernetes")
                ? "text-[var(--accent-default)] font-semibold"
                : "text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-hover)]"
            }`}
            title="Kubernetes"
          >
            <div className="flex items-center gap-3">
              <Layers className="w-4 h-4 shrink-0" />
              {!isCollapsed && (
                <span className="uppercase tracking-wider text-[11px]">
                  Kubernetes
                </span>
              )}
            </div>
            {!isCollapsed &&
              (expandedSections.kubernetes ? (
                <ChevronDown className="w-3.5 h-3.5" />
              ) : (
                <ChevronRight className="w-3.5 h-3.5" />
              ))}
          </button>

          {!isCollapsed && expandedSections.kubernetes && (
            <div className="ml-7 pl-2 border-l border-[var(--border-subtle)] space-y-1 mt-1">
              {[
                { name: "Cluster Overview", href: "/kubernetes" },
                { name: "ArgoCD / GitOps", href: "/argocd" },
              ].map((sub) => {
                const active = pathname === sub.href;
                return (
                  <Link
                    key={sub.href}
                    href={sub.href}
                    className={`block px-2.5 py-1.5 rounded text-[11px] transition-colors ${
                      active
                        ? "text-[var(--accent-default)] font-semibold bg-cyan-500/10"
                        : "text-[var(--text-muted)] hover:text-[var(--text-primary)]"
                    }`}
                  >
                    {active && <span className="mr-1.5">•</span>}
                    {sub.name}
                  </Link>
                );
              })}
            </div>
          )}
        </div>

        {/* ArgoCD GitOps Root Link */}
        <Link
          href="/argocd"
          className={`flex items-center gap-3 px-3 py-2.5 rounded-[var(--radius-md)] transition-all select-none ${
            isRouteActive("/argocd")
              ? "bg-cyan-500/10 text-[var(--accent-default)] border border-cyan-500/30 font-semibold shadow-sm shadow-[var(--accent-glow)]"
              : "text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-hover)]"
          }`}
          title="ArgoCD GitOps"
        >
          <GitPullRequest className="w-4 h-4 shrink-0" />
          {!isCollapsed && (
            <span className="flex-1 uppercase tracking-wider text-[11px]">
              {isRouteActive("/argocd") && <span className="mr-1.5">•</span>}
              ArgoCD GitOps
            </span>
          )}
        </Link>

        {/* Docker Accordion */}
        <div className="pt-1">
          <button
            onClick={() => toggleSection("docker")}
            className={`w-full flex items-center justify-between px-3 py-2.5 rounded-[var(--radius-md)] transition-colors cursor-pointer select-none ${
              isRouteActive("/docker")
                ? "text-[var(--accent-default)] font-semibold"
                : "text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-hover)]"
            }`}
            title="Docker"
          >
            <div className="flex items-center gap-3">
              <Container className="w-4 h-4 shrink-0" />
              {!isCollapsed && (
                <span className="uppercase tracking-wider text-[11px]">
                  Docker
                </span>
              )}
            </div>
            {!isCollapsed &&
              (expandedSections.docker ? (
                <ChevronDown className="w-3.5 h-3.5" />
              ) : (
                <ChevronRight className="w-3.5 h-3.5" />
              ))}
          </button>

          {!isCollapsed && expandedSections.docker && (
            <div className="ml-7 pl-2 border-l border-[var(--border-subtle)] space-y-1 mt-1">
              {[
                { name: "Host Overview", href: "/docker" },
                { name: "Containers & Images", href: "/docker/containers" },
                { name: "Networks & Storage", href: "/docker/networks" },
                { name: "Compose Stacks", href: "/docker/compose" },
              ].map((sub) => {
                const active = pathname === sub.href;
                return (
                  <Link
                    key={sub.href}
                    href={sub.href}
                    className={`block px-2.5 py-1.5 rounded text-[11px] transition-colors ${
                      active
                        ? "text-[var(--accent-default)] font-semibold bg-cyan-500/10"
                        : "text-[var(--text-muted)] hover:text-[var(--text-primary)]"
                    }`}
                  >
                    {active && <span className="mr-1.5">•</span>}
                    {sub.name}
                  </Link>
                );
              })}
            </div>
          )}
        </div>

        {/* Incidents */}
        <Link
          href="/incidents"
          className={`flex items-center gap-3 px-3 py-2.5 rounded-[var(--radius-md)] transition-all select-none ${
            isRouteActive("/incidents")
              ? "bg-red-500/10 text-red-400 border border-red-500/30 font-semibold"
              : "text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-hover)]"
          }`}
          title="Incidents"
        >
          <AlertTriangle className="w-4 h-4 shrink-0 text-red-400" />
          {!isCollapsed && (
            <span className="flex-1 uppercase tracking-wider text-[11px]">
              {isRouteActive("/incidents") && <span className="mr-1.5">•</span>}
              Incidents
            </span>
          )}
        </Link>

        {/* Settings */}
        <Link
          href="/settings"
          className={`flex items-center gap-3 px-3 py-2.5 rounded-[var(--radius-md)] transition-all select-none ${
            isRouteActive("/settings")
              ? "bg-cyan-500/10 text-[var(--accent-default)] border border-cyan-500/30 font-semibold"
              : "text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-hover)]"
          }`}
          title="Settings"
        >
          <Settings className="w-4 h-4 shrink-0" />
          {!isCollapsed && (
            <span className="flex-1 uppercase tracking-wider text-[11px]">
              {isRouteActive("/settings") && <span className="mr-1.5">•</span>}
              Settings
            </span>
          )}
        </Link>
      </nav>

      {/* Exit Button Footer */}
      <div className="p-3 border-t border-[var(--border-subtle)] shrink-0">
        <button
          onClick={handleExit}
          className="w-full flex items-center gap-3 px-3 py-2 rounded-[var(--radius-md)] text-red-400 hover:text-red-300 hover:bg-red-500/10 transition-colors font-mono text-xs cursor-pointer select-none"
          title="Exit platform"
        >
          <LogOut className="w-4 h-4 shrink-0" />
          {!isCollapsed && (
            <span className="uppercase tracking-wider font-semibold">
              ← Exit
            </span>
          )}
        </button>
      </div>
    </aside>
  );
}

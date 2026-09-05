"use client";

import * as React from "react";
import { useRouter } from "next/navigation";
import {
  Search,
  Clock,
  Wrench,
  Sun,
  Moon,
  Bell,
  CheckCheck,
  Trash2,
  LogOut,
  User,
  Shield,
  ChevronDown,
  Sparkles,
} from "lucide-react";
import { useThemeStore } from "../../store/theme-store";
import { useNotificationStore } from "../../store/notification-store";
import { useAuth } from "../../hooks/use-auth";
import { Modal } from "../ui/modal";
import { Button } from "../ui/button";
import { Badge } from "../ui/badge";

export function Header() {
  const router = useRouter();
  const { theme, toggleTheme } = useThemeStore();
  const { notifications, unreadCount, markAllAsRead, clearAll } =
    useNotificationStore();
  const { user, logout } = useAuth();

  const [searchQuery, setSearchQuery] = React.useState("");
  const [timeRange, setTimeRange] = React.useState("Last 1 Hour");
  const [isTimeDropdownOpen, setIsTimeDropdownOpen] = React.useState(false);
  const [isNotifOpen, setIsNotifOpen] = React.useState(false);
  const [isUserMenuOpen, setIsUserMenuOpen] = React.useState(false);
  const [isQuickFixModalOpen, setIsQuickFixModalOpen] = React.useState(false);
  const [quickFixRunning, setQuickFixRunning] = React.useState<string | null>(
    null
  );

  const timeRanges = [
    "Last 15 Minutes",
    "Last 1 Hour",
    "Last 6 Hours",
    "Last 24 Hours",
    "Last 7 Days",
  ];

  const handleLogout = async () => {
    await logout();
    router.push("/login");
  };

  const handleRunQuickFix = (fixName: string) => {
    setQuickFixRunning(fixName);
    setTimeout(() => {
      setQuickFixRunning(null);
      setIsQuickFixModalOpen(false);
    }, 1500);
  };

  const getUserInitials = () => {
    if (!user?.name) return "AD";
    const parts = user.name.trim().split(" ");
    if (parts.length === 1) return parts[0].substring(0, 2).toUpperCase();
    return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
  };

  return (
    <header className="h-16 border-b border-[var(--border-default)] bg-[var(--bg-primary)] px-6 flex items-center justify-between sticky top-0 z-30">
      {/* Global Search Bar */}
      <div className="flex items-center gap-3 w-64 sm:w-80 md:w-96">
        <div className="relative w-full">
          <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search clusters or incidents..."
            className="w-full h-9 pl-9 pr-14 rounded-lg bg-[var(--bg-secondary)] border border-[var(--border-default)] text-xs text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:outline-none focus:border-cyan-500 focus:ring-1 focus:ring-cyan-500/50 transition-all font-sans"
          />
          <kbd className="absolute right-2.5 top-1/2 -translate-y-1/2 px-1.5 py-0.5 rounded text-[10px] font-mono bg-[var(--bg-card)] border border-[var(--border-subtle)] text-[var(--text-muted)]">
            Ctrl K
          </kbd>
        </div>
      </div>

      {/* Action Controls & User Section */}
      <div className="flex items-center gap-3 shrink-0">
        {/* Time Range Selector */}
        <div className="relative shrink-0">
          <button
            onClick={() => setIsTimeDropdownOpen(!isTimeDropdownOpen)}
            className="flex items-center gap-2 h-9 px-3 rounded-lg bg-[var(--bg-secondary)] border border-[var(--border-default)] hover:border-cyan-500/40 text-xs font-mono text-[var(--text-secondary)] hover:text-[var(--text-primary)] transition-colors cursor-pointer whitespace-nowrap shrink-0"
          >
            <Clock className="w-3.5 h-3.5 text-cyan-400 shrink-0" />
            <span className="whitespace-nowrap">{timeRange}</span>
            <ChevronDown className="w-3.5 h-3.5 text-[var(--text-muted)] shrink-0" />
          </button>

          {isTimeDropdownOpen && (
            <div className="absolute right-0 mt-1.5 w-44 rounded-lg bg-[var(--bg-card)] border border-[var(--border-default)] shadow-xl py-1 z-50 text-xs font-mono">
              {timeRanges.map((range) => (
                <button
                  key={range}
                  onClick={() => {
                    setTimeRange(range);
                    setIsTimeDropdownOpen(false);
                  }}
                  className={`w-full text-left px-3 py-2 hover:bg-[var(--bg-hover)] transition-colors ${
                    timeRange === range
                      ? "text-cyan-400 font-semibold bg-cyan-500/10"
                      : "text-[var(--text-secondary)]"
                  }`}
                >
                  {range}
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Quick Fix Button */}
        <button
          onClick={() => setIsQuickFixModalOpen(true)}
          className="flex items-center gap-1.5 h-9 px-3 rounded-lg bg-cyan-500 hover:bg-cyan-400 text-black font-semibold text-xs transition-all shadow-md shadow-cyan-500/20 active:scale-95 cursor-pointer"
        >
          <Wrench className="w-3.5 h-3.5" />
          <span>+ Quick Fix</span>
        </button>

        {/* Theme Toggle Button */}
        <button
          onClick={toggleTheme}
          aria-label="Toggle Theme"
          className="p-2 rounded-lg bg-[var(--bg-secondary)] border border-[var(--border-default)] hover:bg-[var(--bg-hover)] text-[var(--text-secondary)] hover:text-[var(--text-primary)] transition-colors cursor-pointer"
          title={`Switch to ${theme === "dark" ? "Light" : "Dark"} Mode`}
        >
          {theme === "dark" ? (
            <Sun className="w-4 h-4 text-amber-400" />
          ) : (
            <Moon className="w-4 h-4 text-indigo-400" />
          )}
        </button>

        {/* Notification Bell */}
        <div className="relative">
          <button
            onClick={() => setIsNotifOpen(!isNotifOpen)}
            aria-label="Notifications"
            className="relative p-2 rounded-lg bg-[var(--bg-secondary)] border border-[var(--border-default)] hover:bg-[var(--bg-hover)] text-[var(--text-secondary)] hover:text-[var(--text-primary)] transition-colors cursor-pointer"
          >
            <Bell className="w-4 h-4" />
            {unreadCount > 0 && (
              <span className="absolute -top-1 -right-1 flex items-center justify-center min-w-[16px] h-4 px-1 rounded-full bg-red-500 text-white text-[9px] font-bold animate-pulse">
                {unreadCount}
              </span>
            )}
          </button>

          {isNotifOpen && (
            <div className="absolute right-0 mt-2 w-80 rounded-xl bg-[var(--bg-card)] border border-[var(--border-default)] shadow-2xl z-50 overflow-hidden">
              <div className="flex items-center justify-between px-4 py-3 border-b border-[var(--border-subtle)] bg-[var(--bg-secondary)]">
                <div className="flex items-center gap-2">
                  <span className="text-xs font-bold text-[var(--text-primary)]">
                    System Alerts
                  </span>
                  <Badge variant="cyan" size="sm">
                    {unreadCount} new
                  </Badge>
                </div>
                <div className="flex items-center gap-1">
                  <button
                    onClick={markAllAsRead}
                    title="Mark all as read"
                    className="p-1 text-[var(--text-muted)] hover:text-[var(--text-primary)]"
                  >
                    <CheckCheck className="w-3.5 h-3.5" />
                  </button>
                  <button
                    onClick={clearAll}
                    title="Clear all"
                    className="p-1 text-[var(--text-muted)] hover:text-red-400"
                  >
                    <Trash2 className="w-3.5 h-3.5" />
                  </button>
                </div>
              </div>

              <div className="max-h-72 overflow-y-auto divide-y divide-[var(--border-subtle)]">
                {notifications.length === 0 ? (
                  <div className="py-6 text-center text-xs text-[var(--text-muted)]">
                    No active alerts
                  </div>
                ) : (
                  notifications.map((n) => (
                    <div
                      key={n.id}
                      className={`p-3 text-left transition-colors ${
                        !n.read ? "bg-cyan-500/5" : ""
                      }`}
                    >
                      <div className="flex items-center justify-between mb-1">
                        <span className="text-xs font-semibold text-[var(--text-primary)]">
                          {n.title}
                        </span>
                        <span className="text-[10px] font-mono text-[var(--text-muted)]">
                          {n.timestamp}
                        </span>
                      </div>
                      <p className="text-[11px] text-[var(--text-secondary)] leading-relaxed">
                        {n.message}
                      </p>
                    </div>
                  ))
                )}
              </div>
            </div>
          )}
        </div>

        {/* User Profile Avatar Dropdown */}
        <div className="relative pl-1">
          <button
            onClick={() => setIsUserMenuOpen(!isUserMenuOpen)}
            className="flex items-center gap-2.5 p-1 rounded-lg hover:bg-[var(--bg-hover)] transition-colors cursor-pointer"
          >
            <div className="w-8 h-8 rounded-lg bg-gradient-to-tr from-pink-500 to-cyan-400 flex items-center justify-center font-black text-xs text-white shadow-sm shadow-cyan-500/20">
              {getUserInitials()}
            </div>
            <div className="hidden md:flex flex-col text-left leading-tight">
              <span className="text-xs font-bold text-[var(--text-primary)]">
                {user?.name || "Administrator"}
              </span>
              <span className="text-[10px] font-mono text-cyan-400 uppercase tracking-wider">
                {user?.role || "ADMIN"}
              </span>
            </div>
            <ChevronDown className="w-3.5 h-3.5 text-[var(--text-muted)]" />
          </button>

          {isUserMenuOpen && (
            <div className="absolute right-0 mt-2 w-56 rounded-xl bg-[var(--bg-card)] border border-[var(--border-default)] shadow-2xl py-2 z-50 text-xs">
              <div className="px-4 py-2 border-b border-[var(--border-subtle)]">
                <div className="font-semibold text-[var(--text-primary)]">
                  {user?.name || "CIFO Administrator"}
                </div>
                <div className="text-[11px] text-[var(--text-muted)] truncate">
                  {user?.email || "admin@cifo.local"}
                </div>
                <div className="mt-1.5">
                  <Badge
                    variant={
                      user?.role === "admin"
                        ? "critical"
                        : user?.role === "devops"
                        ? "cyan"
                        : "neutral"
                    }
                    size="sm"
                  >
                    Role: {user?.role || "admin"}
                  </Badge>
                </div>
              </div>

              <div className="py-1">
                <button
                  onClick={() => {
                    setIsUserMenuOpen(false);
                    router.push("/settings");
                  }}
                  className="w-full flex items-center gap-2.5 px-4 py-2 text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-hover)]"
                >
                  <User className="w-3.5 h-3.5" />
                  <span>Account Settings</span>
                </button>
                <button
                  onClick={() => {
                    setIsUserMenuOpen(false);
                    router.push("/settings");
                  }}
                  className="w-full flex items-center gap-2.5 px-4 py-2 text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-hover)]"
                >
                  <Shield className="w-3.5 h-3.5" />
                  <span>Security & Roles</span>
                </button>
              </div>

              <div className="pt-1 border-t border-[var(--border-subtle)]">
                <button
                  onClick={handleLogout}
                  className="w-full flex items-center gap-2.5 px-4 py-2 text-red-400 hover:bg-red-500/10 font-semibold"
                >
                  <LogOut className="w-3.5 h-3.5" />
                  <span>Sign Out</span>
                </button>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Quick Fix Modal */}
      <Modal
        isOpen={isQuickFixModalOpen}
        onClose={() => setIsQuickFixModalOpen(false)}
        title="Automated Quick Remediation"
        description="Select an AI-assisted quick fix to apply to the active cluster."
      >
        <div className="space-y-3 py-2 text-left">
          <div className="p-3 rounded-lg border border-[var(--border-default)] bg-[var(--bg-secondary)] flex items-center justify-between">
            <div>
              <div className="text-xs font-semibold text-[var(--text-primary)]">
                Restart CrashLoop Backoff Pods
              </div>
              <div className="text-[11px] text-[var(--text-muted)]">
                Target: 3 failing containers in production-cifo-1
              </div>
            </div>
            <Button
              size="sm"
              variant="quickfix"
              isLoading={quickFixRunning === "pods"}
              onClick={() => handleRunQuickFix("pods")}
            >
              <Sparkles className="w-3 h-3" />
              Apply
            </Button>
          </div>

          <div className="p-3 rounded-lg border border-[var(--border-default)] bg-[var(--bg-secondary)] flex items-center justify-between">
            <div>
              <div className="text-xs font-semibold text-[var(--text-primary)]">
                Flush Redis Transient Cache
              </div>
              <div className="text-[11px] text-[var(--text-muted)]">
                Free memory & purge expired session tokens
              </div>
            </div>
            <Button
              size="sm"
              variant="outline"
              isLoading={quickFixRunning === "redis"}
              onClick={() => handleRunQuickFix("redis")}
            >
              Run Flush
            </Button>
          </div>

          <div className="p-3 rounded-lg border border-[var(--border-default)] bg-[var(--bg-secondary)] flex items-center justify-between">
            <div>
              <div className="text-xs font-semibold text-[var(--text-primary)]">
                Sync ArgoCD Application Drift
              </div>
              <div className="text-[11px] text-[var(--text-muted)]">
                Reconcile out-of-sync resources with Git repo
              </div>
            </div>
            <Button
              size="sm"
              variant="outline"
              isLoading={quickFixRunning === "argocd"}
              onClick={() => handleRunQuickFix("argocd")}
            >
              Force Sync
            </Button>
          </div>
        </div>
      </Modal>
    </header>
  );
}

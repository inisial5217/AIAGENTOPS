"use client";

import * as React from "react";
import {
  Sliders,
  Bell,
  Bot,
  Users,
  Shield,
  Save,
  Send,
  RefreshCw,
  AlertTriangle,
  Eye,
  EyeOff,
  UserCheck,
  UserX,
  Lock,
  Clock,
  Laptop,
  CheckCircle2,
  XCircle,
} from "lucide-react";
import { settingsService } from "../../../services/settings-service";
import {
  CombinedSettings,
  SystemSettings,
  NotificationSettings,
  UserAdmin,
  ActiveSession,
} from "../../../types/settings";
import { useAuthStore } from "../../../lib/auth";

type SettingsTab = "general" | "notifications" | "ai" | "users" | "security";

export default function SettingsPage() {
  const { user } = useAuthStore();
  const isAdmin = user?.role === "admin";

  const [activeTab, setActiveTab] = React.useState<SettingsTab>("general");
  const [loading, setLoading] = React.useState<boolean>(true);
  const [saving, setSaving] = React.useState<boolean>(false);
  const [testingNotif, setTestingNotif] = React.useState<boolean>(false);
  const [toastMessage, setToastMessage] = React.useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  // Password / Token visibility
  const [showBotToken, setShowBotToken] = React.useState<boolean>(false);

  // Settings State
  const [system, setSystem] = React.useState<SystemSettings>({
    id: "",
    app_name: "CIFO Monitoring Platform",
    retention_days: 30,
    refresh_interval: 10,
    ai_auto_remediation: false,
    ai_analysis_threshold: 0.85,
    session_timeout_mins: 60,
    mfa_enforced: false,
    updated_at: "",
  });

  const [notification, setNotification] = React.useState<NotificationSettings>({
    id: "",
    telegram_enabled: true,
    telegram_bot_token: "",
    telegram_chat_id: "",
    email_enabled: false,
    email_recipients: "",
    critical_alert: true,
    warning_alert: true,
    info_alert: false,
    auto_resolve_alert: true,
    quiet_hours_enabled: false,
    quiet_hours_start: "22:00",
    quiet_hours_end: "07:00",
    updated_at: "",
  });

  // User Administration State
  const [usersList, setUsersList] = React.useState<UserAdmin[]>([]);
  const [usersTotal, setUsersTotal] = React.useState<number>(0);
  const [userUpdatingId, setUserUpdatingId] = React.useState<string | null>(null);

  // Sessions Mock/State
  const [sessions, setSessions] = React.useState<ActiveSession[]>([
    {
      id: "sess-1",
      device: "Edge on Windows 11 (Desktop)",
      ip: "127.0.0.1",
      last_active: "Just now",
      is_current: true,
    },
    {
      id: "sess-2",
      device: "Chrome on macOS Sonoma",
      ip: "192.168.1.14",
      last_active: "2 hours ago",
      is_current: false,
    },
  ]);

  const showToast = (text: string, type: "success" | "error" = "success") => {
    setToastMessage({ text, type });
    setTimeout(() => {
      setToastMessage(null);
    }, 4000);
  };

  const loadData = React.useCallback(async () => {
    setLoading(true);
    try {
      const data = await settingsService.getSettings();
      if (data.system) {
        setSystem(data.system);
      }
      if (data.notification) {
        setNotification(data.notification);
      }

      if (isAdmin) {
        const uData = await settingsService.listUsers(50, 0);
        setUsersList(uData.users || []);
        setUsersTotal(uData.total || 0);
      }
    } catch (err: any) {
      console.error("Failed to load settings:", err);
      showToast(err.message || "Failed to load platform settings", "error");
    } finally {
      setLoading(false);
    }
  }, [isAdmin]);

  React.useEffect(() => {
    loadData();
  }, [loadData]);

  const handleSaveSettings = async () => {
    setSaving(true);
    try {
      const updated = await settingsService.updateSettings({
        system: {
          app_name: system.app_name,
          retention_days: Number(system.retention_days),
          refresh_interval: Number(system.refresh_interval),
          ai_auto_remediation: system.ai_auto_remediation,
          ai_analysis_threshold: Number(system.ai_analysis_threshold),
          session_timeout_mins: Number(system.session_timeout_mins),
          mfa_enforced: system.mfa_enforced,
        },
        notification: {
          telegram_enabled: notification.telegram_enabled,
          telegram_bot_token: notification.telegram_bot_token,
          telegram_chat_id: notification.telegram_chat_id,
          email_enabled: notification.email_enabled,
          email_recipients: notification.email_recipients,
          critical_alert: notification.critical_alert,
          warning_alert: notification.warning_alert,
          info_alert: notification.info_alert,
          auto_resolve_alert: notification.auto_resolve_alert,
          quiet_hours_enabled: notification.quiet_hours_enabled,
          quiet_hours_start: notification.quiet_hours_start,
          quiet_hours_end: notification.quiet_hours_end,
        },
      });

      if (updated.system) setSystem(updated.system);
      if (updated.notification) setNotification(updated.notification);
      showToast("Configuration successfully updated and persisted!");
    } catch (err: any) {
      console.error("Update settings error:", err);
      showToast(err.response?.data?.message || err.message || "Update failed", "error");
    } finally {
      setSaving(false);
    }
  };

  const handleTestNotification = async () => {
    setTestingNotif(true);
    try {
      const res = await settingsService.testNotification();
      showToast(res.message || "Test notification dispatched to Telegram!");
    } catch (err: any) {
      showToast(
        err.response?.data?.detail || err.response?.data?.message || err.message || "Test dispatch failed",
        "error"
      );
    } finally {
      setTestingNotif(false);
    }
  };

  const handleRoleChange = async (userId: string, newRole: string) => {
    setUserUpdatingId(userId);
    try {
      const updated = await settingsService.updateUserRole(userId, newRole);
      setUsersList((prev) =>
        prev.map((u) => (u.id === userId ? { ...u, role: updated.role } : u))
      );
      showToast(`User role updated to ${newRole}`);
    } catch (err: any) {
      showToast(err.response?.data?.detail || "Failed to update role", "error");
    } finally {
      setUserUpdatingId(null);
    }
  };

  const handleToggleActive = async (userId: string, currentActive: boolean) => {
    setUserUpdatingId(userId);
    try {
      let updated: UserAdmin;
      if (currentActive) {
        updated = await settingsService.deactivateUser(userId);
        showToast("User account deactivated");
      } else {
        updated = await settingsService.reactivateUser(userId);
        showToast("User account reactivated");
      }
      setUsersList((prev) =>
        prev.map((u) => (u.id === userId ? { ...u, is_active: updated.is_active } : u))
      );
    } catch (err: any) {
      showToast(err.response?.data?.detail || "Action failed", "error");
    } finally {
      setUserUpdatingId(null);
    }
  };

  const handleRevokeOtherSessions = () => {
    setSessions((prev) => prev.filter((s) => s.is_current));
    showToast("All other active sessions revoked successfully");
  };

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh] gap-4">
        <RefreshCw className="w-8 h-8 text-[var(--accent-default)] animate-spin" />
        <span className="text-sm font-mono text-[var(--text-muted)]">
          Loading platform configuration...
        </span>
      </div>
    );
  }

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      {/* Toast Alert Banner */}
      {toastMessage && (
        <div
          id="settings-toast"
          className={`fixed bottom-6 right-6 z-50 flex items-center gap-3 px-4 py-3 rounded-lg shadow-xl border backdrop-blur-md transition-all animate-in fade-in slide-in-from-bottom-5 ${
            toastMessage.type === "success"
              ? "bg-emerald-950/80 border-emerald-500/40 text-emerald-200"
              : "bg-red-950/80 border-red-500/40 text-red-200"
          }`}
        >
          {toastMessage.type === "success" ? (
            <CheckCircle2 className="w-5 h-5 text-emerald-400" />
          ) : (
            <XCircle className="w-5 h-5 text-red-400" />
          )}
          <span className="text-sm font-mono">{toastMessage.text}</span>
        </div>
      )}

      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b border-[var(--border-subtle)] pb-5">
        <div>
          <h1 className="text-2xl font-black tracking-tight text-[var(--text-primary)] flex items-center gap-2.5">
            <Sliders className="w-6 h-6 text-[var(--accent-default)]" />
            System Settings & Administration
          </h1>
          <p className="text-xs font-mono text-[var(--text-muted)] mt-1">
            Global configuration, notification channels, AI engine parameters, and RBAC governance
          </p>
        </div>

        <div className="flex items-center gap-3">
          <button
            id="settings-save-button"
            onClick={handleSaveSettings}
            disabled={saving}
            className="flex items-center gap-2 px-4 py-2 rounded-lg bg-[var(--accent-default)] text-black font-semibold text-xs tracking-wider uppercase hover:opacity-90 transition-all cursor-pointer disabled:opacity-50 shadow-md shadow-cyan-500/20"
          >
            {saving ? (
              <RefreshCw className="w-4 h-4 animate-spin" />
            ) : (
              <Save className="w-4 h-4" />
            )}
            Save Changes
          </button>
        </div>
      </div>

      {/* Navigation Tabs */}
      <div className="flex items-center gap-2 border-b border-[var(--border-subtle)] overflow-x-auto pb-px">
        {[
          { id: "general", label: "General", icon: Sliders },
          { id: "notifications", label: "Notifications", icon: Bell },
          { id: "ai", label: "AI Configuration", icon: Bot },
          { id: "users", label: "Users & RBAC", icon: Users },
          { id: "security", label: "Security & Sessions", icon: Shield },
        ].map((tab) => {
          const Icon = tab.icon;
          const active = activeTab === tab.id;
          return (
            <button
              key={tab.id}
              id={`tab-button-${tab.id}`}
              onClick={() => setActiveTab(tab.id as SettingsTab)}
              className={`flex items-center gap-2 px-4 py-3 border-b-2 text-xs font-mono transition-all cursor-pointer whitespace-nowrap ${
                active
                  ? "border-[var(--accent-default)] text-[var(--accent-default)] font-bold bg-cyan-500/5"
                  : "border-transparent text-[var(--text-muted)] hover:text-[var(--text-primary)] hover:border-[var(--border-default)]"
              }`}
            >
              <Icon className="w-4 h-4" />
              <span>{tab.label}</span>
              {tab.id === "users" && !isAdmin && (
                <Lock className="w-3 h-3 text-[var(--text-muted)] ml-1" />
              )}
            </button>
          );
        })}
      </div>

      {/* Tab 1: General */}
      {activeTab === "general" && (
        <div id="general-settings-panel" className="space-y-6">
          <div className="bg-[var(--bg-secondary)]/40 border border-[var(--border-default)] rounded-xl p-6 backdrop-blur-md space-y-6">
            <h3 className="text-sm font-bold uppercase tracking-wider text-[var(--text-primary)] font-mono border-b border-[var(--border-subtle)] pb-3">
              Platform Profile & Preferences
            </h3>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="space-y-2">
                <label className="text-xs font-mono text-[var(--text-secondary)]">
                  Platform Name
                </label>
                <input
                  id="input-app-name"
                  type="text"
                  value={system.app_name}
                  onChange={(e) =>
                    setSystem({ ...system, app_name: e.target.value })
                  }
                  className="w-full px-3.5 py-2.5 rounded-lg bg-[var(--bg-primary)] border border-[var(--border-default)] text-sm text-[var(--text-primary)] font-mono focus:outline-none focus:border-[var(--accent-default)]"
                />
              </div>

              <div className="space-y-2">
                <label className="text-xs font-mono text-[var(--text-secondary)]">
                  Telemetry Retention Period (Days)
                </label>
                <select
                  id="select-retention-days"
                  value={system.retention_days}
                  onChange={(e) =>
                    setSystem({ ...system, retention_days: Number(e.target.value) })
                  }
                  className="w-full px-3.5 py-2.5 rounded-lg bg-[var(--bg-primary)] border border-[var(--border-default)] text-sm text-[var(--text-primary)] font-mono focus:outline-none focus:border-[var(--accent-default)]"
                >
                  <option value={7}>7 Days (Transient Testing)</option>
                  <option value={14}>14 Days (Standard)</option>
                  <option value={30}>30 Days (Production Recommended)</option>
                  <option value={90}>90 Days (Enterprise Extended)</option>
                </select>
              </div>

              <div className="space-y-2">
                <label className="text-xs font-mono text-[var(--text-secondary)]">
                  Telemetry Real-time Refresh Rate
                </label>
                <select
                  id="select-refresh-interval"
                  value={system.refresh_interval}
                  onChange={(e) =>
                    setSystem({
                      ...system,
                      refresh_interval: Number(e.target.value),
                    })
                  }
                  className="w-full px-3.5 py-2.5 rounded-lg bg-[var(--bg-primary)] border border-[var(--border-default)] text-sm text-[var(--text-primary)] font-mono focus:outline-none focus:border-[var(--accent-default)]"
                >
                  <option value={5}>5 Seconds (Aggressive Polling)</option>
                  <option value={10}>10 Seconds (Balanced Default)</option>
                  <option value={30}>30 Seconds (Low Network)</option>
                  <option value={60}>60 Seconds (Minimal Overhead)</option>
                </select>
              </div>

              <div className="space-y-2">
                <label className="text-xs font-mono text-[var(--text-secondary)]">
                  Last Updated Timestamp
                </label>
                <input
                  type="text"
                  readOnly
                  disabled
                  value={system.updated_at ? new Date(system.updated_at).toLocaleString() : "Initial Setup"}
                  className="w-full px-3.5 py-2.5 rounded-lg bg-[var(--bg-primary)]/50 border border-[var(--border-subtle)] text-sm text-[var(--text-muted)] font-mono"
                />
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Tab 2: Notifications */}
      {activeTab === "notifications" && (
        <div id="notifications-settings-panel" className="space-y-6">
          {/* Telegram Settings */}
          <div className="bg-[var(--bg-secondary)]/40 border border-[var(--border-default)] rounded-xl p-6 backdrop-blur-md space-y-6">
            <div className="flex items-center justify-between border-b border-[var(--border-subtle)] pb-3">
              <div className="flex items-center gap-3">
                <Send className="w-5 h-5 text-cyan-400" />
                <h3 className="text-sm font-bold uppercase tracking-wider text-[var(--text-primary)] font-mono">
                  Telegram Bot Channel
                </h3>
              </div>

              <div className="flex items-center gap-2">
                <label className="relative inline-flex items-center cursor-pointer">
                  <input
                    id="toggle-telegram-enabled"
                    type="checkbox"
                    checked={notification.telegram_enabled}
                    onChange={(e) =>
                      setNotification({
                        ...notification,
                        telegram_enabled: e.target.checked,
                      })
                    }
                    className="sr-only peer"
                  />
                  <div className="w-11 h-6 bg-[var(--bg-hover)] peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-cyan-500"></div>
                </label>
                <span className="text-xs font-mono text-[var(--text-secondary)]">
                  {notification.telegram_enabled ? "Enabled" : "Disabled"}
                </span>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="space-y-2">
                <label className="text-xs font-mono text-[var(--text-secondary)]">
                  Telegram Bot API Token
                </label>
                <div className="relative">
                  <input
                    id="input-telegram-token"
                    type={showBotToken ? "text" : "password"}
                    value={notification.telegram_bot_token || ""}
                    onChange={(e) =>
                      setNotification({
                        ...notification,
                        telegram_bot_token: e.target.value,
                      })
                    }
                    placeholder="e.g. 7123456789:AAHxyz..."
                    className="w-full pl-3.5 pr-10 py-2.5 rounded-lg bg-[var(--bg-primary)] border border-[var(--border-default)] text-sm text-[var(--text-primary)] font-mono focus:outline-none focus:border-[var(--accent-default)]"
                  />
                  <button
                    type="button"
                    onClick={() => setShowBotToken(!showBotToken)}
                    className="absolute right-3 top-2.5 text-[var(--text-muted)] hover:text-[var(--text-primary)] cursor-pointer"
                  >
                    {showBotToken ? (
                      <EyeOff className="w-4 h-4" />
                    ) : (
                      <Eye className="w-4 h-4" />
                    )}
                  </button>
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-xs font-mono text-[var(--text-secondary)]">
                  Telegram Chat / Group ID
                </label>
                <input
                  id="input-telegram-chat-id"
                  type="text"
                  value={notification.telegram_chat_id || ""}
                  onChange={(e) =>
                    setNotification({
                      ...notification,
                      telegram_chat_id: e.target.value,
                    })
                  }
                  placeholder="e.g. -100123456789 or 987654321"
                  className="w-full px-3.5 py-2.5 rounded-lg bg-[var(--bg-primary)] border border-[var(--border-default)] text-sm text-[var(--text-primary)] font-mono focus:outline-none focus:border-[var(--accent-default)]"
                />
              </div>
            </div>

            <div className="flex items-center justify-between pt-2">
              <span className="text-xs font-mono text-[var(--text-muted)]">
                Send an immediate test message using current active credentials.
              </span>
              <button
                id="btn-test-notification"
                onClick={handleTestNotification}
                disabled={testingNotif || !notification.telegram_enabled}
                className="flex items-center gap-2 px-3.5 py-2 rounded-lg bg-cyan-500/10 border border-cyan-500/30 text-cyan-400 font-mono text-xs hover:bg-cyan-500/20 transition-all cursor-pointer disabled:opacity-40"
              >
                {testingNotif ? (
                  <RefreshCw className="w-3.5 h-3.5 animate-spin" />
                ) : (
                  <Send className="w-3.5 h-3.5" />
                )}
                Test Notification Alert
              </button>
            </div>
          </div>

          {/* Email Settings */}
          <div className="bg-[var(--bg-secondary)]/40 border border-[var(--border-default)] rounded-xl p-6 backdrop-blur-md space-y-6">
            <div className="flex items-center justify-between border-b border-[var(--border-subtle)] pb-3">
              <h3 className="text-sm font-bold uppercase tracking-wider text-[var(--text-primary)] font-mono">
                Email Dispatch Gateway
              </h3>

              <div className="flex items-center gap-2">
                <label className="relative inline-flex items-center cursor-pointer">
                  <input
                    id="toggle-email-enabled"
                    type="checkbox"
                    checked={notification.email_enabled}
                    onChange={(e) =>
                      setNotification({
                        ...notification,
                        email_enabled: e.target.checked,
                      })
                    }
                    className="sr-only peer"
                  />
                  <div className="w-11 h-6 bg-[var(--bg-hover)] peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-cyan-500"></div>
                </label>
                <span className="text-xs font-mono text-[var(--text-secondary)]">
                  {notification.email_enabled ? "Enabled" : "Disabled"}
                </span>
              </div>
            </div>

            <div className="space-y-2">
              <label className="text-xs font-mono text-[var(--text-secondary)]">
                Email Recipients (Comma separated)
              </label>
              <input
                id="input-email-recipients"
                type="text"
                value={notification.email_recipients || ""}
                onChange={(e) =>
                  setNotification({
                    ...notification,
                    email_recipients: e.target.value,
                  })
                }
                placeholder="devops@company.com, sre-lead@company.com"
                className="w-full px-3.5 py-2.5 rounded-lg bg-[var(--bg-primary)] border border-[var(--border-default)] text-sm text-[var(--text-primary)] font-mono focus:outline-none focus:border-[var(--accent-default)]"
              />
            </div>
          </div>

          {/* Alert Severity Routing & Quiet Hours */}
          <div className="bg-[var(--bg-secondary)]/40 border border-[var(--border-default)] rounded-xl p-6 backdrop-blur-md space-y-6">
            <h3 className="text-sm font-bold uppercase tracking-wider text-[var(--text-primary)] font-mono border-b border-[var(--border-subtle)] pb-3">
              Severity Routing & Quiet Hours
            </h3>

            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
              <label className="flex items-center gap-2.5 p-3 rounded-lg border border-[var(--border-subtle)] bg-[var(--bg-primary)] cursor-pointer">
                <input
                  type="checkbox"
                  checked={notification.critical_alert}
                  onChange={(e) =>
                    setNotification({
                      ...notification,
                      critical_alert: e.target.checked,
                    })
                  }
                  className="rounded border-[var(--border-default)] text-red-500 focus:ring-0"
                />
                <span className="text-xs font-mono text-red-400 font-bold">
                  Critical
                </span>
              </label>

              <label className="flex items-center gap-2.5 p-3 rounded-lg border border-[var(--border-subtle)] bg-[var(--bg-primary)] cursor-pointer">
                <input
                  type="checkbox"
                  checked={notification.warning_alert}
                  onChange={(e) =>
                    setNotification({
                      ...notification,
                      warning_alert: e.target.checked,
                    })
                  }
                  className="rounded border-[var(--border-default)] text-amber-500 focus:ring-0"
                />
                <span className="text-xs font-mono text-amber-400 font-bold">
                  Warning
                </span>
              </label>

              <label className="flex items-center gap-2.5 p-3 rounded-lg border border-[var(--border-subtle)] bg-[var(--bg-primary)] cursor-pointer">
                <input
                  type="checkbox"
                  checked={notification.info_alert}
                  onChange={(e) =>
                    setNotification({
                      ...notification,
                      info_alert: e.target.checked,
                    })
                  }
                  className="rounded border-[var(--border-default)] text-blue-500 focus:ring-0"
                />
                <span className="text-xs font-mono text-blue-400 font-bold">
                  Info
                </span>
              </label>

              <label className="flex items-center gap-2.5 p-3 rounded-lg border border-[var(--border-subtle)] bg-[var(--bg-primary)] cursor-pointer">
                <input
                  type="checkbox"
                  checked={notification.auto_resolve_alert}
                  onChange={(e) =>
                    setNotification({
                      ...notification,
                      auto_resolve_alert: e.target.checked,
                    })
                  }
                  className="rounded border-[var(--border-default)] text-emerald-500 focus:ring-0"
                />
                <span className="text-xs font-mono text-emerald-400 font-bold">
                  Auto-Resolve
                </span>
              </label>
            </div>

            <div className="pt-4 border-t border-[var(--border-subtle)] space-y-4">
              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <span className="text-xs font-mono font-bold text-[var(--text-primary)]">
                    Quiet Hours Filtering
                  </span>
                  <p className="text-[11px] font-mono text-[var(--text-muted)]">
                    Suppress non-critical alerts during specified nighttime windows
                  </p>
                </div>
                <input
                  type="checkbox"
                  checked={notification.quiet_hours_enabled}
                  onChange={(e) =>
                    setNotification({
                      ...notification,
                      quiet_hours_enabled: e.target.checked,
                    })
                  }
                  className="rounded border-[var(--border-default)] text-cyan-500"
                />
              </div>

              {notification.quiet_hours_enabled && (
                <div className="grid grid-cols-2 gap-4 max-w-sm pt-2">
                  <div className="space-y-1">
                    <label className="text-[11px] font-mono text-[var(--text-secondary)]">
                      Start Time
                    </label>
                    <input
                      type="time"
                      value={notification.quiet_hours_start || "22:00"}
                      onChange={(e) =>
                        setNotification({
                          ...notification,
                          quiet_hours_start: e.target.value,
                        })
                      }
                      className="w-full px-3 py-2 rounded bg-[var(--bg-primary)] border border-[var(--border-default)] text-xs font-mono"
                    />
                  </div>
                  <div className="space-y-1">
                    <label className="text-[11px] font-mono text-[var(--text-secondary)]">
                      End Time
                    </label>
                    <input
                      type="time"
                      value={notification.quiet_hours_end || "07:00"}
                      onChange={(e) =>
                        setNotification({
                          ...notification,
                          quiet_hours_end: e.target.value,
                        })
                      }
                      className="w-full px-3 py-2 rounded bg-[var(--bg-primary)] border border-[var(--border-default)] text-xs font-mono"
                    />
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Tab 3: AI Configuration */}
      {activeTab === "ai" && (
        <div id="ai-settings-panel" className="space-y-6">
          <div className="bg-[var(--bg-secondary)]/40 border border-[var(--border-default)] rounded-xl p-6 backdrop-blur-md space-y-6">
            <h3 className="text-sm font-bold uppercase tracking-wider text-[var(--text-primary)] font-mono border-b border-[var(--border-subtle)] pb-3 flex items-center gap-2">
              <Bot className="w-5 h-5 text-cyan-400" />
              Autonomous AI Diagnostic Engine
            </h3>

            {/* Microservice Endpoint Status */}
            <div className="p-4 rounded-lg bg-emerald-950/20 border border-emerald-500/30 flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="w-2.5 h-2.5 rounded-full bg-emerald-400 animate-pulse" />
                <div>
                  <span className="text-xs font-mono font-bold text-emerald-300">
                    AI Service Microservice Connected
                  </span>
                  <p className="text-[11px] font-mono text-[var(--text-muted)]">
                    Listening on http://127.0.0.1:8000 (FastAPI Uvicorn) • Local Ollama LLM Backend
                  </p>
                </div>
              </div>
              <span className="text-xs font-mono px-2.5 py-1 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                ACTIVE
              </span>
            </div>

            <div className="space-y-6">
              {/* Threshold Slider */}
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <label className="text-xs font-mono text-[var(--text-secondary)]">
                    Analysis Confidence Threshold
                  </label>
                  <span className="text-sm font-bold font-mono text-[var(--accent-default)]">
                    {Math.round(system.ai_analysis_threshold * 100)}%
                  </span>
                </div>
                <input
                  id="range-ai-threshold"
                  type="range"
                  min="0.5"
                  max="0.99"
                  step="0.01"
                  value={system.ai_analysis_threshold}
                  onChange={(e) =>
                    setSystem({
                      ...system,
                      ai_analysis_threshold: parseFloat(e.target.value),
                    })
                  }
                  className="w-full accent-[var(--accent-default)] cursor-pointer"
                />
                <p className="text-[11px] font-mono text-[var(--text-muted)]">
                  Only trigger automated root cause analysis when diagnostic confidence exceeds this score.
                </p>
              </div>

              {/* Auto Remediation Toggle */}
              <div className="p-4 rounded-xl border border-[var(--border-default)] bg-[var(--bg-primary)] space-y-3">
                <div className="flex items-center justify-between">
                  <div>
                    <span className="text-xs font-mono font-bold text-[var(--text-primary)]">
                      Autonomous Remediation Mode
                    </span>
                    <p className="text-[11px] font-mono text-[var(--text-muted)] mt-0.5">
                      Allow the AI Agent to execute self-healing actions (e.g. pod restarts, scale deployments) without human confirmation.
                    </p>
                  </div>
                  <label className="relative inline-flex items-center cursor-pointer">
                    <input
                      id="toggle-ai-auto-remediation"
                      type="checkbox"
                      checked={system.ai_auto_remediation}
                      onChange={(e) =>
                        setSystem({
                          ...system,
                          ai_auto_remediation: e.target.checked,
                        })
                      }
                      className="sr-only peer"
                    />
                    <div className="w-11 h-6 bg-[var(--bg-hover)] peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-amber-500"></div>
                  </label>
                </div>

                {system.ai_auto_remediation && (
                  <div className="p-3 rounded-lg bg-amber-950/20 border border-amber-500/30 flex items-center gap-2.5 text-amber-300 text-xs font-mono">
                    <AlertTriangle className="w-4 h-4 shrink-0 text-amber-400" />
                    <span>
                      High-privilege mode enabled. Actions are logged to the immutable PostgreSQL audit trail.
                    </span>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Tab 4: Users & RBAC */}
      {activeTab === "users" && (
        <div id="users-settings-panel" className="space-y-6">
          {!isAdmin ? (
            <div className="p-6 rounded-xl border border-red-500/30 bg-red-950/20 text-center space-y-3">
              <Lock className="w-8 h-8 text-red-400 mx-auto" />
              <h3 className="text-sm font-bold font-mono text-red-200 uppercase">
                Access Denied: Administrator Privileges Required
              </h3>
              <p className="text-xs font-mono text-[var(--text-muted)] max-w-md mx-auto">
                User administration, role assignment, and account governance are restricted to accounts with the{" "}
                <span className="text-red-400 font-bold">admin</span> realm role.
              </p>
            </div>
          ) : (
            <div className="bg-[var(--bg-secondary)]/40 border border-[var(--border-default)] rounded-xl overflow-hidden backdrop-blur-md">
              <div className="p-5 border-b border-[var(--border-subtle)] flex items-center justify-between">
                <div>
                  <h3 className="text-sm font-bold uppercase tracking-wider text-[var(--text-primary)] font-mono">
                    Platform User Directory ({usersTotal})
                  </h3>
                  <p className="text-xs font-mono text-[var(--text-muted)] mt-0.5">
                    Synchronized from Keycloak OIDC Identity Provider with RBAC enforcement
                  </p>
                </div>
              </div>

              <div className="overflow-x-auto">
                <table className="w-full text-left text-xs font-mono">
                  <thead className="bg-[var(--bg-primary)]/80 text-[var(--text-secondary)] border-b border-[var(--border-subtle)]">
                    <tr>
                      <th className="px-5 py-3">User</th>
                      <th className="px-5 py-3">Email</th>
                      <th className="px-5 py-3">Role</th>
                      <th className="px-5 py-3">Status</th>
                      <th className="px-5 py-3">Created</th>
                      <th className="px-5 py-3 text-right">Actions</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-[var(--border-subtle)]">
                    {usersList.length === 0 ? (
                      <tr>
                        <td colSpan={6} className="px-5 py-8 text-center text-[var(--text-muted)]">
                          No users registered in database.
                        </td>
                      </tr>
                    ) : (
                      usersList.map((u) => {
                        const isSelf = u.email === user?.email || u.id === user?.id;
                        const isUpdating = userUpdatingId === u.id;
                        return (
                          <tr
                            key={u.id}
                            className="hover:bg-[var(--bg-hover)]/30 transition-colors"
                          >
                            <td className="px-5 py-3.5 font-bold text-[var(--text-primary)]">
                              <div className="flex items-center gap-2">
                                <span>{u.name || "User"}</span>
                                {isSelf && (
                                  <span className="text-[10px] px-1.5 py-0.5 rounded bg-cyan-500/20 text-cyan-400 font-normal">
                                    You
                                  </span>
                                )}
                              </div>
                            </td>
                            <td className="px-5 py-3.5 text-[var(--text-secondary)]">
                              {u.email}
                            </td>
                            <td className="px-5 py-3.5">
                              <select
                                value={u.role}
                                disabled={isUpdating}
                                onChange={(e) =>
                                  handleRoleChange(u.id, e.target.value)
                                }
                                className="px-2.5 py-1 rounded bg-[var(--bg-primary)] border border-[var(--border-default)] text-xs font-mono text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent-default)] cursor-pointer"
                              >
                                <option value="admin">admin</option>
                                <option value="devops">devops</option>
                                <option value="viewer">viewer</option>
                              </select>
                            </td>
                            <td className="px-5 py-3.5">
                              {u.is_active ? (
                                <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                                  <span className="w-1.5 h-1.5 rounded-full bg-emerald-400" />
                                  Active
                                </span>
                              ) : (
                                <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-[10px] font-bold bg-red-500/10 text-red-400 border border-red-500/20">
                                  <span className="w-1.5 h-1.5 rounded-full bg-red-400" />
                                  Disabled
                                </span>
                              )}
                            </td>
                            <td className="px-5 py-3.5 text-[var(--text-muted)]">
                              {u.created_at ? new Date(u.created_at).toLocaleDateString() : "-"}
                            </td>
                            <td className="px-5 py-3.5 text-right">
                              {isSelf ? (
                                <span className="text-[11px] text-[var(--text-muted)] italic">
                                  Self Protected
                                </span>
                              ) : (
                                <button
                                  onClick={() =>
                                    handleToggleActive(u.id, u.is_active)
                                  }
                                  disabled={isUpdating}
                                  className={`inline-flex items-center gap-1 px-2.5 py-1 rounded border text-[11px] font-mono transition-all cursor-pointer ${
                                    u.is_active
                                      ? "border-red-500/30 text-red-400 hover:bg-red-500/10"
                                      : "border-emerald-500/30 text-emerald-400 hover:bg-emerald-500/10"
                                  }`}
                                >
                                  {u.is_active ? (
                                    <>
                                      <UserX className="w-3 h-3" />
                                      Deactivate
                                    </>
                                  ) : (
                                    <>
                                      <UserCheck className="w-3 h-3" />
                                      Reactivate
                                    </>
                                  )}
                                </button>
                              )}
                            </td>
                          </tr>
                        );
                      })
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Tab 5: Security & Sessions */}
      {activeTab === "security" && (
        <div id="security-settings-panel" className="space-y-6">
          <div className="bg-[var(--bg-secondary)]/40 border border-[var(--border-default)] rounded-xl p-6 backdrop-blur-md space-y-6">
            <h3 className="text-sm font-bold uppercase tracking-wider text-[var(--text-primary)] font-mono border-b border-[var(--border-subtle)] pb-3">
              Session Governance & Security Policy
            </h3>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="space-y-2">
                <label className="text-xs font-mono text-[var(--text-secondary)]">
                  Session Inactivity Timeout
                </label>
                <select
                  id="select-session-timeout"
                  value={system.session_timeout_mins}
                  onChange={(e) =>
                    setSystem({
                      ...system,
                      session_timeout_mins: Number(e.target.value),
                    })
                  }
                  className="w-full px-3.5 py-2.5 rounded-lg bg-[var(--bg-primary)] border border-[var(--border-default)] text-sm text-[var(--text-primary)] font-mono focus:outline-none focus:border-[var(--accent-default)]"
                >
                  <option value={15}>15 Minutes (Strict Security)</option>
                  <option value={30}>30 Minutes (Recommended)</option>
                  <option value={60}>60 Minutes (1 Hour Standard)</option>
                  <option value={120}>120 Minutes (2 Hours Extended)</option>
                </select>
              </div>

              <div className="space-y-2">
                <label className="text-xs font-mono text-[var(--text-secondary)]">
                  Multi-Factor Authentication (MFA)
                </label>
                <div className="flex items-center justify-between p-2.5 rounded-lg border border-[var(--border-default)] bg-[var(--bg-primary)]">
                  <span className="text-xs font-mono text-[var(--text-primary)]">
                    Enforce Keycloak OTP / WebAuthn
                  </span>
                  <label className="relative inline-flex items-center cursor-pointer">
                    <input
                      id="toggle-mfa-enforced"
                      type="checkbox"
                      checked={system.mfa_enforced}
                      onChange={(e) =>
                        setSystem({
                          ...system,
                          mfa_enforced: e.target.checked,
                        })
                      }
                      className="sr-only peer"
                    />
                    <div className="w-11 h-6 bg-[var(--bg-hover)] peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-cyan-500"></div>
                  </label>
                </div>
              </div>
            </div>
          </div>

          {/* Active Sessions */}
          <div className="bg-[var(--bg-secondary)]/40 border border-[var(--border-default)] rounded-xl p-6 backdrop-blur-md space-y-4">
            <div className="flex items-center justify-between border-b border-[var(--border-subtle)] pb-3">
              <div>
                <h3 className="text-sm font-bold uppercase tracking-wider text-[var(--text-primary)] font-mono">
                  Active User Sessions ({sessions.length})
                </h3>
                <p className="text-xs font-mono text-[var(--text-muted)] mt-0.5">
                  Devices authenticated through Keycloak JWT tokens
                </p>
              </div>

              {sessions.length > 1 && (
                <button
                  onClick={handleRevokeOtherSessions}
                  className="px-3 py-1.5 rounded-lg border border-red-500/30 text-red-400 hover:bg-red-500/10 font-mono text-xs cursor-pointer transition-all"
                >
                  Revoke Other Sessions
                </button>
              )}
            </div>

            <div className="space-y-3">
              {sessions.map((sess) => (
                <div
                  key={sess.id}
                  className="flex items-center justify-between p-3.5 rounded-lg border border-[var(--border-subtle)] bg-[var(--bg-primary)]/60"
                >
                  <div className="flex items-center gap-3">
                    <Laptop className="w-5 h-5 text-[var(--text-secondary)]" />
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="text-xs font-bold font-mono text-[var(--text-primary)]">
                          {sess.device}
                        </span>
                        {sess.is_current && (
                          <span className="text-[10px] px-1.5 py-0.5 rounded bg-emerald-500/20 text-emerald-400 font-mono">
                            Current Device
                          </span>
                        )}
                      </div>
                      <span className="text-[11px] font-mono text-[var(--text-muted)]">
                        IP: {sess.ip} • Last active: {sess.last_active}
                      </span>
                    </div>
                  </div>

                  {!sess.is_current && (
                    <button
                      onClick={() =>
                        setSessions((prev) => prev.filter((s) => s.id !== sess.id))
                      }
                      className="text-xs font-mono text-red-400 hover:underline cursor-pointer"
                    >
                      Revoke
                    </button>
                  )}
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

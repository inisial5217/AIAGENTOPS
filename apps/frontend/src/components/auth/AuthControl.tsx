"use client";

import { useEffect, useState } from "react";
import { useAuthStore } from "../../lib/auth";
import { apiClient } from "../../lib/api";
import {
  Shield,
  ShieldAlert,
  ShieldCheck,
  User as UserIcon,
  LogOut,
  Lock,
  CheckCircle2,
  AlertCircle,
  KeyRound,
} from "lucide-react";

export function AuthControl() {
  const { user, token, isLoading, error, login, logout, initAuth } =
    useAuthStore();
  const [testResult, setTestResult] = useState<{
    status: number | null;
    message: string | null;
    isError: boolean;
  } | null>(null);
  const [customUser, setCustomUser] = useState("");
  const [customPass, setCustomPass] = useState("");

  useEffect(() => {
    initAuth();
  }, [initAuth]);

  const handleQuickLogin = async (u: string, p: string) => {
    setTestResult(null);
    await login(u, p);
  };

  const handleCustomLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!customUser || !customPass) return;
    setTestResult(null);
    await login(customUser, customPass);
  };

  const handleTestAdminEndpoint = async () => {
    try {
      const res = await apiClient.get("/api/v1/admin/users");
      setTestResult({
        status: res.status,
        message: `HTTP ${res.status} OK: Successfully fetched ${res.data?.total} registered users from database.`,
        isError: false,
      });
    } catch (err: any) {
      const status = err.response?.status || 500;
      const detail =
        err.response?.data?.detail || err.message || "Request failed";
      setTestResult({
        status,
        message: `HTTP ${status} Forbidden: ${detail}`,
        isError: true,
      });
    }
  };

  const roleColors: Record<string, string> = {
    admin: "bg-red-500/20 text-red-400 border-red-500/30",
    devops: "bg-cyan-500/20 text-cyan-400 border-cyan-500/30",
    viewer: "bg-emerald-500/20 text-emerald-400 border-emerald-500/30",
  };

  return (
    <div className="w-full mt-6 p-6 rounded-xl bg-[var(--bg-card)] border border-[var(--border-default)] shadow-xl text-left">
      <div className="flex items-center justify-between pb-4 border-b border-[var(--border-subtle)]">
        <div className="flex items-center gap-2">
          <KeyRound className="w-5 h-5 text-[var(--accent-default)]" />
          <h2 className="text-base font-semibold text-[var(--text-primary)]">
            Enterprise IAM & RBAC (Keycloak OIDC)
          </h2>
        </div>
        <span
          className={`text-xs px-2.5 py-0.5 rounded-full font-mono border ${
            user
              ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/30"
              : "bg-yellow-500/10 text-yellow-400 border-yellow-500/30"
          }`}
        >
          {user ? "AUTHENTICATED" : "GUEST (UNAUTHENTICATED)"}
        </span>
      </div>

      {error && (
        <div className="mt-4 p-3 rounded-lg bg-red-500/10 border border-red-500/30 text-red-400 text-xs flex items-center gap-2">
          <AlertCircle className="w-4 h-4 shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {user ? (
        <div className="mt-4 space-y-4">
          <div className="flex items-center justify-between p-4 rounded-lg bg-[var(--bg-secondary)] border border-[var(--border-subtle)]">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-[var(--bg-hover)] flex items-center justify-center text-[var(--accent-default)]">
                <UserIcon className="w-5 h-5" />
              </div>
              <div>
                <div className="flex items-center gap-2">
                  <span className="font-medium text-sm text-[var(--text-primary)]">
                    {user.name}
                  </span>
                  <span
                    className={`text-[10px] uppercase font-bold tracking-wider px-2 py-0.5 rounded-md border ${
                      roleColors[user.role] || "bg-gray-800 text-gray-400"
                    }`}
                  >
                    {user.role}
                  </span>
                </div>
                <span className="text-xs text-[var(--text-secondary)] block font-mono">
                  {user.email}
                </span>
                {user.keycloak_id && (
                  <span className="text-[10px] text-[var(--text-muted)] block font-mono mt-0.5">
                    KC-ID: {user.keycloak_id}
                  </span>
                )}
              </div>
            </div>

            <button
              onClick={() => logout()}
              className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium text-red-400 bg-red-500/10 hover:bg-red-500/20 border border-red-500/20 transition-colors"
            >
              <LogOut className="w-3.5 h-3.5" />
              Logout
            </button>
          </div>

          <div className="p-4 rounded-lg bg-[var(--bg-secondary)] border border-[var(--border-subtle)] space-y-3">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-xs font-semibold text-[var(--text-primary)]">
                  Live RBAC Policy Verification
                </h3>
                <p className="text-[11px] text-[var(--text-secondary)]">
                  Verify role permissions against protected Admin endpoint{" "}
                  <code className="text-[var(--accent-default)] font-mono">
                    GET /api/v1/admin/users
                  </code>
                </p>
              </div>
              <button
                onClick={handleTestAdminEndpoint}
                className="px-3 py-1.5 rounded-lg text-xs font-medium bg-[var(--accent-default)] text-[var(--text-inverse)] hover:bg-[var(--accent-hover)] transition-colors shadow-sm"
              >
                Test Admin Access
              </button>
            </div>

            {testResult && (
              <div
                className={`p-3 rounded-md text-xs border flex items-start gap-2 ${
                  testResult.isError
                    ? "bg-red-500/10 border-red-500/30 text-red-400"
                    : "bg-emerald-500/10 border-emerald-500/30 text-emerald-400"
                }`}
              >
                {testResult.isError ? (
                  <ShieldAlert className="w-4 h-4 shrink-0 mt-0.5" />
                ) : (
                  <ShieldCheck className="w-4 h-4 shrink-0 mt-0.5" />
                )}
                <div>
                  <span className="font-semibold block">
                    {testResult.isError
                      ? "RBAC Enforced (Access Denied)"
                      : "Permission Granted"}
                  </span>
                  <span className="font-mono text-[11px]">
                    {testResult.message}
                  </span>
                </div>
              </div>
            )}
          </div>
        </div>
      ) : (
        <div className="mt-4 space-y-4">
          <div>
            <span className="text-xs font-medium text-[var(--text-secondary)] block mb-2">
              Quick Login Demo Profiles (Keycloak Realm: cifo)
            </span>
            <div className="grid grid-cols-3 gap-2">
              <button
                disabled={isLoading}
                onClick={() =>
                  handleQuickLogin("admin@cifo.local", "admin123")
                }
                className="p-3 rounded-lg bg-[var(--bg-secondary)] hover:bg-[var(--bg-hover)] border border-[var(--border-subtle)] text-left transition-all hover:border-red-500/40 group"
              >
                <div className="flex items-center justify-between mb-1">
                  <span className="text-xs font-bold text-red-400">
                    Admin
                  </span>
                  <Shield className="w-3.5 h-3.5 text-red-400" />
                </div>
                <span className="text-[11px] text-[var(--text-muted)] block font-mono truncate">
                  admin@cifo.local
                </span>
                <span className="text-[10px] text-[var(--status-success)] block mt-1 font-mono">
                  Full Access
                </span>
              </button>

              <button
                disabled={isLoading}
                onClick={() =>
                  handleQuickLogin("devops@cifo.local", "devops123")
                }
                className="p-3 rounded-lg bg-[var(--bg-secondary)] hover:bg-[var(--bg-hover)] border border-[var(--border-subtle)] text-left transition-all hover:border-cyan-500/40 group"
              >
                <div className="flex items-center justify-between mb-1">
                  <span className="text-xs font-bold text-cyan-400">
                    DevOps
                  </span>
                  <ShieldCheck className="w-3.5 h-3.5 text-cyan-400" />
                </div>
                <span className="text-[11px] text-[var(--text-muted)] block font-mono truncate">
                  devops@cifo.local
                </span>
                <span className="text-[10px] text-[var(--accent-default)] block mt-1 font-mono">
                  Deploy & Ops
                </span>
              </button>

              <button
                disabled={isLoading}
                onClick={() =>
                  handleQuickLogin("viewer@cifo.local", "viewer123")
                }
                className="p-3 rounded-lg bg-[var(--bg-secondary)] hover:bg-[var(--bg-hover)] border border-[var(--border-subtle)] text-left transition-all hover:border-emerald-500/40 group"
              >
                <div className="flex items-center justify-between mb-1">
                  <span className="text-xs font-bold text-emerald-400">
                    Viewer
                  </span>
                  <UserIcon className="w-3.5 h-3.5 text-emerald-400" />
                </div>
                <span className="text-[11px] text-[var(--text-muted)] block font-mono truncate">
                  viewer@cifo.local
                </span>
                <span className="text-[10px] text-[var(--text-secondary)] block mt-1 font-mono">
                  Read Only
                </span>
              </button>
            </div>
          </div>

          <form
            onSubmit={handleCustomLogin}
            className="pt-3 border-t border-[var(--border-subtle)] flex items-center gap-2"
          >
            <input
              type="text"
              placeholder="Username or email"
              value={customUser}
              onChange={(e) => setCustomUser(e.target.value)}
              className="flex-1 px-3 py-1.5 rounded-lg bg-[var(--bg-secondary)] border border-[var(--border-subtle)] text-xs text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent-default)]"
            />
            <input
              type="password"
              placeholder="Password"
              value={customPass}
              onChange={(e) => setCustomPass(e.target.value)}
              className="flex-1 px-3 py-1.5 rounded-lg bg-[var(--bg-secondary)] border border-[var(--border-subtle)] text-xs text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent-default)]"
            />
            <button
              type="submit"
              disabled={isLoading}
              className="px-4 py-1.5 rounded-lg bg-[var(--accent-default)] text-[var(--text-inverse)] text-xs font-semibold hover:bg-[var(--accent-hover)] transition-colors disabled:opacity-50"
            >
              {isLoading ? "Signing in..." : "Login"}
            </button>
          </form>
        </div>
      )}
    </div>
  );
}

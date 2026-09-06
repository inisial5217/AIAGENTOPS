"use client";

import * as React from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "../../../hooks/use-auth";
import {
  Shield,
  KeyRound,
  ArrowRight,
  AlertCircle,
  Sparkles,
  Lock,
  User,
  ExternalLink,
} from "lucide-react";
import { Button } from "../../../components/ui/button";
import { Input } from "../../../components/ui/input";
import { Badge } from "../../../components/ui/badge";

export default function LoginPage() {
  const router = useRouter();
  const { login, loginWithKeycloak, isLoading, error, isAuthenticated } =
    useAuth();

  const [username, setUsername] = React.useState("admin@cifo.local");
  const [password, setPassword] = React.useState("admin123");
  const [activeProfile, setActiveProfile] = React.useState<string | null>("Admin");

  // If already authenticated, redirect to monitoring
  React.useEffect(() => {
    if (isAuthenticated) {
      router.push("/monitoring");
    }
  }, [isAuthenticated, router]);

  const handleSubmit = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    const success = await login(username, password);
    if (success) {
      router.push("/monitoring");
    }
  };

  const handleDemoProfile = async (
    roleUser: string,
    rolePass: string,
    label: string
  ) => {
    setUsername(roleUser);
    setPassword(rolePass);
    setActiveProfile(label);
    const success = await login(roleUser, rolePass);
    if (success) {
      router.push("/monitoring");
    }
  };

  return (
    <div className="min-h-screen bg-[var(--bg-primary)] flex flex-col justify-center items-center p-4 relative overflow-hidden">
      {/* Background Decorative Neon Glows */}
      <div className="absolute top-1/4 -left-32 w-96 h-96 bg-pink-500/10 rounded-full blur-3xl pointer-events-none" />
      <div className="absolute bottom-1/4 -right-32 w-96 h-96 bg-cyan-500/10 rounded-full blur-3xl pointer-events-none" />

      <div className="w-full max-w-md bg-[var(--bg-card)] border border-[var(--border-default)] rounded-2xl p-8 shadow-2xl relative z-10 text-center">
        {/* Brand Logo & Title */}
        <div className="flex flex-col items-center mb-6">
          <div className="relative flex items-center justify-center w-12 h-12 rounded-xl bg-gradient-to-br from-pink-500/20 via-cyan-500/20 to-transparent border border-pink-500/40 mb-3 shadow-lg shadow-pink-500/10">
            <span className="text-lg font-black tracking-tighter text-white">
              C<span className="text-[var(--accent-pink)]">I</span>F<span className="text-[var(--accent-default)]">O</span>
            </span>
            <span className="absolute -top-1 -right-1 w-2.5 h-2.5 rounded-full bg-[var(--accent-pink)] animate-ping" />
          </div>
          <h1 className="text-xl font-bold tracking-tight text-[var(--text-primary)]">
            CIFO Command Center
          </h1>
          <p className="text-xs text-[var(--text-muted)] font-mono mt-1">
            Enterprise Multi-Cluster AIOps Platform
          </p>
        </div>

        {/* 1-Click Demo Profiles */}
        <div className="mb-6 text-left">
          <div className="flex items-center justify-between mb-2">
            <span className="text-[11px] font-mono text-[var(--text-muted)] uppercase tracking-wider">
              1-Click Demo Profiles
            </span>
            <Badge variant="cyan" size="sm">
              <Sparkles className="w-2.5 h-2.5" /> Instant Access
            </Badge>
          </div>

          <div className="grid grid-cols-3 gap-2">
            <button
              type="button"
              onClick={() =>
                handleDemoProfile("admin@cifo.local", "admin123", "Admin")
              }
              className={`p-2 rounded-lg border text-left transition-all cursor-pointer ${
                username === "admin@cifo.local"
                  ? "border-pink-500/50 bg-pink-500/10 text-pink-300 shadow-sm"
                  : "border-[var(--border-subtle)] bg-[var(--bg-secondary)] text-[var(--text-secondary)] hover:border-[var(--border-default)]"
              }`}
            >
              <div className="text-[11px] font-bold">Admin</div>
              <div className="text-[9px] font-mono text-[var(--text-muted)]">
                Full Control
              </div>
            </button>

            <button
              type="button"
              onClick={() =>
                handleDemoProfile("devops@cifo.local", "devops123", "DevOps")
              }
              className={`p-2 rounded-lg border text-left transition-all cursor-pointer ${
                username === "devops@cifo.local"
                  ? "border-cyan-500/50 bg-cyan-500/10 text-cyan-300 shadow-sm"
                  : "border-[var(--border-subtle)] bg-[var(--bg-secondary)] text-[var(--text-secondary)] hover:border-[var(--border-default)]"
              }`}
            >
              <div className="text-[11px] font-bold">DevOps</div>
              <div className="text-[9px] font-mono text-[var(--text-muted)]">
                Remediator
              </div>
            </button>

            <button
              type="button"
              onClick={() =>
                handleDemoProfile("viewer@cifo.local", "viewer123", "Viewer")
              }
              className={`p-2 rounded-lg border text-left transition-all cursor-pointer ${
                username === "viewer@cifo.local"
                  ? "border-emerald-500/50 bg-emerald-500/10 text-emerald-300 shadow-sm"
                  : "border-[var(--border-subtle)] bg-[var(--bg-secondary)] text-[var(--text-secondary)] hover:border-[var(--border-default)]"
              }`}
            >
              <div className="text-[11px] font-bold">Viewer</div>
              <div className="text-[9px] font-mono text-[var(--text-muted)]">
                Read-Only
              </div>
            </button>
          </div>
        </div>

        {/* Error Alert */}
        {error && (
          <div className="mb-4 p-3 rounded-lg bg-red-500/10 border border-red-500/30 text-red-400 text-xs flex items-center gap-2 text-left">
            <AlertCircle className="w-4 h-4 shrink-0" />
            <span>{error}</span>
          </div>
        )}

        {/* Direct Login Form */}
        <form onSubmit={handleSubmit} className="space-y-4 text-left">
          <Input
            label="Username / ID"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            placeholder="e.g. admin"
            prefix={<User className="w-3.5 h-3.5" />}
            required
          />

          <Input
            label="Password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="••••••••"
            prefix={<Lock className="w-3.5 h-3.5" />}
            required
          />

          <Button
            type="submit"
            variant="primary"
            className="w-full justify-center h-10 font-bold"
            isLoading={isLoading}
          >
            <span>Sign In to Dashboard</span>
            <ArrowRight className="w-4 h-4" />
          </Button>
        </form>

        {/* Enterprise SSO Divider */}
        <div className="relative my-6">
          <div className="absolute inset-0 flex items-center">
            <div className="w-full border-t border-[var(--border-subtle)]" />
          </div>
          <div className="relative flex justify-center text-[10px] uppercase font-mono">
            <span className="bg-[var(--bg-card)] px-2 text-[var(--text-muted)]">
              Or Enterprise Identity
            </span>
          </div>
        </div>

        {/* Keycloak SSO Button */}
        <Button
          type="button"
          variant="outline"
          onClick={loginWithKeycloak}
          className="w-full justify-center h-10 font-medium text-xs border-[var(--border-default)] hover:border-cyan-500/50"
        >
          <Shield className="w-4 h-4 text-cyan-400" />
          <span>Single Sign-On (Keycloak OIDC)</span>
          <ExternalLink className="w-3 h-3 text-[var(--text-muted)]" />
        </Button>

        {/* Footer info */}
        <div className="mt-6 text-center text-[10px] font-mono text-[var(--text-muted)]">
          Protected by Keycloak JWT &bull; TLS 1.3 &bull; CIFO v2.1.0
        </div>
      </div>
    </div>
  );
}

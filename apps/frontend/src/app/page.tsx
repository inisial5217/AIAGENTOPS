import { AuthControl } from "../components/auth/AuthControl";

export default function HomePage() {
  return (
    <main className="min-h-screen flex flex-col items-center justify-center p-6 text-center">
      <div className="max-w-2xl w-full bg-[var(--bg-card)] border border-[var(--border-default)] rounded-xl p-8 shadow-2xl">
        <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full text-xs font-medium bg-[var(--accent-glow)] text-[var(--accent-default)] border border-[var(--accent-default)]/30 mb-6">
          <span className="w-2 h-2 rounded-full bg-[var(--accent-default)] animate-pulse" />
          CIFO Enterprise AIOps Platform
        </div>
        <h1 className="text-3xl font-bold tracking-tight text-[var(--text-primary)] mb-3">
          Command Center Initialized
        </h1>
        <p className="text-[var(--text-secondary)] text-sm mb-6 leading-relaxed">
          Platform pemantauan real-time untuk Docker, ArgoCD, dan Kubernetes dengan asistensi otonom AI ops.
        </p>

        <div className="grid grid-cols-4 gap-3 pt-4 border-t border-[var(--border-default)] text-left">
          <div className="p-3 rounded-lg bg-[var(--bg-secondary)] border border-[var(--border-subtle)]">
            <span className="text-xs text-[var(--text-muted)] block">Backend Go</span>
            <span className="text-sm font-semibold text-[var(--status-success)]">Echo v4 Ready</span>
          </div>
          <div className="p-3 rounded-lg bg-[var(--bg-secondary)] border border-[var(--border-subtle)]">
            <span className="text-xs text-[var(--text-muted)] block">Keycloak IAM</span>
            <span className="text-sm font-semibold text-[var(--accent-default)]">OIDC / JWT</span>
          </div>
          <div className="p-3 rounded-lg bg-[var(--bg-secondary)] border border-[var(--border-subtle)]">
            <span className="text-xs text-[var(--text-muted)] block">AI Service</span>
            <span className="text-sm font-semibold text-[var(--status-warning)]">Python 3.12</span>
          </div>
          <div className="p-3 rounded-lg bg-[var(--bg-secondary)] border border-[var(--border-subtle)]">
            <span className="text-xs text-[var(--text-muted)] block">Frontend</span>
            <span className="text-sm font-semibold text-[var(--status-info)]">Next.js 16</span>
          </div>
        </div>

        <AuthControl />
      </div>
    </main>
  );
}

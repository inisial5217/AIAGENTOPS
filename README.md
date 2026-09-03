# CIFO Enterprise Monitoring & AI Ops Platform

Enterprise infrastructure monitoring platform for **ArgoCD (Kubernetes)** and **Docker**, featuring an autonomous multi-model AIOps assistant (Gemini, OpenAI, Claude, Ollama) and Telegram incident alerting.

## Technology Stack

- **Frontend**: Next.js 14 (App Router, TypeScript Strict), Tailwind CSS, Radix UI, Zustand, TanStack Query v5, Apache ECharts v5, WebSocket.
- **Backend**: Go 1.22+ (Echo v4, pgxpool, go-redis, slog, Asynq, gRPC).
- **AI Service**: Python 3.12 (FastAPI, LangChain, Multi-Model Circuit Breaker, Zero-Trust Human-in-the-Loop Tools).
- **Databases & Observability**: PostgreSQL 16, Redis 7, VictoriaMetrics, Grafana Loki, Grafana Tempo, Prometheus, Alertmanager.
- **DevOps**: Docker, K3d (local K8s), ArgoCD, Helm, GitHub Actions.

## Directory Structure

```text
/apps
  /frontend          # Next.js 14 web app
  /backend           # Go 1.22+ core services
  /ai-service        # Python 3.12 AIOps agent
/packages
  /api-contracts     # OpenAPI & Protobuf specifications
  /theme             # Centralized design tokens (colors, typography)
/infrastructure
  /local-testbed     # Docker compose & K3d manifests
  /kubernetes        # Production Helm charts
/tests
  /integration       # Integration test suites
  /e2e               # Playwright end-to-end tests
  /load              # K6 performance tests
/scripts             # Bootstrap & operational scripts
```

## Quickstart

1. **Install Prerequisites**:
   - Docker & Docker Compose
   - Go 1.22+
   - Node.js 20+
   - Python 3.12+
   - K3d & kubectl

2. **Start Infrastructure**:
   ```bash
   make dev-infra
   ```

3. **Run Services**:
   ```bash
   make dev-backend   # Port 8080
   make dev-frontend  # Port 3000
   make dev-ai        # Port 50051 / 8000
   ```

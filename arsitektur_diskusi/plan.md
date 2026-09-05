# CIFO Monitoring Platform - Master Implementation Plan
# Status: DRAFT - Menunggu Persetujuan
# Terakhir diperbarui: 2026-09-03
# Referensi: arsitektur_sistem.md v2.0, audit_arsitektur.md v2.0

Dokumen ini adalah rencana implementasi end-to-end untuk membangun platform monitoring enterprise CIFO. Setiap fase disusun berdasarkan urutan dependensi. Tidak ada fase yang boleh dimulai sebelum prasyarat dari fase sebelumnya terpenuhi. Setiap tugas memiliki kriteria selesai yang harus dipenuhi tanpa pengecualian.

---

## ATURAN EKSEKUSI

1. Setiap fase harus diselesaikan secara berurutan (Fase 0 sebelum Fase 1, dst). Sub-tugas dalam satu fase boleh dikerjakan paralel jika tidak ada dependensi silang.
2. Tidak ada tugas yang boleh ditandai selesai tanpa memenuhi seluruh kriteria penerimaan (Acceptance Criteria).
3. Tidak ada data dummy atau mock data. Semua data harus berasal dari sumber nyata (Docker daemon lokal, K3d cluster, ArgoCD lokal).
4. Tidak ada credential yang di-hardcode. Gunakan environment variables atau Vault sejak awal.
5. Semua kode harus lolos linting sebelum dianggap selesai.
6. Komentar kode: 1-4 kata, tanpa simbol dekoratif.

---

## FASE 0: BOOTSTRAP PROYEK & MONOREPO SCAFFOLDING
**Tujuan**: Membuat seluruh struktur direktori monorepo, menginisialisasi setiap layanan (frontend, backend, ai-service), dan mengkonfigurasi tooling dasar (linting, formatting, git hooks).
**Prasyarat**: Tidak ada.
**Estimasi**: 1-2 hari.

### Tugas 0.1: Buat Struktur Direktori Root
Buat seluruh tree direktori sesuai arsitektur_sistem.md Bagian 9. Ini mencakup:
- `/apps/frontend`
- `/apps/backend`
- `/apps/ai-service`
- `/packages/api-contracts`
- `/packages/eslint-config`
- `/packages/theme`
- `/infrastructure/local-testbed`
- `/infrastructure/local-testbed/k3d`
- `/infrastructure/local-testbed/argocd`
- `/infrastructure/local-testbed/prometheus`
- `/infrastructure/local-testbed/alertmanager`
- `/infrastructure/kubernetes/charts`
- `/infrastructure/kubernetes/base`
- `/infrastructure/kubernetes/overlays/staging`
- `/infrastructure/kubernetes/overlays/production`
- `/infrastructure/terraform/modules`
- `/infrastructure/terraform/environments`
- `/infrastructure/security/network-policies`
- `/infrastructure/security/rbac`
- `/infrastructure/security/vault`
- `/tests/integration`
- `/tests/e2e`
- `/tests/load`
- `/scripts`
- `/docs`
- `/docs/adr`
- `.github/workflows`

### Tugas 0.2: Inisialisasi Git Repository
- `git init` di root
- Buat `.gitignore` yang mencakup: node_modules, dist, .next, __pycache__, .env, *.log, vendor/, bin/, coverage/, .vscode/ (kecuali settings.json), .idea/
- Buat `.editorconfig`:
  - root = true
  - charset = utf-8
  - indent_style = space
  - indent_size = 2 (yaml, json, ts, tsx, css)
  - indent_size = 4 (go, py)
  - end_of_line = lf
  - insert_final_newline = true
  - trim_trailing_whitespace = true
- Commit awal: `chore: init monorepo structure`

### Tugas 0.3: Inisialisasi Backend Go
- Jalankan `go mod init github.com/cifo-monitoring/backend` di `/apps/backend`
- Buat `go.work` di root monorepo yang mereferensi `./apps/backend` dan `./tests/integration`
- Install dependensi awal:
  - `github.com/labstack/echo/v4` (web framework)
  - `github.com/jackc/pgx/v5` (PostgreSQL driver + pool)
  - `github.com/redis/go-redis/v9` (Redis client)
  - `github.com/go-playground/validator/v10` (input validation)
  - `github.com/stretchr/testify` (testing)
  - `log/slog` (built-in Go 1.22+, structured logging)
  - `golang.org/x/sync` (errgroup untuk graceful shutdown)
- Buat file entrypoint kosong: `/apps/backend/cmd/server/main.go` dan `/apps/backend/cmd/worker/main.go`
- Buat file stub untuk setiap package di `/internal`: config, handler, middleware, model, repository, service, integration, ws
- Buat `/apps/backend/pkg/apperror/error.go` dengan definisi `AppError` struct
- Buat `/apps/backend/pkg/logger/logger.go` dengan wrapper `slog` dasar
- Buat `/apps/backend/pkg/validator/validator.go` dengan fungsi validasi generik
- Pastikan `go build ./...` berhasil tanpa error

### Tugas 0.4: Inisialisasi Frontend Next.js
- Jalankan `npx -y create-next-app@latest ./` di `/apps/frontend` dengan opsi:
  - TypeScript: yes
  - Tailwind CSS: yes
  - ESLint: yes
  - App Router: yes
  - src directory: yes
  - Import alias: @/*
- Install dependensi tambahan:
  - `zustand` (state management)
  - `@tanstack/react-query` (server state)
  - `axios` (HTTP client)
  - `echarts` + `echarts-for-react` (data visualization)
  - `@radix-ui/react-dialog` (modal)
  - `@radix-ui/react-dropdown-menu` (dropdown)
  - `@radix-ui/react-toast` (toast notification)
  - `@radix-ui/react-tooltip` (tooltip)
  - `clsx` (conditional classnames)
  - `date-fns` (date formatting)
- Install dev dependencies:
  - `vitest` + `@testing-library/react` + `@testing-library/jest-dom`
  - `prettier` + `prettier-plugin-tailwindcss`
- Buat `tailwind.config.ts` dengan konfigurasi tema dark mode (class strategy)
- Buat `/src/styles/globals.css` dengan CSS custom properties untuk dark/light theme
- Buat `/src/styles/tokens.css` dengan design tokens (warna, spacing, radius, font)
- Hapus semua boilerplate default Next.js (sample page, icon, dll)
- Pastikan `npm run build` berhasil tanpa error

### Tugas 0.5: Inisialisasi AI Service Python
- Buat `/apps/ai-service/pyproject.toml` dengan konfigurasi project metadata
- Buat `/apps/ai-service/requirements.lock` dengan dependensi:
  - `fastapi` + `uvicorn[standard]`
  - `grpcio` + `grpcio-tools` + `protobuf`
  - `langchain` + `langchain-google-genai` + `langchain-openai` + `langchain-anthropic`
  - `pydantic` + `pydantic-settings`
  - `httpx` (async HTTP client)
  - `ruff` (linting + formatting)
  - `pytest` + `pytest-asyncio`
- Buat file stub: `main.py`, dan sub-folder agent, tools, providers, prompts, config
- Buat `/apps/ai-service/app/config/settings.py` dengan Pydantic Settings class (env-based config)
- Pastikan struktur dapat di-import tanpa error

### Tugas 0.6: Buat File Konfigurasi Root
- `Makefile` dengan target:
  - `make dev-backend`: jalankan backend Go dengan hot reload (menggunakan `air`)
  - `make dev-frontend`: jalankan Next.js dev server
  - `make dev-ai`: jalankan FastAPI dev server
  - `make dev-infra`: jalankan docker-compose local-testbed
  - `make dev`: jalankan semua di atas secara paralel
  - `make test`: jalankan semua unit test (Go + Frontend + Python)
  - `make lint`: jalankan semua linter
  - `make build`: build semua Docker images
  - `make migrate-up`: jalankan database migration
  - `make migrate-down`: rollback migration terakhir
  - `make seed`: jalankan seed data
  - `make generate-types`: generate Go/TS types dari OpenAPI spec
- `turbo.json` dengan pipeline: build, test, lint, dev
- `README.md` dengan:
  - Deskripsi proyek
  - Prasyarat (Go 1.22+, Node 20+, Python 3.12+, Docker, K3d)
  - Langkah quickstart
  - Referensi ke arsitektur dan dokumentasi

### Tugas 0.7: Buat API Contract Awal
- Buat `/packages/api-contracts/openapi.yaml` dengan:
  - Info section (title, version, description)
  - Server definitions (localhost:8080 untuk development)
  - Security schemes (Bearer JWT)
  - Placeholder paths: /healthz, /readyz, /api/v1/auth/login
  - Komponen schema dasar: User, ApiError, HealthResponse
- Buat `/packages/api-contracts/proto/common.proto` dengan message types dasar
- Buat `/packages/api-contracts/proto/ai_service.proto` dengan service definition placeholder

### Tugas 0.8: Buat Design Token & Theme
- Buat `/packages/theme/colors.json`:
  - dark mode colors (background, surface, text, border, accent shades)
  - light mode colors
  - semantic colors (success, warning, error, critical, info)
  - Warna harus sesuai referensi tampilan (dark theme, cyan/teal accent, merah untuk critical)
- Buat `/packages/theme/typography.json`:
  - Font family: Inter (Google Fonts)
  - Font sizes: xs (12px), sm (14px), base (16px), lg (18px), xl (20px), 2xl (24px), 3xl (30px), 4xl (36px)
  - Font weights: regular (400), medium (500), semibold (600), bold (700)

### Kriteria Penerimaan Fase 0:
- [ ] Seluruh struktur direktori sesuai arsitektur_sistem.md Bagian 9 telah dibuat
- [ ] `go build ./...` di backend berhasil tanpa error
- [ ] `npm run build` di frontend berhasil tanpa error
- [ ] Python imports di ai-service tidak error
- [ ] `.gitignore`, `.editorconfig`, `Makefile`, `turbo.json`, `README.md` ada di root
- [ ] API contract awal (OpenAPI + Proto) telah dibuat
- [ ] Design tokens telah didefinisikan
- [ ] Commit awal telah dibuat

---

## FASE 1: INFRASTRUKTUR LOKAL (LOCAL TESTBED)
**Tujuan**: Menyiapkan seluruh dependensi infrastruktur (PostgreSQL, Redis, Prometheus, VictoriaMetrics, Loki, Grafana Tempo, Alertmanager) dan cluster Kubernetes lokal (K3d + ArgoCD) agar pengembangan dapat menggunakan data nyata.
**Prasyarat**: Fase 0 selesai. Docker Desktop terinstal dan berjalan.
**Estimasi**: 1-2 hari.

### Tugas 1.1: Docker Compose untuk Data Services
Buat `/infrastructure/local-testbed/docker-compose.yml`:
- **PostgreSQL 16**: port 5432, volume persisten, healthcheck, env POSTGRES_DB=cifo_db, POSTGRES_USER=cifo_admin, POSTGRES_PASSWORD (dari .env file)
- **Redis 7**: port 6379, requirepass dari .env, healthcheck, volume persisten
- **VictoriaMetrics**: port 8428, volume persisten, retention period 90d, konfigurasi untuk menerima remote write dari Prometheus
- **Prometheus**: port 9090, mount config file prometheus.yml, mount alert-rules.yml, konfigurasi remote_write ke VictoriaMetrics
- **Grafana Loki**: port 3100, konfigurasi local storage (untuk development, bukan S3)
- **Grafana Tempo**: port 3200, konfigurasi local storage
- Network: semua service dalam satu bridge network `cifo-network`
- .env.example file dengan semua variabel yang dibutuhkan (tanpa value sensitif)

### Tugas 1.2: Docker Compose untuk Monitoring Stack
Buat `/infrastructure/local-testbed/docker-compose.monitoring.yml`:
- **Alertmanager**: port 9093, mount alertmanager.yml, konfigurasi webhook receiver ke backend (http://host.docker.internal:8080/api/v1/webhooks/alertmanager)
- **Grafana** (opsional, untuk debugging): port 3000, datasource auto-provisioning (VictoriaMetrics, Loki, Tempo)

### Tugas 1.3: Konfigurasi Prometheus
Buat `/infrastructure/local-testbed/prometheus/prometheus.yml`:
- global scrape_interval: 15s
- scrape_configs:
  - job `cifo-backend`: target localhost:8080/metrics
  - job `docker`: target menggunakan Docker daemon metrics endpoint
  - job `node`: target node exporter (jika ada)
- remote_write ke VictoriaMetrics (http://victoriametrics:8428/api/v1/write)
- rule_files: alert-rules.yml

Buat `/infrastructure/local-testbed/prometheus/alert-rules.yml`:
- Semua 12 alert rules dari arsitektur_sistem.md Bagian 5.4
- Setiap rule harus memiliki: alert name, expr (PromQL), for (duration), labels (severity), annotations (summary, description)

### Tugas 1.4: Konfigurasi Alertmanager
Buat `/infrastructure/local-testbed/alertmanager/alertmanager.yml`:
- route: group_by [alertname, severity], group_wait 30s, group_interval 5m, repeat_interval 4h
- receiver default: webhook ke backend
- receiver critical: webhook ke backend dengan route khusus (group_wait 10s, repeat_interval 1h)

### Tugas 1.5: Setup K3d Cluster
Buat `/infrastructure/local-testbed/k3d/cluster-config.yaml`:
- Cluster name: cifo-dev
- 1 server node, 2 agent nodes
- Port mapping: 6443 (K8s API), 8443 (ingress HTTPS), 8081 (ingress HTTP)
- Volume mount untuk persistent storage
- Disable traefik (akan install sendiri jika perlu, atau gunakan built-in)

Buat `/infrastructure/local-testbed/k3d/setup-cluster.sh`:
- Periksa apakah k3d terinstal, jika tidak beri instruksi instalasi
- Hapus cluster lama jika ada: `k3d cluster delete cifo-dev`
- Buat cluster baru dari config: `k3d cluster create --config cluster-config.yaml`
- Tunggu cluster ready: `kubectl wait --for=condition=Ready nodes --all --timeout=120s`
- Install ArgoCD: `kubectl create namespace argocd && kubectl apply -n argocd -f ../argocd/install.yaml`
- Tunggu ArgoCD ready: `kubectl wait --for=condition=Available deployment/argocd-server -n argocd --timeout=300s`
- Print ArgoCD initial admin password: `kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d`
- Print akses instruksi

### Tugas 1.6: Setup ArgoCD Lokal
Buat `/infrastructure/local-testbed/argocd/install.yaml`:
- Gunakan ArgoCD stable manifest dari GitHub releases resmi
- Atau referensi URL resmi yang akan di-apply langsung

Buat `/infrastructure/local-testbed/argocd/sample-apps/`:
- Buat 2-3 aplikasi Kubernetes sederhana (nginx, httpbin) sebagai target monitoring:
  - `sample-nginx.yaml`: Deployment + Service nginx dengan 2 replika
  - `sample-httpbin.yaml`: Deployment + Service httpbin dengan 1 replika
  - `argocd-app-nginx.yaml`: ArgoCD Application resource yang mengarah ke sample-nginx
  - `argocd-app-httpbin.yaml`: ArgoCD Application resource yang mengarah ke sample-httpbin
- Aplikasi ini akan menjadi objek pemantauan nyata di dashboard kita

### Tugas 1.7: Script Bootstrap Lokal
Buat `/scripts/setup-local.sh` (dan `/scripts/setup-local.ps1` untuk Windows):
- Periksa prasyarat: Docker, Go, Node, Python, K3d, kubectl
- Salin .env.example ke .env jika belum ada
- Jalankan docker-compose up -d
- Tunggu PostgreSQL ready (pg_isready)
- Jalankan database migration (make migrate-up)
- Jalankan seed data (make seed)
- Jalankan setup K3d cluster
- Print status semua layanan

### Tugas 1.8: Seed Data SQL
Buat `/scripts/seed-data.sql`:
- Insert default roles: admin, devops, viewer
- Insert default alert rule configurations
- Insert default system settings (Telegram webhook URL placeholder, AI model preferences)
- Tidak ada data pengguna palsu. Hanya konfigurasi sistem wajib.

### Kriteria Penerimaan Fase 1:
- [x] `docker-compose up -d` berhasil menjalankan semua service (Postgres, Redis, VictoriaMetrics, Prometheus, Loki, Tempo, Alertmanager)
- [x] PostgreSQL dapat diakses di localhost:5432 dan menerima koneksi
- [x] Redis dapat diakses di localhost:6379 dan merespons PING
- [x] Prometheus UI accessible di localhost:9090 dan menampilkan target
- [x] VictoriaMetrics menerima data dari Prometheus remote write
- [x] K3d cluster berjalan: `kubectl get nodes` menampilkan 3 node Ready
- [x] ArgoCD berjalan: `kubectl get pods -n argocd` semua Running
- [x] Sample apps ter-deploy dan ArgoCD menunjukkan status Synced/Healthy
- [x] Alert rules terdaftar di Prometheus UI (/alerts)
- [x] setup-local.sh (atau .ps1) berjalan end-to-end tanpa error

---

## FASE 2: FONDASI BACKEND
**Tujuan**: Membangun kerangka backend Go yang lengkap: server HTTP, koneksi database, migration, middleware, health checks, structured logging, dan konfigurasi. Setelah fase ini, backend dapat menerima request HTTP, terkoneksi ke semua dependensi, dan menjalankan migration.
**Prasyarat**: Fase 0 dan Fase 1 selesai. Semua dependensi lokal berjalan.
**Estimasi**: 3-4 hari.

### Tugas 2.1: Konfigurasi Aplikasi
Implementasi `/apps/backend/internal/config/config.go`:
- Struct `Config` dengan semua field konfigurasi: server port, database DSN, Redis address, Docker host, ArgoCD URL, Telegram token, AI service address, log level
- Parsing dari environment variables menggunakan `os.Getenv` dengan default values
- Validasi wajib: database DSN dan Redis address harus ada, jika tidak aplikasi gagal start
- Fungsi `Load() (*Config, error)` yang dipanggil dari main.go

### Tugas 2.2: Package Logger
Implementasi `/apps/backend/pkg/logger/logger.go`:
- Wrapper tipis di atas `slog` bawaan Go
- Fungsi `New(level string) *slog.Logger` yang mengembalikan JSON handler
- Default fields: service=cifo-backend
- Fungsi helper untuk menambahkan trace_id dan span_id ke context

### Tugas 2.3: Package AppError
Implementasi `/apps/backend/pkg/apperror/error.go`:
- Struct `AppError` dengan field: Code, HTTPStatus, Message, UserMsg, Err
- Method `Error() string` untuk memenuhi interface error
- Method `Unwrap() error` untuk error chaining
- Constructor functions: `New(code string, status int, msg string, userMsg string) *AppError`
- Predefined errors: ErrUnauthorized, ErrForbidden, ErrNotFound, ErrInternal, ErrBadRequest, ErrRateLimit
- Fungsi `FromError(err error) *AppError` untuk mengkonversi error biasa menjadi AppError

### Tugas 2.4: Database Connection Pool
Implementasi koneksi PostgreSQL di `/apps/backend/internal/repository/`:
- Buat file `db.go` dengan fungsi `NewPostgresPool(dsn string) (*pgxpool.Pool, error)`
- Konfigurasi pool: MinConns=5, MaxConns=25, MaxConnLifetime=1h, MaxConnIdleTime=15m, HealthCheckPeriod=30s
- Logging koneksi menggunakan slog
- Fungsi `Close()` untuk graceful shutdown

Implementasi koneksi Redis:
- Buat file `redis.go` dengan fungsi `NewRedisClient(addr string, password string) (*redis.Client, error)`
- Healthcheck: PING saat inisialisasi

### Tugas 2.5: Database Migrations
Buat semua migration files di `/apps/backend/migrations/`:
- `001_create_users.up.sql`:
  ```sql
  CREATE TABLE users (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      email VARCHAR(255) UNIQUE NOT NULL,
      name VARCHAR(255) NOT NULL,
      role VARCHAR(50) NOT NULL DEFAULT 'viewer',
      keycloak_id VARCHAR(255) UNIQUE,
      is_active BOOLEAN NOT NULL DEFAULT true,
      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );
  CREATE INDEX idx_users_email ON users(email);
  CREATE INDEX idx_users_keycloak_id ON users(keycloak_id);
  ```
- `001_create_users.down.sql`: DROP TABLE users
- `002_create_ai_sessions.up.sql`:
  ```sql
  CREATE TABLE ai_sessions (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      user_id UUID NOT NULL REFERENCES users(id),
      status VARCHAR(20) NOT NULL DEFAULT 'active',
      model_preference VARCHAR(50),
      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      last_activity_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );
  CREATE INDEX idx_ai_sessions_user_id ON ai_sessions(user_id);
  CREATE INDEX idx_ai_sessions_status ON ai_sessions(status);

  CREATE TABLE ai_messages (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      session_id UUID NOT NULL REFERENCES ai_sessions(id) ON DELETE CASCADE,
      role VARCHAR(20) NOT NULL,
      content TEXT NOT NULL,
      model_used VARCHAR(50),
      input_tokens INTEGER DEFAULT 0,
      output_tokens INTEGER DEFAULT 0,
      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );
  CREATE INDEX idx_ai_messages_session_id ON ai_messages(session_id);
  ```
- `002_create_ai_sessions.down.sql`: DROP TABLE ai_messages; DROP TABLE ai_sessions
- `003_create_audit_log.up.sql`:
  ```sql
  CREATE TABLE audit_log (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      actor_type VARCHAR(20) NOT NULL,
      actor_id VARCHAR(255) NOT NULL,
      action VARCHAR(50) NOT NULL,
      resource_type VARCHAR(50) NOT NULL,
      resource_id VARCHAR(255),
      details JSONB,
      ip_address INET,
      user_agent TEXT,
      result VARCHAR(20) NOT NULL DEFAULT 'success'
  );
  CREATE INDEX idx_audit_log_timestamp ON audit_log(timestamp);
  CREATE INDEX idx_audit_log_actor_id ON audit_log(actor_id);
  CREATE INDEX idx_audit_log_action ON audit_log(action);

  CREATE TABLE ai_action_audit_log (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      user_id UUID NOT NULL REFERENCES users(id),
      session_id UUID REFERENCES ai_sessions(id),
      prompt_input_hash VARCHAR(64) NOT NULL,
      ai_output_summary TEXT,
      tool_name VARCHAR(100),
      tool_parameters JSONB,
      approval_status VARCHAR(20) NOT NULL DEFAULT 'pending',
      execution_result TEXT,
      model_used VARCHAR(50),
      timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );
  CREATE INDEX idx_ai_action_audit_user ON ai_action_audit_log(user_id);
  CREATE INDEX idx_ai_action_audit_timestamp ON ai_action_audit_log(timestamp);
  ```
- `003_create_audit_log.down.sql`: DROP TABLE ai_action_audit_log; DROP TABLE audit_log
- `004_create_incidents.up.sql`:
  ```sql
  CREATE TABLE incidents (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      title VARCHAR(500) NOT NULL,
      description TEXT,
      severity VARCHAR(20) NOT NULL,
      status VARCHAR(30) NOT NULL DEFAULT 'open',
      source VARCHAR(50) NOT NULL,
      alert_name VARCHAR(100),
      resource_type VARCHAR(50),
      resource_id VARCHAR(255),
      namespace VARCHAR(255),
      rca_summary TEXT,
      acknowledged_by UUID REFERENCES users(id),
      acknowledged_at TIMESTAMPTZ,
      resolved_by UUID REFERENCES users(id),
      resolved_at TIMESTAMPTZ,
      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );
  CREATE INDEX idx_incidents_status ON incidents(status);
  CREATE INDEX idx_incidents_severity ON incidents(severity);
  CREATE INDEX idx_incidents_created_at ON incidents(created_at);
  ```
- `004_create_incidents.down.sql`: DROP TABLE incidents
- `005_create_ai_usage_tracking.up.sql`:
  ```sql
  CREATE TABLE ai_usage_tracking (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      user_id UUID NOT NULL REFERENCES users(id),
      session_id UUID REFERENCES ai_sessions(id),
      model_provider VARCHAR(50) NOT NULL,
      model_name VARCHAR(100) NOT NULL,
      input_tokens INTEGER NOT NULL DEFAULT 0,
      output_tokens INTEGER NOT NULL DEFAULT 0,
      estimated_cost_usd DECIMAL(10, 6) NOT NULL DEFAULT 0,
      timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );
  CREATE INDEX idx_ai_usage_user ON ai_usage_tracking(user_id);
  CREATE INDEX idx_ai_usage_timestamp ON ai_usage_tracking(timestamp);
  CREATE INDEX idx_ai_usage_provider ON ai_usage_tracking(model_provider);
  ```
- `005_create_ai_usage_tracking.down.sql`: DROP TABLE ai_usage_tracking
- `006_create_notification_settings.up.sql`:
  ```sql
  CREATE TABLE notification_settings (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      telegram_bot_token_ref VARCHAR(255),
      telegram_chat_id VARCHAR(100),
      telegram_enabled BOOLEAN NOT NULL DEFAULT false,
      inapp_enabled BOOLEAN NOT NULL DEFAULT true,
      alert_batching_window_seconds INTEGER NOT NULL DEFAULT 120,
      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );
  ```
- `006_create_notification_settings.down.sql`: DROP TABLE notification_settings
- Integrasikan `golang-migrate` untuk menjalankan migration dari CLI dan dari kode

### Tugas 2.6: Middleware Stack
Implementasi `/apps/backend/internal/middleware/`:

- `logger.go`: Log setiap request (method, path, status code, duration, trace_id). Menggunakan slog.
- `recovery.go`: Catch panic, log stack trace, kembalikan 500 Internal Server Error. Jangan biarkan panic crash server.
- `cors.go`: Konfigurasi CORS ketat. Allow origin hanya dari frontend URL (configurable). Allow methods: GET, POST, PUT, PATCH, DELETE, OPTIONS. Allow headers: Authorization, Content-Type, X-Request-ID, X-Idempotency-Key. Max age: 3600.
- `ratelimit.go`: Sliding window rate limiter menggunakan Redis. Default: 100 req/menit per IP. Endpoint AI chat: 20 req/menit per user. Return header X-RateLimit-Remaining dan X-RateLimit-Reset.
- `auth.go`: (Stub saja di fase ini, implementasi penuh di Fase 3). Validasi JWT token dari header Authorization. Ekstrak user_id dan role dari claims. Set ke context.

### Tugas 2.7: Health Check Endpoints
Implementasi `/apps/backend/internal/handler/health_handler.go`:
- `GET /healthz`: Liveness probe. Kembalikan 200 jika proses berjalan.
- `GET /readyz`: Readiness probe. Periksa koneksi ke: PostgreSQL (pool.Ping), Redis (client.Ping), Docker daemon (ping API). Jika salah satu gagal, kembalikan 503 dengan detail komponen yang gagal.

### Tugas 2.8: Server Entrypoint & Graceful Shutdown
Implementasi `/apps/backend/cmd/server/main.go`:
- Load config
- Inisialisasi logger
- Inisialisasi PostgreSQL pool
- Inisialisasi Redis client
- Jalankan database migration (auto-migrate on startup, bisa dinonaktifkan via config)
- Inisialisasi Echo router
- Register semua middleware (logger, recovery, cors)
- Register health check routes
- Start HTTP server
- Graceful shutdown: tangkap SIGTERM/SIGINT via `signal.NotifyContext`. Stop menerima request baru, tunggu aktif selesai (30s timeout), tutup DB pool, tutup Redis, log shutdown complete.

### Tugas 2.9: Prometheus Metrics Endpoint
Implementasi `/apps/backend/internal/handler/metrics_handler.go`:
- Expose `/metrics` endpoint menggunakan `prometheus/client_golang`
- Metrik wajib:
  - `http_requests_total` (counter, labels: method, path, status_code)
  - `http_request_duration_seconds` (histogram, labels: method, path)
  - `active_websocket_connections` (gauge)
  - `db_pool_active_connections` (gauge)
  - `db_pool_idle_connections` (gauge)

### Tugas 2.10: Dockerfile Backend
Buat `/apps/backend/Dockerfile`:
- Multi-stage build: builder (golang:1.22-alpine) dan runtime (alpine:3.19)
- Builder: copy go.mod/go.sum, download deps, copy source, build binary dengan CGO_ENABLED=0
- Runtime: copy binary, buat non-root user, expose port 8080, healthcheck, entrypoint
- Image size target: kurang dari 30MB

### Kriteria Penerimaan Fase 2:
- [x] Backend server start dan listen di port 8080
- [x] `GET /healthz` mengembalikan 200
- [x] `GET /readyz` mengembalikan 200 (semua dependensi connected)
- [x] `GET /metrics` mengembalikan format Prometheus
- [x] Database migration berhasil: semua tabel terbuat (users, ai_sessions, ai_messages, audit_log, ai_action_audit_log, incidents, ai_usage_tracking, notification_settings)
- [x] Seed data berhasil: roles dan konfigurasi default ter-insert
- [x] Graceful shutdown bekerja: SIGTERM mematikan server dengan bersih
- [x] Middleware logger mencatat setiap request ke stdout dalam format JSON
- [x] Middleware recovery menangkap panic tanpa crash
- [x] Rate limiter berfungsi: request ke-101 dalam 1 menit mendapat 429
- [x] Docker build berhasil dan image berjalan

---

## FASE 3: AUTENTIKASI & OTORISASI
**Tujuan**: Mengimplementasikan autentikasi JWT, integrasi Keycloak (atau auth sederhana untuk awal), RBAC enforcement, dan alur login/logout.
**Prasyarat**: Fase 2 selesai.
**Estimasi**: 2-3 hari.

### Tugas 3.1: Keycloak Setup di Docker Compose
- Tambahkan service Keycloak di docker-compose.yml: port 8180, database PostgreSQL terpisah atau shared
- Buat realm `cifo`, client `cifo-frontend` dan `cifo-backend`
- Konfigurasi roles: admin, devops, viewer
- Buat 3 test users (satu per role) untuk development
- Konfigurasi OIDC endpoints

### Tugas 3.2: JWT Middleware (Implementasi Penuh)
Implementasi `/apps/backend/internal/middleware/auth.go`:
- Fetch JWKS dari Keycloak endpoint dan cache secara lokal (refresh setiap 1 jam)
- Validasi JWT signature menggunakan JWKS public key
- Validasi claims: exp (expiry), iss (issuer), aud (audience)
- Ekstrak user info: sub (keycloak_id), email, name, realm_access.roles
- Set user info ke Echo context untuk digunakan handler
- Public endpoints (tanpa auth): /healthz, /readyz, /metrics, /api/v1/auth/callback
- Protected endpoints: semua /api/v1/* lainnya

### Tugas 3.3: RBAC Enforcement
Implementasi RBAC middleware:
- Fungsi `RequireRole(roles ...string) echo.MiddlewareFunc`
- Periksa role user dari context terhadap roles yang diizinkan
- Jika tidak sesuai, kembalikan 403 Forbidden
- Mapping role ke endpoint:
  - Viewer: GET semua endpoint monitoring, docker, argocd, incidents
  - DevOps: Viewer + POST/PUT pada AI chat, incidents acknowledge/resolve
  - Admin: DevOps + PUT/DELETE pada settings, user management

### Tugas 3.4: Auth Handlers
Implementasi `/apps/backend/internal/handler/auth_handler.go`:
- `GET /api/v1/auth/me`: Kembalikan profil user yang sedang login (dari JWT claims)
- `POST /api/v1/auth/logout`: Invalidate refresh token (jika menggunakan token blacklist di Redis)
- Fungsi sinkronisasi user: saat pertama kali login via Keycloak, buat record di tabel users jika belum ada

### Tugas 3.5: Audit Logging untuk Auth Events
- Setiap login berhasil: catat di audit_log (action=login)
- Setiap login gagal: catat di audit_log (action=login_failed)
- Setiap akses endpoint dengan role yang tidak cukup: catat (action=access_denied)

### Kriteria Penerimaan Fase 3:
- [x] Keycloak berjalan dan realm cifo terkonfigurasi
- [x] Request tanpa JWT token ke protected endpoint mendapat 401
- [x] Request dengan JWT token valid mendapat respons yang benar
- [x] Viewer tidak bisa mengakses endpoint Admin (mendapat 403)
- [x] User otomatis dibuat di tabel users saat login pertama
- [x] Auth events tercatat di tabel audit_log
- [x] JWKS cache bekerja (tidak fetch ke Keycloak setiap request)

---

## FASE 4: FONDASI FRONTEND
**Tujuan**: Membangun kerangka frontend: layout utama (sidebar, header), sistem tema, routing, koneksi API, autentikasi, dan komponen UI dasar. Setelah fase ini, pengguna dapat login, melihat layout dashboard kosong dengan navigasi, dan theme toggle berfungsi.
**Prasyarat**: Fase 3 selesai (endpoint auth tersedia).
**Estimasi**: 3-4 hari.

### Tugas 4.1: Sistem Tema & Design Tokens
- Implementasi `/src/styles/globals.css` dan `/src/styles/tokens.css`
- CSS custom properties untuk dark dan light mode:
  - --color-bg-primary, --color-bg-secondary, --color-bg-tertiary
  - --color-text-primary, --color-text-secondary, --color-text-muted
  - --color-border, --color-border-hover
  - --color-accent (cyan/teal sesuai referensi)
  - --color-success, --color-warning, --color-error, --color-critical
  - --radius-sm, --radius-md, --radius-lg
  - --shadow-sm, --shadow-md, --shadow-lg
- Tailwind config yang menggunakan CSS variables di atas
- Font: Inter dari Google Fonts (import di layout.tsx)

### Tugas 4.2: Komponen UI Dasar
Implementasi `/src/components/ui/`:
- `button.tsx`: Variants (primary, secondary, ghost, danger), sizes (sm, md, lg), loading state
- `input.tsx`: Text input dengan label, error message, disabled state
- `modal.tsx`: Dialog berbasis Radix UI, controlled via props
- `toast.tsx`: Toast notification stack berbasis Radix UI Toast, support severity colors
- `badge.tsx`: Status badge (Running, Stopped, Synced, Failed, dll) dengan warna semantik
- `skeleton.tsx`: Skeleton loading placeholder (rectangle, circle, text)
- `card.tsx`: Container card dengan border, shadow, padding variants
- `dropdown.tsx`: Dropdown menu berbasis Radix UI
- `tooltip.tsx`: Tooltip berbasis Radix UI

### Tugas 4.3: Layout Utama
Implementasi `/src/components/layout/`:
- `sidebar.tsx`: Sidebar navigasi kiri. Item: Monitoring, Kubernetes, Docker, Incidents, Settings. Logo CIFO di atas. EXIT di bawah. Collapsible (disimpan di Zustand store). Active item highlighted dengan accent color.
- `header.tsx`: Header bar atas. Judul halaman, search bar (placeholder), time range selector dropdown (Last 1 Hour, Last 6 Hours, Last 24 Hours, Last 7 Days), Quick Fix button, notification bell icon (dengan counter), user avatar dropdown (profile, logout).
- `breadcrumb.tsx`: Breadcrumb navigation

Implementasi `/src/app/layout.tsx`:
- Root layout yang membungkus sidebar dan header
- React Query provider (QueryClientProvider)
- Theme provider (data-theme attribute di html tag)
- Redirect ke login jika tidak terautentikasi

### Tugas 4.4: Zustand Stores
Implementasi `/src/store/`:
- `sidebar-store.ts`: isCollapsed state, toggle function
- `theme-store.ts`: theme (dark/light), toggle function, persist ke localStorage
- `notification-store.ts`: notifications array, addNotification, removeNotification, markAsRead, unreadCount

### Tugas 4.5: API Client & Auth Hook
Implementasi `/src/lib/api-client.ts`:
- Axios instance dengan base URL dari environment variable
- Request interceptor: tambahkan JWT token dari cookie/localStorage ke header Authorization
- Response interceptor: jika 401, redirect ke login. Jika 429, tampilkan toast rate limit. Jika 5xx, tampilkan toast error.

Implementasi `/src/hooks/use-auth.ts`:
- Custom hook yang mengelola state autentikasi
- Login function: redirect ke Keycloak login page
- Logout function: call backend /api/v1/auth/logout, clear token, redirect ke login
- User info dari JWT claims (decode di client-side untuk display, validasi tetap di server)

### Tugas 4.6: Halaman Login
Implementasi `/src/app/(auth)/login/page.tsx`:
- Halaman login yang mengarahkan ke Keycloak OIDC
- Logo CIFO di tengah
- Tombol "Login with SSO"
- Callback handler setelah Keycloak redirect kembali

### Tugas 4.7: Halaman Dashboard Kosong
Implementasi `/src/app/(dashboard)/monitoring/page.tsx`:
- Layout grid yang sesuai referensi tampilan:
  - Baris 1: Stat cards (Total Kontainer, Total Replika, Overall RAM, Kontainer On, Kontainer Off, Active Incidents) - placeholder loading skeleton
  - Baris 2: Chart area (Host Resource Usage) dan System Event Logs - placeholder
  - Baris 3: Agent Architecture section dan AI Chat section - placeholder
- Semua menggunakan komponen Card dengan Skeleton di dalamnya (karena belum ada data)

### Tugas 4.8: Utility Functions
Implementasi `/src/lib/format.ts`:
- `formatBytes(bytes: number)`: konversi ke KB, MB, GB dengan presisi
- `formatUptime(seconds: number)`: konversi ke "2d 5h 30m"
- `formatDate(date: string)`: format ISO ke "2026-09-03 15:00:00"
- `formatRelative(date: string)`: "5 minutes ago", "2 hours ago"
- `formatPercentage(value: number)`: "64.5%"

### Tugas 4.9: Dockerfile Frontend
Buat `/apps/frontend/Dockerfile`:
- Multi-stage: deps (node:20-alpine, install), builder (build Next.js), runner (node:20-alpine, standalone output)
- Konfigurasi standalone output di next.config.ts
- Non-root user, expose port 3000, healthcheck

### Kriteria Penerimaan Fase 4:
- [ ] `npm run dev` menampilkan halaman login
- [ ] Login via Keycloak berhasil dan redirect ke dashboard
- [ ] Sidebar navigasi berfungsi (highlight active, collapse/expand)
- [ ] Header menampilkan user info dan notification bell
- [ ] Dark mode aktif sebagai default, toggle ke light mode berfungsi
- [ ] Dashboard monitoring menampilkan layout grid dengan skeleton loading
- [ ] Logout menghapus token dan redirect ke login
- [ ] API client menambahkan token ke setiap request
- [ ] Error 401 memicu redirect ke login
- [ ] Docker build berhasil dan image berjalan

---

## FASE 5: MONITORING DOCKER
**Tujuan**: Mengimplementasikan pemantauan Docker secara penuh: backend terhubung ke Docker daemon, mengambil data kontainer/image/volume/network nyata, dan frontend menampilkannya di dashboard dan halaman Docker khusus.
**Prasyarat**: Fase 4 selesai. Docker daemon lokal berjalan.
**Estimasi**: 3-4 hari.

### Tugas 5.1: Docker Client Integration (Backend)
Implementasi `/apps/backend/internal/integration/docker_client.go`:
- Koneksi ke Docker daemon via Unix socket (configurable untuk TCP mTLS)
- Menggunakan library `github.com/docker/docker/client`
- Fungsi:
  - `ListContainers(ctx, all bool)`: daftar semua kontainer (running/stopped)
  - `GetContainer(ctx, id string)`: detail satu kontainer (inspect)
  - `GetContainerStats(ctx, id string)`: live stats CPU/Memory/Network/IO
  - `GetContainerLogs(ctx, id string, tail int)`: ambil N baris log terakhir
  - `StreamContainerLogs(ctx, id string)`: stream log real-time (io.Reader)
  - `RestartContainer(ctx, id string)`: restart kontainer
  - `StopContainer(ctx, id string)`: stop kontainer
  - `ListImages(ctx)`: daftar semua images
  - `ListVolumes(ctx)`: daftar semua volumes
  - `ListNetworks(ctx)`: daftar semua networks
  - `GetSystemInfo(ctx)`: info host Docker (CPU, Memory total, kernel, OS)
- Error handling: wrap semua error ke AppError

### Tugas 5.2: Docker Service Layer (Backend)
Implementasi `/apps/backend/internal/service/docker_service.go`:
- Fungsi bisnis yang memanggil docker_client dan mengkonversi response ke model internal
- Menghitung statistik agregat: total containers, running containers, stopped containers, total RAM used
- Caching di Redis: data yang tidak sering berubah (images, volumes) di-cache 60 detik
- Rate limiting pada operasi write (restart, stop)

### Tugas 5.3: Docker Handlers (Backend)
Implementasi `/apps/backend/internal/handler/docker_handler.go`:
- `GET /api/v1/docker/containers`: daftar kontainer dengan filter (status: running/stopped/all)
- `GET /api/v1/docker/containers/:id`: detail satu kontainer
- `GET /api/v1/docker/containers/:id/stats`: live stats
- `GET /api/v1/docker/containers/:id/logs`: logs (query param: tail, since)
- `POST /api/v1/docker/containers/:id/restart`: restart (DevOps/Admin only)
- `POST /api/v1/docker/containers/:id/stop`: stop (Admin only)
- `GET /api/v1/docker/images`: daftar images
- `GET /api/v1/docker/volumes`: daftar volumes
- `GET /api/v1/docker/networks`: daftar networks
- `GET /api/v1/docker/system`: system info (CPU, Memory, Docker version)
- Semua write operations dicatat di audit_log

### Tugas 5.4: Dashboard Monitoring Cards (Frontend)
Implementasi `/src/components/widgets/`:
- `stat-card.tsx`: Kartu statistik dengan angka besar, label, dan sub-label (sesuai referensi tampilan)
- Komponen cards untuk baris pertama dashboard:
  - Total Kontainer (dari Docker API)
  - Total Replika (dari Kubernetes API - placeholder sampai Fase 6)
  - Overall RAM (dari Docker system info)
  - Kontainer On (running count)
  - Kontainer Off (stopped count, warna merah jika ada)
  - Active Incidents (dari incidents API)
- Data fetching menggunakan React Query dengan refetch interval 15 detik

### Tugas 5.5: Host Resource Usage Chart (Frontend)
Implementasi `/src/components/charts/`:
- `cpu-ram-chart.tsx`: ECharts time-series chart untuk CPU Load (%) dan RAM Usage (%)
- Tab buttons: CPU & RAM, Network I/O, AI Tokens
- Data source: backend endpoint yang query VictoriaMetrics/Prometheus
- Auto-refresh setiap 30 detik
- Time range dari header selector (1h, 6h, 24h, 7d)

### Tugas 5.6: System Event Logs (Frontend)
Implementasi `/src/components/terminal/`:
- `event-log.tsx`: Terminal-style log viewer
- Auto-scroll toggle (Auto-Scroll ON/OFF)
- Export button (download log sebagai file teks)
- Warna berdasarkan level: INFO (hijau), WARNING (kuning), CRITICAL (merah), AI-OPS (cyan)
- Data source: backend endpoint atau WebSocket (implementasi penuh di Fase 7)
- Untuk fase ini: polling setiap 5 detik

### Tugas 5.7: Halaman Docker Detail (Frontend)
Implementasi `/src/app/(dashboard)/docker/page.tsx`:
- Tabel daftar kontainer: Name, Image, Status (badge), CPU %, Memory, Uptime, Actions
- Filter bar: status (All, Running, Stopped), search by name
- Klik pada kontainer membuka detail panel/page:
  - Info: container ID, image, ports, environment (masked), created, status
  - Stats: live CPU/Memory chart
  - Logs: log viewer (ambil tail 200 dari API)
  - Actions: Restart button (DevOps/Admin), Stop button (Admin)

### Tugas 5.8: Monitoring Service (Backend)
Implementasi `/apps/backend/internal/service/monitoring_service.go`:
- Fungsi agregasi untuk dashboard stat cards
- Query VictoriaMetrics untuk time-series data (CPU, RAM, Network)
- Menggunakan PromQL via `/api/v1/query_range` endpoint VictoriaMetrics

Implementasi `/apps/backend/internal/handler/monitoring_handler.go`:
- `GET /api/v1/monitoring/stats`: agregat statistik dashboard
- `GET /api/v1/monitoring/metrics/cpu`: time-series CPU usage
- `GET /api/v1/monitoring/metrics/memory`: time-series memory usage
- `GET /api/v1/monitoring/metrics/network`: time-series network I/O

### Kriteria Penerimaan Fase 5:
- [ ] Backend berhasil terhubung ke Docker daemon lokal
- [ ] GET /api/v1/docker/containers mengembalikan daftar kontainer nyata
- [ ] GET /api/v1/docker/containers/:id/stats mengembalikan live stats
- [ ] Dashboard stat cards menampilkan data nyata (bukan dummy)
- [ ] Chart CPU/RAM menampilkan data time-series nyata dari VictoriaMetrics
- [ ] System Event Logs menampilkan log dari backend
- [ ] Halaman Docker Detail menampilkan tabel kontainer dengan data nyata
- [ ] Detail kontainer menampilkan info, stats, dan log
- [ ] Restart kontainer berfungsi (DevOps/Admin)
- [ ] Audit log mencatat operasi restart/stop

---

## FASE 6: MONITORING KUBERNETES & ARGOCD
**Tujuan**: Mengimplementasikan pemantauan Kubernetes (pods, deployments, nodes) dan ArgoCD (applications, sync status, health, deployment history) menggunakan cluster K3d lokal dan ArgoCD yang sudah di-setup di Fase 1.
**Prasyarat**: Fase 5 selesai. K3d cluster dan ArgoCD berjalan.
**Estimasi**: 4-5 hari.

### Tugas 6.1: Kubernetes Client Integration (Backend)
- Menggunakan `k8s.io/client-go` untuk koneksi ke cluster
- Konfigurasi in-cluster (jika dijalankan di dalam K8s) atau kubeconfig (untuk development lokal)
- Fungsi: ListPods, GetPod, GetPodLogs, ListDeployments, GetDeployment, ListNodes, GetNodeMetrics, ListServices, ListEvents
- Watch API untuk event streaming (Pod status changes, deployment rollout)

### Tugas 6.2: ArgoCD Client Integration (Backend)
- Menggunakan ArgoCD REST/gRPC API
- Autentikasi via ArgoCD API token (disimpan di config/env)
- Fungsi: ListApplications, GetApplication, GetApplicationHistory, SyncApplication, GetApplicationManifests, GetApplicationEvents

### Tugas 6.3: Kubernetes & ArgoCD Service Layers (Backend)
- kubernetes_service.go: agregasi data pods/deployments/nodes
- argocd_service.go: agregasi data applications/sync status

### Tugas 6.4: API Endpoints (Backend)
- `GET /api/v1/kubernetes/pods`: daftar pods (filter: namespace, status)
- `GET /api/v1/kubernetes/pods/:namespace/:name`: detail pod
- `GET /api/v1/kubernetes/pods/:namespace/:name/logs`: pod logs
- `GET /api/v1/kubernetes/deployments`: daftar deployments
- `GET /api/v1/kubernetes/deployments/:namespace/:name`: detail deployment
- `POST /api/v1/kubernetes/deployments/:namespace/:name/restart`: restart deployment (DevOps/Admin)
- `PUT /api/v1/kubernetes/deployments/:namespace/:name/scale`: scale deployment (DevOps/Admin)
- `GET /api/v1/kubernetes/nodes`: daftar nodes dengan resource usage
- `GET /api/v1/argocd/applications`: daftar ArgoCD applications
- `GET /api/v1/argocd/applications/:name`: detail application (health, sync, resources)
- `GET /api/v1/argocd/applications/:name/history`: deployment history
- `POST /api/v1/argocd/applications/:name/sync`: trigger sync (DevOps/Admin)

### Tugas 6.5: Halaman Kubernetes (Frontend)
- `/src/app/(dashboard)/kubernetes/page.tsx`
- Tab: Pods, Deployments, Nodes
- Tabel pods: Name, Namespace, Status (badge), Restarts, CPU, Memory, Age, Node
- Tabel deployments: Name, Namespace, Ready Replicas, Available, Age, Actions (restart, scale)
- Node overview: cards per node dengan CPU/Memory utilization bars
- Klik pada pod/deployment membuka detail panel

### Tugas 6.6: ArgoCD Section di Dashboard
- Widget ArgoCD applications di halaman monitoring utama
- Daftar apps dengan sync status (Synced/OutOfSync) dan health status (Healthy/Degraded/Progressing)
- Klik pada app membuka detail: resources tree, sync history timeline, sync/refresh buttons
- Badge warna: Synced+Healthy (hijau), OutOfSync (kuning), Degraded (merah)

### Tugas 6.7: Update Dashboard Stat Cards
- Total Replika: mengambil dari Kubernetes deployments (jumlah total ready replicas)
- Update monitoring_service.go untuk menggabungkan data Docker dan Kubernetes

### Kriteria Penerimaan Fase 6:
- [ ] Backend terhubung ke K3d cluster dan mengambil data pods/deployments/nodes nyata
- [ ] Backend terhubung ke ArgoCD API dan mengambil data applications nyata
- [ ] Halaman Kubernetes menampilkan pods, deployments, nodes dari cluster K3d
- [ ] ArgoCD widget menampilkan sample apps (nginx, httpbin) dengan status sync/health
- [ ] Restart deployment berfungsi melalui UI
- [ ] Scale deployment berfungsi melalui UI
- [ ] Sync ArgoCD app berfungsi melalui UI
- [ ] Stat card "Total Replika" menampilkan data nyata
- [ ] Deployment history menampilkan timeline revisi
- [ ] Semua write operations tercatat di audit_log

---

## FASE 7: REAL-TIME & WEBSOCKET
**Tujuan**: Mengimplementasikan komunikasi real-time: WebSocket hub, log streaming dari Docker dan Kubernetes, system event streaming, dan notifikasi in-app push.
**Prasyarat**: Fase 6 selesai.
**Estimasi**: 3-4 hari.

### Tugas 7.1: WebSocket Hub (Backend)
Implementasi `/apps/backend/internal/ws/`:
- `hub.go`: Central hub yang mengelola semua koneksi WebSocket. Register/Unregister clients. Broadcast messages ke semua atau ke subset clients (berdasarkan channel/topic).
- `client.go`: Per-connection handler. Buffered channel (max 1000 messages). Write pump (goroutine yang mengirim dari channel ke WebSocket). Read pump (goroutine yang membaca dari WebSocket, handle ping/pong). Heartbeat ping setiap 30 detik. Jika pong tidak diterima dalam 60 detik, tutup koneksi.
- `message.go`: Tipe pesan WebSocket (log_entry, notification, metric_update, container_event, pod_event, argocd_event).

### Tugas 7.2: WebSocket Endpoint (Backend)
Implementasi di `websocket_handler.go`:
- `GET /ws`: Upgrade HTTP ke WebSocket. Autentikasi via query param token (karena WebSocket tidak support custom headers). Setelah connect, client mengirim pesan subscribe ke channel yang diinginkan (docker_logs:container_id, k8s_events, notifications, system_events).

### Tugas 7.3: Docker Log Streaming (Backend)
- Saat client subscribe ke `docker_logs:<container_id>`, backend membuka stream ke Docker daemon (ContainerLogs dengan Follow=true)
- Pipe output ke WebSocket client
- Saat client disconnect, tutup Docker log stream
- Backpressure: jika buffer penuh, drop pesan lama

### Tugas 7.4: Kubernetes Event Streaming (Backend)
- Saat client subscribe ke `k8s_events`, backend menggunakan Watch API client-go untuk memantau events di semua namespace
- Kirim event baru ke WebSocket client (Pod created, Pod deleted, Deployment scaled, dll)

### Tugas 7.5: WebSocket Client (Frontend)
Implementasi `/src/lib/ws-client.ts`:
- Class WebSocketClient dengan auto-reconnect (exponential backoff: 1s, 2s, 4s, 8s, max 30s)
- Subscribe/unsubscribe ke channels
- Event emitter pattern untuk menerima pesan per tipe

Implementasi `/src/hooks/use-websocket.ts`:
- React hook yang membungkus WebSocketClient
- Auto-connect saat komponen mount, auto-disconnect saat unmount
- Return: status (connecting, connected, disconnected), sendMessage, lastMessage, subscribe

### Tugas 7.6: Live Log Viewer (Frontend)
Update `/src/components/terminal/`:
- `log-terminal.tsx`: Terminal emulator yang menerima log via WebSocket
- Virtual scrolling untuk performa (ribuan baris log)
- Syntax highlighting berdasarkan level (ERROR=merah, WARN=kuning, INFO=hijau)
- Fitur: auto-scroll toggle, search/filter dalam log, clear, export

### Tugas 7.7: In-App Notification System
- Backend push notifikasi ke WebSocket channel `notifications` saat ada alert baru atau event penting
- Frontend menerima dan menampilkan toast notification
- CRITICAL: toast persistent (tetap sampai diklik)
- WARNING: toast auto-dismiss 10 detik
- INFO: toast auto-dismiss 5 detik
- Notification center (bell icon di header): daftar semua notifikasi, mark as read, link ke incident detail

### Tugas 7.8: Update System Event Logs
- Ganti polling dengan WebSocket untuk System Event Logs di dashboard
- Subscribe ke channel `system_events`
- Data real-time dari Docker events, K8s events, dan alert dari Alertmanager

### Kriteria Penerimaan Fase 7:
- [ ] WebSocket koneksi berhasil dari frontend ke backend
- [ ] Heartbeat ping/pong berfungsi (koneksi tidak timeout)
- [ ] Auto-reconnect berfungsi saat koneksi terputus
- [ ] Docker container logs stream real-time di log terminal
- [ ] Kubernetes events muncul real-time di System Event Logs
- [ ] Toast notification muncul saat ada alert baru
- [ ] CRITICAL toast persistent, WARNING auto-dismiss
- [ ] Notification center menampilkan daftar notifikasi
- [ ] Buffered channel mencegah goroutine blocking
- [ ] Disconnect client menutup Docker log stream (tidak ada goroutine leak)

---

## FASE 8: ALERTING & INCIDENT MANAGEMENT
**Tujuan**: Mengimplementasikan penerimaan alert dari Alertmanager, pembuatan incident otomatis, integrasi Telegram Bot, incident lifecycle management, dan halaman incidents di frontend.
**Prasyarat**: Fase 7 selesai.
**Estimasi**: 3-4 hari.

### Tugas 8.1: Alertmanager Webhook Receiver (Backend)
- `POST /api/v1/webhooks/alertmanager`: menerima payload alert dari Alertmanager
- Parse payload: alertname, severity, status (firing/resolved), labels, annotations
- Buat incident di database jika status=firing
- Update incident status ke resolved jika status=resolved
- Trigger notifikasi ke Telegram dan in-app notification

### Tugas 8.2: Incident Service (Backend)
Implementasi `/apps/backend/internal/service/incident_service.go`:
- CreateIncident: buat incident baru dari alert
- ListIncidents: daftar incidents dengan filter (status, severity, date range)
- GetIncident: detail incident
- AcknowledgeIncident: ubah status ke acknowledged, catat user dan waktu
- ResolveIncident: ubah status ke resolved, catat user dan waktu
- CloseIncident: ubah status ke closed
- EscalateIncident: kirim ulang notifikasi dengan penanda ESCALATED

### Tugas 8.3: Incident Escalation Background Job
- Asynq job yang berjalan setiap menit
- Periksa incidents dengan status=open yang created_at lebih dari 15 menit lalu
- Untuk setiap incident yang belum di-acknowledge: escalate (kirim ulang ke Telegram dengan prefix ESCALATED)

### Tugas 8.4: Telegram Bot Integration (Backend)
Implementasi `/apps/backend/internal/integration/telegram_client.go`:
- Kirim pesan via Telegram Bot API (HTTP POST ke api.telegram.org)
- Format pesan Markdown (tanpa emoji):
  ```
  [CRITICAL] ContainerOOMKilled
  Resource: payment-gateway (Docker)
  Time: 2026-09-03 15:30:00 UTC
  Description: Container killed due to OOM
  Action Required: Acknowledge in dashboard
  ```
- Alert batching: jika lebih dari 3 alert dalam 2 menit, kirim ringkasan
- Rate limiting: max 30 pesan per menit
- Retry queue: jika pengiriman gagal, simpan di Redis queue dan retry saat koneksi pulih

Implementasi `/apps/backend/internal/service/telegram_service.go`:
- Business logic untuk formatting dan batching
- Asynq task untuk pengiriman asinkron

### Tugas 8.5: Notification Service (Backend)
Implementasi `/apps/backend/internal/service/notification_service.go`:
- Fungsi `Notify(incident Incident)`:
  - Kirim ke Telegram (via Asynq async task)
  - Push ke WebSocket (in-app notification)
  - Simpan di notification history di database

### Tugas 8.6: Incident Handlers (Backend)
- `GET /api/v1/incidents`: daftar incidents (filter: status, severity, page, limit)
- `GET /api/v1/incidents/:id`: detail incident
- `POST /api/v1/incidents/:id/acknowledge`: acknowledge (DevOps/Admin)
- `POST /api/v1/incidents/:id/resolve`: resolve (DevOps/Admin)
- `POST /api/v1/incidents/:id/close`: close (Admin)

### Tugas 8.7: Halaman Incidents (Frontend)
Implementasi `/src/app/(dashboard)/incidents/page.tsx`:
- Tabel incidents: Title, Severity (badge), Status (badge), Source, Created, Acknowledged By, Actions
- Filter: status (Open, Acknowledged, Investigating, Resolved, Closed), severity (Critical, Warning, Info)
- Sort by: created_at (newest first default)
- Klik pada incident membuka detail:
  - Incident info (title, description, severity, source alert)
  - Timeline: created -> acknowledged -> resolved -> closed (dengan timestamps dan actor)
  - Related resource link (kontainer/pod/ArgoCD app)
  - AI RCA summary (placeholder untuk Fase 9)
  - Action buttons: Acknowledge, Resolve, Close (berdasarkan role)
- Auto-refresh via React Query (15 detik)
- Real-time update via WebSocket (incident baru muncul langsung)

### Kriteria Penerimaan Fase 8:
- [ ] Alertmanager mengirim alert ke backend webhook dan incident terbuat di database
- [ ] Telegram Bot mengirim notifikasi dengan format yang benar
- [ ] Alert batching berfungsi (alert storm menghasilkan ringkasan, bukan spam)
- [ ] Incident lifecycle berfungsi: Open -> Acknowledged -> Resolved -> Closed
- [ ] Escalation berfungsi: incident yang tidak di-acknowledge dalam 15 menit di-escalate
- [ ] Halaman incidents menampilkan daftar incidents nyata
- [ ] Detail incident menampilkan timeline lengkap
- [ ] Acknowledge/Resolve/Close berfungsi dari UI
- [ ] Notifikasi in-app muncul real-time saat incident baru dibuat
- [ ] Semua aksi incident tercatat di audit_log
- [ ] Retry queue bekerja: pesan Telegram yang gagal dikirim ulang

---

## FASE 9: AI SERVICE & CHAT AGENT
**Tujuan**: Mengimplementasikan AI Service (Python), multi-model provider, tool calling, conversation memory, prompt injection defense, dan integrasi chat di frontend.
**Prasyarat**: Fase 8 selesai.
**Estimasi**: 5-7 hari.

### Tugas 9.1: AI Service Foundation (Python)
- Implementasi `/apps/ai-service/app/main.py`: FastAPI app dengan gRPC server
- Implementasi `/apps/ai-service/app/config/settings.py`: Pydantic settings
- gRPC service definition dan code generation dari proto file

### Tugas 9.2: LLM Provider Adapters
Implementasi `/apps/ai-service/app/providers/`:
- `base.py`: Abstract base class dengan method `chat(messages, tools) -> response`
- `google_provider.py`: Google AI Studio (Gemini) adapter
- `openai_provider.py`: OpenAI adapter
- `anthropic_provider.py`: Anthropic adapter
- `ollama_provider.py`: Ollama local adapter
- Setiap adapter normalize input/output ke format internal yang konsisten

### Tugas 9.3: Model Orchestrator & Circuit Breaker
Implementasi `/apps/ai-service/app/agent/orchestrator.py`:
- Model routing: coba primary (Google), jika gagal -> fallback 1 (OpenAI), jika gagal -> fallback 2 (Anthropic), jika gagal -> fallback 3 (Ollama), jika semua gagal -> return degraded response
- Circuit breaker: 3 failures dalam 60 detik = circuit open, retry setelah 120 detik
- Logging setiap model switch
- Token counting per request/response
- Cost estimation per model

### Tugas 9.4: Tool Definitions
Implementasi `/apps/ai-service/app/tools/`:
- `base.py`: Base tool class dengan: name, description, parameters (JSON Schema), required_role, requires_approval
- `kubectl_tools.py`: Semua Kubernetes tools dari arsitektur D.4 (get_pod_status, get_deployment_info, dll)
- `docker_tools.py`: Semua Docker tools (get_container_logs, list_docker_containers, dll)
- `argocd_tools.py`: Semua ArgoCD tools (get_argocd_app_status, get_argocd_history, dll)
- Tools read-only memanggil backend Go via gRPC
- Tools write memanggil backend Go via gRPC dengan flag requires_approval=true

### Tugas 9.5: Conversation Memory
Implementasi `/apps/ai-service/app/agent/memory.py`:
- Load conversation history dari backend (PostgreSQL via gRPC)
- Sliding window: kirim 20 pesan terakhir ke LLM
- Summarization: jika sesi panjang, ringkas pesan lama dengan model murah (Gemini Flash)
- TTL enforcement: tandai sesi expired setelah 30 menit inaktif

### Tugas 9.6: Prompt Injection Defense
Implementasi `/apps/ai-service/app/agent/sanitizer.py`:
- Daftar pola berbahaya (regex): "ignore previous", "forget your instructions", "act as root", "sudo", "rm -rf"
- Input sanitization: strip/escape pola berbahaya sebelum kirim ke LLM
- Output validation: pastikan tool calls yang dihasilkan AI sesuai schema yang terdefinisi
- Logging setiap deteksi upaya injection

### Tugas 9.7: System Prompts
Buat `/apps/ai-service/app/prompts/`:
- `system_prompt.txt`: Instruksi dasar untuk AI (siapa kamu, apa tugasmu, batasan, format respons)
- `diagnosis_prompt.txt`: Template untuk Root Cause Analysis (analisis log, identifikasi masalah, rekomendasi)
- `summarize_prompt.txt`: Template untuk meringkas percakapan panjang

### Tugas 9.8: AI Handler di Backend Go
Implementasi `/apps/backend/internal/handler/ai_handler.go`:
- `POST /api/v1/ai/chat`: menerima pesan dari frontend, forward ke AI Service via gRPC, kembalikan respons
- `GET /api/v1/ai/sessions`: daftar sesi AI pengguna
- `GET /api/v1/ai/sessions/:id/messages`: riwayat pesan dalam sesi
- `POST /api/v1/ai/tools/:name/approve`: approve eksekusi tool
- `POST /api/v1/ai/tools/:name/reject`: reject eksekusi tool

Implementasi `/apps/backend/internal/service/ai_service.go`:
- gRPC client ke AI Service
- Simpan pesan ke database (ai_messages)
- Simpan usage tracking (ai_usage_tracking)
- Handle tool approval flow

### Tugas 9.9: AI Chat Interface (Frontend)
Implementasi `/src/components/ai-chat/`:
- `chat-container.tsx`: Container utama AI chat (sesuai referensi: pojok kanan bawah dashboard)
- `chat-input.tsx`: Input field dengan tombol send
- `chat-bubble.tsx`: Bubble pesan (user: kanan, AI: kiri). Markdown rendering untuk respons AI.
- `tool-approval.tsx`: Card konfirmasi ketika AI meminta persetujuan untuk menjalankan perintah. Menampilkan: tool name, parameters, tombol Approve/Reject.
- `model-indicator.tsx`: Indikator model yang sedang aktif (Gemini 1.5 Active, dengan dot hijau)

### Tugas 9.10: Auto-Diagnosis saat Alert CRITICAL
- Saat incident CRITICAL dibuat (dari Alertmanager), backend otomatis mengirim konteks log ke AI Service
- AI menganalisis dan menghasilkan RCA
- RCA disimpan di incident (field rca_summary)
- RCA ditampilkan di detail incident page
- User dapat melanjutkan diskusi dengan AI tentang incident tersebut di chat

### Tugas 9.11: AI Usage Tracking & Cost Display
- Implementasi endpoint `GET /api/v1/ai/usage`: statistik penggunaan (total tokens, estimated cost, per model breakdown)
- Di Settings page: tampilkan grafik penggunaan AI per hari, per model, estimasi biaya

### Tugas 9.12: Dockerfile AI Service
Buat `/apps/ai-service/Dockerfile`:
- Base: python:3.12-slim
- Install dependencies, copy source, non-root user, expose port 50051 (gRPC), healthcheck

### Kriteria Penerimaan Fase 9:
- [ ] AI Service berjalan dan menerima request gRPC
- [ ] Chat AI berfungsi: kirim pesan, terima respons dari Gemini
- [ ] Fallback berfungsi: jika Gemini gagal, otomatis beralih ke model berikutnya
- [ ] Circuit breaker berfungsi: 3 failures = circuit open
- [ ] Degraded mode: jika semua model gagal, pesan eksplisit ditampilkan
- [ ] Read-only tools berfungsi: AI bisa membaca status pod, container logs
- [ ] Write tools meminta approval: AI menampilkan konfirmasi sebelum restart/scale
- [ ] Approval/Rejection berfungsi dari UI
- [ ] Conversation memory berfungsi: AI mengingat konteks percakapan sebelumnya
- [ ] Prompt injection terdeteksi dan diblokir
- [ ] Auto-diagnosis CRITICAL incident menghasilkan RCA
- [ ] Usage tracking mencatat tokens dan estimasi biaya
- [ ] Semua aksi AI tercatat di ai_action_audit_log

---

## FASE 10: HALAMAN SETTINGS & ADMINISTRASI
**Tujuan**: Mengimplementasikan halaman settings untuk konfigurasi sistem, manajemen pengguna, konfigurasi Telegram, dan preferensi AI.
**Prasyarat**: Fase 9 selesai.
**Estimasi**: 2-3 hari.

### Tugas 10.1: Settings API (Backend)
- `GET /api/v1/settings`: konfigurasi sistem saat ini (Admin only)
- `PUT /api/v1/settings`: update konfigurasi (Admin only)
- `GET /api/v1/settings/users`: daftar pengguna (Admin only)
- `PUT /api/v1/settings/users/:id/role`: ubah role pengguna (Admin only)
- `DELETE /api/v1/settings/users/:id`: nonaktifkan pengguna (Admin only)

### Tugas 10.2: Halaman Settings (Frontend)
Implementasi `/src/app/(dashboard)/settings/page.tsx`:
- Tab: General, Notifications, AI Configuration, Users, Security
- General: tema default, bahasa, timezone
- Notifications: konfigurasi Telegram (bot token reference, chat ID, enable/disable), alert batching window
- AI Configuration: model preference order, budget ceiling, usage statistics chart
- Users: tabel pengguna, ubah role, nonaktifkan akun (Admin only)
- Security: force logout all sessions, regenerate API keys

### Kriteria Penerimaan Fase 10:
- [ ] Halaman settings menampilkan konfigurasi sistem
- [ ] Admin dapat mengubah konfigurasi Telegram
- [ ] Admin dapat mengubah role pengguna
- [ ] Admin dapat menonaktifkan pengguna
- [ ] AI configuration menampilkan penggunaan dan estimasi biaya
- [ ] Perubahan settings tercatat di audit_log

---

## FASE 11: OBSERVABILITY (TRACING & LOGGING)
**Tujuan**: Mengimplementasikan distributed tracing (OpenTelemetry), memastikan semua log terstruktur dan ter-korelasi dengan trace_id, dan memverifikasi seluruh observability stack (metrik, log, trace) berfungsi end-to-end.
**Prasyarat**: Fase 10 selesai.
**Estimasi**: 2-3 hari.

### Tugas 11.1: OpenTelemetry di Backend Go
- Install `go.opentelemetry.io/otel` dan eksporter Tempo
- Inisialisasi tracer provider di main.go
- Instrumentasi semua HTTP handler (Echo middleware)
- Instrumentasi database queries (pgx tracing hooks)
- Propagasi trace context ke AI Service (via gRPC metadata)

### Tugas 11.2: OpenTelemetry di AI Service Python
- Install `opentelemetry-sdk` dan eksporter Tempo
- Instrumentasi FastAPI dan gRPC server
- Propagasi trace context dari incoming gRPC ke LLM API calls

### Tugas 11.3: Log-Trace Correlation
- Pastikan setiap log entry mengandung trace_id dan span_id
- Konfigurasi Grafana datasource linking: dari Loki log entry bisa klik ke Tempo trace view

### Tugas 11.4: Verifikasi End-to-End
- Trigger satu alur lengkap: user login -> buka dashboard -> kirim chat AI -> AI baca logs -> AI rekomendasikan restart -> user approve -> restart dieksekusi
- Verifikasi: trace_id yang sama muncul di log backend, log AI service, dan trace Tempo
- Verifikasi: metrik Prometheus/VictoriaMetrics mencatat request count dan duration

### Kriteria Penerimaan Fase 11:
- [ ] Setiap request memiliki trace_id unik
- [ ] Trace muncul di Grafana Tempo dengan semua span (backend, AI service, database)
- [ ] Log di Loki mengandung trace_id yang sama
- [ ] Klik dari log ke trace berfungsi di Grafana
- [ ] Metrik HTTP request duration tercatat dengan benar

---

## FASE 12: SECURITY HARDENING
**Tujuan**: Memperkuat keamanan seluruh sistem: Vault integration, Docker socket proxy, NetworkPolicy, security headers, vulnerability scanning, dan penetration testing dasar.
**Prasyarat**: Fase 11 selesai.
**Estimasi**: 3-4 hari.

### Tugas 12.1: HashiCorp Vault Setup
- Tambahkan Vault di docker-compose
- Buat policies untuk backend dan AI service
- Migrasi semua credential dari .env ke Vault
- Implementasi Vault client di backend Go (vault_client.go)
- Implementasi Vault reader di AI Service Python

### Tugas 12.2: Docker Socket Proxy
- Tambahkan Tecnativa docker-socket-proxy di docker-compose
- Konfigurasi: hanya izinkan GET (inspect, stats, logs) dan POST terbatas (restart)
- Update docker_client.go untuk terhubung ke proxy, bukan langsung ke socket

### Tugas 12.3: Kubernetes Security Manifests
- Buat NetworkPolicy manifests di `/infrastructure/security/network-policies/`
- Buat RBAC manifests di `/infrastructure/security/rbac/` (ServiceAccount cifo-ai-agent-sa dengan ClusterRole terbatas)
- Apply dan verifikasi di K3d cluster

### Tugas 12.4: Security Headers
- Pastikan semua response mengandung: X-Content-Type-Options, X-Frame-Options, X-XSS-Protection, Content-Security-Policy, Strict-Transport-Security, Referrer-Policy

### Tugas 12.5: Security Scanning
- Jalankan `gosec ./...` di backend Go
- Jalankan `trivy image` pada semua Docker images
- Jalankan `gitleaks detect` pada seluruh repository
- Fix semua temuan CRITICAL dan HIGH

### Kriteria Penerimaan Fase 12:
- [ ] Semua credential dibaca dari Vault (bukan .env)
- [ ] Docker socket proxy berfungsi dan membatasi akses
- [ ] NetworkPolicy ter-apply di K3d (pod di namespace cifo-frontend tidak bisa akses database)
- [ ] AI ServiceAccount tidak bisa delete namespace (verifikasi manual)
- [ ] Security headers ada di semua response
- [ ] gosec: 0 temuan HIGH/CRITICAL
- [ ] trivy: 0 temuan CRITICAL pada images
- [ ] gitleaks: 0 credential leak terdeteksi

---

## FASE 13: TESTING KOMPREHENSIF
**Tujuan**: Menulis dan menjalankan semua level test: unit, integration, E2E, dan load test. Memastikan coverage memenuhi target.
**Prasyarat**: Fase 12 selesai.
**Estimasi**: 4-5 hari.

### Tugas 13.1: Unit Tests Backend Go
- Test untuk setiap package di `/internal/service/`: docker_service_test.go, argocd_service_test.go, incident_service_test.go, notification_service_test.go, ai_service_test.go
- Test untuk `/pkg/apperror/error_test.go`
- Test untuk `/pkg/validator/validator_test.go`
- Target coverage: 70% untuk package service dan repository

### Tugas 13.2: Unit Tests Frontend
- Test untuk komponen interaktif: Button, Modal, Toast, ChatInput, ToolApproval
- Test untuk hooks: use-auth, use-websocket
- Test untuk store: sidebar-store, notification-store, theme-store
- Target coverage: 60%

### Tugas 13.3: Unit Tests AI Service
- Test untuk orchestrator (circuit breaker, model switching)
- Test untuk sanitizer (prompt injection detection)
- Test untuk tool definitions (schema validation)

### Tugas 13.4: Integration Tests
- `auth_api_test.go`: test login flow dengan Keycloak test container
- `docker_api_test.go`: test Docker endpoints terhadap Docker daemon
- `argocd_api_test.go`: test ArgoCD endpoints terhadap API mock atau test instance

### Tugas 13.5: E2E Tests (Playwright)
- `login.spec.ts`: login via Keycloak, verifikasi redirect ke dashboard
- `dashboard.spec.ts`: verifikasi stat cards, charts, event logs muncul
- `docker.spec.ts`: navigasi ke halaman Docker, lihat kontainer, klik detail
- `kubernetes.spec.ts`: navigasi ke halaman Kubernetes, lihat pods
- `incidents.spec.ts`: lihat daftar incidents, acknowledge incident
- `ai-chat.spec.ts`: kirim pesan ke AI, terima respons, approve tool execution

### Tugas 13.6: Load Tests (K6)
- `websocket-stress.js`: 500 concurrent WebSocket connections, verifikasi semua menerima messages
- `api-throughput.js`: 1000 req/detik ke GET /api/v1/docker/containers selama 5 menit, verifikasi p99 latency < 200ms

### Kriteria Penerimaan Fase 13:
- [ ] `go test ./...` pass dengan coverage >= 70% untuk service/repository
- [ ] `npx vitest run` pass dengan coverage >= 60% untuk komponen
- [ ] `pytest` pass untuk AI service
- [ ] Semua integration tests pass
- [ ] Semua Playwright E2E tests pass
- [ ] K6 load test: p99 latency < 200ms pada 1000 req/detik
- [ ] K6 WebSocket test: 500 connections stable selama 5 menit

---

## FASE 14: CI/CD PIPELINE
**Tujuan**: Mengimplementasikan GitHub Actions pipeline lengkap: lint, test, security scan, build, push, deploy ke staging via ArgoCD.
**Prasyarat**: Fase 13 selesai.
**Estimasi**: 2-3 hari.

### Tugas 14.1: GitHub Actions CI
Buat `.github/workflows/ci.yml`:
- Trigger: push ke main, pull request ke main
- Jobs (paralel where possible):
  - lint-backend: golangci-lint
  - lint-frontend: eslint + prettier check
  - lint-ai: ruff check
  - lint-docker: hadolint
  - test-backend: go test dengan coverage report
  - test-frontend: vitest dengan coverage report
  - test-ai: pytest
  - test-integration: testcontainers
  - security-scan: gosec, trivy, gitleaks
  - build-images: Docker build (hanya jika tests pass)

### Tugas 14.2: Deploy to Staging Workflow
Buat `.github/workflows/deploy-staging.yml`:
- Trigger: merge ke main (setelah CI pass)
- Push images ke container registry
- Update image tags di ArgoCD application manifests (config repo)
- ArgoCD auto-sync ke staging

### Tugas 14.3: Deploy to Production Workflow
Buat `.github/workflows/deploy-production.yml`:
- Trigger: manual (workflow_dispatch) dengan approval
- Update image tags di production overlay
- ArgoCD sync ke production (manual sync, bukan auto)

### Tugas 14.4: Helm Charts
Buat Helm charts di `/infrastructure/kubernetes/charts/`:
- cifo-frontend: Deployment, Service, Ingress, HPA
- cifo-backend: Deployment, Service, ConfigMap, HPA, ServiceAccount
- cifo-ai-service: Deployment, Service, ServiceAccount
- cifo-data: StatefulSet untuk PostgreSQL, Deployment untuk Redis

### Kriteria Penerimaan Fase 14:
- [ ] CI pipeline berjalan otomatis saat push/PR
- [ ] Semua jobs pass (lint, test, security scan, build)
- [ ] Images ter-push ke container registry
- [ ] Deploy ke staging otomatis saat merge ke main
- [ ] Deploy ke production memerlukan approval manual
- [ ] Helm charts dapat di-install di cluster

---

## FASE 15: DOKUMENTASI & PRODUCTION READINESS
**Tujuan**: Melengkapi semua dokumentasi teknis, runbook operasional, security policy, dan memastikan sistem siap untuk deployment production.
**Prasyarat**: Fase 14 selesai.
**Estimasi**: 2-3 hari.

### Tugas 15.1: Dokumentasi API
- Generate API reference dari OpenAPI spec
- Contoh request/response untuk setiap endpoint
- Penjelasan autentikasi dan rate limiting

### Tugas 15.2: Dokumentasi AI Agent
- Daftar semua tools dengan parameter dan contoh
- Batasan dan izin per role
- Contoh percakapan (use cases)

### Tugas 15.3: Deployment Guide
- Prasyarat infrastruktur (Kubernetes cluster, PostgreSQL, Redis, dll)
- Langkah deployment step-by-step
- Konfigurasi environment variables
- Checklist post-deployment

### Tugas 15.4: Incident Response Runbook
- Prosedur penanganan incident per severity
- Eskalasi matrix
- Contact information

### Tugas 15.5: Security Policy
- Kebijakan akses dan RBAC
- Prosedur rotasi credential
- Prosedur penanganan breach

### Tugas 15.6: Architecture Decision Records
- Finalisasi semua ADR:
  - 001: Mengapa Echo bukan Fiber
  - 002: Mengapa VictoriaMetrics bukan Prometheus standalone
  - 003: Mengapa AI Service terpisah (Python)
  - 004: Mengapa multi-model dengan circuit breaker
  - (tambahan ADR dari keputusan selama implementasi)

### Tugas 15.7: Final Verification
- Jalankan seluruh test suite (unit, integration, E2E, load)
- Jalankan security scan (gosec, trivy, gitleaks)
- Review semua audit_log entries
- Verifikasi backup/restore PostgreSQL
- Verifikasi semua alert rules trigger notifikasi yang benar
- Verifikasi AI chat berfungsi end-to-end
- Demo walkthrough seluruh fitur

### Kriteria Penerimaan Fase 15 (Final):
- [ ] Semua dokumentasi lengkap dan akurat
- [ ] Semua test pass
- [ ] Semua security scan bersih
- [ ] Backup/restore terverifikasi
- [ ] Alert -> Telegram flow terverifikasi
- [ ] AI chat end-to-end terverifikasi
- [ ] Sistem siap untuk deployment production

---

## RINGKASAN FASE

| Fase | Nama | Estimasi | Dependensi |
|------|------|----------|------------|
| 0 | Bootstrap Proyek & Monorepo | 1-2 hari | - |
| 1 | Infrastruktur Lokal (Testbed) | 1-2 hari | Fase 0 |
| 2 | Fondasi Backend | 3-4 hari | Fase 0, 1 |
| 3 | Autentikasi & Otorisasi | 2-3 hari | Fase 2 |
| 4 | Fondasi Frontend | 3-4 hari | Fase 3 |
| 5 | Monitoring Docker | 3-4 hari | Fase 4 |
| 6 | Monitoring Kubernetes & ArgoCD | 4-5 hari | Fase 5 |
| 7 | Real-time & WebSocket | 3-4 hari | Fase 6 |
| 8 | Alerting & Incident Management | 3-4 hari | Fase 7 |
| 9 | AI Service & Chat Agent | 5-7 hari | Fase 8 |
| 10 | Settings & Administrasi | 2-3 hari | Fase 9 |
| 11 | Observability (Tracing) | 2-3 hari | Fase 10 |
| 12 | Security Hardening | 3-4 hari | Fase 11 |
| 13 | Testing Komprehensif | 4-5 hari | Fase 12 |
| 14 | CI/CD Pipeline | 2-3 hari | Fase 13 |
| 15 | Dokumentasi & Production Readiness | 2-3 hari | Fase 14 |
| **Total** | | **40-56 hari** | |

---

## CATATAN PENTING

1. Estimasi waktu di atas adalah untuk satu developer berpengalaman yang bekerja full-time. Dengan tim, waktu bisa diparalelkan secara signifikan.
2. Fase 2-4 (Backend + Auth + Frontend Foundation) adalah yang paling kritis. Jika fondasi ini salah, seluruh bangunan di atasnya akan runtuh. Alokasikan waktu ekstra untuk review di sini.
3. Setiap akhir fase, lakukan review singkat: apakah ada penyimpangan dari arsitektur? Jika ada, perbarui dokumen arsitektur.
4. Jangan skip testing (Fase 13) atau security hardening (Fase 12). Hutang teknis di area ini akan menjadi bom waktu.
5. Data dummy tidak boleh ada. Jika testbed lokal belum siap, tunggu sampai siap. Jangan buat mock data "sementara" yang tidak pernah dihapus.

# Struktur Lengkap & Kamus Direktori Proyek CIFO Platform (`agent v2`)

**Dokumen**: Referensi Arsitektur & Peta Navigasi Berkas Repositori  
**Lokasi**: `d:\agent v2\arsitektur_diskusi\struktur_folder_lengkap.md`  
**Tanggal Penyusunan**: 2026-09-07  
**Cakupan**: 100% Seluruh Folder dan Berkas pada Repositori (484 Berkas Terlacak)  
**Tujuan**: Panduan komprehensif bagi arsitek, engineer, dan operator untuk memahami letak, peran, dan isi dari setiap berkas serta folder dalam monorepo platform.

---

## 1. Peta Pohon Direktori (Annotated Architecture Tree)

```text
agent v2/                                     # Root direktori monorepo CIFO Platform (Enterprise IT & AIOps)
├── .github/                                  # Konfigurasi otomasi GitHub Actions & CI/CD pipeline
│   └── workflows/                            # Definisi pipeline otomasi CI/CD YAML
│       ├── .gitkeep                          # Menjaga direktori tetap terlacak oleh Git
│       ├── ci.yml                            # Pipeline CI utama: multi-linter, multi-testing, security scan, docker build matrix
│       ├── deploy-production.yml             # Workflow promosi staging ke production dengan manual approval gate
│       └── deploy-staging.yml                # Workflow continuous delivery otomatis ke staging via GHCR & ArgoCD sync
├── .editorconfig                             # Konvensi format kode lintas editor (indentasi 2/4 spasi, LF, UTF-8)
├── .env.example                              # Template environment variable global tingkat root
├── .gitignore                                # Pola pengabaian berkas binary, temporary data, caches, dan secrets
├── .gitleaksignore                           # Allowlist untuk scanner gitleaks (mencegah false positive token dummy uji)
├── .golangci.yml                             # Konfigurasi linter Go terpusat (govet, staticcheck, errcheck, ineffassign)
├── Makefile                                  # Task runner root untuk build, test, docker, dan migration
├── README.md                                 # Dokumentasi ikhtisar platform, arsitektur, dan cara menjalankan
├── go.work                                   # Workspace file Go 1.22+ mengikat modul apps/backend
├── go.work.sum                               # Checksum verifikasi dependensi modul workspace Go
├── package.json                              # Konfigurasi root monorepo Node.js / Turborepo
├── package-lock.json                         # Kunci versi dependensi monorepo Node.js
├── turbo.json                                # Konfigurasi pipeline build caching Turborepo
│
├── apps/                                     # Direktori aplikasi / layanan mandiri (Microservices)
│   ├── ai-service/                           # Microservice AI Python 3.12 (FastAPI, Multi-Model LLM, Agentic Tools)
│   ├── backend/                              # Core Backend Engine Go (Echo, PostgreSQL pgxpool, WebSocket, OpenTelemetry)
│   └── frontend/                             # Modern Web Dashboard Next.js 16 / React 19 (Cyberpunk Glassmorphism)
│
├── arsitektur_diskusi/                       # Dokumentasi fondasi arsitektur sistem dan instruksi agen
│   ├── agent_instructions.md                 # Aturan baku agentic coding, standar kode, dan disiplin pengembangan
│   ├── arsitektur_sistem.md                  # Master blueprint arsitektur end-to-end, topologi jaringan, dan diagram data
│   ├── audit_arsitektur.md                   # Catatan audit kesesuaian dan kelayakan teknis platform
│   ├── plan.md                               # Master roadmap teknis Fase 0 sampai Fase 15 beserta kriteria penerimaan
│   └── struktur_folder_lengkap.md            # [DOKUMEN INI] Pemetaan isi lengkap setiap file dan folder di repositori
│
├── bin/                                      # Direktori binary lokal pendukung pengujian (gitignored)
│   └── k6.exe                                # Binary eksekusi load test K6 untuk benchmark throughput dan stress test
│
├── docs/                                     # Dokumentasi formal teknis, standar keamanan, dan SOP operasional
│   ├── adr/                                  # Architecture Decision Records (ADR)
│   │   ├── .gitkeep                          # Menjaga folder adr tetap terlacak oleh Git
│   │   ├── 001-use-echo-over-fiber.md        # ADR: Pemilihan framework HTTP Echo dibandingkan Fiber
│   │   ├── 002-victoriametrics-over-prometheus.md # ADR: Penggunaan VictoriaMetrics sebagai time-series engine utama
│   │   ├── 003-separate-ai-service-python.md # ADR: Pemisahan AI engine ke microservice Python FastAPI
│   │   ├── 004-multi-model-fallback.md       # ADR: Strategi multi-model fallback & circuit breaker pada LLM
│   │   └── 005-zero-mock-data-policy.md      # ADR: Kebijakan mutlak Zero Mock Data pada live runtime
│   ├── runbooks/                             # Standard Operating Procedures (SOP) insiden & pemulihan bencana
│   │   ├── runbook-disaster-recovery.md      # Panduan recovery basis data, klaster K8s, dan failover sistem
│   │   └── runbook-incident-handling.md      # Alur penanganan insiden alert, acknowledged, RCA, dan post-mortem
│   ├── ai-agent-capabilities.md              # Spesifikasi kapabilitas AI agent, tools allowlist, dan approval gate
│   ├── api-reference.md                      # Dokumentasi kontrak lengkap REST API dan WebSocket format envelope
│   ├── deployment-guide.md                   # Petunjuk deployment ke staging dan production berbasis Helm/ArgoCD
│   ├── incident-response.md                  # Kebijakan manajemen siklus insiden, SLA eskalasi, dan level severity
│   └── security-policy.md                    # Kebijakan keamanan enterprise, manajemen rahasia, RBAC, dan audit trail
│
├── implementasi_plan/                        # Spesifikasi implementasi bertahap dan rekap cross-check audit
│   ├── AGENT_HANDOFF_PROGRESS.md             # Status live checkpoint serah terima tugas antar agen pengembang
│   ├── cross_check_f3_f5.md                  # Laporan audit dan optimalisasi Fase 3 sampai Fase 5
│   ├── cross_check_f6_f10.md                 # Laporan audit dan optimalisasi Fase 6 sampai Fase 10
│   ├── cross_check_f10_f14.md                # Laporan audit dan optimalisasi Fase 10 sampai Fase 14
│   ├── f0-1.md                               # Rencana teknis Fase 0 (Setup Repositori) & Fase 1 (Testbed Lokal)
│   ├── f2.md s/d f15.md                      # Rencana teknis spesifik untuk tiap-tiap Fase 2 hingga Fase 15
│   ├── implementation_plan1 s/d plan8        # Rencana kerja historis iteratif awal pengembangan
│   ├── perbaikan_f3_f5_dan_kesiapan_f6.md    # Dokumen tindak lanjut penyelesaian audit Fase 3-5
│   └── perintahawal.md                       # Catatan instruksi awal dari pengguna mengenai arsitektur dasar
│
├── infrastructure/                           # Infrastructure as Code (IaC), Manifest K8s, GitOps, dan Keamanan
│   ├── kubernetes/                           # Manifest Kubernetes terstruktur (Kustomize & Helm)
│   │   ├── base/                             # Base kustomization manifest Kubernetes platform
│   │   ├── charts/                           # Helm charts platform CIFO
│   │   │   ├── cifo-ai-service/              # Subchart Helm mandiri untuk AI Microservice
│   │   │   ├── cifo-backend/                 # Subchart Helm mandiri untuk Go Backend Core
│   │   │   ├── cifo-data/                    # Subchart Helm mandiri untuk PostgreSQL & Redis
│   │   │   ├── cifo-frontend/                # Subchart Helm mandiri untuk Next.js Frontend
│   │   │   └── cifo-platform/                # Umbrella Helm Chart orkestrator menyeluruh platform
│   │   └── overlays/                         # Environment-specific Kustomize overlays
│   │       ├── production/                   # Konfigurasi overlay production HA (3+ replika, resource tinggi)
│   │       └── staging/                      # Konfigurasi overlay staging (1 replika, logging debug)
│   ├── local-testbed/                        # Lingkungan runtime lokal developer (Docker Compose & Observability)
│   │   ├── alertmanager/                     # Konfigurasi routing alert dan integrasi webhook Alertmanager
│   │   ├── argocd/                           # Manifest CRD dan komponen ArgoCD GitOps lokal K3d
│   │   ├── docker-proxy/                     # Template HAProxy untuk Tecnativa Docker Socket Proxy hardening
│   │   ├── grafana/                          # Provisioning datasource Grafana (Loki ke Tempo trace linking)
│   │   ├── k3d/                              # Skrip inisialisasi dan konfigurasi klaster multi-node K3d
│   │   ├── keycloak/                         # Realm export JSON untuk SSO OIDC & manajemen pengguna Keycloak
│   │   ├── prometheus/                       # Konfigurasi scrape target dan alerting rules Prometheus
│   │   ├── tempo/                            # Konfigurasi distributed tracing ingestion Grafana Tempo
│   │   ├── docker-compose.monitoring.yml     # Standalone compose untuk observability stack
│   │   └── docker-compose.yml                # Master local compose (Postgres, Redis, VM, Loki, Tempo, Vault, dll.)
│   ├── security/                             # Pengerasan keamanan klaster Kubernetes
│   │   ├── network-policies/                 # 5 Enterprise NetworkPolicies (default-deny, isolasi data/backend)
│   │   ├── rbac/                             # ServiceAccount, ClusterRole, dan RoleBinding untuk cifo-ai-agent-sa
│   │   └── vault/                            # Policy HCL pembatasan akses rahasia Vault KV v2
│   └── terraform/                            # Modul otomasi provisi cloud infrastructure (IaC)
│       ├── environments/                     # Konfigurasi environment Terraform (dev, staging, prod)
│       └── modules/                          # Reusable modules (VPC, K8s cluster, database, storage)
│
├── packages/                                 # Shared library dan kontrak antarmuka monorepo
│   ├── api-contracts/                        # Definisi kontrak API terpusat (Single Source of Truth)
│   │   ├── openapi.yaml                      # Spesifikasi OpenAPI 3.0 REST endpoints CIFO Backend
│   │   └── proto/                            # Schema Protobuf gRPC antarlayanan (ai_service & common)
│   ├── eslint-config/                        # Standard linting config bersama monorepo
│   │   └── index.js                          # Export modul konfigurasi ESLint terstandar
│   └── theme/                                # Token desain visual cyberpunk glassmorphism
│       ├── colors.json                       # Palet warna primer, aksen neon, status badge, dan border glass
│       └── typography.json                   # Konfigurasi font family, ukuran, dan line-height UI
│
├── referensi tampilan/                       # Tangkapan layar referensi UI / UX acuan platform
│   └── Screenshot 2026-09-03 *.png           # 10 file visual acuan dashboard, detail container, log modal, chat
│
├── scratch/                                  # Ruang kerja pengujian sementara (gitignored)
│   ├── ratelimit_resp.json                   # Sampel respons JSON hasil uji rate limiting
│   └── test-ws.js                            # Skrip uji coba cepat konektivitas handshake WebSocket
│
├── scripts/                                  # Skrip otomasi shell dan PowerShell untuk seluruh alur sistem
│   ├── build-backend.ps1 & build-linux.ps1   # Skrip kompilasi binary Go server untuk Windows dan Linux
│   ├── check-argocd-pods.ps1 / status.ps1    # Skrip diagnosis kesehatan Pod ArgoCD dan layanan klaster
│   ├── init-vault.ps1                        # Otomasi inisialisasi, unseal, dan seeding secrets HashiCorp Vault
│   ├── run-*-tests.ps1                       # Runner pengujian (backend, frontend, ai, repo, e2e, load, integration)
│   ├── seed-data.sql                         # SQL seed awal konfigurasi sistem wajib dan role default
│   ├── setup-local.ps1 & setup-local.sh      # Skrip instalasi lokal end-to-end (Docker, DB, Redis, K3d, K6)
│   ├── test-phase*.ps1                       # Master verification scripts untuk tiap fase pengembangan (Fase 1-15)
│   └── verify-*.ps1                          # Skrip verifikasi kesehatan rute frontend dan status services
│
└── tests/                                    # Pengujian integrasi level sistem dan uji beban
    ├── e2e/                                  # Folder penampung pengujian E2E tingkat sistem
    ├── integration/                          # Modul Go pengujian integrasi lintas dependensi live
    └── load/                                 # Skenario pengujian beban K6 (API throughput & WebSocket stress)
```

---

## 2. Rincian Katalog Berkas & Fungsinya (File-by-File Breakdown)

### 2.1. Berkas Tingkat Root (Root-Level Files)

| Nama Berkas | Tipe | Penjelasan Isi & Fungsi Teknis |
|---|:---:|---|
| `.editorconfig` | Config | Menstandarkan format teks editor: UTF-8, LF line endings, indentasi 2 spasi (JS/TS/JSON/YAML) dan tab untuk Go. |
| `.env.example` | Config | Template environment variables global: DSN database, Redis, port server, kunci API, dan flag OTel. |
| `.gitignore` | Git | Mengabaikan node_modules, .next, bin, coverage, .gocache, .cache, .env, *.exe, dan temporary test files. |
| `.gitleaksignore` | Security | Mengabaikan string dummy/test token pada pengujian unit agar tidak memicu false-positive alert pada security scan. |
| `.golangci.yml` | Linter | Konfigurasi linter Go: mengaktifkan govet, errcheck, staticcheck, unused, gofmt, dan ineffassign dengan timeout 5m. |
| `Makefile` | Build | Otomasi perintah developer: `make dev`, `make test`, `make build`, `make docker-up`, `make migrate-up`, dan `make seed`. |
| `README.md` | Doc | Dokumentasi pengantar proyek, prasyarat sistem, arsitektur arsitektural, dan panduan quick start lokal. |
| `go.work` | Go | Workspace file Go 1.22+ yang menyatukan modul `apps/backend` dalam satu konteks workspace terpadu. |
| `go.work.sum` | Go | Checksum kriptografis untuk dependensi modul pada workspace Go. |
| `package.json` | Node | Manifest monorepo root mendefinisikan workspaces (`apps/*`, `packages/*`) dan skrip global lint, test, build. |
| `package-lock.json` | Node | Mengunci dependency tree seluruh monorepo Node.js untuk reproduktibilitas build mutlak. |
| `turbo.json` | Build | Konfigurasi build pipeline Turborepo: mendefinisikan dependency graph, caching outputs, dan task parallelism. |

---

### 2.2. Layanan AI Microservice (`apps/ai-service`)

Direktori ini berisi microservice Python 3.12 berbasis **FastAPI** yang mengorkestrasi inferensi LLM, pencegahan prompt injection, fallback multi-model, serta registrasi tools diagnosa dan remediasi.

| Path Berkas | Fungsi & Penjelasan Isi |
|---|---|
| `Dockerfile` | Multi-stage production build (`python:3.12-slim` builder -> runner non-root UID 10001) mengisolasi wheels dependensi. |
| `.dockerignore` | Mengabaikan `__pycache__`, `.pytest_cache`, `.venv`, dan `.env` saat Docker build. |
| `.env.example` | Template variabel lingkungan AI service: Google API Key, OpenAI Key, Anthropic Key, Ollama URL, dan Vault config. |
| `pyproject.toml` | Konfigurasi project metadata Python dan setting pengujian Pytest (`testpaths = ["tests"]`). |
| `requirements.txt` | Daftar dependensi produksi: fastapi, uvicorn, opentelemetry-api, httpx, hvac (Vault), pydantic. |
| `app/__init__.py` | Inisialisasi package Python tingkat aplikasi. |
| `app/main.py` | Entrypoint FastAPI: lifespan startup/shutdown, middleware CORS, OpenTelemetry tracing, dan router endpoints. |
| `app/agent/__init__.py` | Inisialisasi package orkestrator agen AI. |
| `app/agent/circuit_breaker.py` | State machine Circuit Breaker (Closed -> Open -> Half-Open) memutus provider saat 3 kegagalan beruntun dalam 60s. |
| `app/agent/memory.py` | Pengelola riwayat percakapan berbasis sliding window (maks 20 pesan) dengan TTL session 60 menit. |
| `app/agent/orchestrator.py` | Engine multi-model fallback: mencoba Google Gemini -> OpenAI -> Anthropic -> Ollama -> Degraded Mock Mode. |
| `app/agent/sanitizer.py` | Pertahanan Prompt Injection: regex scanner untuk pola override instruksi, jailbreak, sudo, dan destructive SQL/bash. |
| `app/config/__init__.py` | Inisialisasi package konfigurasi. |
| `app/config/settings.py` | Pydantic Settings memvalidasi environment variables dengan casting tipe data otomatis dan default values. |
| `app/core/__init__.py` | Inisialisasi core utilities package. |
| `app/core/telemetry.py` | Instrumentasi OpenTelemetry: TracerProvider, OTLP HTTP span exporter ke Tempo, dan W3C propagator. |
| `app/core/vault.py` | Modul pembaca rahasia HashiCorp Vault KV v2 via REST API dengan in-memory cache dan fallback ke environment variables. |
| `app/prompts/__init__.py` | Inisialisasi package prompt templates. |
| `app/prompts/diagnosis_prompt.txt` | Template prompt sistem untuk analisis akar penyebab insiden (Root Cause Analysis / RCA) dari log dan metrik. |
| `app/prompts/summarize_prompt.txt` | Template prompt peringkasan histori insiden dan batching storm alerts. |
| `app/prompts/system_prompt.txt` | System prompt utama AIOps: persona SRE level staff, batasan keamanan, format output Markdown profesional. |
| `app/providers/__init__.py` | Inisialisasi package LLM providers. |
| `app/providers/base.py` | Abstract Base Class `LLMProvider` mendefinisikan interface standar `generate()` dan `health()`. |
| `app/providers/anthropic_provider.py` | Adaptor integrasi model Anthropic Claude 3.5 Sonnet. |
| `app/providers/google_provider.py` | Adaptor integrasi Google AI Studio (Gemini 2.0 Flash) sebagai penyedia utama berkecepatan tinggi. |
| `app/providers/mock_provider.py` | Standby provider deterministik untuk degraded mode saat koneksi internet / seluruh API eksternal terputus. |
| `app/providers/ollama_provider.py` | Adaptor integrasi runtime lokal Ollama (Llama 3 / Mistral) untuk operasi on-premise terisolasi. |
| `app/providers/openai_provider.py` | Adaptor integrasi OpenAI (GPT-4o) sebagai secondary high-accuracy fallback. |
| `app/tools/__init__.py` | Inisialisasi package agentic tools. |
| `app/tools/base.py` | Model dasar definisi tool, metadata skema JSON, dan penanda `requires_approval` (Human-in-the-Loop). |
| `app/tools/argocd_tools.py` | Definisi tools ArgoCD: `get_argocd_app_status`, `get_argocd_history`, dan `sync_argocd_app` (Write/Approval). |
| `app/tools/docker_tools.py` | Definisi tools Docker: `list_docker_containers`, `get_docker_stats`, `restart_container`, dan `stop_container`. |
| `app/tools/kubectl_tools.py` | Definisi tools Kubernetes: `get_pod_status`, `get_container_logs`, `scale_deployment`, dan `restart_deployment`. |
| `tests/__init__.py` | Inisialisasi package pengujian unit. |
| `tests/test_api.py` | Pengujian unit endpoint HTTP `/healthz`, `/readyz`, `/api/v1/chat`, `/api/v1/diagnose`, `/api/v1/tools`. |
| `tests/test_circuit_breaker.py` | Pengujian transisi state circuit breaker dari Closed ke Open dan mekanisme recovery Half-Open. |
| `tests/test_health.py` | Pengujian probes liveness dan readiness pada microservice AI. |
| `tests/test_memory.py` | Pengujian mekanisme sliding window riwayat pesan dan batas pemotongan tokens. |
| `tests/test_orchestrator.py` | Pengujian alur fallback multi-model orchestrator saat primary provider mengalami timeout/error. |
| `tests/test_sanitizer.py` | Pengujian komprehensif prompt sanitizer terhadap pola injection, bypass, dan validasi tool allowlist. |
| `tests/test_tools.py` | Pengujian validitas skema JSON seluruh 13 tools dan pemisahan tool read-only vs tool write (Human-in-the-loop). |
| `tests/test_vault.py` | Pengujian integrasi Vault reader, mekanisme fallback env, dan injection credentials ke settings. |

---

### 2.3. Layanan Backend Core (`apps/backend`)

Direktori ini berisi core engine backend platform yang dibangun dengan **Golang 1.22+ / Echo Framework** menggunakan pola *Clean Architecture / Hexagonal*.

| Path Berkas | Fungsi & Penjelasan Isi |
|---|---|
| `Dockerfile` | Multi-stage build (`golang:1.22-alpine` builder -> `alpine:3.20` non-root `appuser`) dengan binary stripped `-ldflags="-s -w"`. |
| `.dockerignore` | Mengabaikan `.gocache`, `server.exe`, `.env`, dan files lokal saat build container backend. |
| `.env.example` | Template konfigurasi environment backend (port, DB DSN, Redis, Vault, Docker host, ArgoCD, OTel). |
| `go.mod` & `go.sum` | Definisi modul Go `github.com/cifo-monitoring/backend` dan checksum seluruh library dependensi. |
| `cmd/server/main.go` | Main entrypoint server: wiring dependency injection, pendaftaran middleware global, routing, dan graceful shutdown. |
| `cmd/worker/main.go` | Entrypoint background worker untuk memproses tugas berkala di luar siklus request HTTP. |
| `internal/config/config.go` | Struct `Config`, parsing environment variable dengan nilai default, dan integrasi fallback HashiCorp Vault. |
| `internal/config/config_test.go` | Pengujian unit parsing konfigurasi dan validasi environment wajib. |
| `internal/handler/ai_handler.go` | Echo HTTP handler untuk AI chat, riwayat sesi, approval eksekusi tools, dan trigger AI RCA. |
| `internal/handler/ai_handler_test.go` | Unit test handler AI dengan mocking service layer. |
| `internal/handler/argocd_handler.go` | Echo HTTP handler untuk mengelola aplikasi ArgoCD (list, detail, history, sync, overview). |
| `internal/handler/auth_handler.go` | Echo HTTP handler untuk login SSO/Keycloak, profil `/auth/me`, logout, dan audit listing. |
| `internal/handler/docker_handler.go` | Echo HTTP handler untuk menginspeksi kontainer Docker, stats CPU/memory, logs, dan aksi restart/stop. |
| `internal/handler/health_handler.go` | Handler probe Kubernetes `/healthz` (liveness) dan `/readyz` (readiness) memeriksa DB & Redis. |
| `internal/handler/health_handler_test.go` | Unit test probe liveness & readiness. |
| `internal/handler/incident_handler.go` | Handler webhook Alertmanager v4, lifecycle insiden (ack, resolve, close), dan statistik insiden. |
| `internal/handler/incident_handler_test.go` | Unit test penerimaan webhook alert dan transisi status insiden. |
| `internal/handler/kubernetes_handler.go` | Handler operasi klaster K8s: daftar pods, deployment restart/scale, node metrics, services, overview. |
| `internal/handler/metrics_handler.go` | Handler eksposur metrik internal format Prometheus scraper pada endpoint `/metrics`. |
| `internal/handler/monitoring_handler.go` | Handler agregasi telemetri host (CPU, Memory, Network) dan ringkasan status arsitektur. |
| `internal/handler/settings_handler.go` | Handler pengaturan platform: sistem, notifikasi Telegram, manajemen user RBAC, dan audit trail. |
| `internal/handler/settings_handler_test.go` | Unit test handler settings dan proteksi self-deactivation admin. |
| `internal/handler/websocket_handler.go` | Handler upgrade koneksi HTTP ke WebSocket duplex `/ws` dengan validasi auth handshake. |
| `internal/handler/websocket_handler_test.go` | Unit test upgrade protokol dan pendaftaran client connection. |
| `internal/integration/integration.go` | Factory wrapper pembantu inisialisasi integrasi eksternal. |
| `internal/integration/ai_client.go` | Client HTTP komunikasi ke Python AI Service dilengkapi injeksi OpenTelemetry W3C trace context. |
| `internal/integration/argocd_client.go` | Integrasi k8s dynamic client untuk memanipulasi Custom Resource Definition `Applications` ArgoCD. |
| `internal/integration/docker_client.go` | Client resmi Docker Engine SDK (`client.NewClientWithOpts`) via socket proxy TCP. |
| `internal/integration/k8s_client.go` | Client resmi Kubernetes (`client-go`) untuk manipulasi pods, deployments, nodes, dan events. |
| `internal/integration/telegram_client.go` | Client pengiriman notifikasi pesan Telegram Bot API dengan Markdown formatter. |
| `internal/middleware/auth.go` | Echo middleware memvalidasi token JWT / dev token, mengekstrak user claims ke context. |
| `internal/middleware/cors.go` | Echo middleware menangani header Cross-Origin Resource Sharing (CORS). |
| `internal/middleware/logger.go` | Middleware mencatat log setiap HTTP request via slog lengkap dengan context trace ID. |
| `internal/middleware/ratelimit.go` | Middleware IP rate limiting berbasis Redis Token Bucket (100 req/min default). |
| `internal/middleware/rbac.go` | Middleware penegakan hak akses berbasis role (`admin`, `devops`, `viewer`). |
| `internal/middleware/rbac_test.go` | Unit test penolakan HTTP 403 Forbidden bagi role tanpa hak akses memadai. |
| `internal/middleware/recovery.go` | Middleware menangkap panic runtime dan mengembalikan error RFC 7807 tanpa server crash. |
| `internal/middleware/security_headers.go` | Middleware menyuntikkan 7 header keamanan wajib: nosniff, X-Frame-Options DENY, CSP, HSTS, dll. |
| `internal/middleware/security_headers_test.go` | Unit test verifikasi keberadaan dan nilai header keamanan pada HTTP response. |
| `internal/middleware/tracer.go` | OpenTelemetry Echo middleware: ekstraksi traceparent, pembuatan root span, penyuntikan `X-Trace-Id`. |
| `internal/middleware/tracer_test.go` | Unit test verifikasi span generation dan propagasi header W3C. |
| `internal/model/ai.go` | Domain model sesi percakapan AI, pesan, tool call schemas, approval state, dan usage tracking. |
| `internal/model/argocd.go` | Domain model representasi aplikasi ArgoCD, status sync, health, dan revision history. |
| `internal/model/audit.go` | Domain model immutable audit log mencatat aktor, resource, aksi, IP, dan timestamp. |
| `internal/model/docker.go` | Domain model ringkasan kontainer, spesifikasi image, volume, jaringan, dan stats CPU/RAM. |
| `internal/model/incident.go` | Domain model siklus insiden: alert firing, severity, acknowledged, resolved, dan mitigasi. |
| `internal/model/kubernetes.go` | Domain model Pods, Deployments, Nodes, Services, dan metrik klaster K8s. |
| `internal/model/settings.go` | Domain model pengaturan platform (sistem, preferensi model AI, notifikasi, sesi aktif). |
| `internal/model/user.go` | Domain model akun pengguna, sinkronisasi Keycloak, dan penetapan role. |
| `internal/repository/db.go` | Inisialisasi connection pool PostgreSQL (`pgxpool.Pool`) dilengkapi query tracer hooks. |
| `internal/repository/redis.go` | Inisialisasi client koneksi Redis untuk cache dan rate limiting. |
| `internal/repository/migrator.go` | Engine eksekusi migrasi database SQL up/down otomatis saat backend booting. |
| `internal/repository/ai_repository.go` | Repository PostgreSQL untuk tabel `ai_sessions`, `ai_messages`, dan `ai_action_audit_log`. |
| `internal/repository/ai_repository_test.go` | Unit test operasi database AI repository terhadap PostgreSQL live. |
| `internal/repository/audit_repository.go` | Repository pencatatan dan pencarian riwayat jejak audit pada tabel `audit_log`. |
| `internal/repository/audit_repository_test.go` | Unit test persistensi audit trail. |
| `internal/repository/incident_repository.go` | Repository PostgreSQL untuk tabel `incidents` dan query alert yang belum di-acknowledge. |
| `internal/repository/incident_repository_test.go` | Unit test siklus hidup insiden dan pencarian alert terbuka. |
| `internal/repository/settings_repository.go` | Repository query dan update tabel `system_settings` dan `notification_settings`. |
| `internal/repository/settings_repository_test.go` | Unit test query dan mutasi konfigurasi sistem. |
| `internal/repository/user_repository.go` | Repository manajemen pengguna: CRUD, penetapan role, dan deactivation akun. |
| `internal/repository/user_repository_test.go` | Unit test manipulasi user data di PostgreSQL. |
| `internal/security/vault_client.go` | Implementasi Go client untuk membaca secrets HashiCorp Vault KV v2 dengan sync mutex cache. |
| `internal/security/vault_client_test.go` | Unit test pembacaan rahasia Vault dan fallback konfigurasi. |
| `internal/service/service.go` | Definisi interface umum lapisan bisnis service. |
| `internal/service/ai_service.go` | Logika orkestrasi percakapan AI, validasi tool eksekusi, approval gate, dan delegasi ke AI microservice. |
| `internal/service/ai_service_test.go` | Unit test skenario percakapan AI, auto-RCA insiden, dan approval tool write/read. |
| `internal/service/argocd_service.go` | Logika sinkronisasi aplikasi ArgoCD, inspeksi resource tree, dan riwayat revisi GitOps. |
| `internal/service/argocd_service_test.go` | Unit test operasi sinkronisasi dan penanganan error ArgoCD. |
| `internal/service/auth_service.go` | Logika autentikasi: validasi RSA JWT Keycloak, dev token bypass, dan session management. |
| `internal/service/auth_service_test.go` | Unit test validasi token, ekstraksi claims, dan caching RSA public keys. |
| `internal/service/docker_service.go` | Logika manajemen kontainer: listing, stream logs, monitoring metrik, restart, dan stop kontainer. |
| `internal/service/docker_service_test.go` | Unit test kontrol Docker kontainer dan demux stream log output. |
| `internal/service/incident_service.go` | State machine insiden: pemrosesan webhook Alertmanager, auto-escalation ticker, dan notifikasi. |
| `internal/service/incident_service_test.go` | Unit test alur transisi status insiden dan eskalasi otomatis. |
| `internal/service/jwks.go` | Caching JWKS (JSON Web Key Set) dari endpoint OIDC Keycloak dengan refresh berkala. |
| `internal/service/kubernetes_service.go` | Logika orkestrasi Kubernetes: inspeksi pods, restart deployment (rollout), scale replicas. |
| `internal/service/kubernetes_service_test.go` | Unit test manipulasi deployment K8s dan stream pod logs. |
| `internal/service/monitoring_service.go` | Agregasi telemetri CPU/Memory/Network dan komputasi persentase kesehatan arsitektur. |
| `internal/service/monitoring_service_test.go` | Unit test agregasi data telemetri monitoring. |
| `internal/service/notification_service.go` | Logika perutean notifikasi insiden ke WebSocket realtime dan dispatch bot Telegram. |
| `internal/service/notification_service_test.go` | Unit test pengiriman notifikasi dan penanganan kegagalan gateway. |
| `internal/service/settings_service.go` | Logika tata kelola platform: merge settings, pengujian Telegram alert, dan proteksi akun admin. |
| `internal/service/settings_service_test.go` | Unit test mutasi settings dan pengujian dispatch notifikasi uji. |
| `internal/service/telegram_service.go` | Logika pembentukan format pesan Markdown Telegram, rate limit 30 pesan/menit, dan batching alert storm. |
| `internal/service/telegram_service_test.go` | Unit test format alert individual, batch summary storming, dan retry queue Redis. |
| `internal/ws/client.go` | WebSocket client connection handler: goroutine ReadPump dan WritePump terpisah dengan ping/pong heartbeat. |
| `internal/ws/hub.go` | WebSocket hub: thread-safe channel broker, register/unregister klien, dan broadcast topic terarah. |
| `internal/ws/hub_test.go` | Unit test konkurensi WebSocket Hub dan isolasi pesan per topic channel. |
| `internal/ws/message.go` | Definisi struktur pesan WebSocket (`WSMessage`, `EventPayload`, `LogPayload`, `NotificationPayload`). |
| `internal/ws/streamer.go` | Streamer asinkron mengalirkan live log Docker dan K8s ke channel WebSocket saat klien tersubscribe. |
| `migrations/001_create_users.*.sql` | Skrip migrasi DDL pembuatan dan rollback tabel `users`. |
| `migrations/002_create_ai_sessions.*.sql` | Skrip migrasi DDL tabel `ai_sessions` dan `ai_messages`. |
| `migrations/003_create_audit_log.*.sql` | Skrip migrasi DDL tabel `audit_log` dan `ai_action_audit_log`. |
| `migrations/004_create_incidents.*.sql` | Skrip migrasi DDL tabel `incidents`. |
| `migrations/005_create_ai_usage_tracking.*.sql` | Skrip migrasi DDL tabel `ai_usage_tracking` (tokens, cost USD). |
| `migrations/006_create_notification_settings.*.sql` | Skrip migrasi DDL tabel `notification_settings`. |
| `migrations/007_create_notifications.*.sql` | Skrip migrasi DDL tabel `notifications` (riwayat in-app alert). |
| `migrations/008_create_system_settings.*.sql` | Skrip migrasi DDL tabel `system_settings` (konfigurasi global dan budget AI). |
| `pkg/apperror/error.go` | Definisi `AppError` dengan HTTP status, error code, user message, dan unwrap interface. |
| `pkg/apperror/error_test.go` | Unit test pembuatan dan chaining `AppError`. |
| `pkg/logger/logger.go` | Wrapper structured logging `slog` dengan kustom handler penyuntik `trace_id` dan `span_id`. |
| `pkg/logger/logger_test.go` | Unit test format JSON logger dan context trace extraction. |
| `pkg/telemetry/db_tracer.go` | Implementasi `pgx.QueryTracer` hook menyematkan span OpenTelemetry pada setiap query database. |
| `pkg/telemetry/tracer.go` | Tracer provider OpenTelemetry Go dengan exporter OTLP HTTP ke Tempo. |
| `pkg/telemetry/tracer_test.go` | Unit test inisialisasi OTel TracerProvider dan lifecycle shutdown. |
| `pkg/validator/validator.go` | Helper validasi input payload HTTP request. |
| `pkg/validator/validator_test.go` | Unit test validator aturan field. |
| `tests/integration/argocd_api_test.go` | Pengujian integrasi live API ArgoCD terhadap endpoint backend port 8080. |
| `tests/integration/auth_api_test.go` | Pengujian integrasi autentikasi live Keycloak / dev token di port 8080. |
| `tests/integration/docker_api_test.go` | Pengujian integrasi operasi Docker API live melalui socket proxy. |

---

### 2.4. Layanan Frontend Dashboard (`apps/frontend`)

Direktori ini berisi antarmuka pengguna berbasis **Next.js 16 / React 19** dengan estetika modern *Cyberpunk Glassmorphism*.

| Path Berkas | Fungsi & Penjelasan Isi |
|---|---|
| `Dockerfile` | Multi-stage production build Next.js output standalone, non-root user `nextjs:nodejs`. |
| `.dockerignore` | Mengabaikan node_modules, .next, coverage, dan temporary files saat Docker build. |
| `eslint.config.mjs` | Konfigurasi ESLint 9 Flat Config dengan aturan teroptimasi untuk TypeScript dan React 19. |
| `next.config.ts` | Konfigurasi Next.js: output standalone, strict mode, dan pengaturan header keamanan. |
| `package.json` | Manifest frontend: dependensi React 19, Tailwind CSS v4, Lucide icons, TanStack Query, Zustand, ECharts. |
| `playwright.config.ts` | Konfigurasi framework E2E Playwright: browser Chromium/Chrome headless, baseURL http://127.0.0.1:3001. |
| `postcss.config.mjs` | Konfigurasi PostCSS untuk kompilasi Tailwind CSS. |
| `vitest.config.mts` | Konfigurasi Vitest: jsdom environment, v8 coverage engine, `isolate: false` untuk kecepatan tinggi. |
| `vitest.setup.ts` | Setup lifecycle pengujian: mendaftarkan auto-cleanup DOM setelah tiap tes dijalankan. |
| `e2e/01-login.spec.ts` | E2E Test: Halaman login, input kredensial, SSO Keycloak button, dan profil dev-token 1-klik. |
| `e2e/02-dashboard.spec.ts` | E2E Test: Halaman utama, kartu KPI, widget status arsitektur, dan grafik metrik. |
| `e2e/03-docker.spec.ts` | E2E Test: Tabel kontainer Docker, filter status, pencarian, dan modal inspect detail. |
| `e2e/04-kubernetes.spec.ts` | E2E Test: Tabel Pods, filter namespace, dialog terminal viewer log pod. |
| `e2e/05-incidents.spec.ts` | E2E Test: Tabel insiden, badge severity, modal detail insiden, dan tombol aksi status. |
| `e2e/06-ai-chat.spec.ts` | E2E Test: Drawer floating AI chat, tombol trigger, pemilihan prompt cepat, pesan terkirim. |
| `src/app/globals.css` | Stylesheet global: styling scrollbar cyberpunk, efek neon glow, grid background, dan animasi radar. |
| `src/app/layout.tsx` | Root layout aplikasi menyematkan font Geist/Inter, QueryClientProvider, dan ToastProvider. |
| `src/app/page.tsx` | Landing page root yang otomatis mengarahkan ke `/login` atau `/monitoring` sesuai status sesi. |
| `src/app/page.test.tsx` | Unit test redirect root page. |
| `src/app/(auth)/login/page.tsx` | Halaman login bernuansa glassmorphism: form otentikasi, Keycloak SSO, dan selector Dev Profile cepat. |
| `src/app/(dashboard)/layout.tsx` | Shell layout dashboard: Sidebar navigasi, Header bar, floating AI assistant, dan toast center. |
| `src/app/(dashboard)/monitoring/page.tsx` | Dashboard utama: KPI cards (containers, pods, apps, alerts), resource usage charts, event logs. |
| `src/app/(dashboard)/docker/page.tsx` | Halaman inventaris Docker: tabs kontainer, images, volumes, networks, dan system info. |
| `src/app/(dashboard)/docker/containers/page.tsx` | Tampilan spesifik daftar kontainer Docker lengkap dengan status chip dan aksi restart/stop. |
| `src/app/(dashboard)/docker/images/page.tsx` | Tampilan daftar Docker images, tag, ukuran byte, dan creation date. |
| `src/app/(dashboard)/docker/networks/page.tsx` | Tampilan daftar Docker networks (bridge, host, overlay, cifo-network). |
| `src/app/(dashboard)/docker/volumes/page.tsx` | Tampilan daftar Docker volumes dan mountpoint. |
| `src/app/(dashboard)/kubernetes/page.tsx` | Halaman manajemen Kubernetes: tab Pods, Deployments, Nodes, Services, dan pod log viewer dialog. |
| `src/app/(dashboard)/argocd/page.tsx` | Halaman monitoring ArgoCD: status sync (Synced/OutOfSync), health (Healthy/Degraded), resource tree. |
| `src/app/(dashboard)/incidents/page.tsx` | Halaman insiden: filter severity/status, aksi Acknowledge/Resolve/Close, dan tombol Run AI RCA. |
| `src/app/(dashboard)/incidents/page.test.tsx` | Unit test rendering halaman insiden dan KPI cards. |
| `src/app/(dashboard)/settings/page.tsx` | Halaman pengaturan 5 tab: General, Notifications, AI Config, Users & RBAC, Security & Sessions. |
| `src/app/(dashboard)/settings/page.test.tsx` | Unit test rendering tab navigasi dan kontrol form pengaturan. |
| `src/components/ai-chat/chat-bubble.tsx` | Komponen bubble pesan chat: pesan user vs respons asisten dengan format Markdown dan code block. |
| `src/components/ai-chat/chat-container.tsx` | Floating drawer obrolan AI: pemilihan model/provider, histori sesi, dan area scroll percakapan. |
| `src/components/ai-chat/chat-container.test.tsx` | Unit test interaksi buka/tutup drawer chat AI. |
| `src/components/ai-chat/chat-input.tsx` | Input bar chat AI: multiline auto-grow, shortcut enter to send, dan chip prompt sugesti. |
| `src/components/ai-chat/chat-input.test.tsx` | Unit test input teks chat dan submit event. |
| `src/components/ai-chat/model-indicator.tsx` | Badge indikator model aktif: menampilkan provider (Google/OpenAI/Ollama) dan status circuit breaker. |
| `src/components/ai-chat/tool-approval.tsx` | Card persetujuan Human-in-the-Loop untuk aksi destruktif (restart/scale/stop) dengan tombol Approve/Reject. |
| `src/components/ai-chat/tool-approval.test.tsx` | Unit test rendering parameter tool dan trigger approval callback. |
| `src/components/auth/AuthControl.tsx` | Komponen seleksi profil cepat autentikasi (Admin, DevOps, Viewer) untuk pengujian lokal. |
| `src/components/layout/breadcrumb.tsx` | Navigasi breadcrumb dinamis sesuai path URL saat ini. |
| `src/components/layout/docker-nav-tabs.tsx` | Tab navigasi sekunder sub-halaman Docker (Containers, Images, Networks, Volumes). |
| `src/components/layout/header.tsx` | Top header bar: status koneksi WebSocket, live clock, notification bell dropdown, dan user profile menu. |
| `src/components/layout/sidebar.tsx` | Navigasi sidebar collapsible bernuansa glassmorphic dengan ikon Lucide untuk setiap modul. |
| `src/components/providers/app-provider.tsx` | Master provider menggabungkan seluruh context provider aplikasi. |
| `src/components/providers/notification-toast-provider.tsx` | Provider toast alert real-time yang tersambung ke WebSocket topic `notifications`. |
| `src/components/providers/query-provider.tsx` | Provider TanStack React Query dengan caching dan background refetching otomatis. |
| `src/components/terminal/log-terminal.tsx` | Virtualized console log viewer: syntax highlighting, text search, auto-scroll lock, dan ekspor .log. |
| `src/components/terminal/log-terminal.test.tsx` | Unit test terminal log viewer dan interaksi pencarian. |
| `src/components/ui/badge.tsx` | Komponen badge status neon: success, warning, critical, info, primary, neutral. |
| `src/components/ui/button.tsx` | Komponen tombol interaktif: varian primary, secondary, outline, ghost, destructive, cyberpunk glow. |
| `src/components/ui/button.test.tsx` | Unit test varian tombol dan event onClick. |
| `src/components/ui/card.tsx` | Komponen container card glassmorphism bergradasi gelap dengan border halus. |
| `src/components/ui/card.test.tsx` | Unit test render card komponen. |
| `src/components/ui/dropdown.tsx` | Komponen menu dropdown interaktif berbasis Radix UI. |
| `src/components/ui/input.tsx` | Komponen input teks cyberpunk dengan focus ring neon. |
| `src/components/ui/input.test.tsx` | Unit test input komponen. |
| `src/components/ui/modal.tsx` | Dialog modal dialog overlay dengan animasi fade-in dan backdrop blur. |
| `src/components/ui/modal.test.tsx` | Unit test buka/tutup modal dialog. |
| `src/components/ui/skeleton.tsx` | Shimmer placeholder animasi untuk loading state antarmuka. |
| `src/components/ui/toast.tsx` | Komponen toast alert mengambang berbasis Radix UI Toast. |
| `src/components/ui/toast.test.tsx` | Unit test render toast notification. |
| `src/components/ui/tooltip.tsx` | Komponen petunjuk tooltip saat hover mouse. |
| `src/components/widgets/ai-assistant-widget.tsx` | Mini widget AI assistant pada dashboard untuk trigger quick diagnosis insiden aktif. |
| `src/components/widgets/architecture-status.tsx` | Visual status kesehatan arsitektur (Core Backend, AI Service, Docker Engine, Kubernetes, ArgoCD). |
| `src/components/widgets/argocd-status-widget.tsx` | Widget ringkasan aplikasi GitOps ArgoCD pada dashboard utama. |
| `src/components/widgets/container-detail-modal.tsx` | Modal mendalam inspeksi kontainer: environment variables, port bindings, resource limits, dan tombol aksi. |
| `src/components/widgets/host-resource-usage.tsx` | Grafik telemetri ECharts menampilkan tren utilisasi CPU, Memory, dan Network secara real-time. |
| `src/components/widgets/stat-card.tsx` | Kartu metrik KPI berhiaskan ikon neon, angka utama, sublabel, dan badge tren delta persentase. |
| `src/components/widgets/stat-card.test.tsx` | Unit test rendering stat card dan loading shimmer state. |
| `src/components/widgets/system-event-logs.tsx` | Feed event audit sistem real-time yang dialirkan langsung dari WebSocket broker. |
| `src/hooks/use-auth.ts` | Custom hook pengelolaan status login, dev token switching, user roles, dan logout. |
| `src/hooks/use-auth.test.ts` | Unit test logika state autentikasi dan role extraction. |
| `src/hooks/use-websocket.ts` | Custom hook langganan topic WebSocket dengan auto-reconnect exponential backoff. |
| `src/hooks/use-websocket.test.ts` | Unit test pendaftaran topic dan penanganan pesan WebSocket. |
| `src/lib/api-client.ts` | Instance Axios terpusat dilengkapi request interceptor injeksi JWT Bearer token dan error handler. |
| `src/lib/api.ts` | Helper utilitas pembantu endpoint URL. |
| `src/lib/auth.ts` | Utilitas token penyimpanan localStorage, decode payload JWT, dan evaluasi role RBAC. |
| `src/lib/auth.test.ts` | Unit test decoding token dan pengecekan permission. |
| `src/lib/format.ts` | Fungsi pemformat angka, byte (B, KB, MB, GB), persentase, dan tanggal format lokal Indonesia. |
| `src/lib/format.test.ts` | Unit test fungsi pemformatan byte dan tanggal. |
| `src/lib/ws-client.ts` | Singleton instance WebSocket client mengelola koneksi socket fisik tunggal ke `/ws`. |
| `src/lib/ws-client.test.ts` | Unit test lifecycle koneksi singleton WebSocket. |
| `src/services/ai-service.ts` | Client API domain AI: `sendMessage`, `getModels`, `listSessions`, `approveTool`, `getUsageStats`, `generateIncidentRCA`. |
| `src/services/ai-service.test.ts` | Unit test pemanggilan endpoint AI API dan unwrap response. |
| `src/services/argocd-service.ts` | Client API domain ArgoCD: `listApplications`, `getApplication`, `syncApplication`, `getOverview`. |
| `src/services/argocd-service.test.ts` | Unit test API ArgoCD service. |
| `src/services/docker-service.ts` | Client API domain Docker: `listContainers`, `getContainer`, `restartContainer`, `stopContainer`, `getStats`. |
| `src/services/docker-service.test.ts` | Unit test API Docker service. |
| `src/services/incident-service.ts` | Client API domain Insiden: `listIncidents`, `getIncident`, `acknowledgeIncident`, `resolveIncident`, `closeIncident`. |
| `src/services/incident-service.test.ts` | Unit test API incident service. |
| `src/services/kubernetes-service.ts` | Client API domain K8s: `listPods`, `getPodLogs`, `listDeployments`, `restartDeployment`, `scaleDeployment`. |
| `src/services/kubernetes-service.test.ts` | Unit test API kubernetes service. |
| `src/services/settings-service.ts` | Client API domain Settings: `getSettings`, `updateSettings`, `testNotification`, `listUsers`, `updateUserRole`. |
| `src/services/settings-service.test.ts` | Unit test API settings service. |
| `src/store/notification-store.ts` | Zustand store mengelola antrean notifikasi toast aktif dan riwayat notifikasi badge bell header. |
| `src/store/notification-store.test.ts` | Unit test penambahan dan pembersihan notifikasi di store. |
| `src/store/sidebar-store.ts` | Zustand store mengelola state collapse / expand sidebar navigasi. |
| `src/store/sidebar-store.test.ts` | Unit test toggle state sidebar. |
| `src/store/theme-store.ts` | Zustand store mengelola preferensi tema gelap cyberpunk. |
| `src/store/theme-store.test.ts` | Unit test persistensi tema. |
| `src/styles/tokens.css` | CSS variables: definisi warna HEX/HSL, border glass, neon glow effects, dan spacing tokens. |
| `src/types/ai.ts` | Interface TypeScript: sesi obrolan, pesan, definisi tool, provider, dan statistik penggunaan AI. |
| `src/types/argocd.ts` | Interface TypeScript: aplikasi ArgoCD, status sync, health, dan tree resource. |
| `src/types/docker.ts` | Interface TypeScript: kontainer, image, network, volume, dan metrik stats kontainer. |
| `src/types/incident.ts` | Interface TypeScript: entitas insiden, level severity, status transisi, dan payload alert. |
| `src/types/kubernetes.ts` | Interface TypeScript: Pod, Deployment, Node, Service, dan overview klaster K8s. |
| `src/types/settings.ts` | Interface TypeScript: SystemSettings, NotificationSettings, CombinedSettings, UserAdmin, ActiveSession. |
| `src/types/websocket.ts` | Interface TypeScript: WSMessage, EventPayload, LogPayload, NotificationPayload. |

---

### 2.5. Lapisan Infrastruktur, Keamanan, & GitOps (`infrastructure/`)

Direktori ini berisi seluruh definisi **Infrastructure as Code (IaC)**, Helm charts, manifest Kubernetes, pengerasan keamanan, dan testbed lokal.

| Path Berkas | Fungsi & Penjelasan Isi |
|---|---|
| `kubernetes/base/kustomization.yaml` | Manifest Kustomize base mengagregasi resources platform default. |
| `kubernetes/overlays/staging/kustomization.yaml` | Overlay Kustomize lingkungan Staging (namespace `cifo-staging`, tag `staging-latest`). |
| `kubernetes/overlays/production/kustomization.yaml` | Overlay Kustomize lingkungan Production (namespace `cifo-production`, tag release `v1.0.0`). |
| `kubernetes/charts/cifo-platform/Chart.yaml` | Metadata Umbrella Helm Chart v1.0.0 mengorkestrasi seluruh komponen CIFO Platform. |
| `kubernetes/charts/cifo-platform/values.yaml` | Nilai konfigurasi default Helm chart (resources, probes, ports, env). |
| `kubernetes/charts/cifo-platform/values-staging.yaml` | Override konfigurasi Helm untuk staging (1 replika per layanan, debug log). |
| `kubernetes/charts/cifo-platform/values-production.yaml` | Override konfigurasi Helm production HA (3-10 replika, HPA CPU 70%, PDB min 1). |
| `kubernetes/charts/cifo-platform/templates/_helpers.tpl` | Helper template Go templating untuk standardisasi label Kubernetes dan nama resource. |
| `kubernetes/charts/cifo-platform/templates/frontend/*` | Template Helm frontend: `deployment.yaml`, `service.yaml`, `hpa.yaml`, `pdb.yaml`. |
| `kubernetes/charts/cifo-platform/templates/backend/*` | Template Helm backend: `deployment.yaml`, `service.yaml`, `configmap.yaml`, `serviceaccount.yaml`, `hpa.yaml`, `pdb.yaml`. |
| `kubernetes/charts/cifo-platform/templates/ai-service/*` | Template Helm AI service: `deployment.yaml`, `service.yaml`, `serviceaccount.yaml` (`cifo-ai-agent-sa`), `hpa.yaml`. |
| `kubernetes/charts/cifo-platform/templates/data/*` | Template Helm data layer: `postgres-statefulset.yaml` (PVC 10Gi), `postgres-service.yaml`, `redis-deployment.yaml`, `redis-service.yaml`. |
| `kubernetes/charts/cifo-platform/templates/ingress/*` | Template Helm ingress: `ingress.yaml` konfigurasi Traefik / Ingress TLS 1.3. |
| `kubernetes/charts/cifo-platform/templates/security/*` | Template Helm security: `networkpolicies.yaml` menerapkan Zero-Trust NetworkPolicies di klaster. |
| `kubernetes/charts/cifo-ai-service/Chart.yaml` | Standalone subchart Helm untuk microservice AI. |
| `kubernetes/charts/cifo-backend/Chart.yaml` | Standalone subchart Helm untuk backend core Go. |
| `kubernetes/charts/cifo-data/Chart.yaml` | Standalone subchart Helm untuk PostgreSQL & Redis data layer. |
| `kubernetes/charts/cifo-frontend/Chart.yaml` | Standalone subchart Helm untuk frontend dashboard. |
| `local-testbed/docker-compose.yml` | Master compose testbed: Postgres 16, Redis 7, VictoriaMetrics, Loki, Tempo, Prometheus, Alertmanager, Keycloak, Vault, Docker Proxy. |
| `local-testbed/docker-compose.monitoring.yml` | Standalone compose testbed khusus untuk komponen observability dan monitoring. |
| `local-testbed/alertmanager/alertmanager.yml` | Konfigurasi rute notifikasi Alertmanager meneruskan alert ke endpoint webhook backend Go. |
| `local-testbed/argocd/crds.yaml` | Custom Resource Definitions resmi ArgoCD (Applications, AppProjects). |
| `local-testbed/argocd/install.yaml` | Manifest instalasi lengkap komponen controller, server, dan repo-server ArgoCD. |
| `local-testbed/argocd/sample-apps/*.yaml` | Sample aplikasi microservice (`nginx` dan `httpbin`) yang dikelola otomatis oleh ArgoCD GitOps di K3d. |
| `local-testbed/docker-proxy/haproxy.cfg.template` | Template HAProxy membatasi Docker socket Unix: GET & restart diperbolehkan, DELETE/EXEC/secrets di-block 403. |
| `local-testbed/grafana/.../datasources.yaml` | Provisioning datasource Grafana: Prometheus, VictoriaMetrics, Loki, dan Tempo dengan `derivedFields` trace linking. |
| `local-testbed/k3d/cluster-config.yaml` | Konfigurasi klaster K3d lokal: 1 server node, 2 agent worker nodes, port mapping API dan Ingress. |
| `local-testbed/k3d/setup-cluster.ps1` & `.sh` | Skrip otomasi pembuatan klaster K3d dan instalasi ArgoCD untuk Windows dan Linux. |
| `local-testbed/keycloak/cifo-realm.json` | Konfigurasi Realm Keycloak: client `cifo-frontend`, roles (admin, devops, viewer), dan dev users. |
| `local-testbed/prometheus/alert-rules.yml` | Definisi rule alert Prometheus: HighCPUUsage (>85%), HighMemoryUsage (>90%), ContainerDown, PodCrashLooping. |
| `local-testbed/prometheus/prometheus.yml` | Konfigurasi scrape target Prometheus ke backend, cAdvisor, node-exporter, dan remote_write ke VictoriaMetrics. |
| `local-testbed/tempo/tempo.yaml` | Konfigurasi distributed tracing Tempo: OTLP HTTP receiver (:4318), block retention, dan storage lokal. |
| `security/network-policies/00-default-deny-all.yaml` | NetworkPolicy: Memblokir seluruh lalu lintas masuk (ingress) dan keluar (egress) secara default pada namespace. |
| `security/network-policies/01-cifo-frontend-netpol.yaml` | NetworkPolicy: Mengizinkan frontend diakses dari Ingress dan hanya boleh berbicara ke backend (:8080). |
| `security/network-policies/02-cifo-backend-netpol.yaml` | NetworkPolicy: Backend boleh menerima traffic dari frontend dan boleh menghubungi DB, Redis, AI service, dan K8s API. |
| `security/network-policies/03-cifo-ai-service-netpol.yaml` | NetworkPolicy: AI service hanya boleh menerima request dari backend dan hanya boleh egress ke Internet (LLM APIs). |
| `security/network-policies/04-cifo-data-netpol.yaml` | NetworkPolicy: Database PostgreSQL dan Redis HANYA boleh diakses oleh backend; akses langsung dari AI service ditolak. |
| `security/rbac/01-cifo-ai-agent-sa.yaml` | Manifest pembuatan ServiceAccount `cifo-ai-agent-sa` untuk agen AI. |
| `security/rbac/02-cifo-ai-agent-role.yaml` | Manifest ClusterRole `cifo-ai-agent-role`: hak baca pods/namespaces, hak patch deployments, dilarang delete & secrets. |
| `security/rbac/03-cifo-ai-agent-binding.yaml` | Manifest ClusterRoleBinding menghubungkan ServiceAccount `cifo-ai-agent-sa` dengan role yang telah dibatasi. |
| `security/vault/cifo-backend-policy.hcl` | HCL Policy Vault: hak baca rahasia path `secret/data/cifo/backend` dan `secret/data/cifo/shared`. |
| `security/vault/cifo-ai-policy.hcl` | HCL Policy Vault: hak baca rahasia path `secret/data/cifo/ai-service` dan `secret/data/cifo/shared`. |
| `terraform/environments/*` | Konfigurasi Terraform per environment staging dan production. |
| `terraform/modules/*` | Modul Terraform reusable untuk provisi komputasi, storage, dan jaringan. |

---

### 2.6. Shared Packages (`packages/`)

| Path Berkas | Fungsi & Penjelasan Isi |
|---|---|
| `api-contracts/openapi.yaml` | Spesifikasi OpenAPI 3.0 mendefinisikan seluruh endpoint REST backend, model payload, dan kode respons HTTP. |
| `api-contracts/proto/ai_service.proto` | Definisi skema Protobuf gRPC untuk komunikasi biner berkinerja tinggi backend ke AI microservice. |
| `api-contracts/proto/common.proto` | Definisi pesan gRPC Protobuf umum (status, metadata, timestamps). |
| `eslint-config/index.js` | Modul konfigurasi ESLint terstandar yang diimpor oleh paket-paket JavaScript/TypeScript di monorepo. |
| `theme/colors.json` | Token warna resmi: palet latar gelap (#0b0f19), aksen cyan neon (#00f0ff), pink magenta (#f43f5e), status green (#10b981). |
| `theme/typography.json` | Token tipografi: konfigurasi font-family Geist/Inter, skala modular ukuran huruf (xs s/d 4xl), dan bobot font. |

---

### 2.7. Skrip Otomasi & Verifikasi Operasional (`scripts/`)

| Nama Berkas Skrip | Fungsi & Penjelasan Isi |
|---|---|
| `build-backend.ps1` | Mengompilasi binary `server.exe` backend Go lokal dengan cache terisolasi. |
| `build-linux.ps1` | Mengompilasi binary backend Linux ELF 64-bit untuk deployment container. |
| `check-argocd-pods.ps1` | Memeriksa status kesehatan Pods ArgoCD di namespace `argocd` klaster K3d. |
| `check-coverage.ps1` | Mengevaluasi persentase statement coverage backend dan memverifikasi batas threshold >= 70%. |
| `check-status.ps1` | Memeriksa status seluruh container Docker lokal, port open, dan ketersediaan layanan. |
| `complete-fase1.ps1` | Otomasi bootstrap dan verifikasi akhir kelayakan Fase 1 (Local Testbed). |
| `debug-k3d.ps1` | Skrip diagnosis jaringan dan status node klaster K3d lokal. |
| `docker-helper.ps1` | Helper pembantu inspeksi container Docker lokal. |
| `exec.ps1` | Wrapper eksekusi perintah shell dengan penanganan error terstruktur. |
| `fix-network.ps1` | Skrip perbaikan routing jaringan Docker Compose pada sistem operasi Windows. |
| `generate-api-types.ps1` & `.sh` | Otomasi generate tipe TypeScript frontend langsung dari spesifikasi OpenAPI backend. |
| `init-vault.ps1` | Otomasi inisialisasi HashiCorp Vault lokal, unseal, pengaktifan engine KV v2, apply HCL policies, dan seeding secrets. |
| `quick-check.ps1` | Pemeriksaan cepat ketersediaan HTTP endpoint backend, frontend, dan AI service. |
| `run-ai-tests.ps1` | Menjalankan 24 unit test AI service menggunakan Pytest dengan konfigurasi Vault dummy. |
| `run-backend-tests.ps1` | Menjalankan seluruh unit test service backend Go dengan perhitungan coverage. |
| `run-e2e-tests.ps1` | Menjalankan seluruh test suite Playwright E2E pada browser headless. |
| `run-frontend-tests.ps1` | Menjalankan unit test frontend menggunakan Vitest dengan V8 coverage engine. |
| `run-integration-tests.ps1` | Menjalankan integration test Go terhadap server backend live di port 8080. |
| `run-load-tests.ps1` | Menjalankan load test K6 (API Throughput 1000 req/s & WebSocket Stress 500 VUs). |
| `run-repo-tests.ps1` | Menjalankan pengujian repository layer backend Go terhadap database PostgreSQL live. |
| `seed-data.sql` | Skrip SQL memasukkan data awal: default roles (admin, devops, viewer) dan system settings dasar. |
| `setup-k6.ps1` | Skrip download dan setup binary K6 load testing ke direktori `bin/k6.exe`. |
| `setup-local.ps1` & `setup-local.sh` | Skrip instalasi lokal lengkap (Docker Compose, DB migrations, K3d, ArgoCD, sample apps). |
| `start-backend.ps1` | Skrip menjalankan backend Go server di background dengan variabel lingkungan lokal. |
| `start-frontend.ps1` | Skrip menjalankan Next.js dev server pada port 3001 di background. |
| `status-fase1.ps1` | Skrip evaluasi kesiapan testbed lokal Fase 1. |
| `test-db-conn.ps1` | Menguji konektivitas TCP dan autentikasi ke database PostgreSQL `cifo_db`. |
| `test-monitoring.ps1` | Menguji endpoint metrik monitoring dan ketersediaan data time-series VictoriaMetrics. |
| `test-phase3-auth.ps1` | Verifikasi kepatuhan autentikasi, token validation, dan penegakan role RBAC Fase 3. |
| `test-phase6-backend.ps1` | Verifikasi integrasi Kubernetes & ArgoCD client backend Fase 6. |
| `test-phase7-websocket.ps1` | Verifikasi real-time WebSocket broker, subscription topic, dan log streaming Fase 7. |
| `test-phase8-alerts.ps1` | Verifikasi Alertmanager webhook, siklus insiden, dan bot Telegram Fase 8. |
| `test-phase9-ai.ps1` | Verifikasi multi-model LLM fallback, prompt injection sanitizer, dan AI RCA Fase 9. |
| `test-phase10-settings.ps1` | Verifikasi komprehensif halaman Settings, zero-mock active sessions, dan audit log Fase 10. |
| `test-phase11-tracing.ps1` | Verifikasi distributed tracing OpenTelemetry, traceparent propagation, dan Grafana linking Fase 11. |
| `test-phase12-security.ps1` | Verifikasi 22 butir uji keamanan: Vault, Docker Proxy, K8s RBAC, NetPols, dan SecurityHeaders Fase 12. |
| `test-phase13-comprehensive.ps1` | Master orchestrator pengujian piramida QA: Unit (Backend/FE/AI), Integration, E2E, Load Testing Fase 13. |
| `test-phase14-cicd.ps1` | Verifikasi multi-stage Dockerfiles, GitHub Actions workflows, Umbrella Helm Chart, Kustomize Fase 14. |
| `test-phase15-production-readiness.ps1` | Master readiness validator untuk simulasi dry-run, failover, dan audit kesiapan produksi Fase 15. |
| `test-rate-limit.ps1` | Menguji efektivitas penolakan HTTP 429 Too Many Requests saat request melebihi kuota rate limiter. |
| `verify-frontend-routes.ps1` | Menguji ketersediaan seluruh rute halaman Next.js (/monitoring, /docker, /kubernetes, dll.). |
| `verify-services.ps1` | Memeriksa ketersediaan port dan respons HTTP seluruh microservices CIFO Platform. |
| `wait-for-docker.ps1` | Utility pembantu menunggu Docker Engine siap menerima koneksi sebelum bootstrap. |

---

### 2.8. Pengujian Tingkat Sistem (`tests/`)

| Path Berkas | Fungsi & Penjelasan Isi |
|---|---|
| `tests/e2e/.gitkeep` | Direktori penampung pengujian E2E tingkat sistem. |
| `tests/integration/go.mod` | Modul Go independen untuk integration test suites. |
| `tests/integration/.gitkeep` | Menjaga folder integrasi tetap terlacak oleh Git. |
| `tests/load/api-throughput.js` | Skenario K6 menguji throughput API backend pada beban 1000 req/s selama 10 detik (target p99 < 200ms). |
| `tests/load/websocket-stress.js` | Skenario K6 menguji ketahanan koneksi WebSocket pada beban 500 concurrent VUs (target p95 < 1000ms). |
| `tests/load/.gitkeep` | Menjaga folder load testing tetap terlacak oleh Git. |

---

## 3. Ringkasan Kepatuhan & Kebersihan Struktur

1. **Prinsip Zero Clutter**:
   - Seluruh binary kompilasi lokal (`*.exe`), file cache (`.gocache/`, `.cache/`, `.pytest_cache/`), dan test output (`coverage/`, `test-results/`) telah dimasukkan ke dalam [.gitignore](file:///d:/agent%20v2/.gitignore).
2. **Keterbacaan dan Penamaan Konsisten**:
   - Direktori kode sumber menggunakan penamaan *kebab-case* yang bersih (`ai-service`, `api-contracts`, `network-policies`, `local-testbed`).
3. **Pemisahan Peran yang Tegas (Clean Separation of Concerns)**:
   - Kode aplikasi (`apps/`), infrastruktur & orkestrasi klaster (`infrastructure/`), dokumentasi formal (`docs/`), skrip otomasi (`scripts/`), dan kontrak antarmuka bersama (`packages/`) terpisah secara modular tanpa dependensi sirkular.
4. **Kesiapan Audit Enterprise**:
   - Repositori ini sepenuhnya terdokumentasi dan siap diaudit oleh tim DevOps, SRE, maupun Security Auditor untuk *production handover*.

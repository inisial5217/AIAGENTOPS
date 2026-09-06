# CIFO Platform — Panduan & Dokumen Handoff Komprehensif (Fase 0 s.d. Fase 8)

> **Dokumen Handoff untuk Agent AI Baru / Sesi Lanjutan**  
> **Repository**: [https://github.com/inisial5217/AIAGENTOPS](https://github.com/inisial5217/AIAGENTOPS)  
> **Tanggal Pembuatan**: 2026-09-07  
> **Status Terkini**: **Fase 0 s.d. Fase 8 SELESAI 100% (Terverifikasi & Siap Menuju Fase 9)**  
> **Peran Wajib Agent**: Senior Principal Software Architect, Full-Stack Developer, DevOps & SRE Specialist, Senior QA Analyst, dan UI/UX Designer.

---

## 1. Instruksi Peran & Standar Kerja Wajib (Mandatory Persona & Engineering Standards)

Jika Anda adalah Agent AI yang membaca dokumen ini untuk melanjutkan proyek CIFO (*Cloud Infrastructure & Operations Platform*), Anda **WAJIB** mengadopsi standar keprofesionalan dan disiplin rekayasa perangkat lunak tertinggi:

### 1.1 Persona & Pola Pikir (Mindset)
1. **Ahli Arsitektur (Principal Architect)**: Selalu membaca dan menganalisis secara mendalam 3 dokumen panduan di folder [`arsitektur_diskusi/`](file:///d:/agent%20v2/arsitektur_diskusi/) sebelum mengambil tindakan atau memulai fase baru:
   - [`arsitektur_diskusi/arsitektur_sistem.md`](file:///d:/agent%20v2/arsitektur_diskusi/arsitektur_sistem.md): Arsitektur global CIFO, pola integrasi, diagram data flow, dan spesifikasi teknologi.
   - [`arsitektur_diskusi/plan.md`](file:///d:/agent%20v2/arsitektur_diskusi/plan.md): Rencana bertahap dari Fase 0 hingga Fase 12 beserta kriteria penerimaan spesifik.
   - [`arsitektur_diskusi/agent_instructions.md`](file:///d:/agent%20v2/arsitektur_diskusi/agent_instructions.md): Aturan penulisan kode, konvensi penamaan, dan batasan operasional.
2. **Programmer & Developer Ahli**: Menulis kode yang terstruktur, modular, efisien, dan bersih (*Clean Architecture*). Memisahkan concern antara handler, service, repository, dan model domain.
3. **DevOps & SRE Specialist**: Memastikan sistem terisolasi dengan baik di Docker/K8s, menggunakan connection pool aman, bebas dari kebocoran goroutine/koneksi, dan memiliki mekanisme fallback serta retry queue.
4. **Senior QA Analyst**: Setiap perubahan backend **WAJIB** memiliki unit test Go (`go test ./...`), dan setiap komponen frontend **WAJIB** memiliki unit test Vitest (`npm test`). Skrip pengujian integrasi live harus disiapkan di `scripts/`.
5. **UI/UX Designer**: Menjaga estetika antarmuka modern bernuansa *Cyberpunk Dark Mode / Sleek Enterprise Glassmorphism*. Menggunakan design tokens terpusat di `tokens.css`, visual feedback real-time, badge semantik, serta micro-animations. Hindari UI template murahan atau bare minimum MVP.

### 1.2 Aturan Emas Pengembangan (Golden Rules)
- **ZERO MOCK DATA**: **DILARANG KERAS** menggunakan data tiruan (mock list, hardcoded array, fake metrics) pada implementasi produksi backend maupun frontend. Seluruh data metrik, pod Kubernetes, kontainer Docker, insiden, dan event log **harus bersumber langsung dari database PostgreSQL nyata (`cifo_db`), Redis (`cifo-redis`), Docker Daemon, K3d cluster (`cifo-dev`), Prometheus, dan Alertmanager**.
- **Standar Komentar Kode Go**: Komentar kode pada Go **WAJIB 1-4 kata saja dalam bahasa Inggris** (contoh: `// parse request payload`, `// handle database error`). **DILARANG** menggunakan baris dekoratif atau separator karakter (seperti `// ==================`).
- **Penanganan Error Terstruktur**: Gunakan typed errors dari package [`pkg/apperror`](file:///d:/agent%20v2/apps/backend/pkg/apperror) (`NewNotFound`, `NewBadRequest`, `NewUnauthorized`, `NewForbidden`, `NewInternal`).
- **Disiplin Concurrency & Clean Goroutine**: Semua background loop dan WebSocket streamer harus mendengarkan `ctx.Done()`. Selalu gunakan buffered channel dengan strategi non-blocking drop untuk slow consumer, dan daftarkan hook pembersihan (`onTopicEmpty`) saat tidak ada lagi klien yang berlangganan.
- **Disiplin Fase (STRICT)**: **JANGAN PERNAH** melompat atau memulai fase berikutnya sebelum pengguna memberikan perintah eksplisit. Setiap penyelesaian fase **WAJIB** didokumentasikan di `implementasi_plan/f<nomor_fase>.md` dan dirangkum di `walkthrough.md`.

---

## 2. Struktur Repositori & Teknologi Inti

```text
d:\agent v2\
├── .git/                                 # Git version control (Remote: origin/main)
├── .gitignore                            # Konfigurasi ignore (node_modules, bin, .next, .gocache, dll)
├── go.work, go.work.sum                  # Go Workspace multi-module
├── arsitektur_diskusi/                   # PANDUAN UTAMA PROYEK
│   ├── arsitektur_sistem.md              # Arsitektur sistem menyeluruh
│   ├── plan.md                           # Rencana detail Fase 0 s.d. Fase 12
│   └── agent_instructions.md             # Konvensi dan instruksi teknis
├── implementasi_plan/                    # DOKUMENTASI HISTORIS PENYELESAIAN FASE
│   ├── f0-1.md                           # Dokumentasi Fase 0 & 1
│   ├── f2.md                             # Dokumentasi Fase 2 (Core Observability)
│   ├── f3.md                             # Dokumentasi Fase 3 (Auth & RBAC)
│   ├── f4.md                             # Dokumentasi Fase 4 (Frontend Core & Dashboard)
│   ├── f5.md                             # Dokumentasi Fase 5 (Docker Management)
│   ├── f6.md                             # Dokumentasi Fase 6 (Kubernetes & ArgoCD)
│   ├── f7.md                             # Dokumentasi Fase 7 (Real-Time & WebSocket)
│   └── f8.md                             # Dokumentasi Fase 8 (Alerting & Incident Management)
├── apps/
│   ├── backend/                          # Backend API Engine (Go 1.24, Chi, pgxpool, go-redis)
│   │   ├── cmd/server/main.go            # Entrypoint HTTP Server (:8080) & WS Hub
│   │   ├── cmd/worker/main.go            # Entrypoint Standalone Worker
│   │   ├── internal/config/              # Manajemen env & konfigurasi
│   │   ├── internal/handler/             # HTTP & WebSocket handlers
│   │   ├── internal/integration/         # Klien eksternal (Docker, K8s client-go, ArgoCD, Telegram)
│   │   ├── internal/middleware/          # Auth JWT, RBAC, CORS, Recovery, Rate Limiter
│   │   ├── internal/model/               # Domain data structures & contracts
│   │   ├── internal/repository/          # Data Access Layer PostgreSQL
│   │   ├── internal/service/             # Core Business Logic & State Machines
│   │   ├── internal/ws/                  # WebSocket Engine (Hub, Client, Streamer, Message)
│   │   ├── migrations/                   # SQL Migrations (001 s.d. 007)
│   │   └── pkg/                          # Utilitas bersama (apperror, validator)
│   ├── frontend/                         # Modern Web UI (Next.js 16, React 19, Turbopack, Zustand, Radix UI)
│   │   ├── src/app/(auth)/login/         # Halaman Login Keycloak / Dev fallback
│   │   ├── src/app/(dashboard)/          # Dashboard Shell & Layout
│   │   │   ├── monitoring/               # Live Telemetry & Metrics Gauges
│   │   │   ├── docker/                   # Docker Management (containers, images, networks, volumes)
│   │   │   ├── kubernetes/               # K8s Management (Pods, Deployments, Services, Namespaces)
│   │   │   ├── argocd/                   # GitOps Application Sync & Resource Inspection
│   │   │   ├── incidents/                # Incident Management & 4-Stage Lifecycle Stepper
│   │   │   └── settings/                 # Konfigurasi profil & preferensi
│   │   ├── src/components/               # UI components, layout, widgets, terminal, toast
│   │   ├── src/hooks/                    # React hooks (useWebSocket, dll)
│   │   ├── src/lib/                      # API client Axios, WebSocket singleton, utils
│   │   ├── src/services/                 # Layanan pemanggil API backend
│   │   ├── src/store/                    # Zustand state management (theme, sidebar, notification)
│   │   └── src/types/                    # Definisi tipe TypeScript
│   └── ai-service/                       # Microservice AI (Target Fase 9: Python 3.11, FastAPI, LangGraph)
├── infrastructure/local-testbed/         # Komposisi Infrastruktur Lokal
│   ├── docker-compose.yml                # PostgreSQL, Redis, Keycloak, Prometheus, Grafana, Alertmanager, ArgoCD
│   ├── alertmanager/                     # Konfigurasi Alertmanager webhook
│   ├── prometheus/                       # Prometheus scrapers & alert rules
│   └── argocd/                           # ArgoCD manifest & sample apps
└── scripts/                              # Otomasi & Skrip Pengujian Integrasi
    ├── start-backend.ps1                 # Menjalankan backend server Go (:8080)
    ├── seed-data.sql                     # Seeding data pengguna & dummy metric
    ├── test-monitoring.ps1               # Pengujian integrasi metrik
    ├── test-phase6-backend.ps1           # Pengujian integrasi K8s & ArgoCD
    ├── test-phase7-websocket.ps1         # Pengujian integrasi WebSocket & Event Watcher
    └── test-phase8-alerts.ps1            # Pengujian integrasi Alertmanager & Incident Lifecycle
```

---

## 3. Rangkuman Progres Fase 0 s.d. Fase 8 (Completed Milestones)

### Fase 0 & 1: Fondasi Arsitektur, Monorepo & Skema Basis Data
- **Infrastruktur Lokal**: Docker Compose terpadu mengelola `cifo-postgres` (:5432), `cifo-redis` (:6379), Keycloak (:8080), Prometheus (:9090), Grafana (:3000), Alertmanager (:9093), dan ArgoCD (:8082). Cluster K3d `cifo-dev` terpasang.
- **Skema Basis Data**: Migrasi 001 s.d. 004 membuat tabel `users`, `roles`, `audit_log`, `metrics_history`, `alerts`, dan `incidents`.
- **Keycloak Auth**: Konfigurasi Realm `cifo` dengan Roles: `admin`, `devops`, dan `viewer`.

### Fase 2: Observabilitas Inti & Metrik Host/Runtime
- **Prometheus Collector**: Scraper Go runtime metrics (goroutines, heap memory, GC pause) dan telemetri host (CPU load, memory usage, disk I/O, network traffic).
- **Health Check Endpoints**: Liveness (`/healthz`) dan Readiness (`/readyz`) memverifikasi status koneksi PostgreSQL dan Redis.

### Fase 3: Autentikasi, Otorisasi RBAC & Audit Trail
- **JWT Middleware**: Validasi access token dari Keycloak dengan dukungan dev-tokens (`dev-token-admin`, `dev-token-devops`, `dev-token-viewer`) pada lingkungan development.
- **Role Enforcement**: Guard `RequireRole()` memverifikasi hak akses pada setiap endpoint sensitif.
- **Audit Logging**: Mencatat seluruh aktivitas mutasi ke tabel `audit_log` (User ID, Role, Action, IP Address, Metadata).

### Fase 4: Frontend Core, Design System & Monitoring Dashboard
- **Next.js 16 & Turbopack**: Setup modern Next.js 16 App Router dengan React 19 dan TypeScript.
- **Design Tokens (`tokens.css`)**: Palet Cyberpunk Dark Glassmorphism, HSL border glow, semantic colors (Red Error, Amber Warn, Green Info, Cyan Accent).
- **Monitoring Page (`/monitoring`)**: Menampilkan 4 KPI Stat Cards, visual Telemetry Gauges (CPU, Memory, Disk), grafik Prometheus, dan Host Resource Usage.

### Fase 5: Docker Container Management Engine
- **Docker Engine SDK**: Integrasi langsung ke `//./pipe/docker_engine` (Windows Named Pipe) tanpa mock.
- **Manajemen Kontainer**: Endpoint inventarisasi kontainer, filtering berdasarkan status, dan aksi siklus hidup (`start`, `stop`, `restart`) dengan audit log.
- **Halaman `/docker/containers`**: Tabel interaktif, navigasi tab (Images, Networks, Volumes), dan Inspect Modal.

### Fase 6: Kubernetes & GitOps Engine (K3d & ArgoCD)
- **Kubernetes Client (`client-go`)**: Terhubung langsung ke K3d cluster `cifo-dev` via kubeconfig lokal.
- **Manajemen Sumber Daya K8s**: Listing Pods, Deployments, Services, Namespaces, dan mutasi scaling replika deployment dengan batasan RBAC.
- **Integrasi ArgoCD**: Klien HTTP mengonsumsi ArgoCD REST API (`/api/v1/applications`), sinkronisasi GitOps on-demand, dan modal inspeksi pohon resource aplikasi.
- **Halaman `/kubernetes` & `/argocd`**: Antarmuka responsif dengan filter namespace, status sinkronisasi, dan tombol aksi terisolasi.

### Fase 7: Komunikasi Real-Time & Bi-directional WebSocket
- **WebSocket Protocol (RFC 6455)**: Menggunakan `github.com/gorilla/websocket` v1.5.3 dengan endpoint `GET /ws?token=...`.
- **Pub/Sub Topic Hub**: Topik terisolasi `notifications`, `system_events`, `docker_logs:<container_id>`, dan `k8s_events`.
- **Zero Goroutine Leak**: Pompa ganda `readPump` dan `writePump`, heartbeat ping (30s) / pong (60s), dan pembatalan otomatis upstream Docker stream saat topic kosong (`onTopicEmpty`).
- **Backpressure Protection**: Buffered channel berkapasitas 1000 pesan dengan kebijakan non-blocking drop saat client lag.
- **Frontend Live Components**:
  - `WebSocketClient` Singleton dengan *exponential backoff reconnect* (1s s.d. 30s) dan otomatisasi re-subscribe topic.
  - Cyberpunk Dark `LogTerminal` dengan syntax highlighting dinamis, search filter, auto-scroll toggle, buffer clear, dan download file log.
  - `NotificationToastProvider` dengan aturan persistensi waktu: `CRITICAL` bersifat persisten, `WARNING` auto-dismiss 10 detik, dan `INFO` auto-dismiss 5 detik.

### Fase 8: Alerting & Incident Management
- **Migrasi Database 007**: Membuat tabel `notifications` (`id`, `incident_id`, `channel`, `recipient`, `title`, `message`, `severity`, `status`, `error_message`, `created_at`) dan memperluas tabel `incidents` dengan kolom `closed_by` dan `closed_at`.
- **Alertmanager Webhook Receiver**: Endpoint `POST /api/v1/webhooks/alertmanager` memproses payload `firing` (membuat insiden baru / deduplikasi) dan `resolved` (otomatis menyelesaikan insiden aktif).
- **Incident Lifecycle State Machine**:
  - Alur: `Open` -> `Acknowledged` -> `Resolved` -> `Closed`.
  - Penegakan Peran (RBAC): *DevOps* dan *Admin* berwenang melakukan acknowledge dan resolve; Penutupan (*Close*) **HANYA** boleh dilakukan oleh peran *Admin*.
  - Audit Trail: Setiap transisi status dicatat otomatis ke tabel `audit_log`.
- **Integrasi Telegram Bot**:
  - Format Markdown bersih tanpa emoji sesuai panduan desain sistem.
  - Deteksi badai alert (*alert storming* > 3 alert dalam 2 menit) yang diagregasi menjadi pesan ringkasan tunggal.
  - Antrean coba ulang (*retry queue*) berbasis Redis pada key `cifo:telegram:retry_queue`.
  - Rate limiting internal maksimal 30 permintaan/menit dengan fallback offline aman.
- **Background Escalation Worker**: Berjalan setiap 1 menit untuk memeriksa insiden `Open` yang belum di-acknowledge setelah > 15 menit, lalu menaikkan severity menjadi `critical` dengan prefiks `[ESCALATED]` dan menyiarkan notifikasi darurat.
- **In-App Live Push**: Pembuatan insiden otomatis menyiarkan payload ke topik WebSocket `notifications`, memicu refresh real-time pada dashboard frontend.
- **Dashboard `/incidents`**:
  - 5 KPI Stat Cards (Total, Open & Firing, Acknowledged, Resolved, MTTR).
  - Filter bar (Tabs status, filter severity, search dinamis).
  - Tabel responsif dengan Quick Actions (*Ack*, *Resolve*, *Close*).
  - Modal Detail Komprehensif: 4-stage visual lifecycle timeline stepper, metadata & link runbook, tautan navigasi langsung ke sumber daya Docker/K8s terkait, histori pengiriman notifikasi, dan kartu placeholder *Phase 9 AI Root Cause Analysis (RCA)*.

---

## 4. Status Pengujian, Verifikasi & Metrik Mutu

Semua kode telah diuji secara komprehensif tanpa toleransi error:

1. **Pengujian Unit Backend Go**:
   - Perintah: `go test ./...`
   - Status: **100% LULUS (36+ tests)** across config, handlers, middleware, services, ws, dan validator.
2. **Pengujian Unit Frontend Vitest**:
   - Perintah: `npx vitest run`
   - Status: **100% LULUS (52 tests across 14 suites)**.
3. **Kompilasi Produksi Next.js**:
   - Perintah: `npm run build`
   - Status: **100% SUKSES (Turbopack, exit code 0)**, 13 static pages prerendered (`/`, `/monitoring`, `/docker`, `/docker/containers`, `/docker/images`, `/docker/networks`, `/docker/volumes`, `/kubernetes`, `/argocd`, `/incidents`, `/settings`, `/login`, `/_not-found`).
4. **Verifikasi Integrasi Real Scripts**:
   - `scripts/test-phase8-alerts.ps1` -> 100% Lulus memvalidasi webhook firing, deduplikasi, acknowledge, resolve, close, auto-resolution resolved alert, dan pencatatan audit log di PostgreSQL.

---

## 5. Menghadapi Fase 9: AI Service & Chat Agent (Briefing Arsitektur)

> [!CAUTION]
> **PENTING UNTUK AGENT SELANJUTNYA**:
> **JANGAN PERNAH** memulai atau membuat kode untuk Fase 9 sebelum pengguna secara eksplisit memberikan perintah seperti: *"lanjut ke fase 9"*.

Ketika pengguna menginstruksikan untuk memulai Fase 9, berikut adalah panduan arsitektur yang harus dipedomani (berdasarkan `plan.md` Baris 1016-1100+):

### 5.1 Ruang Lingkup Fase 9
1. **Layanan Microservice Python (`apps/ai-service`)**:
   - Runtime: Python 3.11+ dengan framework FastAPI dan Uvicorn.
   - LLM: Google Gemini 2.5 Flash menggunakan API Key (`GOOGLE_API_KEY`).
   - Orkestrasi Agent: LangGraph StateGraph untuk mengelola siklus Reason-Act (ReAct) agent yang mampu memanggil tools secara iteratif.
   - Basis Pengetahuan RAG: ChromaDB / Qdrant lokal untuk mengindeks dokumen runbooks, arsitektur, dan histori insiden masa lalu.
2. **Kumpulan Tools Agent (Function Calling)**:
   - Tool `get_container_logs`: Mengambil log dari backend Go REST API.
   - Tool `get_pod_status`: Memeriksa status pod Kubernetes.
   - Tool `get_prometheus_metrics`: Melakukan query metrik PromQL.
   - Tool `get_incident_context`: Mengambil detail insiden untuk analisis akar masalah.
   - Tool `execute_safe_action`: Menjalankan aksi mitigasi yang telah disetujui (misal restart pod, scale deployment).
3. **Automated Root Cause Analysis (RCA)**:
   - Saat insiden dibuka pada dashboard `/incidents`, operator dapat mengklik tombol investigasi AI.
   - AI Service menganalisis log kontainer terkait, anomali metrik sebelum alert terpicu, dan histori kejadian serupa untuk menghasilkan ringkasan akar masalah (*Root Cause Hypothesis*) serta langkah remediasi (*Remediation Steps*).
4. **Interactive SRE Chat Drawer pada Frontend**:
   - Antarmuka chat AI responsif di bagian kanan layar atau modal terapung dengan streaming token response, syntax highlighting kode bash/yaml, dan tombol aksi satu-klik (*Action Buttons*).

---

## 6. Panduan Menjalankan Sistem pada Komputer Baru

Jika repositori ini di-clone ke komputer baru:

### 6.1 Prasyarat Lingkungan
- Docker Desktop aktif (mendukung Windows Containers / WSL2 Engine).
- Go versi 1.24 atau lebih baru.
- Node.js versi 18 atau lebih baru (npm aktif).
- Python versi 3.11 atau lebih baru (uv / pip).
- K3d CLI untuk cluster Kubernetes lokal (`k3d cluster create cifo-dev`).

### 6.2 Langkah Menjalankan
1. **Menyalakan Testbed Docker Compose**:
   ```powershell
   cd "d:\agent v2\infrastructure\local-testbed"
   docker compose up -d
   ```
2. **Memverifikasi Migrasi Basis Data**:
   Database `cifo_db` telah siap dengan migrasi 001 s.d. 007 di folder `apps/backend/migrations`.
3. **Menjalankan Backend Go**:
   ```powershell
   cd "d:\agent v2"
   powershell -ExecutionPolicy Bypass -File scripts\start-backend.ps1
   ```
   Backend akan mendengarkan pada `http://127.0.0.1:8080`.
4. **Menjalankan Frontend Next.js**:
   ```powershell
   cd "d:\agent v2\apps\frontend"
   npm install
   npm run dev
   ```
   Frontend akan dapat diakses pada `http://localhost:3001`.
5. **Kredensial Default**:
   - Admin: `admin@cifo.local` / `Admin123!`
   - DevOps: `devops@cifo.local` / `DevOps123!`
   - Token Dev: `dev-token-admin`, `dev-token-devops`, `dev-token-viewer`

---
*Dokumen ini merupakan checkpoint resmi penyelesaian Fase 8. Seluruh riwayat dan verifikasi tersimpan rapi dan dapat dipertanggungjawabkan.*

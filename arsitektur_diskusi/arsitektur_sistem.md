# Enterprise IT Monitoring & AI Ops Platform - Architecture Document
# Versi: 2.0 (Post-Audit Revision)
# Terakhir diperbarui: 2026-09-03

Dokumen ini adalah rancangan arsitektur komprehensif untuk platform pemantauan infrastruktur kelas enterprise dengan fokus utama pada pemantauan **ArgoCD (Kubernetes)** dan **Docker**. Platform ini dilengkapi dengan asisten AI (AIOps) canggih yang mendukung multi-model (fallback mechanism) dan sistem notifikasi insiden via Telegram.

Dokumen ini telah diperbarui berdasarkan hasil audit arsitektur yang mengidentifikasi 36+ temuan kelemahan. Setiap temuan telah diintegrasikan ke dalam rancangan.

---

## 1. Visi, Lingkup & Batasan Proyek

### 1.1 Visi
Platform ini adalah pusat komando (command center) infrastruktur, bukan sekadar dashboard pasif. Tujuannya adalah memberikan visibilitas mendalam, kemampuan diagnosis otomatis, dan eksekusi remediasi terkontrol terhadap seluruh siklus hidup kontainer (Docker) dan status sinkronisasi GitOps (ArgoCD/Kubernetes).

### 1.2 Lingkup Inti
- Pemantauan real-time status Docker containers, images, volumes, dan networks
- Pemantauan real-time ArgoCD applications, sync status, health status, dan deployment history
- Pemantauan resource Kubernetes (Pods, Deployments, Services, Nodes)
- AI Agent yang mampu melakukan diagnosis dan eksekusi perintah terbatas
- Alerting via Telegram Bot dan in-app notification
- Audit trail lengkap untuk setiap tindakan manusia dan AI

### 1.3 Batasan (Out of Scope)
- Bukan pengganti Grafana untuk visualisasi metrik umum
- Tidak menangani CI pipeline (hanya CD via ArgoCD)
- Tidak menangani manajemen cluster Kubernetes secara penuh (bukan pengganti Rancher/Lens)
- Tidak menyimpan atau memproses data pelanggan akhir (hanya data infrastruktur)

### 1.4 Lingkungan Pengujian Lokal (Local Testbed)
Untuk memastikan pengembangan yang aman dan terisolasi, kita menyediakan lingkungan pengujian lokal yang terdiri dari:
- Docker daemon lokal sebagai sumber data kontainer nyata
- Cluster Kubernetes lokal menggunakan K3d (ringan, cepat, berbasis Docker)
- ArgoCD yang di-deploy di dalam cluster K3d tersebut
- Semua data yang ditampilkan di dashboard adalah data nyata dari testbed ini, bukan data dummy atau halusinasi

---

## 2. Pilihan Teknologi (Technology Stack)

### A. Frontend (Antarmuka Pengguna)
- **Framework Utama**: Next.js 14+ (React 18) dengan TypeScript strict mode. Menggunakan App Router untuk routing berbasis file system yang terstruktur.
- **Styling**: Tailwind CSS v3 untuk utilitas cepat, digabungkan dengan komponen Radix UI Primitives untuk aksesibilitas (keyboard navigation, screen reader). Tema di-lock ke Dark Mode sebagai default, dengan opsi Light Mode via CSS custom properties toggle.
- **State Management**: Zustand untuk state UI global yang ringan (sidebar collapse, modal state). TanStack Query (React Query) v5 untuk server state (fetching, caching, background refetching metrik setiap 10 detik).
- **Real-time Communication**: WebSocket untuk streaming log kontainer dan event Kubernetes. Implementasi heartbeat ping/pong setiap 30 detik. Client memiliki reconnection otomatis dengan exponential backoff (1s, 2s, 4s, 8s, max 30s). Server menggunakan buffered channel per koneksi (max 1000 pesan), drop pesan lama jika buffer penuh untuk mencegah goroutine blocking.
- **Data Visualization**: Apache ECharts v5. Dipilih karena kemampuan menangani puluhan ribu titik data time-series tanpa lag. Semua chart instance menggunakan resize observer untuk responsivitas.
- **Notifikasi In-App**: Notification center di header dashboard. Toast notification untuk alert CRITICAL bersifat persistent (tidak auto-dismiss) sampai di-acknowledge pengguna. Menggunakan koneksi WebSocket yang sama untuk push notification real-time.
- **Loading & Error States**: Setiap widget memiliki tiga state visual: skeleton placeholder (loading), pesan error dengan tombol retry (error), dan tampilan kosong informatif (empty). React Error Boundary membungkus setiap widget agar crash pada satu komponen tidak merobohkan halaman.
- **Responsivitas**: Dua breakpoint minimum: desktop (1280px+) dan tablet (768px-1279px). Pada tablet, sidebar berubah menjadi collapsible drawer. Grafik tetap responsif via ECharts resize observer.

### B. Backend (Layanan Inti)
- **Bahasa Pemrograman**: Golang (Go) 1.22+. Bahasa standar untuk cloud-native tooling.
- **Pola Arsitektur**: Modular Monolith berbasis Domain-Driven Design (DDD). Setiap domain (monitoring, agent, notification, auth) adalah modul independen dengan batas yang jelas. Siap dipecah menjadi microservices jika beban meningkat tanpa refaktor besar.
- **Framework Web**: Echo v4 atau Chi. Dipilih karena berbasis `net/http` standar Go, menjamin kompatibilitas penuh dengan seluruh ekosistem middleware Go (termasuk library OIDC, gRPC-gateway, dan OpenTelemetry instrumentation). Catatan: Fiber (fasthttp) ditolak setelah evaluasi audit karena inkompatibilitas ekosistem.
- **Komunikasi Internal**: gRPC untuk komunikasi performa tinggi antar modul internal dan ke AI Service. REST API (JSON) untuk konsumsi Frontend.
- **Background Jobs**: Asynq (berbasis Redis) untuk memproses antrean tugas asinkron: pengiriman notifikasi Telegram, sinkronisasi massal data historis, pembersihan sesi AI yang expired.
- **Graceful Shutdown**: Entrypoint menggunakan `signal.NotifyContext` untuk menangkap SIGTERM/SIGINT. Urutan shutdown: (1) stop menerima request baru, (2) tunggu request aktif selesai (timeout 30 detik), (3) tutup koneksi WebSocket aktif, (4) kirim sinyal shutdown ke Asynq worker, (5) tutup connection pool database, (6) flush pending logs.
- **Health Checks**: Endpoint `/healthz` (liveness probe) memeriksa apakah proses masih responsif. Endpoint `/readyz` (readiness probe) memeriksa koneksi aktif ke PostgreSQL, Redis, Docker daemon, dan ArgoCD API. Kubernetes probes dikonfigurasi untuk memanggil endpoint ini.

### C. Database & Data Pipeline
- **Database Relasional**: PostgreSQL 16+.
  - Menyimpan: profil pengguna, konfigurasi RBAC, pengaturan webhook, histori interaksi AI, audit log, konfigurasi alerting.
  - Enkripsi at Rest: Aktifkan Transparent Data Encryption (TDE) atau gunakan encrypted disk volumes pada level storage.
  - Backup: WAL archiving untuk Point-in-Time Recovery (PITR). Backup otomatis harian via `pgBackRest`. Simpan backup di object storage terpisah (berbeda region/availability zone).
  - RTO (Recovery Time Objective): maksimal 1 jam. RPO (Recovery Point Objective): maksimal 5 menit.
  - Schema Migration: Menggunakan `golang-migrate` atau `Atlas`. Semua perubahan schema berupa file migration yang di-version control di folder `/backend/migrations`. Tidak boleh ada perubahan schema manual di produksi.
- **Database Time-Series**: VictoriaMetrics (menggantikan Prometheus standalone).
  - Alasan: VictoriaMetrics kompatibel penuh dengan PromQL dan Prometheus remote write, tetapi memiliki built-in long-term retention, kompresi data yang lebih baik (hingga 10x lebih hemat storage), dan performa query yang lebih tinggi.
  - Prometheus tetap digunakan sebagai scraper/collector, tapi data dikirim via remote write ke VictoriaMetrics untuk penyimpanan jangka panjang.
  - Retention policy: raw data 90 hari, downsampled data (5 menit agregasi) 1 tahun.
- **Log Aggregation**: Grafana Loki v3.
  - Format log: structured JSON (wajib). Setiap baris log mengandung: `timestamp`, `level`, `msg`, `trace_id`, `service`, `component`.
  - Retention policy berdasarkan level: ERROR/CRITICAL 90 hari, WARNING 30 hari, INFO/DEBUG 7 hari.
  - Storage backend: S3-compatible object storage untuk raw log chunks.
  - Enkripsi: encrypted volumes pada storage layer.
- **Cache & Message Queue**: Redis 7+.
  - Fungsi: rate-limiting API (sliding window), caching respons query berat, session store, dan message broker untuk Asynq.
  - Keamanan: `requirepass` wajib diaktifkan. TLS client authentication untuk koneksi dari backend. Tidak boleh terekspos ke jaringan publik.

### D. Agen AI (AI-Ops & Multi-Model Chat)

#### D.1 Arsitektur Layanan AI
AI Service dipisahkan sebagai layanan independen menggunakan **Python 3.12 + FastAPI + LangChain**. Alasan: ekosistem LangChain Python jauh lebih matang dan lengkap dibandingkan binding Go (`langchaingo`). Komunikasi dengan Backend Go via gRPC internal. Ini memberi fleksibilitas penuh untuk library AI Python tanpa mengorbankan performa Backend Go.

#### D.2 Strategi Multi-Model (High Availability AI)
- **Primary**: Google AI Studio (Gemini 2.0 Flash / 1.5 Pro). Context window besar, cocok untuk analisis log panjang.
- **Fallback 1**: OpenAI (GPT-4o). Jika Gemini rate limited atau gangguan.
- **Fallback 2**: Anthropic (Claude 3.5 Sonnet). Jika OpenAI juga tidak tersedia.
- **Fallback 3 (Offline)**: Ollama (Llama 3 / Mistral) untuk mode offline atau data sangat rahasia yang tidak boleh keluar dari jaringan internal.
- **Circuit Breaker**: Menggunakan pattern circuit breaker (library `sony/gobreaker` atau implementasi kustom di Python). Jika satu provider gagal 3 kali berturut-turut dalam 60 detik, circuit terbuka dan request dialihkan ke fallback berikutnya. Circuit mencoba menutup kembali setelah 120 detik.
- **Degraded Mode**: Jika SEMUA model (Google, OpenAI, Anthropic, Ollama) gagal bersamaan, AI chat menampilkan pesan eksplisit: "Fitur AI sedang tidak tersedia. Silakan gunakan dashboard manual." Insiden ini di-log dan alert Telegram dikirim ke administrator.

#### D.3 Conversation Memory & Session Management
- Tabel `ai_sessions` di PostgreSQL: `id`, `user_id`, `created_at`, `last_activity_at`, `status` (active/expired/closed).
- Tabel `ai_messages` di PostgreSQL: `id`, `session_id`, `role` (user/assistant/system), `content`, `model_used`, `input_tokens`, `output_tokens`, `timestamp`.
- TTL sesi: 30 menit tanpa aktivitas, sesi otomatis ditandai expired oleh background job Asynq.
- Context Window Management: Riwayat yang dikirim ke LLM dibatasi 20 pesan terakhir. Pesan yang lebih lama diringkas (summarized) oleh model murah (Gemini Flash) dan disertakan sebagai satu pesan konteks ringkasan.
- Pre-processing Log: Backend hanya mengirimkan baris log yang mengandung flag ERROR/CRITICAL/FATAL ditambah 50 baris sebelumnya sebagai konteks relevan. Log INFO/DEBUG difilter keluar.

#### D.4 Tool Calling / Function Calling (Definisi Terstruktur)
Setiap tool didefinisikan sebagai struct dengan schema yang ketat:

**Read-Only Tools (tanpa approval)**:
| Tool Name | Deskripsi | Parameter |
|-----------|-----------|-----------|
| `get_pod_status` | Status pod di namespace | `namespace`, `pod_name` (opsional) |
| `get_container_logs` | Log kontainer Docker | `container_id`, `tail_lines` (default 100) |
| `get_argocd_app_status` | Status sync ArgoCD app | `app_name` |
| `get_deployment_info` | Detail deployment K8s | `namespace`, `deployment_name` |
| `get_node_resources` | Utilisasi CPU/RAM node | `node_name` (opsional) |
| `get_docker_stats` | Statistik kontainer Docker | `container_id` (opsional) |
| `list_docker_containers` | Daftar kontainer aktif | `status_filter` (running/stopped/all) |
| `get_argocd_history` | Riwayat deployment ArgoCD | `app_name`, `limit` (default 10) |

**Write Tools (wajib approval via Human-in-the-loop)**:
| Tool Name | Deskripsi | Parameter | Required Role |
|-----------|-----------|-----------|---------------|
| `restart_deployment` | Restart deployment K8s | `namespace`, `deployment_name` | DevOps, Admin |
| `scale_deployment` | Scale replika deployment | `namespace`, `deployment_name`, `replicas` | DevOps, Admin |
| `sync_argocd_app` | Trigger sync ArgoCD | `app_name`, `prune` (boolean) | DevOps, Admin |
| `restart_container` | Restart kontainer Docker | `container_id` | DevOps, Admin |
| `stop_container` | Stop kontainer Docker | `container_id` | Admin |

**Perintah yang Dilarang Keras (Hardcoded Blocklist)**:
- `kubectl delete namespace`
- `kubectl delete node`
- `docker system prune`
- `docker volume rm` (tanpa konfirmasi ganda)
- Semua perintah yang mengandung `--force` pada resource kritis

**Alur Eksekusi Write Tool**:
1. AI mengidentifikasi kebutuhan aksi tulis dari percakapan
2. AI menghasilkan structured output (JSON) dengan nama tool dan parameter
3. Backend memvalidasi: (a) tool ada di allowlist, (b) parameter sesuai schema, (c) user memiliki role yang diperlukan
4. Frontend menampilkan konfirmasi eksplisit kepada pengguna: "AI ingin menjalankan: restart_deployment(namespace=production, deployment_name=payment-gateway). Setuju?"
5. Pengguna mengkonfirmasi atau menolak
6. Jika disetujui, backend mengeksekusi perintah
7. Hasil dicatat di `ai_action_audit_log`

#### D.5 Cost Control & Budget Tracking
- Tabel `ai_usage_tracking`: `id`, `user_id`, `session_id`, `model_provider`, `model_name`, `input_tokens`, `output_tokens`, `estimated_cost_usd`, `timestamp`.
- Budget ceiling per bulan per organisasi (dikonfigurasi di settings).
- Jika penggunaan mencapai 80% budget: peringatan dikirim ke admin.
- Jika penggunaan mencapai 100% budget: model otomatis beralih ke model termurah yang tersedia, atau fitur AI dinonaktifkan sampai reset bulan berikutnya.

#### D.6 Pertahanan Prompt Injection
- **Input Sanitization Layer**: Sebelum prompt dikirim ke LLM, layer sanitasi memfilter pola berbahaya yang diketahui (instruksi override seperti "ignore previous instructions", "act as root").
- **Allowlist Command Execution**: Output AI yang mengandung perintah eksekusi HARUS melalui parser deterministik yang mencocokkan terhadap daftar tool yang terdefinisi (lihat D.4). Perintah di luar daftar ditolak.
- **Output Validation**: Respons AI yang ditampilkan ke pengguna harus di-escape untuk mencegah XSS jika mengandung kode HTML/JavaScript.
- **Sandboxed Execution**: Perintah yang dieksekusi oleh AI dijalankan dengan akun layanan Kubernetes terpisah (`cifo-ai-agent-sa`) yang memiliki RBAC terbatas (lihat bagian Keamanan).

### E. Integrasi Pihak Ketiga & DevOps
- **Docker Integration**: Komunikasi via Docker Engine API. Untuk lokal: Unix Socket (`/var/run/docker.sock`). Untuk remote: TCP dengan mTLS (mutual TLS) wajib. Tidak ada koneksi Docker tanpa enkripsi.
- **ArgoCD Integration**: Mengkonsumsi ArgoCD gRPC/REST API. Autentikasi via ArgoCD API token yang disimpan di Vault. Polling status setiap 15 detik atau menggunakan ArgoCD Notification Webhook untuk event-driven updates.
- **Telegram Alerting**: Webhook-based Telegram Bot. Format notifikasi: teks rapi (Markdown), tanpa emoji, dengan penanda prioritas (CRITICAL, WARNING, INFO). Rate limiting pengiriman: maksimal 30 pesan per menit per chat group untuk menghindari Telegram API throttling.
- **Prometheus/Alertmanager**: Prometheus scrape endpoint `/metrics` dari Backend Go dan Docker daemon. Alertmanager menerima alert rules dan mengirim webhook ke Backend, yang kemudian meneruskan ke Telegram dan in-app notification.

---

## 3. Topologi & Arsitektur Jaringan (Deployment)

### 3.1 Komponen Runtime
- **Reverse Proxy**: Traefik sebagai gerbang utama. Menangani terminasi SSL/TLS, routing berbasis host/path, dan rate limiting layer pertama.
- **Service Mesh**: (Fase 2) Istio atau Linkerd untuk mTLS antar layanan di dalam cluster Kubernetes.
- **Lokasi Data**: Semua layanan database dan caching di-deploy dalam private subnet. Tidak ada port database yang terekspos ke internet publik. Akses hanya melalui VPN atau bastion host.

### 3.2 Kubernetes Network Policies
Setiap namespace memiliki NetworkPolicy yang membatasi komunikasi:
- **namespace: cifo-frontend**: Hanya boleh berkomunikasi ke `cifo-backend` (port 8080) dan Traefik ingress.
- **namespace: cifo-backend**: Boleh berkomunikasi ke PostgreSQL (port 5432), Redis (port 6379), VictoriaMetrics (port 8428), Loki (port 3100), ArgoCD API (port 443), Docker daemon, dan `cifo-ai-service` (port 50051 gRPC).
- **namespace: cifo-ai-service**: Boleh berkomunikasi ke `cifo-backend` (port 50051 gRPC), dan egress ke endpoint LLM eksternal (Google, OpenAI, Anthropic). Tidak boleh langsung ke database.
- **namespace: cifo-data**: PostgreSQL, Redis, VictoriaMetrics, Loki. Hanya menerima koneksi dari `cifo-backend`. Deny-all ingress dari namespace lain.

### 3.3 Diagram Alur Jaringan
```
Internet
    |
    v
[Traefik Ingress + TLS Termination]
    |
    +---> [Frontend (Next.js SSR)] <--- Static Assets (CDN)
    |
    +---> [Backend (Go)] ----gRPC----> [AI Service (Python)]
              |                             |
              +---> PostgreSQL              +---> Google AI Studio
              +---> Redis                   +---> OpenAI API
              +---> VictoriaMetrics         +---> Anthropic API
              +---> Loki                    +---> Ollama (internal)
              +---> Docker Engine API
              +---> ArgoCD API
              +---> Telegram Bot API
```

---

## 4. Keamanan Siber (Enterprise Security)

### 4.1 Autentikasi & Otorisasi
- **Identity Provider**: Keycloak dalam mode High Availability (minimal 2 replika, shared PostgreSQL session store). Mendukung SSO perusahaan via SAML 2.0 dan OpenID Connect (OIDC).
- **Token Strategy**: JWT access token (TTL 15 menit) + refresh token (TTL 7 hari). Backend memvalidasi JWT signature secara lokal (tanpa roundtrip ke Keycloak untuk setiap request) menggunakan JWKS endpoint yang di-cache.
- **Role-Based Access Control (RBAC)**: Tiga role utama:
  - **Admin**: Akses penuh. Dapat mengkonfigurasi sistem, mengelola pengguna, mengeksekusi semua perintah AI.
  - **DevOps**: Dapat melihat semua data monitoring, mengeksekusi perintah AI terbatas (restart, scale, sync). Tidak dapat mengelola pengguna atau mengubah konfigurasi keamanan.
  - **Viewer**: Hanya baca. Dapat melihat dashboard dan log. Tidak dapat mengeksekusi perintah apapun melalui AI atau antarmuka.
- **Multi-Factor Authentication (MFA)**: Wajib untuk role Admin dan DevOps. Opsional untuk Viewer. Didukung via Keycloak (TOTP, WebAuthn).

### 4.2 Manajemen Rahasia (Secrets Management)
- **HashiCorp Vault**: Semua credential disimpan dan diambil secara dinamis:
  - Kunci API Google AI Studio, OpenAI, Anthropic
  - Password database PostgreSQL
  - Redis password
  - Token Telegram Bot
  - ArgoCD API token
  - Keycloak admin credential
- **Rotasi Kunci**: Vault Dynamic Secrets dengan TTL pendek untuk database credentials (rotasi setiap 1 jam). Untuk API key yang tidak mendukung dynamic rotation (Google AI Studio), gunakan Vault Transit Engine untuk wrapping dan rotasi wrapper berkala.
- **Injeksi Runtime**: Aplikasi membaca secrets dari Vault saat startup dan me-refresh secara periodik. Tidak ada credential di environment variables, config files, atau source code.

### 4.3 Enkripsi
- **In Transit**: TLS 1.3 untuk semua komunikasi eksternal. mTLS untuk komunikasi internal antar layanan (Backend-to-AI-Service, Backend-to-Database).
- **At Rest**: PostgreSQL menggunakan encrypted disk volumes (LUKS/dm-crypt atau cloud-native encryption). Redis dikonfigurasi dengan TLS client authentication. Loki storage backend menggunakan encrypted S3 buckets (SSE-S3 atau SSE-KMS). VictoriaMetrics data directory pada encrypted volumes.

### 4.4 Pencegahan Serangan
- **WAF (Web Application Firewall)**: ModSecurity atau cloud-native WAF di depan Traefik untuk mendeteksi dan memblokir pola serangan umum (SQLi, XSS, path traversal).
- **CORS**: Strict origin whitelist. Hanya domain frontend yang diizinkan.
- **Rate Limiting**: Dua layer:
  - Layer 1 (Traefik): Global rate limit per IP (100 req/detik).
  - Layer 2 (Backend middleware): Per-user per-endpoint menggunakan Redis sliding window. Endpoint `/api/ai/chat` dibatasi 20 req/menit per user.
- **Input Validation**: Semua input HTTP di-validasi menggunakan struct tags Go (`validator` library). Request body yang melebihi ukuran maksimal (misal 1MB) ditolak.
- **CSP (Content Security Policy)**: Header CSP ketat pada respons Frontend untuk mencegah XSS dan data exfiltration.
- **HSTS**: HTTP Strict Transport Security diaktifkan dengan max-age minimal 1 tahun.

### 4.5 Audit Trail
- Tabel `audit_log` (append-only, tidak boleh UPDATE/DELETE):
  - `id` (UUID)
  - `timestamp` (with timezone)
  - `actor_type` (user/ai_agent/system)
  - `actor_id` (user_id atau service_account_name)
  - `action` (enum: login, view, execute, configure, approve, reject)
  - `resource_type` (deployment, container, argocd_app, setting)
  - `resource_id`
  - `details` (JSONB, berisi parameter spesifik)
  - `ip_address`
  - `user_agent`
  - `result` (success/failure)
- Tabel `ai_action_audit_log` (khusus aksi AI, append-only):
  - `id`, `user_id`, `session_id`, `prompt_input_hash` (SHA-256, bukan plaintext untuk privasi), `ai_output_summary`, `tool_name`, `tool_parameters` (JSONB), `approval_status` (pending/approved/rejected), `execution_result`, `model_used`, `timestamp`

### 4.6 AI Agent Security (Zero-Trust AI)
- Kubernetes ServiceAccount khusus: `cifo-ai-agent-sa`
- ClusterRole terbatas:
  - Allowed: get/list/watch pada pods, deployments, services, events, logs
  - Allowed: patch pada deployments (hanya untuk restart dan scale, di-enforce oleh Admission Webhook)
  - Denied: delete pada semua resource
  - Denied: create/delete namespace
  - Denied: akses ke secrets dan configmaps
- Docker akses AI agent: read-only (inspect, logs, stats). Tidak boleh stop/remove/exec.
- Semua perintah dari AI melewati validasi backend (bukan direct execution)

---

## 5. Observability & Monitoring Stack

### 5.1 Metrik (Metrics)
- **Collector**: Prometheus scrape config untuk mengumpulkan metrik dari:
  - Backend Go (`/metrics` endpoint): goroutine count, heap memory, HTTP request latency histogram, active WebSocket connections
  - Docker daemon: container CPU/memory/network/disk via cAdvisor
  - Kubernetes nodes: kubelet metrics
  - ArgoCD: application sync status metrics
  - PostgreSQL: pg_exporter
  - Redis: redis_exporter
- **Storage**: VictoriaMetrics (long-term, PromQL compatible)
- **Self-Monitoring**: Backend Go wajib mengekspos metrik internalnya sendiri. Sistem monitoring tidak boleh buta terhadap statusnya sendiri.

### 5.2 Logging (Logs)
- **Format**: Structured JSON via `slog` (standard library Go 1.22+). Setiap baris log wajib mengandung:
  ```json
  {
    "timestamp": "2026-09-03T15:00:00.000Z",
    "level": "ERROR",
    "msg": "failed to connect",
    "trace_id": "abc123",
    "span_id": "def456",
    "service": "backend",
    "component": "docker_client"
  }
  ```
- **Pipeline**: Aplikasi menulis ke stdout (standar 12-factor app). Loki Promtail atau Alloy mengambil log dari stdout kontainer dan mengirim ke Loki.
- **Kontainer Log Streaming**: Log dari Docker containers dan Kubernetes pods di-stream ke frontend via WebSocket. Backend bertindak sebagai proxy yang mem-buffer dan meneruskan log stream.

### 5.3 Tracing (Distributed Tracing)
- **Library**: OpenTelemetry SDK di Backend Go dan AI Service Python.
- **Backend**: Grafana Tempo (terintegrasi baik dengan Loki dan VictoriaMetrics di Grafana stack).
- **Propagasi**: Setiap HTTP/gRPC request membawa `trace_id` dan `span_id` yang di-propagasi ke semua layanan downstream (termasuk ke AI Service).
- **Correlation**: `trace_id` yang sama digunakan di log (Loki) dan trace (Tempo), memungkinkan navigasi langsung dari log entry ke trace visual.

### 5.4 Alerting Rules (Prometheus/Alertmanager)
Definisi alert rules yang wajib ada sejak awal:

| Alert Name | Kondisi | Severity | Aksi |
|------------|---------|----------|------|
| `ContainerCpuHigh` | CPU kontainer > 85% selama 5 menit | WARNING | Telegram + In-App |
| `ContainerCpuCritical` | CPU kontainer > 95% selama 3 menit | CRITICAL | Telegram + In-App + AI Analysis |
| `ContainerOOMKilled` | Kontainer mati karena OOM | CRITICAL | Telegram + In-App |
| `ArgoCDSyncFailed` | Sync gagal | CRITICAL | Telegram + In-App |
| `ArgoCDAppDegraded` | App status Degraded | WARNING | Telegram + In-App |
| `PodCrashLooping` | Restart > 3x dalam 10 menit | CRITICAL | Telegram + In-App + AI Analysis |
| `DiskUsageHigh` | Disk > 80% | WARNING | Telegram + In-App |
| `DiskUsageCritical` | Disk > 90% | CRITICAL | Telegram + In-App |
| `NodeNotReady` | Node status NotReady | CRITICAL | Telegram + In-App |
| `BackendHealthCheckFailed` | /healthz gagal > 30 detik | CRITICAL | Telegram (langsung, bypass backend) |
| `AIServiceUnavailable` | Semua model AI gagal | WARNING | Telegram + In-App |
| `HighErrorRate` | Error rate > 5% dalam 5 menit | WARNING | Telegram + In-App |

Route semua alert ke Backend webhook, yang kemudian meneruskan ke Telegram Bot dan in-app notification center secara paralel.

---

## 6. Strategi Testing (Quality Assurance)

### 6.1 Piramida Testing
- **Unit Test**:
  - Go: `testing` + `testify` (assertions & mocks). Target coverage: minimal 70% untuk package `service` dan `repository`.
  - Frontend: `vitest` + `@testing-library/react`. Target coverage: minimal 60% untuk komponen interaktif.
- **Integration Test**:
  - Go: Menggunakan `testcontainers-go` untuk spin up PostgreSQL dan Redis dalam container saat testing. Test endpoint API terhadap database sungguhan.
  - Folder: `/tests/integration`
- **End-to-End Test**:
  - Playwright untuk alur pengguna kritis: login, navigasi dashboard, kirim perintah AI, terima notifikasi, approve aksi AI.
  - Folder: `/tests/e2e`
- **Load Test**:
  - K6 untuk memastikan backend mampu menangani volume metrik dan koneksi WebSocket tinggi.
  - Skenario: 500 concurrent WebSocket connections streaming log, 1000 req/detik ke API metrik.
  - Folder: `/tests/load`

### 6.2 Contract Testing
- File OpenAPI 3.1 / Swagger di `/packages/api-contracts/openapi.yaml` sebagai sumber kebenaran tunggal.
- File Protobuf di `/packages/api-contracts/proto/` untuk gRPC contracts.
- Generate Go server stubs dan TypeScript client types dari file yang sama.
- CI memvalidasi bahwa implementasi sesuai dengan kontrak. Jika tidak sesuai, build gagal.

---

## 7. CI/CD Pipeline

### 7.1 Tooling
- **CI**: GitHub Actions (atau GitLab CI).
- **Container Registry**: Harbor (self-hosted) atau cloud-native (ECR/GCR/ACR).
- **Deployment**: GitOps via ArgoCD. Backend tidak melakukan deployment langsung. Perubahan image tag di-commit ke config repo, ArgoCD auto-sync.

### 7.2 Pipeline Stages
```
[Push to main/PR]
    |
    v
[1. Lint & Format]
    - golangci-lint (Go)
    - eslint + prettier (Frontend)
    - ruff (Python AI Service)
    - hadolint (Dockerfiles)
    |
    v
[2. Unit Test] (paralel per layanan)
    - go test ./...
    - npx vitest run
    - pytest
    |
    v
[3. Integration Test]
    - testcontainers (database tests)
    |
    v
[4. Security Scan]
    - trivy (container image vulnerability scan)
    - gosec (Go security linter)
    - gitleaks (credential leak detection in code)
    |
    v
[5. Build Docker Images] (hanya layanan yang berubah, via Turborepo/change detection)
    |
    v
[6. Push to Container Registry]
    |
    v
[7. Update ArgoCD App Manifest] (commit new image tag to config repo)
    |
    v
[ArgoCD Auto-Sync to Staging] --> [Manual Approval] --> [ArgoCD Sync to Production]
```

### 7.3 Monorepo Tooling
- **Frontend packages**: Turborepo untuk task orchestration dan caching.
- **Go modules**: Go Workspaces (`go.work`) untuk mengelola multi-module dalam satu repo.
- **Change Detection**: CI hanya membangun ulang layanan yang berubah. Menggunakan path filter di GitHub Actions atau Turborepo cache.

---

## 8. Standar Pengembangan (Coding Standards)

### 8.1 Komentar Kode
Komentar harus sangat ringkas, jelas, dan profesional. Format yang disetujui (1-4 kata, tanpa simbol dekoratif):
- Benar: `// init db pool`
- Benar: `// validate token`
- Benar: `// parse ws message`
- Salah: `// --- Fungsi untuk mengambil data dari database ---`
- Salah: `// *** IMPORTANT: This handles the connection ***`

### 8.2 Error Handling (Go)
- Definisikan custom error types di `/pkg/apperror`:
  ```go
  type AppError struct {
      Code       string // "AUTH_001", "DOCKER_002"
      HTTPStatus int    // 401, 500
      Message    string // internal log message
      UserMsg    string // safe message for API response
      Err        error  // wrapped original error
  }
  ```
- Gunakan error wrapping: `fmt.Errorf("connect db: %w", err)` untuk menjaga stack trace.
- Hindari `panic()`. Selalu kembalikan error dan tangani di level atas.
- Tidak ada hardcoded credential di manapun.

### 8.3 Git Workflow
- **Branching**: trunk-based development. Branch `main` selalu deployable.
- **Feature branches**: `feature/<ticket-id>-<short-description>`
- **Commit messages**: Conventional Commits (`feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `chore:`)
- **PR Requirements**: Minimal 1 reviewer approval, semua CI checks pass, no credential leaks detected.

### 8.4 Data Integrity
- Tidak ada penggunaan mock data atau data dummy di lingkungan development. Semua data berasal dari local testbed (Docker daemon + K3d cluster + ArgoCD).
- Seed data (`/scripts/seed-data.sql`) hanya berisi konfigurasi sistem wajib (default roles, default alert rules), bukan data palsu.

---

## 9. Struktur Folder Monorepo (Lengkap)

```text
/cifo-monitoring-platform
│
├── /apps
│   ├── /frontend                          # Next.js 14 web application
│   │   ├── /public                        # Static assets (favicon, manifest)
│   │   ├── /src
│   │   │   ├── /app                       # App Router (pages, layouts, loading, error)
│   │   │   │   ├── /(auth)                # Route group: login, callback
│   │   │   │   ├── /(dashboard)           # Route group: monitoring, kubernetes, docker
│   │   │   │   │   ├── /monitoring        # Halaman monitoring utama
│   │   │   │   │   ├── /kubernetes        # Detail pods, deployments, nodes
│   │   │   │   │   ├── /docker            # Detail containers, images, volumes
│   │   │   │   │   ├── /incidents         # Daftar dan detail insiden
│   │   │   │   │   └── /settings          # Konfigurasi pengguna dan sistem
│   │   │   │   ├── layout.tsx             # Root layout (sidebar, header)
│   │   │   │   └── not-found.tsx          # 404 page
│   │   │   ├── /components
│   │   │   │   ├── /ui                    # Primitif (Button, Input, Modal, Toast)
│   │   │   │   ├── /charts               # ECharts wrappers (CpuChart, MemoryChart)
│   │   │   │   ├── /terminal             # Log terminal emulator
│   │   │   │   ├── /ai-chat              # Chat interface (input, bubble, toolbar)
│   │   │   │   ├── /notifications        # Notification center, toast stack
│   │   │   │   ├── /layout               # Sidebar, Header, Breadcrumb
│   │   │   │   └── /widgets              # Dashboard cards (ContainerCount, etc.)
│   │   │   ├── /hooks
│   │   │   │   ├── use-websocket.ts       # WebSocket dengan reconnect
│   │   │   │   ├── use-metrics.ts         # Polling metrik via React Query
│   │   │   │   └── use-auth.ts            # Auth state dan redirect
│   │   │   ├── /lib
│   │   │   │   ├── api-client.ts          # Axios instance dengan interceptors
│   │   │   │   ├── ws-client.ts           # WebSocket client class
│   │   │   │   └── format.ts             # Formatter (bytes, dates, uptime)
│   │   │   ├── /store
│   │   │   │   ├── sidebar-store.ts       # Sidebar collapse state
│   │   │   │   ├── notification-store.ts  # In-app notification queue
│   │   │   │   └── theme-store.ts         # Dark/Light mode toggle
│   │   │   ├── /types
│   │   │   │   ├── api.ts                 # Generated dari OpenAPI spec
│   │   │   │   ├── websocket.ts           # WS message types
│   │   │   │   └── models.ts              # Domain model interfaces
│   │   │   └── /styles
│   │   │       ├── globals.css            # CSS reset, custom properties, tema
│   │   │       └── tokens.css             # Design tokens (colors, spacing, fonts)
│   │   ├── Dockerfile
│   │   ├── next.config.ts
│   │   ├── tailwind.config.ts
│   │   ├── tsconfig.json
│   │   └── package.json
│   │
│   ├── /backend                           # Golang core service
│   │   ├── /cmd
│   │   │   ├── /server                    # Entrypoint HTTP/gRPC server
│   │   │   │   └── main.go
│   │   │   └── /worker                    # Entrypoint Asynq background worker
│   │   │       └── main.go
│   │   ├── /internal
│   │   │   ├── /config                    # Env parsing, Vault integration
│   │   │   │   └── config.go
│   │   │   ├── /handler                   # HTTP handlers (Fiber/Echo/Chi)
│   │   │   │   ├── auth_handler.go
│   │   │   │   ├── monitoring_handler.go
│   │   │   │   ├── docker_handler.go
│   │   │   │   ├── argocd_handler.go
│   │   │   │   ├── ai_handler.go
│   │   │   │   ├── incident_handler.go
│   │   │   │   ├── websocket_handler.go
│   │   │   │   └── health_handler.go
│   │   │   ├── /middleware
│   │   │   │   ├── auth.go                # JWT validation, RBAC enforcement
│   │   │   │   ├── ratelimit.go           # Redis sliding window
│   │   │   │   ├── logger.go              # Request logging with trace_id
│   │   │   │   ├── cors.go                # Strict CORS config
│   │   │   │   └── recovery.go            # Panic recovery (catch, log, 500)
│   │   │   ├── /model
│   │   │   │   ├── user.go
│   │   │   │   ├── container.go
│   │   │   │   ├── deployment.go
│   │   │   │   ├── argocd_app.go
│   │   │   │   ├── incident.go
│   │   │   │   ├── ai_session.go
│   │   │   │   ├── ai_message.go
│   │   │   │   └── audit_log.go
│   │   │   ├── /repository
│   │   │   │   ├── user_repo.go
│   │   │   │   ├── incident_repo.go
│   │   │   │   ├── ai_session_repo.go
│   │   │   │   ├── audit_repo.go
│   │   │   │   └── cache_repo.go          # Redis abstraction
│   │   │   ├── /service
│   │   │   │   ├── auth_service.go
│   │   │   │   ├── monitoring_service.go
│   │   │   │   ├── docker_service.go
│   │   │   │   ├── argocd_service.go
│   │   │   │   ├── ai_service.go          # gRPC client to AI Service
│   │   │   │   ├── notification_service.go
│   │   │   │   ├── incident_service.go
│   │   │   │   └── telegram_service.go
│   │   │   ├── /integration
│   │   │   │   ├── docker_client.go       # Docker Engine API wrapper
│   │   │   │   ├── argocd_client.go       # ArgoCD API wrapper
│   │   │   │   ├── telegram_client.go     # Telegram Bot API wrapper
│   │   │   │   ├── prometheus_client.go   # PromQL query executor
│   │   │   │   └── vault_client.go        # Vault secret fetcher
│   │   │   └── /ws
│   │   │       ├── hub.go                 # WebSocket connection manager
│   │   │       ├── client.go              # Per-connection handler
│   │   │       └── message.go             # WS message types
│   │   ├── /migrations
│   │   │   ├── 001_create_users.up.sql
│   │   │   ├── 001_create_users.down.sql
│   │   │   ├── 002_create_ai_sessions.up.sql
│   │   │   ├── 002_create_ai_sessions.down.sql
│   │   │   ├── 003_create_audit_log.up.sql
│   │   │   ├── 003_create_audit_log.down.sql
│   │   │   ├── 004_create_incidents.up.sql
│   │   │   └── 004_create_incidents.down.sql
│   │   ├── /pkg
│   │   │   ├── /apperror                  # Custom error types
│   │   │   │   └── error.go
│   │   │   ├── /logger                    # Structured slog wrapper
│   │   │   │   └── logger.go
│   │   │   └── /validator                 # Input validation helpers
│   │   │       └── validator.go
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── go.sum
│   │
│   └── /ai-service                        # Python AI Agent service
│       ├── /app
│       │   ├── main.py                    # FastAPI + gRPC entrypoint
│       │   ├── /agent
│       │   │   ├── orchestrator.py        # Model routing, circuit breaker
│       │   │   ├── memory.py              # Session & conversation history
│       │   │   └── sanitizer.py           # Prompt injection defense
│       │   ├── /tools
│       │   │   ├── base.py                # Base tool interface
│       │   │   ├── kubectl_tools.py       # K8s read/write tools
│       │   │   ├── docker_tools.py        # Docker read/write tools
│       │   │   └── argocd_tools.py        # ArgoCD read/write tools
│       │   ├── /providers
│       │   │   ├── base.py                # Abstract LLM provider
│       │   │   ├── google_provider.py     # Google AI Studio adapter
│       │   │   ├── openai_provider.py     # OpenAI adapter
│       │   │   ├── anthropic_provider.py  # Anthropic adapter
│       │   │   └── ollama_provider.py     # Local Ollama adapter
│       │   ├── /prompts
│       │   │   ├── system_prompt.txt      # Base system instructions
│       │   │   ├── diagnosis_prompt.txt   # Root cause analysis template
│       │   │   └── summarize_prompt.txt   # Conversation summarization
│       │   └── /config
│       │       └── settings.py            # Pydantic settings from env/Vault
│       ├── /proto
│       │   └── ai_service.proto           # gRPC service definition
│       ├── Dockerfile
│       ├── pyproject.toml
│       └── requirements.lock
│
├── /packages                              # Shared libraries
│   ├── /api-contracts
│   │   ├── openapi.yaml                   # REST API specification (source of truth)
│   │   └── /proto
│   │       ├── ai_service.proto
│   │       └── common.proto
│   ├── /eslint-config
│   │   └── index.js                       # Shared ESLint rules
│   └── /theme
│       ├── colors.json                    # Design tokens
│       └── typography.json                # Font definitions
│
├── /infrastructure
│   ├── /local-testbed
│   │   ├── docker-compose.yml             # Postgres, Redis, VictoriaMetrics, Loki, Prometheus
│   │   ├── docker-compose.monitoring.yml  # Grafana, Tempo, Alertmanager
│   │   ├── /k3d
│   │   │   ├── cluster-config.yaml        # K3d cluster definition
│   │   │   └── setup-cluster.sh           # Create cluster + install ArgoCD
│   │   ├── /argocd
│   │   │   ├── install.yaml               # ArgoCD installation manifest
│   │   │   └── sample-apps/               # Sample K8s apps for testing monitoring
│   │   ├── /prometheus
│   │   │   ├── prometheus.yml             # Scrape config
│   │   │   └── alert-rules.yml            # Alerting rules definition
│   │   └── /alertmanager
│   │       └── alertmanager.yml           # Route alerts to backend webhook
│   │
│   ├── /kubernetes                        # Production manifests (Helm Charts)
│   │   ├── /charts
│   │   │   ├── /cifo-frontend
│   │   │   ├── /cifo-backend
│   │   │   ├── /cifo-ai-service
│   │   │   └── /cifo-data                 # PostgreSQL, Redis, VictoriaMetrics
│   │   ├── /base                          # Kustomize base
│   │   └── /overlays
│   │       ├── /staging
│   │       └── /production
│   │
│   ├── /terraform                         # Cloud resource provisioning
│   │   ├── /modules
│   │   │   ├── /vpc
│   │   │   ├── /eks                       # Kubernetes cluster
│   │   │   ├── /rds                       # Managed PostgreSQL
│   │   │   └── /s3                        # Object storage for Loki/backups
│   │   ├── /environments
│   │   │   ├── /staging
│   │   │   └── /production
│   │   └── backend.tf                     # Terraform state backend (S3 + DynamoDB)
│   │
│   └── /security
│       ├── /network-policies              # K8s NetworkPolicy manifests
│       ├── /rbac                          # K8s RBAC for AI agent ServiceAccount
│       └── /vault                         # Vault policy definitions
│
├── /tests
│   ├── /integration                       # API integration tests
│   │   ├── docker_api_test.go
│   │   ├── argocd_api_test.go
│   │   └── auth_api_test.go
│   ├── /e2e                               # Playwright E2E tests
│   │   ├── login.spec.ts
│   │   ├── dashboard.spec.ts
│   │   ├── ai-chat.spec.ts
│   │   └── playwright.config.ts
│   └── /load                              # K6 load tests
│       ├── websocket-stress.js
│       └── api-throughput.js
│
├── /scripts
│   ├── setup-local.sh                     # Bootstrap local dev environment
│   ├── setup-local.ps1                    # PowerShell variant for Windows
│   ├── generate-certs.sh                  # Generate mTLS certificates
│   ├── seed-data.sql                      # System config seed (roles, alert rules)
│   ├── generate-api-types.sh              # Generate Go/TS types from OpenAPI
│   └── run-all-tests.sh                   # Run unit + integration + e2e
│
├── /docs
│   ├── architecture.md                    # Dokumen ini
│   ├── audit-findings.md                  # Hasil audit arsitektur
│   ├── api-reference.md                   # REST API documentation
│   ├── ai-agent-capabilities.md           # AI tools, permissions, limits
│   ├── deployment-guide.md                # Production deployment runbook
│   ├── incident-response.md               # Incident handling procedures
│   ├── security-policy.md                 # Security policies dan compliance
│   └── /adr                               # Architecture Decision Records
│       ├── 001-use-echo-over-fiber.md
│       ├── 002-victoriametrics-over-prometheus.md
│       ├── 003-separate-ai-service-python.md
│       └── 004-multi-model-fallback.md
│
├── .github
│   └── /workflows
│       ├── ci.yml                         # Lint, test, build, scan
│       ├── deploy-staging.yml             # Auto-deploy to staging on merge
│       └── deploy-production.yml          # Manual approval deploy to production
│
├── .gitignore
├── .editorconfig                          # Consistent formatting across IDEs
├── go.work                                # Go Workspaces for multi-module
├── turbo.json                             # Turborepo configuration
├── Makefile                               # Shortcuts: make dev, make test, make build
└── README.md                              # Project overview dan quickstart
```

---

## 10. Alur Kerja Operasional (Workflow)

### 10.1 Alur Monitoring & Alert
1. Prometheus scrape metrik dari Docker daemon, Kubernetes nodes, dan ArgoCD setiap 15 detik
2. VictoriaMetrics menyimpan metrik untuk long-term query
3. Alertmanager mengevaluasi alert rules
4. Jika threshold terpenuhi, Alertmanager mengirim webhook ke Backend
5. Backend memproses alert: simpan ke database sebagai incident, kirim notifikasi ke Telegram Bot, push in-app notification via WebSocket
6. Frontend menerima notifikasi dan menampilkan toast (CRITICAL: persistent, WARNING: 10 detik auto-dismiss)

### 10.2 Alur AI Diagnosis
1. Alert CRITICAL diterima oleh Backend
2. Backend secara otomatis mengirimkan konteks insiden (log ERROR/CRITICAL + 50 baris sebelumnya) ke AI Service
3. AI Service menganalisis menggunakan model aktif (Gemini/OpenAI/Anthropic)
4. AI menghasilkan Root Cause Analysis (RCA) dan rekomendasi tindakan
5. RCA disimpan di database dan ditampilkan di incident detail page
6. Jika AI merekomendasikan tindakan (restart, scale), pengguna harus approve secara manual

### 10.3 Alur Backup & Recovery
1. pgBackRest menjalankan backup inkremental setiap 6 jam, full backup setiap minggu
2. WAL archiving aktif terus-menerus untuk PITR
3. Backup disimpan di object storage terpisah (region berbeda)
4. Restore test dijalankan secara otomatis setiap bulan di lingkungan staging
5. RTO: 1 jam, RPO: 5 menit

---

## 11. Analisis Kritis & Saran Strategis

### Saran 1: Pola Circuit Breaker untuk AI
Integrasi LLM sangat rentan terhadap rate limit atau latensi jaringan. Wajib mengimplementasikan circuit breaker. Jika endpoint Google AI Studio gagal 3x beruntun dalam 60 detik, circuit terbuka dan request dialihkan ke fallback model tanpa intervensi manual.

### Saran 2: Streaming Log via Persistent Connection
Membaca log dari ratusan kontainer Docker/ArgoCD dengan HTTP Polling akan menghabiskan CPU dan bandwidth. Gunakan koneksi TCP persisten dan teruskan data langsung ke klien via WebSocket. Backend bertindak sebagai multiplexer yang menggabungkan multiple log streams.

### Saran 3: Self-Monitoring
Sistem monitoring tidak boleh buta terhadap statusnya sendiri. Backend Go wajib mengekspos metrik internal (goroutine count, heap memory, active DB connections, active WebSocket connections) via endpoint `/metrics`.

### Saran 4: Context Window Management
Menyuntikkan seluruh riwayat log ke model AI akan memicu error batas token. Backend hanya mengirimkan baris log ERROR/CRITICAL/FATAL ditambah 50 baris sebelumnya. Log INFO/DEBUG difilter keluar sebelum dikirim ke AI.

### Saran 5: Zero-Trust AI
AI tidak boleh memiliki akses `cluster-admin`. RBAC Kubernetes khusus dibuat untuk AI ServiceAccount. Izin baca luas, izin tulis sangat terbatas. Larangan keras untuk operasi penghapusan resource. Semua perintah dari AI melewati validasi backend.

### Saran 6: Token Refresh dan Session Invalidation
JWT access token harus pendek (15 menit) untuk meminimalkan dampak kebocoran. Refresh token disimpan di HttpOnly cookie. Backend harus mampu menginvalidasi semua sesi aktif dari satu user (force logout) untuk kasus akun yang dikompromikan.

### Saran 7: Database Connection Pooling
PostgreSQL connection pool harus dikonfigurasi dengan cermat. Gunakan `pgxpool` di Go dengan min 5, max 25 koneksi per instance backend. Monitor pool stats via metrik Prometheus untuk mendeteksi connection leak.

### Saran 8: Docker Socket Security
Mount Docker socket (`/var/run/docker.sock`) ke kontainer backend adalah risiko keamanan tinggi. Pertimbangkan menggunakan Docker socket proxy (seperti Tecnativa docker-socket-proxy) yang membatasi API call yang diizinkan (hanya GET untuk read-only, POST terbatas untuk restart).

### Saran 9: Idempotency pada Write Operations
Semua write operation (restart, scale, sync) harus idempoten. Jika pengguna menekan tombol approve dua kali karena lag jaringan, operasi tidak boleh dieksekusi ganda. Gunakan idempotency key per request.

### Saran 10: Telegram Bot Rate Limiting dan Failover
Telegram API membatasi pengiriman pesan. Jika terjadi alert storm (puluhan alert dalam waktu singkat), pesan harus di-batch dan di-deduplikasi. Kirim ringkasan ("5 CRITICAL alerts dalam 2 menit terakhir") daripada 5 pesan terpisah. Jika Telegram API tidak tersedia, log alert tetap tersimpan di database dan dikirim ulang saat koneksi pulih.

### Saran 11: Image Vulnerability Scanning di CI
Setiap Docker image yang di-build harus dipindai menggunakan Trivy sebelum di-push ke registry. Image dengan vulnerability CRITICAL harus gagal build (pipeline dihentikan). Image dengan HIGH mendapat warning tapi tetap bisa lanjut ke staging (tidak ke production).

### Saran 12: Incident Lifecycle Management
Setiap incident yang dibuat dari alert harus memiliki lifecycle yang jelas: Open -> Acknowledged -> Investigating -> Resolved -> Closed. Transisi state dicatat di audit log. Incident yang tidak di-acknowledge dalam 15 menit akan di-escalate (kirim ulang ke Telegram dengan penanda ESCALATED).

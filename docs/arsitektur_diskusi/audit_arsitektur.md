# Audit Arsitektur: Temuan Kelemahan & Evaluasi Kritis
# Versi: 2.0 (Diperbarui setelah revisi arsitektur)

Dokumen ini adalah hasil analisis mendalam terhadap `arsitektur_sistem.md`. Setiap temuan dikategorikan berdasarkan domain keahlian dan tingkat keparahan.

Skala Keparahan:
- **P0 (Kritis)**: Harus diperbaiki sebelum implementasi dimulai. Bisa menyebabkan kegagalan sistem atau kebocoran keamanan.
- **P1 (Tinggi)**: Harus direncanakan di awal. Akan menjadi masalah besar jika ditunda.
- **P2 (Sedang)**: Perlu dipertimbangkan. Bisa ditangani di iterasi kedua tapi berisiko jika diabaikan total.

Status:
- **RESOLVED**: Temuan telah ditutup di arsitektur v2.0
- **NEW**: Temuan baru yang teridentifikasi saat revisi

---

## BAGIAN 1: KEAMANAN SIBER (Cybersecurity Audit)

### 1.1 [P0] Tidak Ada Strategi Enkripsi Data at Rest -- RESOLVED
Dokumen awal hanya membahas enkripsi data in transit (TLS/mTLS). Tidak ada rencana untuk enkripsi data yang tersimpan.

**Resolusi**: Arsitektur v2.0 Bagian 4.3 mendefinisikan enkripsi at-rest untuk PostgreSQL (encrypted volumes), Redis (TLS client auth), Loki (encrypted S3), dan VictoriaMetrics (encrypted volumes).

### 1.2 [P0] Prompt Injection pada AI Agent -- RESOLVED
Agen AI memiliki akses ke Docker dan Kubernetes API tanpa pertahanan prompt injection.

**Resolusi**: Arsitektur v2.0 Bagian D.6 mendefinisikan: input sanitization layer, allowlist command execution, output validation, dan sandboxed execution dengan ServiceAccount terbatas.

### 1.3 [P0] Tidak Ada Audit Log untuk Aksi AI -- RESOLVED
Tidak ada mekanisme pencatatan aksi yang dieksekusi AI.

**Resolusi**: Arsitektur v2.0 Bagian 4.5 mendefinisikan dua tabel audit: `audit_log` (umum) dan `ai_action_audit_log` (khusus AI), keduanya append-only.

### 1.4 [P1] Token API AI Tersimpan di Satu Titik Kegagalan -- RESOLVED
Tidak ada desain untuk rotasi otomatis kunci API.

**Resolusi**: Arsitektur v2.0 Bagian 4.2 mendefinisikan Vault Dynamic Secrets dengan TTL pendek dan Transit Engine untuk wrapping.

### 1.5 [P1] Tidak Ada Network Policy Kubernetes -- RESOLVED
Tidak ada rancangan Kubernetes NetworkPolicy.

**Resolusi**: Arsitektur v2.0 Bagian 3.2 mendefinisikan NetworkPolicy per namespace dengan aturan komunikasi yang spesifik.

### 1.6 [P1] Keycloak Sebagai Single Point of Failure -- RESOLVED
Jika Keycloak mati, seluruh sistem tidak bisa melakukan autentikasi.

**Resolusi**: Arsitektur v2.0 Bagian 4.1 mendefinisikan Keycloak HA (2 replika), JWT local validation via JWKS cache, dan token caching.

### 1.7 [P2] Tidak Ada Rate Limiting Spesifik untuk Endpoint AI Chat -- RESOLVED
Rate limiting umum disebut tapi endpoint AI chat butuh perlakuan khusus.

**Resolusi**: Arsitektur v2.0 Bagian 4.4 mendefinisikan dua layer rate limiting, dengan endpoint `/api/ai/chat` dibatasi 20 req/menit per user.

### 1.8 [P0] [NEW] Docker Socket Exposure
Mount Docker socket (`/var/run/docker.sock`) ke kontainer backend adalah risiko keamanan tinggi. Penyerang yang mendapat akses ke kontainer backend bisa mendapat kontrol penuh atas Docker host.

**Resolusi**: Arsitektur v2.0 Saran 8 merekomendasikan Docker socket proxy (Tecnativa docker-socket-proxy) untuk membatasi API call yang diizinkan.

### 1.9 [P1] [NEW] Tidak Ada MFA (Multi-Factor Authentication)
Dokumen awal tidak menyebutkan MFA untuk akun dengan privilege tinggi.

**Resolusi**: Arsitektur v2.0 Bagian 4.1 mewajibkan MFA untuk role Admin dan DevOps via Keycloak (TOTP, WebAuthn).

### 1.10 [P1] [NEW] Tidak Ada CSP dan HSTS Headers
Tidak ada definisi security headers untuk response HTTP.

**Resolusi**: Arsitektur v2.0 Bagian 4.4 mendefinisikan CSP ketat dan HSTS dengan max-age minimal 1 tahun.

### 1.11 [P1] [NEW] JWT Token Tanpa Invalidation Strategy
JWT bersifat stateless. Jika akun dikompromikan, token aktif tidak bisa direvoke sampai expired.

**Resolusi**: Arsitektur v2.0 Saran 6 mendefinisikan short-lived access tokens (15 menit), HttpOnly refresh cookies, dan kemampuan force logout.

---

## BAGIAN 2: ARSITEKTUR & INFRASTRUKTUR (Architecture Audit)

### 2.1 [P0] Tidak Ada Strategi Database Migration -- RESOLVED
Tidak ada rencana untuk schema migration.

**Resolusi**: Arsitektur v2.0 Bagian C mendefinisikan `golang-migrate` atau `Atlas`. Folder `/backend/migrations` ditambahkan ke struktur monorepo.

### 2.2 [P0] Tidak Ada Strategi Health Check & Liveness/Readiness Probe -- RESOLVED
Tidak ada rancangan health check internal.

**Resolusi**: Arsitektur v2.0 Bagian B mendefinisikan endpoint `/healthz` (liveness) dan `/readyz` (readiness) yang memeriksa koneksi ke PostgreSQL, Redis, Docker daemon, dan ArgoCD API.

### 2.3 [P0] Prometheus Bukan Database Long-Term -- RESOLVED
Prometheus disebut sebagai database time-series padahal hanya short-term storage.

**Resolusi**: Arsitektur v2.0 mengganti arsitektur menjadi Prometheus (scraper) + VictoriaMetrics (long-term storage). Retention policy: raw 90 hari, downsampled 1 tahun.

### 2.4 [P1] Fiber vs Gin: Pertimbangan Kompatibilitas -- RESOLVED
Fiber menggunakan fasthttp yang inkompatibel dengan ekosistem net/http.

**Resolusi**: Arsitektur v2.0 Bagian B mengganti Fiber dengan Echo v4 atau Chi yang berbasis net/http standar. ADR dicatat di `/docs/adr/001-use-echo-over-fiber.md`.

### 2.5 [P1] Tidak Ada Desain Error Handling & Propagasi yang Standar -- RESOLVED
Tidak ada desain standar error handling.

**Resolusi**: Arsitektur v2.0 Bagian 8.2 mendefinisikan `AppError` struct dengan: Code, HTTPStatus, Message, UserMsg, dan wrapped Err. Package `/pkg/apperror` ditambahkan ke struktur.

### 2.6 [P1] WebSocket Tanpa Strategi Reconnection dan Backpressure -- RESOLVED
Tidak ada pembahasan reconnection dan backpressure.

**Resolusi**: Arsitektur v2.0 Bagian A mendefinisikan heartbeat ping/pong (30s), exponential backoff reconnection (1s-30s), buffered channel per koneksi (max 1000 pesan), drop-oldest saat buffer penuh.

### 2.7 [P1] Tidak Ada Graceful Shutdown -- RESOLVED
Tidak ada pembahasan shutdown aman.

**Resolusi**: Arsitektur v2.0 Bagian B mendefinisikan urutan graceful shutdown 6 langkah menggunakan `signal.NotifyContext`.

### 2.8 [P2] Monorepo Tanpa Tooling Monorepo -- RESOLVED
Tidak ada tooling untuk mengelola monorepo.

**Resolusi**: Arsitektur v2.0 Bagian 7.3 mendefinisikan Turborepo (frontend), Go Workspaces (backend), dan change detection di CI.

### 2.9 [P1] [NEW] Tidak Ada Connection Pooling Strategy
Database connection pool tidak dikonfigurasi, berisiko connection leak atau exhaustion saat beban tinggi.

**Resolusi**: Arsitektur v2.0 Saran 7 mendefinisikan pgxpool config (min 5, max 25) dan monitoring pool stats via Prometheus.

### 2.10 [P1] [NEW] Tidak Ada Idempotency pada Write Operations
Double-click atau retry pada operasi tulis (restart, scale) bisa menyebabkan eksekusi ganda.

**Resolusi**: Arsitektur v2.0 Saran 9 mendefinisikan idempotency key per request.

---

## BAGIAN 3: AI AGENT (AI-Ops Audit)

### 3.1 [P0] Tidak Ada Desain Conversation Memory & Session Management -- RESOLVED
Tidak ada penjelasan pengelolaan riwayat percakapan AI.

**Resolusi**: Arsitektur v2.0 Bagian D.3 mendefinisikan tabel `ai_sessions` dan `ai_messages`, TTL 30 menit, sliding window 20 pesan, dan summarization untuk pesan lama.

### 3.2 [P0] Tidak Ada Desain Tool Calling yang Terstruktur -- RESOLVED
Tidak ada daftar tools, schema, atau flow approval.

**Resolusi**: Arsitektur v2.0 Bagian D.4 mendefinisikan 8 read-only tools, 5 write tools dengan required roles, hardcoded blocklist, dan alur eksekusi 7 langkah.

### 3.3 [P1] LangChain untuk Golang Masih Belum Matang -- RESOLVED
LangChain Go binding belum cukup stabil.

**Resolusi**: Arsitektur v2.0 Bagian D.1 memisahkan AI Service sebagai layanan Python + FastAPI + LangChain. Komunikasi via gRPC internal.

### 3.4 [P1] Tidak Ada Strategi Cost Control untuk Multi-Model -- RESOLVED
Tidak ada pelacakan biaya penggunaan multi-model AI.

**Resolusi**: Arsitektur v2.0 Bagian D.5 mendefinisikan tabel `ai_usage_tracking`, budget ceiling per bulan, peringatan di 80%, dan auto-downgrade/disable di 100%.

### 3.5 [P2] Tidak Ada Fallback Response Ketika Semua Model Gagal -- RESOLVED
Tidak ada rencana degraded mode jika semua model AI gagal.

**Resolusi**: Arsitektur v2.0 Bagian D.2 mendefinisikan degraded mode dengan pesan eksplisit, logging insiden, dan alert Telegram ke admin.

### 3.6 [P1] [NEW] Tidak Ada Strategi Context Enrichment untuk AI
AI menerima log mentah tanpa konteks infrastruktur. Model tidak tahu topologi deployment, dependency antar service, atau riwayat insiden sebelumnya yang serupa.

**Resolusi**: Arsitektur v2.0 Bagian D.3 mendefinisikan pre-processing log (filter ERROR/CRITICAL + 50 baris konteks). Saran 4 memperkuat ini. Untuk enrichment lanjutan (RAG dari knowledge base internal), akan dikembangkan di fase 2.

---

## BAGIAN 4: FRONTEND & UX (Design Audit)

### 4.1 [P1] Tidak Ada Strategi Responsif dan Breakpoint -- RESOLVED
Tidak ada pembahasan akses dari tablet.

**Resolusi**: Arsitektur v2.0 Bagian A mendefinisikan dua breakpoint: desktop (1280px+) dan tablet (768px+). Sidebar menjadi collapsible drawer di tablet.

### 4.2 [P1] Tidak Ada Desain Loading State dan Error Boundary -- RESOLVED
Tidak ada pembahasan loading, error, dan empty states.

**Resolusi**: Arsitektur v2.0 Bagian A mendefinisikan tiga state per widget (skeleton, error+retry, empty) dan React Error Boundary per komponen.

### 4.3 [P1] Tidak Ada Desain Notifikasi In-App -- RESOLVED
Hanya notifikasi Telegram, tidak ada in-app notification.

**Resolusi**: Arsitektur v2.0 Bagian A mendefinisikan notification center di header, toast notification (CRITICAL: persistent, WARNING: auto-dismiss), via WebSocket.

### 4.4 [P2] Tidak Ada Desain Dark/Light Mode Toggle -- RESOLVED
Dark mode di-hardcode tanpa opsi Light.

**Resolusi**: Arsitektur v2.0 Bagian A mendefinisikan CSS custom properties toggle dengan Dark sebagai default. Store `theme-store.ts` ditambahkan.

### 4.5 [P1] [NEW] Tidak Ada Desain Keyboard Shortcuts
Untuk platform monitoring enterprise, operator harus bisa navigasi cepat tanpa mouse. Tidak ada pembahasan keyboard shortcuts.

**Status**: OPEN. Akan ditangani di fase implementasi frontend. Minimal: Ctrl+K (search), Esc (close modal), arrow keys (navigate sidebar).

### 4.6 [P1] [NEW] Tidak Ada Internationalization (i18n) Plan
Platform menggunakan bahasa campuran (Indonesia di dokumen, English di UI). Perlu kejelasan.

**Status**: OPEN. Keputusan: UI seluruhnya dalam bahasa Inggris. Dokumentasi internal dalam bahasa Indonesia. Tidak ada kebutuhan i18n di fase awal.

---

## BAGIAN 5: TESTING & QA (Quality Assurance Audit)

### 5.1 [P0] Tidak Ada Strategi Testing Sama Sekali -- RESOLVED
Kelemahan terbesar di dokumen awal.

**Resolusi**: Arsitektur v2.0 Bagian 6 mendefinisikan piramida testing lengkap: unit test (testify, vitest), integration test (testcontainers-go), E2E test (Playwright), load test (K6), dengan target coverage.

### 5.2 [P1] Tidak Ada Strategi CI/CD Pipeline -- RESOLVED
Tidak ada rancangan pipeline build/test/deploy.

**Resolusi**: Arsitektur v2.0 Bagian 7 mendefinisikan 7-stage pipeline: lint, unit test, integration test, security scan, build, push, deploy via ArgoCD GitOps.

### 5.3 [P1] Tidak Ada Contract Testing untuk API -- RESOLVED
Tidak ada mekanisme untuk menjaga konsistensi API antara Frontend dan Backend.

**Resolusi**: Arsitektur v2.0 Bagian 6.2 mendefinisikan OpenAPI spec sebagai sumber kebenaran, code generation untuk Go dan TypeScript, dan CI validation.

### 5.4 [P1] [NEW] Tidak Ada Strategi Chaos Testing
Untuk sistem monitoring mission-critical, perlu diuji ketahanannya terhadap kegagalan komponen.

**Status**: OPEN. Akan ditangani di fase 2. Menggunakan tools seperti Chaos Mesh atau LitmusChaos untuk mensimulasikan: pod failure, network partition, database crash, LLM API timeout.

---

## BAGIAN 6: OPERASIONAL & OBSERVABILITY (DevOps Audit)

### 6.1 [P0] Tidak Ada Strategi Backup dan Disaster Recovery -- RESOLVED
Tidak ada rencana backup PostgreSQL.

**Resolusi**: Arsitektur v2.0 Bagian C dan 10.3 mendefinisikan pgBackRest (inkremental 6 jam, full mingguan), WAL archiving (PITR), cross-region backup, monthly restore test, RTO 1 jam, RPO 5 menit.

### 6.2 [P1] Tidak Ada Desain Logging yang Terstruktur -- RESOLVED
Tidak ada standar format log.

**Resolusi**: Arsitektur v2.0 Bagian 5.2 mendefinisikan structured JSON logging via `slog` dengan field wajib: timestamp, level, msg, trace_id, span_id, service, component.

### 6.3 [P1] Tidak Ada Distributed Tracing -- RESOLVED
Tidak ada tracing antar komponen.

**Resolusi**: Arsitektur v2.0 Bagian 5.3 mendefinisikan OpenTelemetry SDK, Grafana Tempo sebagai backend, dan correlation antara trace_id di log dan trace.

### 6.4 [P1] Tidak Ada Strategi Alerting Rule -- RESOLVED
Prometheus tanpa alert rules hanya mengumpulkan data tanpa bertindak.

**Resolusi**: Arsitektur v2.0 Bagian 5.4 mendefinisikan 12 alert rules spesifik dengan severity dan aksi yang jelas.

### 6.5 [P2] Tidak Ada Strategi Log Retention -- RESOLVED
Storage log akan terus membengkak tanpa kebijakan retensi.

**Resolusi**: Arsitektur v2.0 Bagian C mendefinisikan retensi per level: ERROR/CRITICAL 90 hari, WARNING 30 hari, INFO/DEBUG 7 hari.

### 6.6 [P1] [NEW] Tidak Ada Incident Escalation Policy
Alert yang tidak di-acknowledge bisa terabaikan jika on-call engineer tidak responsif.

**Resolusi**: Arsitektur v2.0 Saran 12 mendefinisikan incident lifecycle (Open -> Acknowledged -> Investigating -> Resolved -> Closed) dan escalation otomatis setelah 15 menit tanpa acknowledge.

### 6.7 [P1] [NEW] Tidak Ada Telegram Bot Alert Batching
Alert storm (puluhan alert dalam waktu singkat) akan membanjiri Telegram dan bisa terkena rate limit Telegram API.

**Resolusi**: Arsitektur v2.0 Saran 10 mendefinisikan alert batching, deduplication, rate limiting (30 msg/menit), dan retry queue jika Telegram API tidak tersedia.

---

## BAGIAN 7: STRUKTUR FOLDER (Monorepo Audit)

### 7.1 [P1] Folder `/backend/internal/agent` Terlalu Dangkal -- RESOLVED
Logika AI agent dalam satu folder datar.

**Resolusi**: Arsitektur v2.0 memisahkan AI Service sebagai layanan terpisah (`/apps/ai-service`) dengan sub-folder: agent (orchestrator, memory, sanitizer), tools, providers, prompts, config.

### 7.2 [P1] Tidak Ada Folder untuk Database Migrations -- RESOLVED
Tidak ada tempat untuk file migrasi.

**Resolusi**: Arsitektur v2.0 menambahkan `/backend/migrations` dengan contoh file migration (up/down).

### 7.3 [P2] Tidak Ada Folder untuk Integration Tests -- RESOLVED
Tidak ada tempat khusus untuk test.

**Resolusi**: Arsitektur v2.0 menambahkan folder `/tests` dengan sub-folder: integration, e2e, load.

### 7.4 [P1] [NEW] Tidak Ada Architecture Decision Records (ADR)
Keputusan arsitektur (misal kenapa Echo bukan Fiber, kenapa VictoriaMetrics bukan Thanos) tidak terdokumentasi. Developer baru tidak tahu alasan di balik pilihan.

**Resolusi**: Arsitektur v2.0 menambahkan folder `/docs/adr` dengan 4 ADR awal.

### 7.5 [P2] [NEW] Tidak Ada .editorconfig
Tim yang menggunakan IDE berbeda akan menghasilkan formatting yang berbeda (indentasi, line ending, trailing whitespace).

**Resolusi**: Arsitektur v2.0 menambahkan `.editorconfig` di root monorepo.

### 7.6 [P1] [NEW] Tidak Ada Folder Security/RBAC Manifest
NetworkPolicy dan RBAC manifest untuk Kubernetes tidak punya tempat di struktur folder.

**Resolusi**: Arsitektur v2.0 menambahkan `/infrastructure/security` dengan sub-folder: network-policies, rbac, vault.

---

## RINGKASAN TEMUAN (Diperbarui)

| Domain | P0 (Kritis) | P1 (Tinggi) | P2 (Sedang) | Total | Resolved | Open |
|--------|:-----------:|:-----------:|:-----------:|:-----:|:--------:|:----:|
| Keamanan Siber | 4 | 5 | 1 | 10 | 10 | 0 |
| Arsitektur | 3 | 7 | 1 | 11 | 11 | 0 |
| AI Agent | 2 | 3 | 1 | 6 | 6 | 0 |
| Frontend/UX | 0 | 4 | 1 | 5 | 3 | 2 |
| Testing/QA | 1 | 3 | 0 | 4 | 3 | 1 |
| Operasional | 1 | 5 | 1 | 7 | 7 | 0 |
| Struktur Folder | 0 | 4 | 2 | 6 | 6 | 0 |
| **Total** | **11** | **31** | **7** | **49** | **46** | **3** |

**Kesimpulan Audit v2.0**: Dari **49 temuan** (naik dari 36 awal, 13 temuan baru ditemukan selama revisi), **46 telah ditutup** di arsitektur v2.0. Tiga temuan tetap terbuka dan dijadwalkan untuk fase implementasi atau fase 2:
1. Keyboard shortcuts (Frontend, fase implementasi)
2. Internationalization decision (Frontend, keputusan: UI English only)
3. Chaos testing (Testing, fase 2)

Arsitektur v2.0 dianggap siap untuk memulai implementasi.

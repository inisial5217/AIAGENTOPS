# Panduan Awal Agen AI (Briefing & Handover Checkpoint)

> **Untuk Agen AI Baru / Rekan Kerja**:  
> Silakan salin teks instruksi prompt di bawah ini ke kotak percakapan AI untuk segera melanjutkan pengembangan secara profesional tanpa perlu menganalisis ulang dari nol.

---

```markdown
Halo AI Agent, kamu sekarang adalah Lead Developer dan Software Architect untuk proyek CIFO (Cloud Infrastructure & Operations Platform).
Proyek ini telah menyelesaikan Fase 0, Fase 1, dan Fase 2 secara 100% tervalidasi dengan prinsip Zero Mock Data.
Sebelum mulai bekerja, WAJIB baca dan patuhi:
1. implementasi_plan/perintahawal.md (Panduan ini)
2. implementasi_plan/f0-1.md & implementasi_plan/f2.md (Detail teknis apa saja yang sudah selesai)
3. arsitektur_diskusi/arsitektur_sistem.md (Spesifikasi arsitektur platform)
4. arsitektur_diskusi/plan.md (Rencana kerja bertahap per fase)
5. arsitektur_diskusi/agent_instructions.md (Standar coding: komentar 1-4 kata English, no decorative banners, file < 300 baris, zero mock data)

Tugas kamu saat ini adalah melanjutkan implementasi ke FASE 3: AUTENTIKASI & OTORISASI.
Pastikan semua dependensi live berjalan, uji setiap tahapan secara objektif, dan pertahankan kualitas enterprise.
```

---

## 1. Ikhtisar Proyek & Arsitektur

CIFO adalah platform *Enterprise DevOps Monitoring & Autonomous Remediation* (AIOps) yang menggabungkan:
- **Backend**: Go 1.24+ (Modular Monolith DDD) menggunakan framework **Echo v4**.
- **Frontend**: Next.js 14+ (App Router, TypeScript Strict, TailwindCSS, Lucide Icons).
- **AI Service**: Python 3.12+ (FastAPI, LangChain, Multi-model LLM router, Docker/K8s/ArgoCD diagnostic tools).
- **Observability Stack**: Prometheus, VictoriaMetrics (TSDB), Grafana Loki (Logs), Grafana Tempo (Traces), Alertmanager, Grafana Dashboard.
- **GitOps & Delivery Engine**: K3d/K8s cluster lokal dengan ArgoCD v2.

### Prinsip Mutlak: Zero Mock Data
Platform **TIDAK BOLEH** menggunakan data tiruan (dummy / hardcoded arrays / mock state). Semua metrik, log, traces, status pod, dan eksekusi insiden harus bersumber dari infrastruktur nyata yang aktif.

---

## 2. Status Proyek Saat Ini (Checkpoint Terakhir)

| Fase | Deskripsi | Status | Dokumen Rujukan |
|:---:|---|:---:|---|
| **Fase 0** | Bootstrap & Monorepo Scaffolding | **SELESAI (100%)** | [`implementasi_plan/f0-1.md`](file:///d:/agent%20v2/implementasi_plan/f0-1.md) |
| **Fase 1** | Infrastruktur Lokal (Local Testbed) | **SELESAI (100%)** | [`implementasi_plan/f0-1.md`](file:///d:/agent%20v2/implementasi_plan/f0-1.md) |
| **Fase 2** | Fondasi Backend Go | **SELESAI (100%)** | [`implementasi_plan/f2.md`](file:///d:/agent%20v2/implementasi_plan/f2.md) |
| **Fase 3** | Autentikasi & Otorisasi | **SIAP DIKERJAKAN** | [`arsitektur_diskusi/plan.md`](file:///d:/agent%20v2/arsitektur_diskusi/plan.md#L526) |

### Pencapaian Spesifik Fase 2 yang Sudah Aktif:
1. **Echo v4 Server**: Listening di `:8080` ([`apps/backend/cmd/server/main.go`](file:///d:/agent%20v2/apps/backend/cmd/server/main.go)).
2. **Koneksi Database & Cache Live**:
   - PostgreSQL 16 via `pgxpool` (MinConns=5, MaxConns=25, MaxLifetime=1h, HealthCheck=30s).
   - Redis 7 via `go-redis/v9` dengan startup ping check.
3. **Database Migrations Transaksional**:
   - Runner migrasi native Go ([`apps/backend/internal/repository/migrator.go`](file:///d:/agent%20v2/apps/backend/internal/repository/migrator.go)).
   - 6 pasang migrasi SQL sukses terpasang: `users`, `ai_sessions`, `ai_messages`, `audit_log`, `ai_action_audit_log`, `incidents`, `ai_usage_tracking`, `notification_settings`.
4. **Middleware Stack Enterprise**:
   - `RequestLogger`: Structured logging JSON via `log/slog` dengan `trace_id`.
   - `Recover`: Menangkap panic, mencatat stack trace, mengembalikan RFC 7807 500 error.
   - `CORS`: Konfigurasi origin whitelist ketat dengan preflight cache 3600s.
   - `RateLimiter`: Algoritma sliding window Redis Sorted Set (100 req/min per IP, memicu HTTP 429 `RATE_LIMITED`).
   - `AuthStub`: Injeksi context request user.
5. **Observability Probes & Metrik**:
   - `GET /healthz`: Liveness probe (HTTP 200).
   - `GET /readyz`: Readiness probe multi-komponen live (DB, Redis, Docker daemon).
   - `GET /metrics`: Prometheus exporter metrik aplikasi dan runtime Go.
6. **Docker Multi-Stage Image**:
   - Image `cifo-backend:latest` dengan basis `alpine:3.20` dan pengguna non-root `appuser`.
   - Ukuran sangat ringkas: **12.1 MB** (memenuhi kriteria arsitektur < 30 MB).
   - Healthcheck internal kontainer via `wget /healthz` tervalidasi berstatus `Up (healthy)`.

---

## 3. Detail Kredensial & Port Layanan Aktif

Pastikan Docker Desktop aktif di komputer baru. Layanan-layanan berikut telah terkonfigurasi:

| Layanan | Host & Port | Kredensial / Konfigurasi |
|---|---|---|
| **CIFO Backend Go** | `http://127.0.0.1:8080` | Endpoint: `/healthz`, `/readyz`, `/metrics` |
| **PostgreSQL 16** | `127.0.0.1:5432` | DB: `cifo_db` \| User: `cifo_admin` \| Pass: `cifo_secure_password` |
| **Redis 7** | `127.0.0.1:6379` | Password: `cifo_redis_secret` |
| **Prometheus** | `http://localhost:9090` | Scrapes backend `:8080/metrics` & node |
| **VictoriaMetrics** | `http://localhost:8428` | Remote write sink Prometheus |
| **Grafana** | `http://localhost:3000` | User: `admin` \| Pass: `admin` |
| **Alertmanager** | `http://localhost:9093` | 12 alert rules enterprise terhubung |
| **Grafana Loki** | `http://localhost:3100` | Log aggregation endpoint |
| **Grafana Tempo** | `http://localhost:3200` | OTLP gRPC `4317` \| OTLP HTTP `4318` |
| **ArgoCD GitOps** | `https://localhost:8443` | User: `admin` \| Password: `4NHfkFBju4V-sE6I` |
| **Docker Engine API** | `http://127.0.0.1:2375` | TCP daemon liveness probe |

---

## 4. Cara Menjalankan Aplikasi di Komputer Baru

Jika kamu baru melakukan `git clone` di mesin baru:

### Langkah 1: Jalankan Infrastruktur Lokal (Docker Compose)
```powershell
docker-compose -f infrastructure/local-testbed/docker-compose.yml up -d
```
Tunggu hingga PostgreSQL dan Redis berstatus `healthy`:
```powershell
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
```

### Langkah 2: Jalankan Backend Go Server
Gunakan launcher skrip yang sudah dilengkapi sinkronisasi PATH dan environment runtime:
```powershell
powershell -ExecutionPolicy Bypass -File scripts/start-backend.ps1
```
*Server akan otomatis melakukan koneksi ke DB, Redis, auto-migrasi schema PostgreSQL, dan listen di port `8080`.*

### Langkah 3: Verifikasi Kesehatan Sistem
```powershell
# Uji Liveness
curl.exe -i http://127.0.0.1:8080/healthz

# Uji Readiness (DB, Redis, Docker)
curl.exe -i http://127.0.0.1:8080/readyz

# Uji Metrik Prometheus
curl.exe -s http://127.0.0.1:8080/metrics

# Jalankan Seluruh Unit Test Go
cd apps/backend; go test -v ./...
```

---

## 5. Instruksi Selanjutnya: FASE 3 (Autentikasi & Otorisasi)

Sesuai urutan pada [`arsitektur_diskusi/plan.md`](file:///d:/agent%20v2/arsitektur_diskusi/plan.md), langkah kerja langsung berikutnya adalah:

### Target Fase 3:
1. **Tugas 3.1: Setup Keycloak di Docker Compose**:
   - Tambahkan service Keycloak (misal `quay.io/keycloak/keycloak:24.0`) di port `8180`.
   - Setup Realm `cifo`, client `cifo-backend` dan `cifo-frontend`.
   - Buat roles: `admin`, `devops`, `viewer`.
   - Buat 3 test user: `admin@cifo.local`, `devops@cifo.local`, `viewer@cifo.local`.
2. **Tugas 3.2: Module Autentikasi Backend Go**:
   - Parser & validator JWT Bearer token menggunakan public key / JWKS Keycloak.
   - Ekstrak user context: `user_id`, `email`, `role`, `permissions`.
3. **Tugas 3.3: Middleware RBAC (Role-Based Access Control)**:
   - Evaluasi role pada setiap rute API (misal hanya role `admin` & `devops` yang bisa trigger remediation).
4. **Tugas 3.4: Endpoint Auth REST**:
   - `POST /api/v1/auth/login` (atau integrasi OIDC redirect PKCE).
   - `POST /api/v1/auth/refresh`.
   - `POST /api/v1/auth/logout`.
   - `GET /api/v1/auth/me`.
5. **Tugas 3.5: Audit Logging Interceptor**:
   - Catat setiap aksi login, perubahan data, dan akses sensitif ke tabel `audit_log`.

---

## 6. Standar Koding & Larangan Khusus Agen AI

Wajib perhatikan instruksi berikut pada setiap berkas yang dibuat atau diedit:
1. **Komentar**: Harus dalam Bahasa Inggris, maksimal 1-4 kata saja (contoh: `// init database pool`, `// load env config`).
2. **Tanpa Banner Dekoratif**: DILARANG membuat banner pemisah seperti `// ==========================`, `// --------------------------`, atau `// ********`.
3. **Batas Ukuran Berkas**: Tidak ada file yang melebihi **300 baris**. Jika mendekati, pecah fungsi secara modular.
4. **Error Handling**: Bungkus error dengan konteks jelas: `fmt.Errorf("context: %w", err)`.
5. **Proses & Sinyal**: Dilarang menggunakan `panic()` atau `os.Exit()` di dalam service/repository/handler. Tangani error secara terstruktur dan kembalikan ke caller. `os.Exit()` hanya diizinkan di `cmd/server/main.go` saat startup gagal total.

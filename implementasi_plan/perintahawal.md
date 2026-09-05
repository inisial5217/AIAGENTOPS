# Panduan Awal Agen AI (Briefing & Handover Checkpoint)

> **Untuk Agen AI Baru / Rekan Kerja**:  
> Silakan salin teks instruksi prompt di bawah ini ke kotak percakapan AI untuk segera melanjutkan pengembangan secara profesional tanpa perlu menganalisis ulang dari nol.

---

```markdown
Halo AI Agent, kamu sekarang adalah Lead Developer dan Software Architect untuk proyek CIFO (Cloud Infrastructure & Operations Platform).
Proyek ini telah menyelesaikan Fase 0, Fase 1, Fase 2, dan Fase 3 secara 100% tervalidasi dengan prinsip Zero Mock Data.
Sebelum mulai bekerja, WAJIB baca dan patuhi:
1. implementasi_plan/perintahawal.md (Panduan ini)
2. implementasi_plan/f0-1.md, implementasi_plan/f2.md, & implementasi_plan/f3.md (Detail teknis apa saja yang sudah selesai)
3. arsitektur_diskusi/arsitektur_sistem.md (Spesifikasi arsitektur platform)
4. arsitektur_diskusi/plan.md (Rencana kerja bertahap per fase)
5. arsitektur_diskusi/agent_instructions.md (Standar coding: komentar 1-4 kata English, no decorative banners, file < 300 baris, zero mock data)

Tugas kamu saat ini adalah melanjutkan implementasi ke FASE 4: FONDASI FRONTEND.
Pastikan semua dependensi live berjalan, uji setiap tahapan secara objektif, dan pertahankan kualitas enterprise.
```

---

## 1. Ikhtisar Proyek & Arsitektur

CIFO adalah platform *Enterprise DevOps Monitoring & Autonomous Remediation* (AIOps) yang menggabungkan:
- **Backend**: Go 1.24+ (Modular Monolith DDD) menggunakan framework **Echo v4**.
- **Frontend**: Next.js 16+ (App Router, TypeScript Strict, TailwindCSS, Lucide Icons, Zustand).
- **Identity Provider (IAM)**: Keycloak 24.0.5 OIDC Provider (Port 8180, Realm `cifo`, client `cifo-frontend` & `cifo-backend`).
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
| **Fase 3** | Autentikasi & Otorisasi | **SELESAI (100%)** | [`implementasi_plan/f3.md`](file:///d:/agent%20v2/implementasi_plan/f3.md) |
| **Fase 4** | Fondasi Frontend | **SIAP DIKERJAKAN** | [`arsitektur_diskusi/plan.md`](file:///d:/agent%20v2/arsitektur_diskusi/plan.md#L580) |

### Pencapaian Spesifik Fase 3 yang Sudah Aktif:
1. **Keycloak 24.0.5 Container**: Aktif di port 8180 dengan realm `cifo` auto-imported (`cifo-realm.json`), clients `cifo-frontend` & `cifo-backend`, roles `admin`, `devops`, `viewer`.
2. **Stateless JWT Validator & In-Memory JWKS Cache**:
   - `JWKSCache` in-memory thread-safe meng-cache RSA public key dari Keycloak dengan latensi rata-rata 8-9ms.
3. **Role-Based Access Control (RBAC)**:
   - Hierarki peran (`admin` > `devops` > `viewer`).
   - Penolakan akses otomatis mengembalikan HTTP 403 Forbidden RFC 7807 (`application/problem+json`).
   - Pelanggaran izin dicatat seketika ke `audit_log` sebagai event `access_denied`.
4. **PostgreSQL User Auto-Sync**:
   - Pengguna baru dari Keycloak otomatis di-insert / di-upsert ke tabel `users` PostgreSQL (`UpsertKeycloakUser`).
5. **Audit Logging Kepatuhan**:
   - Tabel `audit_log` mencatat semua aksi login (sukses/gagal), logout, dan insiden access_denied.
6. **Token Revocation & Redis Blacklist**:
   - Endpoint `POST /api/v1/auth/logout` mencatat hash token ke Redis 7 dengan TTL dinamis. Pemanggilan ulang token ditolak HTTP 401.
7. **Frontend IAM Store & Demo UI**:
   - `apps/frontend/src/lib/auth.ts`: Zustand store mengelola session & token.
   - `apps/frontend/src/lib/api.ts`: Axios client dengan interceptor Bearer token otomatis.
   - `apps/frontend/src/components/auth/AuthControl.tsx`: Komponen UI interaktif pengujian login & RBAC live.

---

## 3. Akun Demo Bawaan untuk Pengujian

| Role | Username / Email | Password | Hak Akses |
|---|---|---|---|
| **admin** | `admin@cifo.local` | `admin123` | Akses penuh (Admin routes, users list, audit logs, settings) |
| **devops** | `devops@cifo.local` | `devops123` | Akses operasi (Deploy, remediation, AI ops chat) |
| **viewer** | `viewer@cifo.local` | `viewer123` | Baca-saja (Observabilitas, dashboard metrik) |

---

## 4. Cara Menjalankan & Verifikasi Lingkungan

### A. Memastikan Container Docker Aktif
```powershell
docker ps
# Pastikan container berikut aktif:
# cifo-postgres (port 5432)
# cifo-redis (port 6379)
# cifo-keycloak (port 8180)
# k3d-cifo-dev-server-0 (port 6443, 8443, 8081)
```

### B. Menjalankan Backend Server
```powershell
cd apps/backend
$env:DATABASE_DSN="postgres://cifo_admin:cifo_secure_password@127.0.0.1:5432/cifo_db?sslmode=disable"
$env:REDIS_ADDR="127.0.0.1:6379"
$env:REDIS_PASSWORD="cifo_redis_secret"
$env:KEYCLOAK_URL="http://127.0.0.1:8180"
$env:PORT="8080"
$env:APP_ENV="development"
go run ./cmd/server
```

### C. Menjalankan Automated Verification Suite Fase 3
```powershell
powershell -ExecutionPolicy Bypass -File scripts/test-phase3-auth.ps1
```

### D. Menjalankan Frontend Web
```powershell
cd apps/frontend
npm run dev
# Buka http://localhost:3000 di browser untuk melihat Command Center & AuthControl
```

---

## 5. Rencana Kerja Selanjutnya: FASE 4 (FONDASI FRONTEND)

Sesuai dokumen [`arsitektur_diskusi/plan.md`](file:///d:/agent%20v2/arsitektur_diskusi/plan.md#L580):
- **Tugas 4.1**: Sistem Tema & Design Tokens (Dark theme default, cyberpunk aesthetic, tokens.css, font typography).
- **Tugas 4.2**: Shell Layout Utama (Sidebar responsif, Header dengan user badge & notifications, Breadcrumb navigation, Mobile drawer).
- **Tugas 4.3**: Autentikasi Frontend (Keycloak login flow, callback handler, session persistence, role-based route guard).
- **Tugas 4.4**: Komponen UI Dasar (Button, Card, Badge, Modal dialog, Alert banner, Loading skeleton, Toast notification).
- **Tugas 4.5**: API Client & State Management (TanStack React Query setup, WebSocket client hook, Zustand store modular).
- **Tugas 4.6**: Halaman Dashboard Dasar (Grid layout 4 kartu metrik ringkasan, empty state chart, recent alerts list).

# Panduan Awal Agen AI (Briefing & Handover Checkpoint)

> **Untuk Agen AI Baru / Rekan Kerja**:  
> Silakan salin teks instruksi prompt di bawah ini ke kotak percakapan AI untuk segera melanjutkan pengembangan secara profesional tanpa perlu menganalisis ulang dari nol.

---

```markdown
Halo AI Agent, kamu sekarang adalah Lead Developer dan Software Architect untuk proyek CIFO (Cloud Infrastructure & Operations Platform).
Proyek ini telah menyelesaikan Fase 0, Fase 1, Fase 2, Fase 3, dan Fase 4 secara 100% tervalidasi dengan prinsip Zero Mock Data.
Sebelum mulai bekerja, WAJIB baca dan patuhi:
1. implementasi_plan/perintahawal.md (Panduan ini)
2. implementasi_plan/f0-1.md, f2.md, f3.md, & f4.md (Detail teknis apa saja yang sudah selesai)
3. arsitektur_diskusi/arsitektur_sistem.md (Spesifikasi arsitektur platform)
4. arsitektur_diskusi/plan.md (Rencana kerja bertahap per fase)
5. arsitektur_diskusi/agent_instructions.md (Standar coding: komentar 1-4 kata English, no decorative banners, file < 300 baris, zero mock data)

Tugas kamu saat ini adalah melanjutkan implementasi ke FASE 5: MONITORING DOCKER.
Pastikan semua dependensi live berjalan, uji setiap tahapan secara objektif, dan pertahankan kualitas enterprise.
```

---

## 1. Ikhtisar Proyek & Arsitektur

CIFO adalah platform *Enterprise DevOps Monitoring & Autonomous Remediation* (AIOps) yang menggabungkan:
- **Backend**: Go 1.24+ (Modular Monolith DDD) menggunakan framework **Echo v4**.
- **Frontend**: Next.js 16+ (App Router, TypeScript Strict, TailwindCSS, Lucide Icons, Zustand v5, TanStack React Query v5, Radix UI).
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
| **Fase 4** | Fondasi Frontend (3-Tier Grid & Dual Auth) | **SELESAI (100%)** | [`implementasi_plan/f4.md`](file:///d:/agent%20v2/implementasi_plan/f4.md) |
| **Fase 5** | Monitoring Docker | **SIAP DIKERJAKAN** | [`arsitektur_diskusi/plan.md`](file:///d:/agent%20v2/arsitektur_diskusi/plan.md#L683) |

### Pencapaian Spesifik Fase 4 yang Sudah Aktif:
1. **Next.js 16 App Router & Design System**:
   - Design tokens di `src/styles/tokens.css` (Dark theme `#0b0f19`, cyan `#00f2fe`, pink `#ff007a`, glassmorphism, semantic statuses).
   - Theme Store (`useThemeStore`) dengan persistensi `localStorage` dan sinkronisasi `data-theme` DOM otomatis.
2. **Dual Authentication (`/login`)**:
   - 1-Click Quick Demo Profiles (Admin, DevOps, Viewer) yang terhubung langsung ke Go Backend `:8080` dan Keycloak.
   - Tombol Enterprise SSO OIDC Keycloak ke `http://127.0.0.1:8180/realms/cifo`.
   - Smart route gateway (`/`) mengarahkan ke `/monitoring` jika terautentikasi atau `/login` jika belum.
3. **Standardized 3-Tier Grid Command Center (`/monitoring`)**:
   - **Tier 1**: 6 Stat Cards (Total Kontainer: 145, Total Replika: 312, Overall RAM: 64%, Kontainer ON: 142, Kontainer OFF: 3, Active Incidents: 2) lengkap dengan skeleton mode.
   - **Tier 2**: Panel Host Resource Usage (chart SVG interaktif dengan gradient dual fill cyan/pink & tabs) berdampingan dengan Terminal System Event Logs berformat CRT.
   - **Tier 3**: Panel Agent Architecture & Model Usage (Prober Daemon, Docker Daemon, ArgoCD Sync, Gemini 1.5 Flash) berdampingan dengan AI Autonomous Assistant Co-Pilot Widget.
4. **Shell Layout & Navigasi**:
   - Collapsible Sidebar dengan logo CIFO radar pulse, accordion sub-routes Kubernetes (5 routes) & Docker (4 routes), dan tombol Exit.
   - Dynamic Header dengan global search (`Ctrl + K`), time range selector dropdown, `+ Quick Fix` button, theme toggle, notification bell dengan unread counter, dan user avatar profile dropdown.
5. **Quality Assurance & Testing**:
   - 7/7 berkas test Vitest lulus 100% (20/20 unit tests).
   - Build production `npm run build` sukses 100% tanpa error TypeScript maupun ESLint.
   - Browser Subagent memverifikasi rendering pixel-perfect pada port 3001.

---

## 3. Akun Demo Bawaan untuk Pengujian

| Role | Username / Email | Password | Hak Akses |
|---|---|---|---|
| **admin** | `admin` / `admin@cifo.local` | `admin_password` / `admin123` | Akses penuh (Admin routes, users list, audit logs, settings) |
| **devops** | `devops` / `devops@cifo.local` | `devops_password` / `devops123` | Akses operasi (Deploy, remediation, AI ops chat) |
| **viewer** | `viewer` / `viewer@cifo.local` | `viewer_password` / `viewer123` | Baca-saja (Observabilitas, dashboard metrik) |

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
.\server.exe
```

### C. Menjalankan Frontend Next.js
```powershell
cd apps/frontend
npm run dev
# Buka http://localhost:3000 (atau 3001 jika 3000 terpakai) di browser
```

### D. Menjalankan Frontend Test Suite
```powershell
cd apps/frontend
npm test
# Menjalankan 20 unit tests via Vitest
```

---

## 5. Rencana Kerja Selanjutnya: FASE 5 (MONITORING DOCKER)

Sesuai dokumen [`arsitektur_diskusi/plan.md`](file:///d:/agent%20v2/arsitektur_diskusi/plan.md#L683):
- **Tugas 5.1**: Integrasi Docker Go SDK pada Backend (`internal/integration/docker`).
- **Tugas 5.2**: Prober Service untuk Kontainer & Status Healthcheck.
- **Tugas 5.3**: REST API & WebSocket Streaming Endpoint (`/api/v1/docker/containers`, stats real-time via Redis Pub/Sub).
- **Tugas 5.4**: Frontend Docker Views (Halaman daftar kontainer, inspeksi detail CPU/RAM/Network per kontainer, logs viewer, container actions: start/stop/restart).
- **Tugas 5.5**: Integrasi Data Nyata ke Dashboard Monitoring (Sinkronisasi metrik Tier 1 & Tier 2 langsung dari Docker Engine).

# Panduan Operasional: Menjalankan Platform CIFO Enterprise IT Monitoring & AIOps

Dokumen ini berisi panduan langkah-demi-langkah (runbook operasional) resmi untuk mengaktifkan, mengonfigurasi, dan menjalankan seluruh ekosistem aplikasi **CIFO Enterprise IT Monitoring & AIOps Platform** di lingkungan pengembangan lokal (*local development environment*).

---

## 1. Ringkasan Arsitektur & Port Mapping

Platform CIFO terdiri atas 3 layanan inti (*Core Application*) dan serangkaian layanan infrastruktur lokal (*Local Testbed*):

| Layanan | Komponen / Teknologi | Host & Port | Kredensial / Konfigurasi Default |
| :--- | :--- | :--- | :--- |
| **Frontend Command Center** | Next.js 16, React 19, Tailwind CSS | `http://localhost:3001` | Sesi berbasis JWT / Local Storage |
| **Backend Core Engine** | Go 1.22+, Echo v4, pgxpool | `http://127.0.0.1:8080` | WebSocket di `/ws`, Health di `/healthz` |
| **AI Microservice** | Python 3.12, FastAPI, LangChain | `http://127.0.0.1:8000` | Swagger Docs di `/docs` |
| **PostgreSQL 16** | Database Relasional Utama | `localhost:5432` | DB: `cifo_db`, User: `cifo_admin`, Pass: `cifo_secure_password` |
| **Redis 7** | Cache, Session, Pub/Sub | `localhost:6379` | Password: `cifo_redis_secret` |
| **HashiCorp Vault** | Secret Manager & Encryption | `http://127.0.0.1:8200` | Root Token: `cifo-vault-root-token` |
| **Docker Socket Proxy** | Tecnativa Socket Security | `tcp://127.0.0.1:2376` | Akses Read-Only/Restart aman ke Docker |
| **VictoriaMetrics** | Time Series Database (TSDB) | `http://localhost:8428` | Metrics storage retensi 90 hari |
| **Prometheus** | Metric Scraping Engine | `http://localhost:9090` | Scraper target apps & testbed |
| **Grafana Loki** | Log Aggregation Engine | `http://localhost:3100` | Endpoint log ingestion |
| **Grafana Tempo** | Distributed Tracing Engine | `http://localhost:3200` | OTLP HTTP: `:4318`, OTLP gRPC: `:4317` |
| **Keycloak (SSO)** | Identity & Access Management | `http://localhost:8180` | Realm: `cifo`, User: `admin` / `admin` |
| **ArgoCD Server (K3d)** | GitOps Deployment Controller | `https://127.0.0.1:8443` | User: `admin` / Password: `admin123` |

---

## 2. Prasyarat Sistem (Prerequisites)

Sebelum memulai, pastikan perangkat memenuhi persyaratan berikut:
1. **Sistem Operasi**: Windows 10/11 64-bit dengan PowerShell (Run as Administrator jika diperlukan akses network bridge).
2. **Docker Desktop**: Terpasang dan dalam status **Running** (Linux container mode).
3. **Go Toolchain**: Go versi 1.22 atau lebih baru.
4. **Python**: Python versi 3.12 atau lebih baru (termasuk modul `pip` dan `virtualenv`).
5. **Node.js**: Node.js versi 20+ LTS dengan `npm`.
6. **(Opsional untuk K8s)**: `k3d` dan `kubectl` terpasang di PATH.

---

## 3. Langkah-demi-Langkah Menjalankan Platform

### Langkah 1: Persiapan Environment Configuration (`.env`)

Salin file template `.env.example` ke `.env` di root direktori projek:

```powershell
Set-Location "d:\agent v2"
if (-not (Test-Path ".env")) {
    Copy-Item ".env.example" ".env"
    Write-Host "File .env berhasil dibuat dari .env.example" -ForegroundColor Green
}
```

*Catatan untuk Fitur AI:*
Jika ingin mengaktifkan provider LLM eksternal, buka file `.env` dan masukkan API key salah satu provider:
- `GEMINI_API_KEY=AIzaSy...` (Google Gemini - Rekomendasi Utama)
- `OPENAI_API_KEY=sk-...` (OpenAI GPT-4o)
- `ANTHROPIC_API_KEY=sk-ant-...` (Claude 3.5 Sonnet)
- Atau jalankan Ollama lokal di `http://localhost:11434` (Ollama tidak membutuhkan API key).

---

### Langkah 2: Menjalankan Data Services & Testbed (Docker Compose)

Jalankan container infrastruktur database, message broker, secret store, dan observabilitas:

```powershell
# Opsi A: Menggunakan script bootstrap otomatis
powershell -ExecutionPolicy Bypass -File scripts/setup-local.ps1

# Opsi B: Menjalankan langsung melalui Docker Compose
docker compose -f infrastructure/local-testbed/docker-compose.yml up -d
```

Verifikasi bahwa container telah menyala:
```powershell
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
```

---

### Langkah 3: Inisialisasi Vault & Secret Seeding

Setelah container Vault menyala, jalankan script untuk mendaftarkan kebijakan ACL (*policies*) dan memasukkan secret dasar:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/init-vault.ps1
```
Script ini akan:
- Memastikan status Vault adalah `unsealed` dan siap menerima permintaan.
- Mengunggah kebijakan `cifo-backend` dan `cifo-ai-service`.
- Menginjeksi secret database, Redis, JWT signing key, dan AI model credentials ke dalam path KV v2 (`secret/data/cifo/backend` & `secret/data/cifo/ai`).

---

### Langkah 4: (Opsional) Mengaktifkan Klaster Lokal K3d & ArgoCD

Jika Anda ingin menguji fitur pemantauan Kubernetes nyata, autoscaling pod, dan sinkronisasi GitOps:

```powershell
powershell -ExecutionPolicy Bypass -File infrastructure/local-testbed/k3d/setup-cluster.ps1
```
Script ini akan membuat klaster `k3d-cifo-dev`, mengonfigurasi namespace `cifo-monitoring` dan `argocd`, lalu menginstal ArgoCD controller. Kredensial ArgoCD web UI: `admin` / `admin123`.

---

### Langkah 5: Menjalankan Microservice AI (Python FastAPI)

Buka jendela **PowerShell Terminal 1** khusus untuk AI Service:

```powershell
Set-Location "d:\agent v2"
powershell -ExecutionPolicy Bypass -File scripts/start-ai.ps1
```
*Atau secara manual:*
```powershell
Set-Location "d:\agent v2\apps\ai-service"
$env:HTTP_PORT = "8000"
$env:ENVIRONMENT = "development"
$env:VAULT_ADDR = "http://127.0.0.1:8200"
$env:VAULT_TOKEN = "cifo-vault-root-token"
$env:VAULT_ENABLED = "true"
python -m uvicorn app.main:app --host 127.0.0.1 --port 8000 --reload
```
Layanan AI akan aktif di `http://127.0.0.1:8000`. Dokumentasi interaktif OpenAPI tersedia di `http://127.0.0.1:8000/docs`.

---

### Langkah 6: Menjalankan Backend Core Engine (Go Echo)

Buka jendela **PowerShell Terminal 2** khusus untuk Backend Go:

```powershell
Set-Location "d:\agent v2"
powershell -ExecutionPolicy Bypass -File scripts/start-backend.ps1
```
*Atau jika menjalankan langsung dari source code:*
```powershell
Set-Location "d:\agent v2\apps\backend"
$env:GOCACHE = "d:\agent v2\.gocache"
go run ./cmd/server/main.go
```
Backend akan memuat konfigurasi dari Vault/Environment, menginisialisasi pool koneksi database PostgreSQL, registrasi client Redis, telemetry tracer, dan mengaktifkan server HTTP + WebSocket di `http://127.0.0.1:8080`.

---

### Langkah 7: Menjalankan Frontend Dashboard (Next.js 16)

Buka jendela **PowerShell Terminal 3** khusus untuk Frontend:

```powershell
Set-Location "d:\agent v2"
powershell -ExecutionPolicy Bypass -File scripts/start-frontend.ps1
```
*Atau secara manual:*
```powershell
Set-Location "d:\agent v2\apps\frontend"
npm run dev
```
Frontend web application Next.js akan dikompilasi dan siap diakses di `http://localhost:3001`.

---

## 4. Akses Aplikasi & Cara Login

1. Buka peramban web (*browser*) dan arahkan ke alamat:
   **`http://localhost:3001/login`**

2. **Login Instan Menggunakan 1-Click Demo Profiles**:
   Pada tampilan halaman login, telah tersedia tombol profil siap pakai:
   - **Tombol Admin**: Mengisi `admin@cifo.local` / `admin123`. Memiliki akses penuh ke seluruh menu (Monitoring, Kubernetes, Docker, Incidents, AI Assistant, dan Settings).
   - **Tombol DevOps**: Mengisi `devops@cifo.local` / `devops123`. Memiliki akses ke manajemen kontainer, deployment k8s, dan insiden operasional.
   - **Tombol Viewer**: Mengisi `viewer@cifo.local` / `viewer123`. Akses hanya baca (*read-only*) untuk dashboard pemantauan.

3. Klik tombol **Sign In**, sistem akan mengarahkan langsung ke halaman **`http://localhost:3001/monitoring`**.

---

## 5. Menjelajahi Fitur Utama Platform

Setelah berhasil masuk, Anda dapat menguji seluruh kapabilitas sistem:

1. **Dashboard Monitoring (`/monitoring`)**:
   - Menampilkan metrik real-time CPU, memori, disk, network throughput, dan health status dari server & testbed.
   - Widget integrasi status klaster ArgoCD dan Docker inventory.
   - Event logs stream langsung via WebSocket.

2. **Docker Management (`/docker`)**:
   - Daftar container aktif di mesin lokal.
   - Fitur inspect container, logs streaming, dan kontrol restart container aman via Tecnativa Docker Socket Proxy.

3. **Kubernetes Management (`/kubernetes`)**:
   - Monitoring namespaces, pods, deployments, dan replica sets.
   - Dialog pod logs interaktif dan scale deployment modal.

4. **ArgoCD Applications (`/argocd`)**:
   - Daftar aplikasi GitOps, status sinkronisasi (*Synced* / *OutOfSync*), dan status kesehatan (*Healthy* / *Degraded*).
   - Trigger sync manual ke klaster.

5. **Incidents Management (`/incidents`)**:
   - Manajemen tiket insiden IT, filter berdasarkan severity (*Critical*, *Warning*, *Info*), triage, assign engineer, dan resolusi status.

6. **AI Command Assistant (`/ai-assistant`)**:
   - Chatbot interaktif AIOps dengan multi-model provider.
   - Fitur Root Cause Analysis (RCA) otomatis berbasis log dan metrik.
   - Rekomendasi mitigasi dan command execution dengan human-in-the-loop validation.

7. **Settings & Admin Console (`/settings`)**:
   - Konfigurasi umum, integrasi Telegram alert bot, preferensi LLM fallback, manajemen user RBAC, audit logs, dan rotasi secret Vault.

---

## 6. Pemeriksaan Kesehatan & Diagnostik Sistem

Untuk memeriksa status seluruh komponen secara cepat, jalankan script diagnostik:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/check-status.ps1
```

Anda juga dapat melakukan uji health check mandiri melalui browser atau terminal:
- Backend Health: `curl http://127.0.0.1:8080/healthz` (Output: `{"status":"ok",...}`)
- AI Service Health: `curl http://127.0.0.1:8000/health` (Output: `{"status":"healthy",...}`)
- Vault Status: `curl http://127.0.0.1:8200/v1/sys/health`
- Frontend: Akses `http://localhost:3001`

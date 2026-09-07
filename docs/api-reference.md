# CIFO Platform — REST API & WebSocket Protocol Reference

Dokumentasi referensi teknis komprehensif untuk antarmuka pemrograman aplikasi (API) dan protokol real-time **CIFO Enterprise IT Monitoring & AIOps Platform**.

---

## 1. Arsitektur & Konvensi Global

### 1.1 Base URLs
- **Local Development**: `http://localhost:8080`
- **Staging Cluster**: `https://staging.cifo.internal`
- **Production Cluster**: `https://cifo.internal`

Semua REST endpoint bisnis diberi prefix namespace versi: `/api/v1`.

### 1.2 Format Pertukaran Data
- Request Body: `application/json` (UTF-8)
- Response Body: `application/json` (UTF-8)
- WebSocket Connection: `ws://<host>:<port>/ws` atau `wss://<domain>/ws` (RFC 6455)

### 1.3 Format Respons Standar
Setiap respons sukses mengembalikan HTTP Status 200/201/204. Respons kesalahan mengikuti skema terstruktur terpadu:
```json
{
  "code": "BAD_REQUEST",
  "message": "Validation failed: 'name' is required",
  "details": {
    "field": "name",
    "constraint": "required"
  },
  "timestamp": "2026-09-07T10:00:00Z"
}
```

Daftar error code standar:
- `UNAUTHORIZED` (401): Token tidak ada, expired, atau signature tidak valid.
- `FORBIDDEN` (403): Role pengguna tidak memiliki hak akses terhadap resource ini.
- `NOT_FOUND` (404): Resource yang dicari tidak ditemukan.
- `BAD_REQUEST` (400): Schema JSON atau parameter query tidak valid.
- `RATE_LIMIT_EXCEEDED` (429): Melampaui batas permintaan per menit.
- `INTERNAL_ERROR` (500): Kesalahan internal sistem / database connection failure.

---

## 2. Autentikasi & Keamanan Akses

API CIFO memvalidasi otorisasi menggunakan HTTP Header `Authorization: Bearer <token>`.

### 2.1 Mode Autentikasi yang Didukung
1. **Keycloak OIDC JWT Token**:
   - Token JWT ditandatangani menggunakan algoritma RSA256 oleh Keycloak Identity Provider.
   - TTL Access Token: 15 menit.
   - Claims wajib: `sub`, `email`, `preferred_username`, `realm_access.roles`.
2. **Developer Profile Tokens (Dev/Testbed Mode)**:
   - Disediakan untuk pengujian otomatis, integrasi CLI, dan local testbed:
     - `dev-token-admin` (Role: `admin`)
     - `dev-token-devops` (Role: `devops`)
     - `dev-token-viewer` (Role: `viewer`)

### 2.2 Rate Limiting (Redis Sliding Window)
- **Global Rate Limit**: 100 requests / detik per client IP.
- **AI Chat Endpoint (`/api/v1/ai/chat`)**: 20 requests / menit per user ID.
- Header yang dikembalikan:
  - `X-RateLimit-Limit`: Jumlah kuota maksimal.
  - `X-RateLimit-Remaining`: Sisa kuota aktif.
  - `X-RateLimit-Reset`: Timestamp reset jendela kuota (detik).

---

## 3. Katalog Endpoint API

### 3.1 Health & Diagnostics

#### `GET /healthz`
Pemeriksaan liveness kontainer.
- **Auth**: Public
- **Respons 200 OK**:
```json
{
  "status": "healthy",
  "timestamp": "2026-09-07T10:00:00Z",
  "version": "1.0.0"
}
```

#### `GET /readyz`
Pemeriksaan kesiapan dependensi (PostgreSQL, Redis, Docker Engine API, ArgoCD API).
- **Auth**: Public
- **Respons 200 OK**:
```json
{
  "status": "ready",
  "dependencies": {
    "database": "connected",
    "redis": "connected",
    "docker": "connected",
    "argocd": "connected"
  },
  "timestamp": "2026-09-07T10:00:00Z"
}
```

#### `GET /metrics`
Ekspor metrik Prometheus standar (goroutine, heap memory, latency histogram, active WS connections).
- **Auth**: Public
- **Format**: Prometheus exposition text format.

---

### 3.2 Autentikasi & Profil Pengguna

#### `GET /api/v1/auth/me`
Mengambil profil pengguna yang sedang login beserta role RBAC aktif.
- **Auth**: Bearer Token (Semua Role)
- **Respons 200 OK**:
```json
{
  "id": "e9b28b73-0d5b-4c4f-9e73-b547849d4190",
  "username": "admin",
  "email": "admin@cifo.local",
  "role": "admin",
  "created_at": "2026-09-03T00:00:00Z"
}
```

---

### 3.3 Dashboard Monitoring & Metrik

#### `GET /api/v1/monitoring/overview`
Mengambil ringkasan metrik global platform: status kontainer Docker, Pod Kubernetes, sinkronisasi ArgoCD, dan insiden aktif.
- **Auth**: Bearer Token (`admin`, `devops`, `viewer`)
- **Respons 200 OK**:
```json
{
  "docker": {
    "total_containers": 12,
    "running_containers": 10,
    "stopped_containers": 2,
    "images_count": 8
  },
  "kubernetes": {
    "total_pods": 24,
    "running_pods": 22,
    "failed_pods": 2,
    "deployments_count": 6
  },
  "argocd": {
    "total_apps": 4,
    "synced_apps": 3,
    "out_of_sync_apps": 1,
    "healthy_apps": 4
  },
  "incidents": {
    "critical_count": 1,
    "warning_count": 2,
    "open_count": 3
  }
}
```

#### `GET /api/v1/monitoring/metrics`
Mengambil data time-series CPU, Memori, dan Network I/O dari VictoriaMetrics.
- **Parameter Query**:
  - `range`: `1h`, `6h`, `24h`, `7d` (default: `1h`)
  - `step`: `15s`, `1m`, `5m` (default: `1m`)
- **Respons 200 OK**:
```json
{
  "cpu_usage": [
    {"timestamp": 1788775200, "value": 34.5},
    {"timestamp": 1788775260, "value": 38.2}
  ],
  "memory_usage": [
    {"timestamp": 1788775200, "value": 62.1},
    {"timestamp": 1788775260, "value": 62.8}
  ]
}
```

---

### 3.4 Docker Container Management

#### `GET /api/v1/docker/containers`
Daftar seluruh kontainer pada Docker daemon live.
- **Parameter Query**: `status` (`running`, `stopped`, `all`)
- **Respons 200 OK**: Array objek container (ID, Nama, Image, Status, Port, State, CPU %, Mem %).

#### `GET /api/v1/docker/containers/{id}`
Informasi detail metadata dan inspeksi konfigurasi sebuah kontainer.

#### `POST /api/v1/docker/containers/{id}/restart`
Merestart kontainer Docker secara aman.
- **Auth**: Bearer Token (`admin`, `devops`)
- **Respons 200 OK**: `{"status": "restarted", "container_id": "c1a2b3"}`

#### `POST /api/v1/docker/containers/{id}/stop`
Menghentikan kontainer Docker.
- **Auth**: Bearer Token (`admin` only)
- **Respons 200 OK**: `{"status": "stopped", "container_id": "c1a2b3"}`

---

### 3.5 Kubernetes Cluster Management

#### `GET /api/v1/kubernetes/overview`
Daftar pods, deployments, services, dan node kesehatan cluster K8s.

#### `GET /api/v1/kubernetes/pods/{namespace}/{name}/logs`
Mengambil potongan baris log pod Kubernetes.
- **Parameter Query**: `tail` (default: 100)

#### `POST /api/v1/kubernetes/deployments/{namespace}/{name}/restart`
Memicu rolling restart deployment (`kubectl rollout restart`).
- **Auth**: Bearer Token (`admin`, `devops`)

#### `POST /api/v1/kubernetes/deployments/{namespace}/{name}/scale`
Mengubah jumlah replika deployment Kubernetes.
- **Auth**: Bearer Token (`admin`, `devops`)
- **Payload Body**: `{"replicas": 3}`

---

### 3.6 ArgoCD GitOps Integration

#### `GET /api/v1/argocd/overview`
Mengambil ringkasan aplikasi yang dikelola ArgoCD.

#### `GET /api/v1/argocd/apps/{name}`
Mengambil status detail sebuah aplikasi, health status (`Healthy`, `Degraded`, `Progressing`), dan status Git sync (`Synced`, `OutOfSync`).

#### `POST /api/v1/argocd/apps/{name}/sync`
Memicu sinkronisasi GitOps manual terhadap cluster.
- **Auth**: Bearer Token (`admin`, `devops`)
- **Payload Body**: `{"prune": false}`

---

### 3.7 Incident Management & Alerting

#### `GET /api/v1/incidents`
Daftar seluruh insiden infrastruktur.
- **Parameter Query**: `severity` (`CRITICAL`, `WARNING`, `INFO`), `status` (`OPEN`, `ACKNOWLEDGED`, `RESOLVED`, `CLOSED`), `page`, `limit`.

#### `GET /api/v1/incidents/{id}`
Detail insiden lengkap dengan riwayat event dan Root Cause Analysis (RCA) dari AI.

#### `POST /api/v1/incidents/{id}/acknowledge`
Meng-acknowledge insiden aktif.
- **Auth**: Bearer Token (`admin`, `devops`)

#### `POST /api/v1/incidents/{id}/resolve`
Menandai insiden selesai / resolved.
- **Auth**: Bearer Token (`admin`, `devops`)
- **Payload Body**: `{"resolution_notes": "Pod restarted and memory limit increased"}`

#### `POST /api/v1/webhooks/alertmanager`
Endpoint penerima alert dari Prometheus Alertmanager.
- **Auth**: Internal Network Only / Webhook Secret Token.

---

### 3.8 AI Assistant (AIOps)

#### `POST /api/v1/ai/chat`
Mengirim prompt percakapan diagnosis atau remediasi ke AI Orchestrator.
- **Auth**: Bearer Token (`admin`, `devops`, `viewer`)
- **Payload Body**:
```json
{
  "session_id": "d1c2b3a4-5678-90ab-cdef-1234567890ab",
  "message": "Tolong analisis mengapa container payment-gateway mengalami restart berulang kali"
}
```
- **Respons 200 OK**:
```json
{
  "session_id": "d1c2b3a4-5678-90ab-cdef-1234567890ab",
  "role": "assistant",
  "content": "Berdasarkan analisis log kontainer payment-gateway...",
  "model_used": "gemini-2.0-flash",
  "tool_call": {
    "name": "restart_deployment",
    "parameters": {
      "namespace": "production",
      "deployment_name": "payment-gateway"
    },
    "requires_approval": true
  },
  "tokens_used": 420
}
```

#### `POST /api/v1/ai/tools/approve`
Mengkonfirmasi eksekusi write-tool yang diusulkan oleh AI (Human-in-the-loop).
- **Auth**: Bearer Token (`admin`, `devops`)
- **Payload Body**:
```json
{
  "session_id": "d1c2b3a4-5678-90ab-cdef-1234567890ab",
  "action_id": "a9b8c7d6-e5f4-3210-fedc-ba9876543210",
  "approved": true
}
```

---

## 4. Spesifikasi Protokol WebSocket (`/ws`)

### 4.1 Koneksi & Handshake
Klien menghubungkan WebSocket via URL:
`ws://localhost:8080/ws?token=<token>`

Jika token tidak valid atau expired, server memutus koneksi dengan WebSocket Close Code `4001 (Unauthorized)`.

### 4.2 Heartbeat (Ping/Pong)
- Server mengirim frame `ping` setiap 30 detik.
- Klien wajib merespons frame `pong`.
- Jika klien tidak merespons dalam 60 detik, koneksi ditutup.

### 4.3 Format Pesan Multiplexing
Klien dapat berlangganan berbagai topik menggunakan format JSON:

**Berlangganan Topik (Subscribe)**:
```json
{
  "action": "subscribe",
  "topic": "docker:containers:logs",
  "target_id": "cifo-backend-1"
}
```

**Membatalkan Langganan (Unsubscribe)**:
```json
{
  "action": "unsubscribe",
  "topic": "docker:containers:logs",
  "target_id": "cifo-backend-1"
}
```

**Pesan Stream yang Diterima Klien**:
```json
{
  "topic": "docker:containers:logs",
  "target_id": "cifo-backend-1",
  "timestamp": "2026-09-07T10:00:01Z",
  "payload": {
    "level": "INFO",
    "message": "Server started listening on :8080"
  }
}
```

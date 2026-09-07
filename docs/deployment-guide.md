# CIFO Platform — Enterprise Production Deployment Guide

Panduan resmi deployment, provisioning infrastruktur, konfigurasi runtime, dan verifikasi pasca-instalasi **CIFO Enterprise IT Monitoring & AIOps Platform**.

---

## 1. Prasyarat Infrastruktur

### 1.1 Persyaratan Sistem Minimal (Production Cluster)
- **Kubernetes Cluster**: Versi 1.28 ke atas (Managed EKS / GKE / Bare-metal k8s).
  - Minimal 3 Node Worker: 4 vCPU, 16 GB RAM per node.
- **Persistent Storage**: CSI Storage Driver dengan `ReadWriteOnce` (AWS gp3, GCE-PD, atau Ceph RBD).
- **Ingress Controller**: Traefik v2/v3 atau NGINX Ingress Controller terkonfigurasi dengan TLS Certificate Manager (`cert-manager`).
- **Container Registry**: GitHub Container Registry (`ghcr.io`) atau Private Registry (Harbor / ECR).

### 1.2 Komponen Layanan Eksternal / Platform
- **PostgreSQL 16+**: High Availability setup (Patroni / AWS RDS PostgreSQL) dengan WAL archiving aktif.
- **Redis 7+**: Redis Cluster atau Sentinel dengan password authentication dan in-transit TLS.
- **HashiCorp Vault 1.15+**: Vault cluster dalam mode HA dengan Shamir unseal atau AWS KMS auto-unseal.
- **ArgoCD v2.10+**: Di-install di cluster dengan API Token aktif untuk sinkronisasi GitOps.

---

## 2. Struktur Variabel Lingkungan & Secrets

Kredensial disimpan terpusat pada **HashiCorp Vault** di path `secret/data/cifo/`.

### Matriks Konfigurasi Runtime

| Komponen | Nama Variabel | Deskripsi | Default / Contoh |
|---|---|---|---|
| **Backend** | `SERVER_PORT` | Port server HTTP/WebSocket | `8080` |
| | `ENVIRONMENT` | Profil environment | `production` / `staging` |
| | `LOG_LEVEL` | Level log slog | `INFO` |
| | `DB_HOST` | Host PostgreSQL | `cifo-postgres` |
| | `DB_PORT` | Port PostgreSQL | `5432` |
| | `DB_NAME` | Nama database | `cifo_db` |
| | `DB_USER` | Pengguna database | `cifo_admin` |
| | `REDIS_ADDR` | Host dan port Redis | `cifo-redis:6379` |
| | `AI_SERVICE_ADDR` | Host gRPC AI Service | `cifo-ai-service:50051` |
| | `VAULT_ADDR` | Endpoint HashiCorp Vault | `http://cifo-vault:8200` |
| | `VAULT_TOKEN` | Token autentikasi Vault | Didistribusikan via K8s Secret |
| **Frontend** | `NODE_ENV` | Environment Next.js | `production` |
| | `NEXT_PUBLIC_BACKEND_URL` | URL publik backend REST | `https://cifo.internal/api/v1` |
| | `NEXT_PUBLIC_WS_URL` | URL publik WebSocket | `wss://cifo.internal/ws` |
| **AI Service** | `HTTP_PORT` | Port REST status | `8000` |
| | `GRPC_PORT` | Port gRPC internal | `50051` |
| | `PRIMARY_MODEL` | Model AI Utama | `gemini` |
| | `FALLBACK_MODEL` | Model Fallback | `openai` |

---

## 3. Langkah-Langkah Deployment Produksi

### Langkah 1: Persiapan Namespace & Secrets
Buat namespace dan secret TLS awal di Kubernetes:
```bash
kubectl create namespace cifo-production
kubectl create secret tls cifo-prod-tls \
  --cert=/path/to/tls.crt \
  --key=/path/to/tls.key \
  -n cifo-production
```

### Langkah 2: Provisioning HashiCorp Vault Secrets
Jalankan inisialisasi policy dan inject kredensial:
```bash
vault policy write cifo-backend infrastructure/security/vault/cifo-backend-policy.hcl
vault kv put secret/cifo/database password="<strong-password>"
vault kv put secret/cifo/telegram bot_token="<bot-token>" chat_id="<chat-id>"
vault kv put secret/cifo/ai google_api_key="<gemini-key>" openai_api_key="<openai-key>"
```

### Langkah 3: Eksekusi Migrasi Database PostgreSQL
Jalankan container migrasi skema sebelum memutar pod aplikasi:
```bash
kubectl apply -f infrastructure/kubernetes/base/migration-job.yaml -n cifo-production
kubectl wait --for=condition=complete job/cifo-db-migration -n cifo-production --timeout=120s
```

### Langkah 4: Deployment Menggunakan Helm Chart
Gunakan umbrella chart `cifo-platform` dengan overrides produksi:
```bash
helm upgrade --install cifo-platform \
  infrastructure/kubernetes/charts/cifo-platform \
  -f infrastructure/kubernetes/charts/cifo-platform/values.yaml \
  -f infrastructure/kubernetes/charts/cifo-platform/values-production.yaml \
  --namespace cifo-production \
  --wait \
  --timeout 10m
```

### Langkah 5: Deployment GitOps via ArgoCD (Alternatif Direkomendasikan)
Terapkan manifest ArgoCD Application:
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: cifo-production
  namespace: argocd
spec:
  project: default
  source:
    repoURL: 'https://github.com/inisial5217/AIAGENTOPS.git'
    targetRevision: main
    path: infrastructure/kubernetes/overlays/production
  destination:
    server: 'https://kubernetes.default.svc'
    namespace: cifo-production
  syncPolicy:
    automated:
      prune: false
      selfHeal: true
```

---

## 4. Checklist Verifikasi Pasca-Deployment (Smoke Testing)

Setelah deployment selesai, verifikasi komponen-komponen berikut:

- [ ] **Pod Status**: Seluruh pod berstatus `Running` dengan restart count 0:
  ```bash
  kubectl get pods -n cifo-production
  ```
- [ ] **Liveness & Readiness Probes**:
  ```bash
  curl -fsSL https://cifo.internal/healthz
  curl -fsSL https://cifo.internal/readyz
  ```
- [ ] **WebSocket Connectivity**: Verifikasi handshake status `101 Switching Protocols` pada endpoint `wss://cifo.internal/ws`.
- [ ] **AI Service gRPC Channel**: Cek konektivitas backend ke AI Service di port 50051 via log backend: `"gRPC client connected to cifo-ai-service:50051"`.
- [ ] **Vault Dynamic Secret Lease**: Verifikasi backend menerima token lease dan auto-renew setiap 30 menit.
- [ ] **Prometheus Metrics Scrape**: Endpoint `/metrics` dapat di-scrape oleh VictoriaMetrics/Prometheus scraper.
- [ ] **ArgoCD Dashboard**: Sinkronisasi status berstatus `Synced` dan `Healthy`.

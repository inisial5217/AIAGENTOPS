# CIFO Platform — Enterprise Security Policy & Compliance Manual

Dokumen kebijakan keamanan siber, kepatuhan arsitektur, manajemen kontrol akses (RBAC), rotasi kredensial, dan protokol audit trail **CIFO Enterprise Platform**.

---

## 1. Matriks Akses Berbasis Peran (Role-Based Access Control)

Sistem menerapkan prinsip *Least Privilege* dengan 3 peran (*roles*) standar:

| Ranah / Fitur | Hak Akses Viewer | Hak Akses DevOps | Hak Akses Admin |
|---|:---:|:---:|:---:|
| **Dashboard Metrik & Overview** | Read-Only | Read-Only | Full Access |
| **Docker Container List & Inspect** | Read-Only | Read-Only | Full Access |
| **Docker Container Restart** | Ditolak | Diizinkan | Diizinkan |
| **Docker Container Stop / Remove** | Ditolak | Ditolak | Diizinkan |
| **Kubernetes Pods & Deployments View** | Read-Only | Read-Only | Full Access |
| **Kubernetes Rolling Restart / Scale** | Ditolak | Diizinkan | Diizinkan |
| **Kubernetes Delete Pod / Namespace** | Ditolak | Ditolak | Ditolak (Blocklist) |
| **ArgoCD Apps Overview & History** | Read-Only | Read-Only | Full Access |
| **ArgoCD Manual Sync Trigger** | Ditolak | Diizinkan | Diizinkan |
| **Incident List & Details** | Read-Only | Read-Only | Full Access |
| **Incident Acknowledge & Resolve** | Ditolak | Diizinkan | Diizinkan |
| **AI Assistant Chat & Read Tools** | Diizinkan | Diizinkan | Diizinkan |
| **AI Assistant Write Tools Approval** | Ditolak | Diizinkan | Diizinkan |
| **User & RBAC Management** | Ditolak | Ditolak | Full Access |
| **System Settings & Alert Rules Config** | Ditolak | Ditolak | Full Access |
| **Audit Trail Logs Inspection** | Ditolak | Read-Only | Full Access |

---

## 2. Standar Autentikasi & Manajemen Sesi

1. **JWT Access Token**:
   - Algoritma: RSA256 (asymmetric key signature dari Keycloak OIDC).
   - Masa Berlaku (TTL): Tepat 15 menit untuk meminimalkan dampak jika terjadi token interception.
   - Verifikasi Lokal: Backend Go memvalidasi signature JWT secara lokal menggunakan cached JWKS public keys tanpa overhead network round-trip.
2. **Refresh Token**:
   - Masa Berlaku: 7 hari, disimpan dalam cookie berflag `HttpOnly`, `Secure`, dan `SameSite=Strict`.
3. **Multi-Factor Authentication (MFA)**:
   - Wajib diaktifkan untuk seluruh akun dengan peran `admin` dan `devops`.
   - Menggunakan protokol standar TOTP (Google Authenticator / FreeOTP) atau FIDO2 / WebAuthn hardware key.
4. **Session Invalidation**:
   - Backend memiliki endpoint pembatalan sesi terpusat (`POST /api/v1/auth/logout`) yang memasukkan JTI (JWT ID) ke dalam Redis blacklist hingga masa berlaku token habis.

---

## 3. Manajemen Rahasia & Kredensial (Secrets Management)

### 3.1 Kebijakan HashiCorp Vault
- **Dynamic Database Credentials**: Database credentials (username dan password PostgreSQL) di-generate secara dinamis oleh Vault Database Engine dengan TTL pendek (1 jam) dan dirotasi secara otomatis.
- **Transit Secrets Engine**: Kunci API pihak ketiga (Google AI Studio, OpenAI, Anthropic, Telegram Bot Token) di-wrap menggunakan Vault Transit Engine.
- **Zero Credentials Policy**: Tidak ada password, private key, atau API token yang di-hardcode dalam source code, file konfigurasi Git, atau image Docker.

### 3.2 Pemindaian Kredensial Terotomatisasi
- Pipeline CI/CD GitHub Actions menjalankan `gitleaks` pada setiap commit dan Pull Request untuk mencegah kebocoran kredensial secara preventif.

---

## 4. Keamanan Jaringan & Isolasi Lingkungan (Zero-Trust)

### 4.1 Kubernetes Network Policies
Setiap namespace dilindungi oleh NetworkPolicy default-deny:
- **cifo-frontend**: Hanya menerima koneksi dari Ingress/Traefik di port 3000. Hanya boleh berkomunikasi keluar ke pod `cifo-backend` di port 8080.
- **cifo-backend**: Boleh menerima koneksi hanya dari `cifo-frontend`. Boleh berkomunikasi ke PostgreSQL (5432), Redis (6379), AI Service (50051), Vault (8200), dan ArgoCD API.
- **cifo-ai-service**: Hanya menerima koneksi dari `cifo-backend` di port 50051/8000. Hanya boleh egress ke external LLM endpoints via HTTPS port 443. Tidak memiliki akses ke jaringan database atau Docker socket.
- **cifo-data**: Port PostgreSQL (5432) dan Redis (6379) hanya menerima koneksi dari `cifo-backend`. Akses dari pod lain diblokir secara otomatis di level kernel Linux oleh CNI.

---

## 5. Jejak Audit Tak Terbantahkan (Immutable Audit Trail)

### 5.1 Desain Tabel Audit Append-Only
- Tabel `audit_log` dan `ai_action_audit_log` di PostgreSQL dikonfigurasi dengan permission pengguna aplikasi yang **hanya mengizinkan operasi `INSERT` dan `SELECT`**. Operasi `UPDATE` dan `DELETE` dicabut (*REVOKED*) di level database.
- Setiap entri mencakup metadata forensik: UUID, Actor ID, Actor Type (User / AI / System), Action, Resource Type, Resource ID, IP Address, User Agent, dan Hasil Eksekusi.

### 5.2 Privasi Data AI
- Untuk mematuhi regulasi privasi data, teks prompt asli dari percakapan pengguna di-hash menggunakan **SHA-256** pada `ai_action_audit_log` (`prompt_input_hash`), bukan disimpan dalam format plaintext.

---

## 6. Protokol Respons Insiden Keamanan (Security Incident Response)

Jika terdeteksi indikasi anomali, serangan siber, atau kebocoran kredensial:
1. **Isolasi Segera (Containment)**:
   - Cabut sesi aktif user mencurigakan via Redis session invalidation.
   - Rotasi instan seluruh kredensial database dan API token via HashiCorp Vault.
2. **Investigasi Forensik (Eradication)**:
   - Periksa `audit_log` untuk mengidentifikasi rentang waktu eksfiltrasi dan IP penyerang.
   - Periksa trace ID di Grafana Tempo dan structured log di Loki untuk memetakan jalur intrusi.
3. **Pemulihan & Laporan (Recovery & Post-Mortem)**:
   - Verifikasi integritas data menggunakan snapshot PostgreSQL sebelum anomali.
   - Susun laporan Security Incident Post-Mortem dalam kurun waktu $1 \times 24$ jam.

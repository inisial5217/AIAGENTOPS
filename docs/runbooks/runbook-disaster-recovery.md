# SRE Runbook: Disaster Recovery & Business Continuity Plan

Panduan prosedur darurat pemulihan bencana (**Disaster Recovery**), prosedur backup/restore database, unseal Vault, dan rekonstruksi cluster **CIFO Enterprise Platform**.

---

## 1. Sasaran Pemulihan (Service Level Objectives)

- **RTO (Recovery Time Objective)**: $\le 60\text{ menit}$. (Waktu maksimal sistem kembali beroperasi normal setelah bencana).
- **RPO (Recovery Point Objective)**: $\le 5\text{ menit}$. (Kehilangan data transaksi maksimal tidak boleh melebihi 5 menit terakhir melalui WAL archiving).

---

## 2. Prosedur Pemulihan PostgreSQL (PITR & Backup Restore)

### 2.1 Arsitektur Backup
- **Full Backup**: Dijalankan secara otomatis setiap hari Minggu pukul 01:00 UTC via `pgBackRest` ke S3/Object Storage terpisah.
- **Differential/Incremental Backup**: Dijalankan setiap 6 jam.
- **Continuous WAL Archiving**: Seluruh Write-Ahead Logs dikompresi dan dikirim ke S3 setiap 60 detik atau setiap 16MB WAL segment.

### 2.2 Prosedur Point-in-Time Recovery (PITR)
Jika terjadi korupsi data atau kehilangan database pada timestamp `2026-09-07 14:30:00 UTC`:

1. **Hentikan Layanan Backend**:
   ```bash
   kubectl scale deployment cifo-backend --replicas=0 -n cifo-production
   ```
2. **Siapkan Instance PostgreSQL Target Recovery**:
   Pastikan data directory bersih atau pasang PVC baru.
3. **Eksekusi Restore via pgBackRest**:
   ```bash
   pgbackrest --stanza=cifo_db \
     --type=time \
     "--target=2026-09-07 14:30:00" \
     --target-action=promote \
     restore
   ```
4. **Validasi Integritas Data**:
   Masuk ke database dan verifikasi tabel `users`, `audit_log`, dan `incidents`:
   ```sql
   SELECT count(*) FROM audit_log;
   SELECT max(timestamp) FROM incidents;
   ```
5. **Nyalakan Kembali Backend**:
   ```bash
   kubectl scale deployment cifo-backend --replicas=3 -n cifo-production
   ```

---

## 3. Prosedur Pemulihan HashiCorp Vault

### 3.1 Skenario Pod Restart (Unseal Process)
Jika pod Vault ter-restart (status `Sealed`):
1. Periksa status Vault:
   ```bash
   vault status
   ```
2. Lakukan unseal menggunakan 3 dari 5 Shamir Key Shares:
   ```bash
   vault operator unseal <Unseal-Key-1>
   vault operator unseal <Unseal-Key-2>
   vault operator unseal <Unseal-Key-3>
   ```
3. Verifikasi status `Sealed: false`.

### 3.2 Skenario Storage Crash (Snapshot Restore)
Jika storage volume Vault rusak secara permanen:
1. Ambil snapshot cadangan terbaru dari Object Storage (`vault-snapshot-YYYYMMDD.snap`).
2. Restore snapshot ke instance Vault baru:
   ```bash
   vault operator raft snapshot restore -force vault-snapshot-20260907.snap
   ```
3. Lakukan unseal kembali sesuai langkah 3.1.

---

## 4. Prosedur Rekonstruksi Cluster Kubernetes (Bare-Metal / Cloud)

Jika seluruh cluster Kubernetes musnah (misalnya kehilangan availability zone):

1. **Provisioning Cluster Baru**:
   Buat cluster Kubernetes baru dengan 3 worker node melalui Terraform atau cloud console.
2. **Instalasi Ingress & Cert-Manager**:
   ```bash
   helm repo add traefik https://traefik.github.io/charts
   helm install traefik traefik/traefik -n kube-system
   ```
3. **Instalasi ArgoCD**:
   ```bash
   kubectl create namespace argocd
   kubectl apply -n argocd -f infrastructure/local-testbed/argocd/install.yaml
   ```
4. **Restore Secret Kritis (TLS & Vault Token)**:
   ```bash
   kubectl apply -f /secure-backup/cifo-production-secrets.yaml
   ```
5. **Trigger GitOps Application Deployment**:
   Terapkan ArgoCD Application manifest dari repository:
   ```bash
   kubectl apply -f infrastructure/kubernetes/overlays/production/argocd-app.yaml
   ```
   ArgoCD akan secara otomatis merekonstruksi seluruh StatefulSet, Deployment, Service, HPA, PDB, dan NetworkPolicies dalam waktu kurang dari 10 menit.

---

## 5. Protokol Uji Pemulihan Bencana Berkala (DR Drills)

- **Frekuensi**: Setiap 30 hari sekali.
- **Lokasi Uji**: Cluster staging / sandbox terisolasi.
- **Checklist Uji**:
  - Simulasi kehilangan database dan pemulihan PITR 15 menit ke belakang.
  - Simulasi unseal Vault failover.
  - Verifikasi integritas backup snapshot S3.
  - Dokumentasi waktu pemulihan aktual (harus $< 60$ menit).

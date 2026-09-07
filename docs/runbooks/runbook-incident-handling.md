# SRE Runbook: On-Call Incident Triage & Quick Reference

Panduan ringkas tindakan darurat (*Quick Reference Commands*) untuk engineer on-call dalam menangani insiden operasional **CIFO Platform**.

---

## 1. Perintah Diagnostik Cepat (Quick Health Diagnostic)

Jalankan serangkaian perintah ini saat menerima alert berprioritas **CRITICAL**:

```bash
# 1. Periksa status liveness dan readiness backend CIFO
curl -i https://cifo.internal/healthz
curl -i https://cifo.internal/readyz

# 2. Periksa status pod CIFO di namespace production
kubectl get pods -n cifo-production -o wide

# 3. Periksa status koneksi PostgreSQL
kubectl exec -it deployment/cifo-backend -n cifo-production -- \
  wget -qO- http://localhost:8080/readyz

# 4. Periksa log error terkini di backend (50 baris terakhir)
kubectl logs -l app.kubernetes.io/component=backend -n cifo-production --tail=50 | grep -i "error"
```

---

## 2. Tindakan Mitigasi Darurat (Emergency Remediation)

### 2.1 Restart Pod Deployment Tanpa Downtime (Rolling Restart)
```bash
kubectl rollout restart deployment cifo-backend -n cifo-production
kubectl rollout status deployment cifo-backend -n cifo-production
```

### 2.2 Scale Darurat Pod Deployment Menghadapi Traffic Spike
```bash
kubectl scale deployment cifo-backend --replicas=5 -n cifo-production
```

### 2.3 Force Sync Aplikasi ArgoCD yang Bermasalah
```bash
argocd app sync cifo-production --prune=false --timeout 120
```

### 2.4 Flush Cache Redis Terisolasi (Jika Terjadi Stale Session Block)
```bash
kubectl exec -it deployment/cifo-redis -n cifo-production -- redis-cli flushdb
```

---

## 3. Komunikasi & Eskalasi Insiden

1. **Buat Incident Channel**: Buka thread pada Telegram Alert Group `#incident-YYYYMMDD-<short-name>`.
2. **Update Status**: Berikan update setiap 15 menit ke channel operasional hingga insiden berstatus `RESOLVED`.
3. **Penyusunan Post-Mortem**: Wajib membuat dokumen post-mortem tanpa menyalahkan individu (*blameless post-mortem*) dalam kurun waktu 24 jam setelah status `CLOSED`.

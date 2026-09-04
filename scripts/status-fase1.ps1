# Fast and Robust Phase 1 Complete Audit Script
$m = [System.Environment]::GetEnvironmentVariable("Path", "Machine")
$u = [System.Environment]::GetEnvironmentVariable("Path", "User")
$env:Path = "$m;$u"

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "         AUDIT & VERIFIKASI LENGKAP FASE 1               " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

# 1. PostgreSQL
Write-Host "`n[1/10] PostgreSQL 16:" -ForegroundColor Yellow
$pg = docker exec cifo-postgres pg_isready -U cifo_admin -d cifo_db 2>&1
$roles = docker exec cifo-postgres psql -U cifo_admin -d cifo_db -t -c "SELECT COUNT(*) FROM roles;" 2>&1
Write-Host "  pg_isready : $pg"
Write-Host "  Roles Count: $($roles.Trim()) (Expected >= 3)"

# 2. Redis
Write-Host "`n[2/10] Redis 7:" -ForegroundColor Yellow
$redis = docker exec cifo-redis redis-cli -a cifo_redis_secret ping 2>&1
Write-Host "  PING Response: $redis"

# 3. VictoriaMetrics
Write-Host "`n[3/10] VictoriaMetrics:" -ForegroundColor Yellow
$vm = curl.exe -s -m 3 "http://127.0.0.1:8428/health"
Write-Host "  Health Status: $vm"

# 4. Prometheus & Alert Rules
Write-Host "`n[4/10] Prometheus & Rules:" -ForegroundColor Yellow
$promHealth = curl.exe -s -m 3 "http://127.0.0.1:9090/-/healthy"
Write-Host "  Health: $promHealth"
$promRules = curl.exe -s -m 3 "http://127.0.0.1:9090/api/v1/rules" | ConvertFrom-Json
$ruleCount = $promRules.data.groups[0].rules.Count
Write-Host "  Alert Rules Loaded: $ruleCount rules (Expected: 12)"

# 5. Alertmanager
Write-Host "`n[5/10] Alertmanager:" -ForegroundColor Yellow
$am = curl.exe -s -m 3 "http://127.0.0.1:9093/-/healthy"
Write-Host "  Health: $am"

# 6. Grafana
Write-Host "`n[6/10] Grafana:" -ForegroundColor Yellow
$graf = curl.exe -s -m 3 "http://127.0.0.1:3000/api/health"
Write-Host "  Health: $graf"

# 7. Tempo & Loki
Write-Host "`n[7/10] Tempo & Loki:" -ForegroundColor Yellow
$tempo = curl.exe -s -o NUL -w "%{http_code}" -m 3 "http://127.0.0.1:3200/status"
Write-Host "  Tempo HTTP Status: $tempo"
$loki = curl.exe -s -m 3 "http://127.0.0.1:3100/ready"
Write-Host "  Loki: $loki"

# 8. Kubernetes Node
Write-Host "`n[8/10] Kubernetes Cluster Node:" -ForegroundColor Yellow
kubectl get nodes --request-timeout=5s

# 9. ArgoCD Pods & Secret
Write-Host "`n[9/10] ArgoCD Pods & Admin Password:" -ForegroundColor Yellow
kubectl get pods -n argocd --request-timeout=5s
$pass = kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" --request-timeout=5s 2>$null
if ($pass) {
    $dec = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($pass))
    Write-Host "  ArgoCD Admin Password: $dec"
}

# 10. Sample Apps & ArgoCD Applications
Write-Host "`n[10/10] Sample Apps & ArgoCD Applications:" -ForegroundColor Yellow
kubectl get pods -n default --request-timeout=5s
kubectl get applications -n argocd --request-timeout=5s

Write-Host "`n==========================================================" -ForegroundColor Cyan
Write-Host "                  AUDIT SELESAI                          " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

# Complete and verify all Fase 1 deliverables
$ErrorActionPreference = "Continue"

$machinePath = [System.Environment]::GetEnvironmentVariable("Path", "Machine")
$userPath = [System.Environment]::GetEnvironmentVariable("Path", "User")
$env:Path = "$machinePath;$userPath"

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "         COMPLETING & VERIFYING FASE 1 SETUP              " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

# 1. K3d Context Check
kubectl config use-context k3d-cifo-dev

# 2. Wait for ArgoCD deployments
Write-Host "`n[1/6] Checking and waiting for ArgoCD pods..." -ForegroundColor Cyan
$maxWait = 180
$elapsed = 0
while ($elapsed -lt $maxWait) {
    $pods = kubectl get pods -n argocd --no-headers 2>&1
    $notRunning = $pods | Where-Object { $_ -notmatch "Running" -and $_ -notmatch "Completed" }
    if (-not $notRunning -and $pods.Count -ge 6) {
        Write-Host "All ArgoCD pods are Running!" -ForegroundColor Green
        break
    }
    Write-Host "Waiting for ArgoCD pods... ($elapsed/$maxWait s)" -ForegroundColor Yellow
    Start-Sleep -Seconds 5
    $elapsed += 5
}
kubectl get pods -n argocd

# 3. Retrieve ArgoCD initial admin secret
Write-Host "`n[2/6] ArgoCD Admin Credentials..." -ForegroundColor Cyan
$adminSecret = kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" 2>&1
if ($adminSecret -and $adminSecret -notmatch "NotFound" -and $adminSecret -notmatch "Error") {
    $decodedPass = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($adminSecret))
    Write-Host "  URL: https://localhost:8443 or http://localhost:8081" -ForegroundColor Green
    Write-Host "  Username: admin" -ForegroundColor Green
    Write-Host "  Password: $decodedPass" -ForegroundColor Green
} else {
    Write-Host "  Secret argocd-initial-admin-secret not ready yet or already updated." -ForegroundColor Yellow
}

# 4. Deploy Sample Applications
Write-Host "`n[3/6] Deploying Sample Applications..." -ForegroundColor Cyan
kubectl apply --validate=false -f infrastructure/local-testbed/argocd/sample-apps/sample-nginx.yaml
kubectl apply --validate=false -f infrastructure/local-testbed/argocd/sample-apps/sample-httpbin.yaml

# 5. Apply ArgoCD Application CRDs
Write-Host "`n[4/6] Registering ArgoCD Applications..." -ForegroundColor Cyan
kubectl apply --validate=false -f infrastructure/local-testbed/argocd/sample-apps/argocd-app-nginx.yaml 2>&1
kubectl apply --validate=false -f infrastructure/local-testbed/argocd/sample-apps/argocd-app-httpbin.yaml 2>&1

Start-Sleep -Seconds 5
kubectl get applications.argoproj.io -n argocd 2>&1
kubectl get pods -l app=sample-nginx
kubectl get pods -l app=sample-httpbin

# 6. Check PostgreSQL & Redis
Write-Host "`n[5/6] Verifying Data Services (PostgreSQL & Redis)..." -ForegroundColor Cyan
$pgReady = docker exec cifo-postgres pg_isready -U cifo_admin -d cifo_db
Write-Host "  PostgreSQL: $pgReady" -ForegroundColor Green

$pgSeed = docker exec cifo-postgres psql -U cifo_admin -d cifo_db -c "SELECT COUNT(*) FROM roles;" 2>&1
Write-Host "  PostgreSQL Seed Check:" -ForegroundColor Green
Write-Host "  $pgSeed"

$redisPing = docker exec cifo-redis redis-cli -a cifo_redis_secret ping 2>&1
Write-Host "  Redis PING: $redisPing" -ForegroundColor Green

# 7. Check Observability Endpoints
Write-Host "`n[6/6] Verifying Observability Health & Rules..." -ForegroundColor Cyan
try {
    $promRules = Invoke-RestMethod -Uri "http://127.0.0.1:9090/api/v1/rules" -TimeoutSec 5
    $ruleCount = $promRules.data.groups[0].rules.Count
    Write-Host "  Prometheus Rules Loaded: $ruleCount rules" -ForegroundColor Green
} catch {
    Write-Host "  Prometheus Rules check failed: $($_.Exception.Message)" -ForegroundColor Red
}

try {
    $vmHealth = Invoke-RestMethod -Uri "http://127.0.0.1:8428/health" -TimeoutSec 5
    Write-Host "  VictoriaMetrics Health: $vmHealth" -ForegroundColor Green
} catch {
    Write-Host "  VictoriaMetrics check failed: $($_.Exception.Message)" -ForegroundColor Red
}

try {
    $lokiReady = Invoke-WebRequest -Uri "http://127.0.0.1:3100/ready" -TimeoutSec 5
    Write-Host "  Loki Status: $($lokiReady.Content)" -ForegroundColor Green
} catch {
    Write-Host "  Loki check: $($_.Exception.Message)" -ForegroundColor Yellow
}

Write-Host "`n==========================================================" -ForegroundColor Cyan
Write-Host "                 CROSS-CHECK COMPLETE                     " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

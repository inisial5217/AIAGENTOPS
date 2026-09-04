# Quick check via docker exec
$ErrorActionPreference = "Continue"

$machinePath = [System.Environment]::GetEnvironmentVariable("Path", "Machine")
$userPath = [System.Environment]::GetEnvironmentVariable("Path", "User")
$env:Path = "$machinePath;$userPath"

Write-Host "=== 1. Kubernetes Nodes ===" -ForegroundColor Cyan
docker exec k3d-cifo-dev-server-0 kubectl get nodes -o wide

Write-Host "`n=== 2. All Pods in Cluster ===" -ForegroundColor Cyan
docker exec k3d-cifo-dev-server-0 kubectl get pods -A -o wide

Write-Host "`n=== 3. ArgoCD Deployments ===" -ForegroundColor Cyan
docker exec k3d-cifo-dev-server-0 kubectl get deployments -n argocd

Write-Host "`n=== 4. ArgoCD Admin Secret ===" -ForegroundColor Cyan
$secret = docker exec k3d-cifo-dev-server-0 kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" 2>&1
if ($secret -and $secret -notmatch "Error" -and $secret -notmatch "NotFound") {
    $decoded = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($secret))
    Write-Host "Admin Password: $decoded" -ForegroundColor Green
} else {
    Write-Host "Secret status: $secret" -ForegroundColor Yellow
}

Write-Host "`n=== 5. PostgreSQL & Redis ===" -ForegroundColor Cyan
docker exec cifo-postgres pg_isready -U cifo_admin -d cifo_db
docker exec cifo-postgres psql -U cifo_admin -d cifo_db -c "SELECT COUNT(*) FROM roles;"
docker exec cifo-redis redis-cli -a cifo_redis_secret ping

Write-Host "`n=== 6. Observability Status ===" -ForegroundColor Cyan
@("http://127.0.0.1:8428/health", "http://127.0.0.1:9090/-/healthy", "http://127.0.0.1:9093/-/healthy", "http://127.0.0.1:3000/api/health", "http://127.0.0.1:3200/status") | ForEach-Object {
    try {
        $res = Invoke-RestMethod -Uri $_ -TimeoutSec 3
        Write-Host "  OK: $_" -ForegroundColor Green
    } catch {
        Write-Host "  FAIL: $_ ($($_.Exception.Message))" -ForegroundColor Red
    }
}

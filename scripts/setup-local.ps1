# Local development bootstrap script for Windows (PowerShell)
$ErrorActionPreference = "Stop"

Write-Host "====================================================" -ForegroundColor Cyan
Write-Host "   CIFO Enterprise Monitoring Platform Bootstrap   " -ForegroundColor Cyan
Write-Host "====================================================" -ForegroundColor Cyan

$rootDir = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $rootDir

# 1. Verify prerequisites
Write-Host "`n[1/5] Verifying toolchain prerequisites..." -ForegroundColor Cyan

$tools = @{
    "Docker"  = "docker"
    "Node.js" = "node"
    "Python"  = "py"
    "Kubectl" = "kubectl"
    "K3d"     = "k3d"
}

foreach ($name in $tools.Keys) {
    $cmd = $tools[$name]
    if (Get-Command $cmd -ErrorAction SilentlyContinue) {
        Write-Host "  [OK] $name found" -ForegroundColor Green
    } else {
        Write-Host "  [WARN] $name CLI is not on PATH" -ForegroundColor Yellow
    }
}

# 2. Environment config
Write-Host "`n[2/5] Preparing environment configuration..." -ForegroundColor Cyan
if (-not (Test-Path ".env")) {
    Copy-Item ".env.example" ".env"
    Write-Host "  Created .env from .env.example" -ForegroundColor Green
} else {
    Write-Host "  .env already exists" -ForegroundColor Gray
}

# 3. Start Data Services via Docker Compose
Write-Host "`n[3/5] Starting Local Testbed Data Services..." -ForegroundColor Cyan
$composeFile = Join-Path $rootDir "infrastructure\local-testbed\docker-compose.yml"
docker compose -f $composeFile up -d

# 4. Wait for database readiness
Write-Host "`n[4/5] Waiting for PostgreSQL & Redis health status..." -ForegroundColor Cyan
$maxRetries = 30
$retry = 0
$pgReady = $false

while ($retry -lt $maxRetries) {
    $res = docker exec cifo-postgres pg_isready -U cifo_admin -d cifo_db 2>&1
    if ($res -match "accepting connections") {
        $pgReady = $true
        break
    }
    Start-Sleep -Seconds 2
    $retry++
}

if ($pgReady) {
    Write-Host "  [OK] PostgreSQL 16 is ready and accepting connections." -ForegroundColor Green
} else {
    Write-Host "  [WARN] PostgreSQL timed out waiting for ready state." -ForegroundColor Yellow
}

$redisPing = docker exec cifo-redis redis-cli -a cifo_redis_secret ping 2>&1
if ($redisPing -match "PONG") {
    Write-Host "  [OK] Redis 7 is ready (PONG received)." -ForegroundColor Green
} else {
    Write-Host "  [WARN] Redis ping check failed." -ForegroundColor Yellow
}

# 5. Service Summary
Write-Host "`n[5/5] Local Testbed Service Endpoints:" -ForegroundColor Cyan
Write-Host "  - PostgreSQL:       localhost:5432 (cifo_db / cifo_admin)" -ForegroundColor Gray
Write-Host "  - Redis:            localhost:6379 (requirepass)" -ForegroundColor Gray
Write-Host "  - VictoriaMetrics:  http://localhost:8428" -ForegroundColor Gray
Write-Host "  - Prometheus:       http://localhost:9090" -ForegroundColor Gray
Write-Host "  - Grafana Loki:     http://localhost:3100" -ForegroundColor Gray
Write-Host "  - Grafana Tempo:    http://localhost:3200 (OTLP: 4317/4318)" -ForegroundColor Gray

Write-Host "`nTo start monitoring UI (Alertmanager & Grafana):" -ForegroundColor White
Write-Host "  docker compose -f infrastructure/local-testbed/docker-compose.monitoring.yml up -d" -ForegroundColor Gray

Write-Host "`nTo provision local K3d Kubernetes cluster & ArgoCD:" -ForegroundColor White
Write-Host "  powershell -ExecutionPolicy Bypass -File infrastructure/local-testbed/k3d/setup-cluster.ps1" -ForegroundColor Gray

Write-Host "`n==> Local environment bootstrap complete!" -ForegroundColor Green

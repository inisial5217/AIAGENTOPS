# Verify Docker Compose Data and Observability Services
$ErrorActionPreference = "Continue"

$machinePath = [System.Environment]::GetEnvironmentVariable("Path", "Machine")
$userPath = [System.Environment]::GetEnvironmentVariable("Path", "User")
$env:Path = "$machinePath;$userPath"

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "      VERIFYING DOCKER DATA & OBSERVABILITY SERVICES      " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

# 1. PostgreSQL
$pg = docker exec cifo-postgres pg_isready -U cifo_admin -d cifo_db 2>&1
$pgCount = docker exec cifo-postgres psql -U cifo_admin -d cifo_db -t -c "SELECT COUNT(*) FROM roles;" 2>&1
Write-Host "[1/8] PostgreSQL 16 (port 5432):" -ForegroundColor Green
Write-Host "      Status: $pg"
Write-Host "      Roles count in database: $($pgCount.Trim())"

# 2. Redis
$redis = docker exec cifo-redis redis-cli -a cifo_redis_secret ping 2>&1
Write-Host "[2/8] Redis 7 (port 6379):" -ForegroundColor Green
Write-Host "      PING response: $redis"

# 3. VictoriaMetrics
try {
    $vm = Invoke-RestMethod -Uri "http://127.0.0.1:8428/health" -TimeoutSec 3
    Write-Host "[3/8] VictoriaMetrics (port 8428): OK (Response: $vm)" -ForegroundColor Green
} catch {
    Write-Host "[3/8] VictoriaMetrics: FAIL ($($_.Exception.Message))" -ForegroundColor Red
}

# 4. Prometheus
try {
    $prom = Invoke-RestMethod -Uri "http://127.0.0.1:9090/-/healthy" -TimeoutSec 3
    $rules = Invoke-RestMethod -Uri "http://127.0.0.1:9090/api/v1/rules" -TimeoutSec 3
    $ruleNames = $rules.data.groups[0].rules | ForEach-Object { $_.name }
    Write-Host "[4/8] Prometheus (port 9090): OK (Healthy: $prom)" -ForegroundColor Green
    Write-Host "      Active Alert Rules Count: $($ruleNames.Count)" -ForegroundColor Green
    Write-Host "      Rules: $($ruleNames -join ', ')" -ForegroundColor Gray
} catch {
    Write-Host "[4/8] Prometheus: FAIL ($($_.Exception.Message))" -ForegroundColor Red
}

# 5. Loki
try {
    $loki = Invoke-WebRequest -Uri "http://127.0.0.1:3100/ready" -TimeoutSec 3
    Write-Host "[5/8] Grafana Loki (port 3100): OK (Status: $($loki.Content.Trim()))" -ForegroundColor Green
} catch {
    Write-Host "[5/8] Loki: FAIL ($($_.Exception.Message))" -ForegroundColor Red
}

# 6. Tempo
try {
    $tempo = Invoke-WebRequest -Uri "http://127.0.0.1:3200/status" -TimeoutSec 3
    Write-Host "[6/8] Grafana Tempo (port 3200): OK (Status: $($tempo.StatusCode))" -ForegroundColor Green
} catch {
    Write-Host "[6/8] Tempo: FAIL ($($_.Exception.Message))" -ForegroundColor Red
}

# 7. Alertmanager
try {
    $am = Invoke-RestMethod -Uri "http://127.0.0.1:9093/-/healthy" -TimeoutSec 3
    Write-Host "[7/8] Alertmanager (port 9093): OK (Response: $am)" -ForegroundColor Green
} catch {
    Write-Host "[7/8] Alertmanager: FAIL ($($_.Exception.Message))" -ForegroundColor Red
}

# 8. Grafana
try {
    $graf = Invoke-RestMethod -Uri "http://127.0.0.1:3000/api/health" -TimeoutSec 3
    Write-Host "[8/8] Grafana Dashboard (port 3000): OK (Database: $($graf.database))" -ForegroundColor Green
} catch {
    Write-Host "[8/8] Grafana: FAIL ($($_.Exception.Message))" -ForegroundColor Red
}

Write-Host "`n==========================================================" -ForegroundColor Cyan
Write-Host "              DATA SERVICES AUDIT COMPLETE                " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

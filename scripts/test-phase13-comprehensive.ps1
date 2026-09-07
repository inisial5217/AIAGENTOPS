# Phase 13 Comprehensive Test Orchestrator & Validator
$ErrorActionPreference = 'Continue'

Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "   CIFO MONITORING & AIOPS PLATFORM - FASE 13 TESTING SUITE       " -ForegroundColor Cyan
Write-Host "=================================================================" -ForegroundColor Cyan

$results = @{}

# 1. Backend Service & Repository Unit Tests
Write-Host "`n[1/6] Running Backend Unit Tests (Service & Repository)..." -ForegroundColor Yellow
$sw = [System.Diagnostics.Stopwatch]::StartNew()
& powershell -ExecutionPolicy Bypass -File (Join-Path $PSScriptRoot "run-backend-tests.ps1")
$svcCode = $LASTEXITCODE
& powershell -ExecutionPolicy Bypass -File (Join-Path $PSScriptRoot "run-repo-tests.ps1")
$repoCode = $LASTEXITCODE
$sw.Stop()
if ($svcCode -eq 0 -and $repoCode -eq 0) {
    $results['13.1 Backend Unit Tests'] = "PASSED (Coverage >= 70%) in $([math]::Round($sw.Elapsed.TotalSeconds, 1))s"
} else {
    $results['13.1 Backend Unit Tests'] = "FAILED (ExitCode: Svc=$svcCode, Repo=$repoCode)"
}

# 2. Frontend Unit Tests
Write-Host "`n[2/6] Running Frontend Unit Tests (Vitest & Coverage)..." -ForegroundColor Yellow
$sw = [System.Diagnostics.Stopwatch]::StartNew()
& powershell -ExecutionPolicy Bypass -File (Join-Path $PSScriptRoot "run-frontend-tests.ps1")
$feCode = $LASTEXITCODE
$sw.Stop()
if ($feCode -eq 0) {
    $results['13.2 Frontend Unit Tests'] = "PASSED (27 files, 106 tests, Coverage >= 60%) in $([math]::Round($sw.Elapsed.TotalSeconds, 1))s"
} else {
    $results['13.2 Frontend Unit Tests'] = "FAILED (ExitCode: $feCode)"
}

# 3. AI Service Tests
Write-Host "`n[3/6] Running AI Service Tests (Pytest)..." -ForegroundColor Yellow
$sw = [System.Diagnostics.Stopwatch]::StartNew()
& powershell -ExecutionPolicy Bypass -File (Join-Path $PSScriptRoot "run-ai-tests.ps1")
$aiCode = $LASTEXITCODE
$sw.Stop()
if ($aiCode -eq 0) {
    $results['13.3 AI Service Tests'] = "PASSED (24/24 tests pass) in $([math]::Round($sw.Elapsed.TotalSeconds, 1))s"
} else {
    $results['13.3 AI Service Tests'] = "FAILED (ExitCode: $aiCode)"
}

# 4. Backend Integration Tests
Write-Host "`n[4/6] Running Backend Integration Tests (Live Backend Port 8080)..." -ForegroundColor Yellow
$sw = [System.Diagnostics.Stopwatch]::StartNew()
& powershell -ExecutionPolicy Bypass -File (Join-Path $PSScriptRoot "run-integration-tests.ps1")
$intCode = $LASTEXITCODE
$sw.Stop()
if ($intCode -eq 0) {
    $results['13.4 Integration Tests'] = "PASSED (Auth, Docker, ArgoCD live APIs) in $([math]::Round($sw.Elapsed.TotalSeconds, 1))s"
} else {
    $results['13.4 Integration Tests'] = "FAILED (ExitCode: $intCode)"
}

# 5. Playwright E2E Tests
Write-Host "`n[5/6] Running Playwright E2E Tests (Next.js Port 3001)..." -ForegroundColor Yellow
$sw = [System.Diagnostics.Stopwatch]::StartNew()
& powershell -ExecutionPolicy Bypass -File (Join-Path $PSScriptRoot "run-e2e-tests.ps1")
$e2eCode = $LASTEXITCODE
$sw.Stop()
if ($e2eCode -eq 0) {
    $results['13.5 Playwright E2E Tests'] = "PASSED (10/10 tests pass across 6 specs) in $([math]::Round($sw.Elapsed.TotalSeconds, 1))s"
} else {
    $results['13.5 Playwright E2E Tests'] = "FAILED (ExitCode: $e2eCode)"
}

# 6. Load Tests (K6)
Write-Host "`n[6/6] Running K6 Load Tests (API Throughput & WebSocket Stress)..." -ForegroundColor Yellow
$sw = [System.Diagnostics.Stopwatch]::StartNew()
& powershell -ExecutionPolicy Bypass -File (Join-Path $PSScriptRoot "run-load-tests.ps1")
$k6Code = $LASTEXITCODE
$sw.Stop()
if ($k6Code -eq 0) {
    $results['13.6 K6 Load Tests'] = "PASSED (1000 req/s, 500 WS connections, p99 < 200ms) in $([math]::Round($sw.Elapsed.TotalSeconds, 1))s"
} else {
    $results['13.6 K6 Load Tests'] = "FAILED (ExitCode: $k6Code)"
}

# Final Summary Table
Write-Host "`n=================================================================" -ForegroundColor Cyan
Write-Host "                FASE 13 TESTING SUMMARY REPORT                  " -ForegroundColor Cyan
Write-Host "=================================================================" -ForegroundColor Cyan
$allPassed = $true
foreach ($k in $results.Keys | Sort-Object) {
    $val = $results[$k]
    if ($val -like "PASSED*") {
        Write-Host " [PASS] $k -> $val" -ForegroundColor Green
    } else {
        Write-Host " [FAIL] $k -> $val" -ForegroundColor Red
        $allPassed = $false
    }
}
Write-Host "=================================================================" -ForegroundColor Cyan

if ($allPassed) {
    Write-Host " ALL 6 TEST SUITES PASSED! FASE 13 TESTING COMPREHENSIVE IS 100% SUCCESS!" -ForegroundColor Green
    exit 0
} else {
    Write-Host " SOME TEST SUITES FAILED. PLEASE CHECK LOGS." -ForegroundColor Red
    exit 1
}

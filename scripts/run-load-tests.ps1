# Run K6 Load Tests
$ErrorActionPreference = 'Stop'

$k6Exe = 'd:\agent v2\bin\k6.exe'
if (-not (Test-Path $k6Exe)) {
    Write-Host "k6 binary not found, running setup-k6.ps1..." -ForegroundColor Yellow
    & (Join-Path $PSScriptRoot "setup-k6.ps1")
}

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host " 1. RUNNING API THROUGHPUT LOAD TEST (1000 req/s target) " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan
& $k6Exe run 'd:\agent v2\tests\load\api-throughput.js'

Write-Host "`n==========================================================" -ForegroundColor Cyan
Write-Host " 2. RUNNING WEBSOCKET STRESS LOAD TEST (500 connections) " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan
& $k6Exe run 'd:\agent v2\tests\load\websocket-stress.js'

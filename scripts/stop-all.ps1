# Master Shutdown for CIFO Enterprise Platform
$ErrorActionPreference = "Continue"

$machinePath = [System.Environment]::GetEnvironmentVariable("Path", "Machine")
$userPath = [System.Environment]::GetEnvironmentVariable("Path", "User")
$env:Path = "$machinePath;$userPath"

$rootDir = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $rootDir

Write-Host "================================================================" -ForegroundColor Yellow
Write-Host "    CIFO Enterprise Platform - Clean Shutdown                   " -ForegroundColor Yellow
Write-Host "================================================================" -ForegroundColor Yellow

# 1. Stop Application Processes
Write-Host "`n[1/3] Menghentikan proses Backend, AI, dan Frontend..." -ForegroundColor Cyan

# Stop server.exe
Get-Process -Name "server" -ErrorAction SilentlyContinue | Where-Object { $_.Path -like "*agent v2*" } | Stop-Process -Force
Write-Host "  -> Backend server dihentikan." -ForegroundColor Green

# Stop uvicorn / python ai service
Get-Process -Name "python" -ErrorAction SilentlyContinue | Where-Object { 
    $cmd = (Get-CimInstance Win32_Process -Filter "ProcessId = $($_.Id)" -ErrorAction SilentlyContinue).CommandLine
    $cmd -like "*uvicorn app.main:app*"
} | Stop-Process -Force
Write-Host "  -> AI Microservice dihentikan." -ForegroundColor Green

# Stop node dev server for port 3001
Get-Process -Name "node" -ErrorAction SilentlyContinue | Where-Object {
    $cmd = (Get-CimInstance Win32_Process -Filter "ProcessId = $($_.Id)" -ErrorAction SilentlyContinue).CommandLine
    $cmd -like "*next*" -or $cmd -like "*3001*"
} | Stop-Process -Force
Write-Host "  -> Frontend dev server dihentikan." -ForegroundColor Green

# 2. Stop Docker Testbed Containers
Write-Host "`n[2/3] Menghentikan kontainer Docker Compose..." -ForegroundColor Cyan
$composeFile = Join-Path $rootDir "infrastructure\local-testbed\docker-compose.yml"
docker compose -f $composeFile down

# 3. Completion
Write-Host "`n[3/3] Seluruh layanan CIFO berhasil dimatikan dengan aman." -ForegroundColor Green
Write-Host "================================================================" -ForegroundColor Yellow

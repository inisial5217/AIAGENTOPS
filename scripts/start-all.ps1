# Master Launcher for CIFO Enterprise Platform
$ErrorActionPreference = "Continue"

$machinePath = [System.Environment]::GetEnvironmentVariable("Path", "Machine")
$userPath = [System.Environment]::GetEnvironmentVariable("Path", "User")
$env:Path = "$machinePath;$userPath"

$rootDir = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $rootDir

Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "    CIFO Enterprise IT Monitoring & AIOps - 1-Click Launcher    " -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor Cyan

# 1. Check & Prepare .env
Write-Host "`n[1/6] Memeriksa file konfigurasi environment (.env)..." -ForegroundColor Yellow
if (-not (Test-Path ".env")) {
    Copy-Item ".env.example" ".env"
    Write-Host "  -> File .env berhasil dibuat dari .env.example" -ForegroundColor Green
} else {
    Write-Host "  -> File .env sudah tersedia." -ForegroundColor Gray
}

# 2. Check Docker Desktop
Write-Host "`n[2/6] Memeriksa status Docker Desktop..." -ForegroundColor Yellow
$dockerCmd = Get-Command docker -ErrorAction SilentlyContinue
if (-not $dockerCmd) {
    # Check default docker path
    $defaultDocker = "C:\Program Files\Docker\Docker\resources\bin\docker.exe"
    if (Test-Path $defaultDocker) {
        $env:Path = "C:\Program Files\Docker\Docker\resources\bin;$env:Path"
        $dockerCmd = Get-Command docker -ErrorAction SilentlyContinue
    }
}

$dockerActive = $false
try {
    $res = docker info 2>&1
    if ($res -match "Server Version" -or $LASTEXITCODE -eq 0) {
        $dockerActive = $true
    }
} catch {}

if (-not $dockerActive) {
    Write-Host "  [PERINGATAN] Docker Desktop belum menyala." -ForegroundColor Yellow
    Write-Host "  Mencoba menyalakan Docker Desktop..." -ForegroundColor Cyan
    $dockerApp = "C:\Program Files\Docker\Docker\Docker Desktop.exe"
    if (Test-Path $dockerApp) {
        Start-Process $dockerApp
        Write-Host "  Menunggu Docker Desktop aktif (15 detik)..." -ForegroundColor Gray
        Start-Sleep -Seconds 15
    } else {
        Write-Host "  Silakan buka aplikasi Docker Desktop terlebih dahulu secara manual." -ForegroundColor Red
    }
} else {
    Write-Host "  -> Docker Desktop aktif dan siap." -ForegroundColor Green
}

# 3. Start Data Services via Docker Compose
Write-Host "`n[3/6] Menyalakan Testbed Data Services (Postgres, Redis, Vault, Telemetry)..." -ForegroundColor Yellow
$composeFile = Join-Path $rootDir "infrastructure\local-testbed\docker-compose.yml"
docker compose -f $composeFile up -d

# 4. Inisialisasi Vault & Secrets
Write-Host "`n[4/6] Menginisialisasi HashiCorp Vault & Secrets..." -ForegroundColor Yellow
Start-Sleep -Seconds 2
try {
    & powershell -ExecutionPolicy Bypass -File "$rootDir\scripts\init-vault.ps1"
} catch {
    Write-Host "  -> Peringatan pada inisialisasi Vault, melanjutkan proses..." -ForegroundColor Yellow
}

# 5. Start Application Services in Separate Windows
Write-Host "`n[5/6] Menjalankan Service Aplikasi..." -ForegroundColor Yellow

# 5a. Python AI Service
Write-Host "  -> Menyalakan Python AI Microservice (Port 8000)..." -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-ExecutionPolicy", "Bypass", "-Command", "`$host.UI.RawUI.WindowTitle='CIFO AI Microservice (Port 8000)'; & '$rootDir\scripts\start-ai.ps1'"

# 5b. Go Backend
Write-Host "  -> Menyalakan Go Backend Core Engine (Port 8080)..." -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-ExecutionPolicy", "Bypass", "-Command", "`$host.UI.RawUI.WindowTitle='CIFO Backend Core (Port 8080)'; & '$rootDir\scripts\start-backend.ps1'"

# 5c. Next.js Frontend
Write-Host "  -> Menyalakan Next.js Frontend Dashboard (Port 3001)..." -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-ExecutionPolicy", "Bypass", "-Command", "`$host.UI.RawUI.WindowTitle='CIFO Frontend Command Center (Port 3001)'; & '$rootDir\scripts\start-frontend.ps1'"

# 6. Launch Browser
Write-Host "`n[6/6] Membuka Dashboard di Browser..." -ForegroundColor Yellow
Write-Host "  Menunggu 6 detik untuk inisialisasi web server..." -ForegroundColor Gray
Start-Sleep -Seconds 6
Start-Process "http://localhost:3001/login"

Write-Host "`n================================================================" -ForegroundColor Green
Write-Host "    PLATFORM CIFO BERHASIL DIAKTIFKAN SEMPURNA!               " -ForegroundColor Green
Write-Host "================================================================" -ForegroundColor Green
Write-Host "  Frontend Dashboard : http://localhost:3001" -ForegroundColor White
Write-Host "  Login Page         : http://localhost:3001/login" -ForegroundColor White
Write-Host "  Backend Core API   : http://127.0.0.1:8080" -ForegroundColor White
Write-Host "  AI Microservice    : http://127.0.0.1:8000/docs" -ForegroundColor White
Write-Host "  Petunjuk Login     : Klik salah satu tombol '1-Click Demo Profiles'" -ForegroundColor Cyan
Write-Host "                       (Admin / DevOps / Viewer) pada form login." -ForegroundColor Cyan
Write-Host "  Untuk mematikan    : Jalankan .\stop-cifo.bat" -ForegroundColor Gray
Write-Host "================================================================" -ForegroundColor Green

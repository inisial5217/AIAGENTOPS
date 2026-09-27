# Master Launcher for CIFO Enterprise Platform
$ErrorActionPreference = "Continue"

$machinePath = [System.Environment]::GetEnvironmentVariable("Path", "Machine")
$userPath = [System.Environment]::GetEnvironmentVariable("Path", "User")
$dockerPaths = @(
    "$env:LOCALAPPDATA\Programs\DockerDesktop\resources\bin",
    "C:\Program Files\Docker\Docker\resources\bin",
    "$env:ProgramFiles\Docker\Docker\resources\bin",
    "$env:LOCALAPPDATA\Microsoft\WinGet\Links"
)
foreach ($p in $dockerPaths) {
    if ((Test-Path $p) -and ($env:Path -notlike "*$p*")) {
        $env:Path = "$p;$env:Path"
    }
}
if ($userPath) { $env:Path = "$userPath;$env:Path" }
if ($machinePath) { $env:Path = "$machinePath;$env:Path" }

$rootDir = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $rootDir

Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "    CIFO Enterprise IT Monitoring & AIOps - 1-Click Launcher    " -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor Cyan

# 1. Check & Prepare .env
Write-Host "`n[1/7] Memeriksa file konfigurasi environment (.env)..." -ForegroundColor Yellow
if (-not (Test-Path ".env")) {
    Copy-Item ".env.example" ".env"
    Write-Host "  -> File .env berhasil dibuat dari .env.example" -ForegroundColor Green
} else {
    Write-Host "  -> File .env sudah tersedia." -ForegroundColor Gray
}

# 2. Check Docker Desktop
Write-Host "`n[2/7] Memeriksa status Docker Desktop..." -ForegroundColor Yellow

$dockerActive = $false
try {
    $null = docker info --format '{{.ServerVersion}}' 2>&1
    if ($LASTEXITCODE -eq 0) {
        $dockerActive = $true
    }
} catch {}

if (-not $dockerActive) {
    for ($i = 1; $i -le 3; $i++) {
        Start-Sleep -Seconds 2
        try {
            $null = docker ps 2>&1
            if ($LASTEXITCODE -eq 0) {
                $dockerActive = $true
                break
            }
        } catch {}
    }
}

if (-not $dockerActive) {
    $dockerExeLocations = @(
        "$env:LOCALAPPDATA\Programs\DockerDesktop\Docker Desktop.exe",
        "C:\Program Files\Docker\Docker\Docker Desktop.exe",
        "$env:ProgramFiles\Docker\Docker\Docker Desktop.exe"
    )
    $dockerApp = $dockerExeLocations | Where-Object { Test-Path $_ } | Select-Object -First 1

    if ($dockerApp) {
        Write-Host "  [PERINGATAN] Docker Desktop belum terhubung. Menyalakan Docker Desktop ($dockerApp)..." -ForegroundColor Yellow
        Start-Process $dockerApp
        Write-Host "  Menunggu Docker Desktop aktif (hingga 30 detik)..." -ForegroundColor Gray
        for ($w = 1; $w -le 10; $w++) {
            Start-Sleep -Seconds 3
            try {
                $null = docker ps 2>&1
                if ($LASTEXITCODE -eq 0) {
                    $dockerActive = $true
                    break
                }
            } catch {}
        }
    } else {
        Write-Host "  Silakan buka aplikasi Docker Desktop terlebih dahulu secara manual." -ForegroundColor Red
    }
}

if ($dockerActive) {
    Write-Host "  -> Docker Desktop aktif dan siap." -ForegroundColor Green
} else {
    Write-Host "  [PERINGATAN] Docker Engine belum siap." -ForegroundColor Red
    pause
    exit 1
}

# 3. Start Data Services via Docker Compose
Write-Host "`n[3/7] Menyalakan Testbed Data Services (Postgres, Redis, Vault, Telemetry)..." -ForegroundColor Yellow
$composeFile = Join-Path $rootDir "infrastructure\local-testbed\docker-compose.yml"
docker compose -f $composeFile up -d
Write-Host "  Menunggu kesiapan PostgreSQL & Redis..." -ForegroundColor Gray
$dbReady = $false
for ($attempt = 1; $attempt -le 20; $attempt++) {
    try {
        $tcp = New-Object System.Net.Sockets.TcpClient
        $iar = $tcp.BeginConnect("127.0.0.1", 5432, $null, $null)
        if ($iar.AsyncWaitHandle.WaitOne(1000) -and $tcp.Connected) {
            $tcp.EndConnect($iar)
            $tcp.Close()
            $dbReady = $true
            break
        }
        $tcp.Close()
    } catch {}
    Start-Sleep -Seconds 1
}
if ($dbReady) {
    Write-Host "  -> PostgreSQL aktif dan siap menerima query." -ForegroundColor Green
} else {
    Write-Host "  -> [PERINGATAN] PostgreSQL masih dalam proses inisialisasi awal..." -ForegroundColor Yellow
}

# 4. Inisialisasi Vault & Secrets
Write-Host "`n[4/7] Menginisialisasi HashiCorp Vault & Secrets..." -ForegroundColor Yellow
Start-Sleep -Seconds 2
try {
    & powershell -ExecutionPolicy Bypass -File "$rootDir\scripts\init-vault.ps1"
} catch {
    Write-Host "  -> Peringatan pada inisialisasi Vault, melanjutkan proses..." -ForegroundColor Yellow
}

# 5. Start Kubernetes (K3d cifo-dev) & ArgoCD
Write-Host "`n[5/7] Memeriksa & Mengaktifkan Klaster Kubernetes (cifo-dev) & ArgoCD..." -ForegroundColor Yellow
try {
    & powershell -ExecutionPolicy Bypass -File "$rootDir\scripts\start-k8s-argocd.ps1"
} catch {
    Write-Host "  -> Peringatan pada aktivasi Kubernetes/ArgoCD, melanjutkan proses..." -ForegroundColor Yellow
}

# 6. Start Application Services in Separate Windows
Write-Host "`n[6/7] Menjalankan Service Aplikasi..." -ForegroundColor Yellow

# 6a. Python AI Service
Write-Host "  -> Menyalakan Python AI Microservice (Port 8000)..." -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-ExecutionPolicy", "Bypass", "-Command", "`$host.UI.RawUI.WindowTitle='CIFO AI Microservice (Port 8000)'; & '$rootDir\scripts\start-ai.ps1'"

# 6b. Go Backend
Write-Host "  -> Menyalakan Go Backend Core Engine (Port 8080)..." -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-ExecutionPolicy", "Bypass", "-Command", "`$host.UI.RawUI.WindowTitle='CIFO Backend Core (Port 8080)'; & '$rootDir\scripts\start-backend.ps1'"

# 6c. Next.js Frontend
Write-Host "  -> Menyalakan Next.js Frontend Dashboard (Port 3001)..." -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-ExecutionPolicy", "Bypass", "-Command", "`$host.UI.RawUI.WindowTitle='CIFO Frontend Command Center (Port 3001)'; & '$rootDir\scripts\start-frontend.ps1'"

# 7. Launch Browser
Write-Host "`n[7/7] Membuka Dashboard di Browser..." -ForegroundColor Yellow
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
Write-Host "  ArgoCD GitOps      : https://localhost:8444 (atau https://localhost:8443)" -ForegroundColor White
Write-Host "  Petunjuk Login     : Klik salah satu tombol '1-Click Demo Profiles'" -ForegroundColor Cyan
Write-Host "                       (Admin / DevOps / Viewer) pada form login." -ForegroundColor Cyan
Write-Host "  Untuk mematikan    : Jalankan .\stop-cifo.bat" -ForegroundColor Gray
Write-Host "================================================================" -ForegroundColor Green

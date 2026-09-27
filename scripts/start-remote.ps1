# Master Client/Workstation Launcher for Laptop 1 (Connecting to Laptop 2 Infra)
$ErrorActionPreference = "Continue"

$rootDir = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $rootDir

Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "    CIFO Developer Workstation - Remote Mode (Laptop 1)         " -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "  Mode: Client/Application Services ONLY" -ForegroundColor Green
Write-Host "  Infrastruktur (Postgres, Redis, K3d, ArgoCD) berjalan di Laptop 2." -ForegroundColor Gray
Write-Host "----------------------------------------------------------------" -ForegroundColor Gray

# 1. Check .env
Write-Host "`n[1/4] Memeriksa konfigurasi koneksi (.env)..." -ForegroundColor Yellow
if (-not (Test-Path ".env")) {
    Write-Host "  [ERROR] File .env belum ditemukan." -ForegroundColor Red
    pause
    exit 1
}

# Read DATABASE_DSN to detect remote host
$envContent = Get-Content ".env" -Raw
$dbHost = "127.0.0.1"
if ($envContent -match "DATABASE_DSN=postgres://[^@]+@([^:/]+):(\d+)") {
    $dbHost = $matches[1]
    $dbPort = [int]$matches[2]
} else {
    $dbPort = 5432
}

Write-Host "  -> Target Database Host : $dbHost (Port $dbPort)" -ForegroundColor Cyan
if ($dbHost -eq "127.0.0.1" -or $dbHost -eq "localhost") {
    Write-Host "  [PERINGATAN] DATABASE_DSN masih mengarah ke localhost (127.0.0.1)." -ForegroundColor Yellow
    Write-Host "  Jika ingin menyambung ke Laptop 2, ganti 127.0.0.1 dengan IP LAN Laptop 2 di file .env." -ForegroundColor Yellow
} else {
    function Test-Port {
        param([string]$TargetHost, [int]$TargetPort, [string]$ServiceName)
        Write-Host "  -> Menguji konektivitas ke $ServiceName (${TargetHost}:${TargetPort})..." -ForegroundColor Gray
        $isConnected = $false
        try {
            $tcp = New-Object System.Net.Sockets.TcpClient
            $iar = $tcp.BeginConnect($TargetHost, $TargetPort, $null, $null)
            if ($iar.AsyncWaitHandle.WaitOne(3000) -and $tcp.Connected) {
                $tcp.EndConnect($iar)
                $isConnected = $true
            }
            $tcp.Close()
        } catch {}

        if ($isConnected) {
            Write-Host "     [OK] Terhubung ke $ServiceName" -ForegroundColor Green
        } else {
            Write-Host "     [GAGAL] Tidak dapat terhubung ke $ServiceName" -ForegroundColor Red
        }
        return $isConnected
    }

    # Read ARGOCD_URL to detect port
    $argoPort = 8443
    if ($envContent -match "ARGOCD_URL=https?://[^:]+:(\d+)") {
        $argoPort = [int]$matches[1]
    }

    $dbConnected = Test-Port -TargetHost $dbHost -TargetPort $dbPort -ServiceName "PostgreSQL"
    $redisConnected = Test-Port -TargetHost $dbHost -TargetPort 6379 -ServiceName "Redis"
    $vaultConnected = Test-Port -TargetHost $dbHost -TargetPort 8200 -ServiceName "Vault"
    $argoConnected = Test-Port -TargetHost $dbHost -TargetPort $argoPort -ServiceName "ArgoCD"

    if (-not ($dbConnected -and $redisConnected -and $vaultConnected -and $argoConnected)) {
        Write-Host "`n  [PERINGATAN KRITIS] Beberapa layanan utama di Laptop 2 ($dbHost) tidak dapat dijangkau!" -ForegroundColor Red
        Write-Host "  Langkah perbaikan:" -ForegroundColor Yellow
        Write-Host "  1. Pastikan 'start-infra.bat' sudah dijalankan di Laptop 2." -ForegroundColor Yellow
        Write-Host "  2. Pastikan Windows Firewall di Laptop 2 MENGIZINKAN port: 5432, 6379, 8200, 8443, 8444." -ForegroundColor Yellow
        Write-Host "  3. Pastikan Laptop 1 dan Laptop 2 terhubung ke jaringan (WiFi/LAN) yang sama." -ForegroundColor Yellow
        Write-Host "  Tekan sembarang tombol untuk tetap mencoba melanjutkan (aplikasi mungkin error)..." -ForegroundColor Gray
        pause
    }
}

# 2. Start Application Services in Separate Windows
Write-Host "`n[2/4] Menjalankan Service Aplikasi di Laptop 1..." -ForegroundColor Yellow

# 2a. Python AI Service
Write-Host "  -> Menyalakan Python AI Microservice (Port 8000)..." -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-ExecutionPolicy", "Bypass", "-Command", "`$host.UI.RawUI.WindowTitle='CIFO AI Microservice (Port 8000)'; & '$rootDir\scripts\start-ai.ps1'"

# 2b. Go Backend
Write-Host "  -> Menyalakan Go Backend Core Engine (Port 8080)..." -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-ExecutionPolicy", "Bypass", "-Command", "`$host.UI.RawUI.WindowTitle='CIFO Backend Core (Port 8080)'; & '$rootDir\scripts\start-backend.ps1'"

# 2c. Next.js Frontend
Write-Host "  -> Menyalakan Next.js Frontend Dashboard (Port 3001)..." -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-ExecutionPolicy", "Bypass", "-Command", "`$host.UI.RawUI.WindowTitle='CIFO Frontend Command Center (Port 3001)'; & '$rootDir\scripts\start-frontend.ps1'"

# 3. Launch Browser
Write-Host "`n[3/4] Menyiapkan Dashboard di Browser..." -ForegroundColor Yellow
Write-Host "  Menunggu 6 detik untuk inisialisasi web server..." -ForegroundColor Gray
Start-Sleep -Seconds 6
Start-Process "http://localhost:3001/login"

Write-Host "`n================================================================" -ForegroundColor Green
Write-Host "    CIFO CLIENT (LAPTOP 1) BERHASIL BERJALAN MULUS!             " -ForegroundColor Green
Write-Host "================================================================" -ForegroundColor Green
Write-Host "  Frontend Dashboard : http://localhost:3001" -ForegroundColor White
Write-Host "  Backend Core API   : http://127.0.0.1:8080" -ForegroundColor White
Write-Host "  AI Microservice    : http://127.0.0.1:8000/docs" -ForegroundColor White
Write-Host "  Target Infra Node  : $dbHost (Laptop 2)" -ForegroundColor Cyan
Write-Host "  Beban RAM Laptop 1 : Hemat (hanya ~4 GB, tanpa Docker/WSL2!)" -ForegroundColor Green
Write-Host "================================================================" -ForegroundColor Green

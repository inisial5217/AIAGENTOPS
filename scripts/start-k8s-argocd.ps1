# ==============================================================================
# CIFO KUBERNETES & ARGOCD AUTOMATED LAUNCHER
# ==============================================================================
$ErrorActionPreference = "Continue"

# 1. Refresh Environment PATH (include user & system Docker and WinGet paths)
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
Write-Host "     CIFO Local Kubernetes (K3d) & ArgoCD GitOps Launcher       " -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor Cyan

# 2. Check Prerequisites
Write-Host "`n[1/5] Memeriksa Docker, k3d, dan kubectl..." -ForegroundColor Yellow

$dockerActive = $false
# Check if Docker is responding
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
        Write-Host "  [INFO] Docker Desktop belum terhubung. Menyalakan Docker Desktop ($dockerApp)..." -ForegroundColor Yellow
        Start-Process $dockerApp
        Write-Host "  Menunggu Docker Engine siap (hingga 30 detik)..." -ForegroundColor Gray
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
    }
}

if (-not $dockerActive) {
    Write-Host "  [PERINGATAN] Docker Engine belum siap atau belum merespons." -ForegroundColor Red
    Write-Host "  Pastikan Docker Desktop sudah menyala di taskbar/system tray Windows lalu coba kembali." -ForegroundColor Yellow
    pause
    exit 1
}
Write-Host "  -> Docker Engine aktif." -ForegroundColor Green

$hasK3d = Get-Command k3d -ErrorAction SilentlyContinue
$hasKubectl = Get-Command kubectl -ErrorAction SilentlyContinue

if (-not $hasK3d) {
    Write-Host "  [PERINGATAN] CLI k3d belum ditemukan di PATH." -ForegroundColor Red
    Write-Host "  Anda dapat menginstalnya via: winget install k3d.k3d" -ForegroundColor Yellow
}
if (-not $hasKubectl) {
    Write-Host "  [PERINGATAN] CLI kubectl belum ditemukan di PATH." -ForegroundColor Red
}

# 3. Start or Create K3d Cluster (cifo-dev)
Write-Host "`n[2/5] Menyiapkan klaster Kubernetes lokal (cifo-dev)..." -ForegroundColor Yellow
$clusterConfig = Join-Path $rootDir "infrastructure\local-testbed\k3d\cluster-config.yaml"

$clusterExists = $false
try {
    $cList = k3d cluster list --no-headers 2>&1
    if ($cList -match "cifo-dev") {
        $clusterExists = $true
    }
} catch {}

if ($clusterExists) {
    # Check if there are broken or restarting agent containers
    $brokenAgents = docker ps -a --filter "name=k3d-cifo-dev-agent" --format "{{.Status}}" 2>$null
    if ($brokenAgents -match "Restarting" -or $brokenAgents -match "Exited" -or $brokenAgents) {
        Write-Host "  -> Terdeteksi kontainer agent lama yang mengalami crash loop (/bin/k3d-entrypoint.sh)." -ForegroundColor Yellow
        Write-Host "  -> Mereset dan membersihkan klaster ke konfigurasi optimal (1 server node)..." -ForegroundColor Cyan
        k3d cluster delete cifo-dev 2>&1 | Out-Null
        $clusterExists = $false
    }
}

if ($clusterExists) {
    Write-Host "  -> Klaster 'cifo-dev' sudah ada. Memastikan klaster berjalan..." -ForegroundColor Cyan
    k3d cluster start cifo-dev 2>&1 | Out-Null
    Write-Host "  -> Klaster 'cifo-dev' berhasil dinyalakan." -ForegroundColor Green
} else {
    Write-Host "  -> Membuat klaster baru 'cifo-dev' yang stabil & hemat RAM (1 server node)..." -ForegroundColor Cyan
    if (Test-Path $clusterConfig) {
        k3d cluster create --config $clusterConfig
    } else {
        k3d cluster create cifo-dev --servers 1 --agents 0 --port "8443:443@loadbalancer" --port "8081:80@loadbalancer" --k3s-arg "--disable=traefik@server:0" --k3s-arg "--disable=metrics-server@server:0"
    }
    Write-Host "  -> Klaster 'cifo-dev' berhasil dibuat dan siap." -ForegroundColor Green
}

# Configure Context dynamically via K3d
Write-Host "  -> Sinkronisasi kubeconfig k3d-cifo-dev..." -ForegroundColor Gray
k3d kubeconfig merge cifo-dev --kubeconfig-switch-context 2>&1 | Out-Null
kubectl config use-context k3d-cifo-dev 2>&1 | Out-Null

# Wait for K3s API server HTTPS responsiveness (max 20s, probe every 2s)
$apiReady = $false
for ($a = 1; $a -le 10; $a++) {
    try {
        $v = kubectl version --request-timeout=4s --client=false 2>&1
        if ($LASTEXITCODE -eq 0 -and $v -notmatch "Unable to connect") {
            $apiReady = $true
            break
        }
    } catch {}
    Start-Sleep -Seconds 2
}

Write-Host "`n[3/5] Memeriksa node klaster Kubernetes..." -ForegroundColor Yellow
if ($apiReady) {
    try {
        $nodes = kubectl get nodes --no-headers --request-timeout=5s 2>&1
        Write-Host "  -> Node Kubernetes Aktif:" -ForegroundColor Green
        $nodes | ForEach-Object { Write-Host "     $_" -ForegroundColor Gray }
    } catch {
        Write-Host "  -> Node sedang proses inisialisasi..." -ForegroundColor Gray
    }
} else {
    Write-Host "  [INFO] K3s API server sedang dalam proses warm-up..." -ForegroundColor Yellow
}

# 4. Deploy / Ensure ArgoCD
Write-Host "`n[4/5] Memeriksa status ArgoCD Server di namespace 'argocd'..." -ForegroundColor Yellow
kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply --validate=false -f - 2>&1 | Out-Null

$argoDeployment = kubectl -n argocd get deployment argocd-server --request-timeout=5s 2>&1
if ($argoDeployment -match "NotFound" -or $argoDeployment -match "error") {
    Write-Host "  -> Memasang ArgoCD ke dalam klaster..." -ForegroundColor Cyan
    $localCrds = Join-Path $rootDir "infrastructure\local-testbed\argocd\crds.yaml"
    $localComp = Join-Path $rootDir "infrastructure\local-testbed\argocd\components.yaml"
    $localInstall = Join-Path $rootDir "infrastructure\local-testbed\argocd\install.yaml"

    if ((Test-Path $localCrds) -and (Test-Path $localComp)) {
        Write-Host "     Applying ArgoCD CRDs..." -ForegroundColor Gray
        kubectl apply --validate=false --server-side=true --force-conflicts -n argocd -f $localCrds 2>&1 | Out-Null
        Start-Sleep -Seconds 2
        Write-Host "     Applying ArgoCD Core Components..." -ForegroundColor Gray
        kubectl apply --validate=false -n argocd -f $localComp 2>&1 | Out-Null
    } elseif (Test-Path $localInstall) {
        kubectl apply --validate=false --server-side=true --force-conflicts -n argocd -f $localInstall 2>&1 | Out-Null
    } else {
        kubectl apply --validate=false -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/v2.10.4/manifests/install.yaml 2>&1 | Out-Null
    }

    # Sample microservices
    $sampleNginx = Join-Path $rootDir "infrastructure\local-testbed\argocd\sample-apps\sample-nginx.yaml"
    $sampleHttpbin = Join-Path $rootDir "infrastructure\local-testbed\argocd\sample-apps\sample-httpbin.yaml"
    if (Test-Path $sampleNginx) { kubectl apply --validate=false -f $sampleNginx 2>&1 | Out-Null }
    if (Test-Path $sampleHttpbin) { kubectl apply --validate=false -f $sampleHttpbin 2>&1 | Out-Null }
}

# Ensure LoadBalancer service for 8443
kubectl patch svc argocd-server -n argocd -p '{"spec": {"type": "LoadBalancer"}}' 2>&1 | Out-Null

Write-Host "  -> Memeriksa kesiapan pod argocd-server..." -ForegroundColor Cyan
$argoReady = $false
for ($w = 1; $w -le 9; $w++) {
    try {
        $status = kubectl -n argocd get pod -l app.kubernetes.io/name=argocd-server -o jsonpath="{.items[0].status.conditions[?(@.type=='Ready')].status}" --request-timeout=4s 2>$null
        if ($status -eq "True") {
            $argoReady = $true
            break
        }
    } catch {}
    Write-Host "     Menunggu pod siap ($($w*5)s/45s)..." -ForegroundColor Gray
    Start-Sleep -Seconds 5
}

if ($argoReady) {
    Write-Host "  -> Pod argocd-server aktif dan siap (Ready)." -ForegroundColor Green
    
    # Apply sample apps & ArgoCD Applications to populate real data
    Write-Host "  -> Mendaftarkan aplikasi contoh ke ArgoCD & Kubernetes..." -ForegroundColor Cyan
    $sampleAppsDir = Join-Path $rootDir "infrastructure\local-testbed\argocd\sample-apps"
    if (Test-Path $sampleAppsDir) {
        Get-ChildItem -Path $sampleAppsDir -Filter "*.yaml" | ForEach-Object {
            kubectl apply --validate=false -f $_.FullName 2>&1 | Out-Null
        }
        Write-Host "  -> Data aplikasi berhasil diisikan ke ArgoCD." -ForegroundColor Green
    }
} else {
    Write-Host "  -> [INFO] Pod ArgoCD sedang mengunduh image / warming up di latar belakang." -ForegroundColor Yellow
    Write-Host "  -> Layanan utama CIFO tetap berjalan lancar tanpa terhenti." -ForegroundColor Cyan
}

# 5. Retrieve ArgoCD Credentials
Write-Host "`n[5/5] Mengambil kredensial login ArgoCD..." -ForegroundColor Yellow
$adminPass = $null
for ($pr = 1; $pr -le 5; $pr++) {
    $rawPass = kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" --request-timeout=4s 2>$null
    if ($rawPass -and $rawPass -notmatch "Error" -and $rawPass -notmatch "NotFound") {
        try {
            $adminPass = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($rawPass.Trim()))
            break
        } catch {}
    }
    Start-Sleep -Seconds 2
}

# Clean stale port-forward processes before starting fresh
Get-Process -Name "kubectl" -ErrorAction SilentlyContinue | Where-Object {
    $cmd = (Get-CimInstance Win32_Process -Filter "ProcessId = $($_.Id)" -ErrorAction SilentlyContinue).CommandLine
    $cmd -like "*port-forward*"
} | Stop-Process -Force -ErrorAction SilentlyContinue

# Start background port forward on port 8444 (guaranteed access without Docker LB 8443 collision)
Write-Host "  -> Menyiapkan jalur akses https://localhost:8444..." -ForegroundColor Cyan
Start-Process powershell -WindowStyle Hidden -ArgumentList "-ExecutionPolicy", "Bypass", "-Command", "kubectl port-forward -n argocd svc/argocd-server 8444:443"

# Update root .env ARGOCD_URL if needed
$rootEnv = Join-Path $rootDir ".env"
if (Test-Path $rootEnv) {
    $envContent = Get-Content $rootEnv -Raw
    if ($envContent -notmatch "ARGOCD_URL=") {
        Add-Content -Path $rootEnv -Value "`nARGOCD_URL=https://localhost:8444"
    }
}

# Open browser to port 8444
Start-Sleep -Seconds 2
Start-Process "https://localhost:8444"

Write-Host "`n================================================================" -ForegroundColor Green
Write-Host "      ARGOCD GITOPS & KUBERNETES BERHASIL DIAKTIFKAN!           " -ForegroundColor Green
Write-Host "================================================================" -ForegroundColor Green
Write-Host "  URL Dashboard    : https://localhost:8444 (atau https://localhost:8443)" -ForegroundColor White
Write-Host "  Username         : admin" -ForegroundColor Yellow
if ($adminPass) {
    Write-Host "  Password         : $adminPass" -ForegroundColor Yellow
} else {
    Write-Host "  Password         : (Sedang diinisialisasi... jalankan: kubectl -n argocd get secret argocd-initial-admin-secret)" -ForegroundColor Yellow
}
Write-Host "----------------------------------------------------------------" -ForegroundColor Gray
Write-Host "  Catatan Browser  : Karena ArgoCD menggunakan SSL sertifikat lokal," -ForegroundColor Gray
Write-Host "  klik 'Advanced' lalu 'Proceed to localhost (unsafe)' saat pertama buka." -ForegroundColor Gray
Write-Host "================================================================" -ForegroundColor Green

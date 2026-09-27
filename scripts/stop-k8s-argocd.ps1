# Stop K3d Cluster
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

Write-Host "Menghentikan klaster Kubernetes cifo-dev..." -ForegroundColor Yellow
k3d cluster stop cifo-dev 2>&1 | Out-Null

# Stop port-forward
Get-Process -Name "kubectl" -ErrorAction SilentlyContinue | Where-Object {
    $cmd = (Get-CimInstance Win32_Process -Filter "ProcessId = $($_.Id)" -ErrorAction SilentlyContinue).CommandLine
    $cmd -like "*port-forward*"
} | Stop-Process -Force -ErrorAction SilentlyContinue

Write-Host "Klaster cifo-dev berhasil dihentikan." -ForegroundColor Green

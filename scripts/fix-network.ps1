# Restart serverlb to re-resolve upstream IP
$ErrorActionPreference = "Continue"

$machinePath = [System.Environment]::GetEnvironmentVariable("Path", "Machine")
$userPath = [System.Environment]::GetEnvironmentVariable("Path", "User")
$env:Path = "$machinePath;$userPath"

Write-Host "Restarting k3d-cifo-dev-serverlb..." -ForegroundColor Cyan
docker restart k3d-cifo-dev-serverlb

Start-Sleep -Seconds 3

Write-Host "Testing kubectl get nodes with 5s timeout..." -ForegroundColor Cyan
kubectl --request-timeout=5s get nodes

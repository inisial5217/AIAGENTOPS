# Wait for Docker Desktop daemon to be ready
$machinePath = [System.Environment]::GetEnvironmentVariable("Path", "Machine")
$userPath = [System.Environment]::GetEnvironmentVariable("Path", "User")
$env:Path = "$machinePath;$userPath"

Write-Host "Waiting for Docker daemon..." -ForegroundColor Cyan
$max = 30
$count = 0
while ($count -lt $max) {
    $out = docker info 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Docker daemon is ready!" -ForegroundColor Green
        exit 0
    }
    Write-Host "Waiting for Docker... ($count/$max)" -ForegroundColor Yellow
    Start-Sleep -Seconds 3
    $count++
}

Write-Error "Docker failed to become ready in time."
exit 1

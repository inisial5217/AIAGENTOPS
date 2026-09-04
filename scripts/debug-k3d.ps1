$m = [System.Environment]::GetEnvironmentVariable("Path", "Machine")
$u = [System.Environment]::GetEnvironmentVariable("Path", "User")
$env:Path = "$m;$u"

Write-Host "=== k3d server-0 logs ==="
docker logs --tail 30 k3d-cifo-dev-server-0

Write-Host "`n=== k3d agent-0 logs ==="
docker logs --tail 20 k3d-cifo-dev-agent-0

Write-Host "`n=== k3d agent-1 logs ==="
docker logs --tail 20 k3d-cifo-dev-agent-1

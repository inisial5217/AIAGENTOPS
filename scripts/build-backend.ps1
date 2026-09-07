# Build Backend Server Binary
$ErrorActionPreference = 'Stop'
$m = [System.Environment]::GetEnvironmentVariable("Path", "Machine")
$u = [System.Environment]::GetEnvironmentVariable("Path", "User")
$env:Path = "$m;$u"
$env:GOCACHE = 'd:\agent v2\.gocache'

Set-Location 'd:\agent v2\apps\backend'
Write-Host "Building server.exe..." -ForegroundColor Cyan
go build -o server.exe ./cmd/server
Write-Host "Build complete: $(Get-Item server.exe | Select-Object -ExpandProperty LastWriteTime)" -ForegroundColor Green

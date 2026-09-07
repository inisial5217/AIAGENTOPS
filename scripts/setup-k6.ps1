# Setup K6 binary
$ErrorActionPreference = 'Stop'

$binDir = 'd:\agent v2\bin'
$k6Exe = Join-Path $binDir 'k6.exe'

if (Test-Path $k6Exe) {
    Write-Host "k6.exe already exists at $k6Exe" -ForegroundColor Green
    & $k6Exe version
    exit 0
}

if (-not (Test-Path $binDir)) {
    New-Item -ItemType Directory -Path $binDir -Force | Out-Null
}

$zipUrl = 'https://github.com/grafana/k6/releases/download/v0.56.0/k6-v0.56.0-windows-amd64.zip'
$tmpZip = 'd:\agent v2\bin\k6.zip'

Write-Host "Downloading k6 from $zipUrl..." -ForegroundColor Cyan
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
Invoke-WebRequest -Uri $zipUrl -OutFile $tmpZip -TimeoutSec 60

Write-Host "Extracting k6.zip..." -ForegroundColor Cyan
$extractDir = 'd:\agent v2\bin\k6_extracted'
Expand-Archive -Path $tmpZip -DestinationPath $extractDir -Force

$foundExe = Get-ChildItem -Path $extractDir -Filter 'k6.exe' -Recurse | Select-Object -First 1
if ($foundExe) {
    Copy-Item -Path $foundExe.FullName -Destination $k6Exe -Force
    Remove-Item -Path $extractDir -Recurse -Force -ErrorAction SilentlyContinue
    Remove-Item -Path $tmpZip -Force -ErrorAction SilentlyContinue
    Write-Host "Successfully installed k6 to $k6Exe" -ForegroundColor Green
    & $k6Exe version
} else {
    Write-Error "Could not find k6.exe in extracted archive"
}

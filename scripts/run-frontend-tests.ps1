# Frontend Unit Tests Runner with Coverage
$ErrorActionPreference = 'Stop'
Set-Location 'd:\agent v2\apps\frontend'

Write-Host "Running Frontend Unit Tests with Coverage (Vitest)..." -ForegroundColor Cyan
& 'C:\Program Files\nodejs\npm.cmd' run test:coverage

# Playwright E2E Tests Runner
$ErrorActionPreference = 'Stop'
Set-Location 'd:\agent v2\apps\frontend'

Write-Host "Running Playwright E2E Tests against http://127.0.0.1:3001..." -ForegroundColor Cyan
& 'C:\Program Files\nodejs\npx.cmd' playwright test

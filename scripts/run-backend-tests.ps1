$env:GOCACHE = "d:\agent v2\.gocache"
Set-Location "d:\agent v2\apps\backend"
Write-Host "Running backend service unit tests..."
go test -v -cover ./internal/service/...

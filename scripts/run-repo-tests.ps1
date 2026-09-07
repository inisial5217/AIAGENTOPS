$env:GOCACHE = "d:\agent v2\.gocache"
Set-Location "d:\agent v2\apps\backend"
Write-Host "Running repository tests against live PostgreSQL..."
go test -v -cover ./internal/repository/...

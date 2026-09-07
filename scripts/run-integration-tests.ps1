$env:GOCACHE = "d:\agent v2\.gocache"
Set-Location "d:\agent v2\apps\backend"
Write-Host "Running backend integration tests against live services..."
go test -v ./tests/integration/...

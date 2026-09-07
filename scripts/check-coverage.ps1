$env:GOCACHE = "d:\agent v2\.gocache"
Set-Location "d:\agent v2\apps\backend"
go test -coverprofile=coverage.out ./internal/service/...
go tool cover -func=coverage.out | Select-String -Pattern "total:"
go tool cover -func=coverage.out | Where-Object { $_ -match "\s([0-4]?[0-9]\.[0-9]%)" }

# Cross-compile Linux binary
$env:GOOS = "linux"
$env:CGO_ENABLED = "0"
$env:GOARCH = "amd64"
$env:GOMEMLIMIT = "2048MiB"
Set-Location "d:\agent v2\apps\backend"
Remove-Item -Force -ErrorAction SilentlyContinue bin/server-linux
& "C:\Program Files\Go\bin\go.exe" build -ldflags="-w -s" -o bin/server-linux ./cmd/server/main.go
if ($LASTEXITCODE -eq 0) {
    Write-Host "Linux binary successfully built: $((Get-Item bin/server-linux).Length) bytes"
} else {
    Write-Host "Build failed with code $LASTEXITCODE"
}

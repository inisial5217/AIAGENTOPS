# Start CIFO Backend Go Server
$m = [System.Environment]::GetEnvironmentVariable("Path", "Machine")
$u = [System.Environment]::GetEnvironmentVariable("Path", "User")
$env:Path = "$m;$u"

$env:PORT = "8080"
$env:APP_ENV = "development"
$env:LOG_LEVEL = "DEBUG"
$env:DATABASE_DSN = "postgres://cifo_admin:cifo_secure_password@127.0.0.1:5432/cifo_db?sslmode=disable"
$env:REDIS_ADDR = "127.0.0.1:6379"
$env:REDIS_PASSWORD = "cifo_redis_secret"
$env:KEYCLOAK_URL = "http://127.0.0.1:8180"
$env:ALLOWED_ORIGINS = "http://localhost:3000,http://127.0.0.1:3000,http://localhost:3001,http://127.0.0.1:3001"
$env:ARGOCD_URL = "https://127.0.0.1:8443"

Set-Location "d:\agent v2\apps\backend"
if (Test-Path ".\server.exe") {
    & ".\server.exe"
} elseif (Test-Path ".\bin\server.exe") {
    & ".\bin\server.exe"
} else {
    go run ./cmd/server/main.go
}

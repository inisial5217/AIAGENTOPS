# Verify all frontend routes
$ErrorActionPreference = "Continue"

$routes = @(
    "/login",
    "/monitoring",
    "/docker",
    "/docker/images",
    "/docker/volumes",
    "/docker/networks"
)

foreach ($r in $routes) {
    try {
        $res = Invoke-WebRequest -Uri "http://localhost:3001$r" -TimeoutSec 15
        Write-Host "Route $r -> HTTP $($res.StatusCode) OK" -ForegroundColor Green
    } catch {
        Write-Host "Route $r -> FAIL: $($_.Exception.Message)" -ForegroundColor Red
    }
}

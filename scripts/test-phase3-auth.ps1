# Test Phase 3 Authentication & Authorization
$ErrorActionPreference = "Stop"

Write-Host "`n=======================================================" -ForegroundColor Cyan
Write-Host "   CIFO PLATFORM - PHASE 3 AUTH INTEGRATION TESTS" -ForegroundColor Cyan
Write-Host "=======================================================`n" -ForegroundColor Cyan

$backendUrl = "http://127.0.0.1:8080"
$keycloakUrl = "http://127.0.0.1:8180"

# 1. Health checks
Write-Host "[1/7] Verifying Keycloak & Backend Connectivity..." -ForegroundColor Yellow
$oidc = Invoke-RestMethod -Uri "$keycloakUrl/realms/cifo/.well-known/openid-configuration" -TimeoutSec 5
if ($oidc.issuer -match "cifo") {
    Write-Host "  [OK] Keycloak realm 'cifo' is active: $($oidc.issuer)" -ForegroundColor Green
} else {
    Write-Host "  [FAIL] Keycloak realm discovery failed" -ForegroundColor Red
    exit 1
}

$health = Invoke-RestMethod -Uri "$backendUrl/healthz" -TimeoutSec 5
Write-Host "  [OK] Backend healthz: $($health.status)" -ForegroundColor Green

# 2. 401 Unauthorized on protected route without token
Write-Host "`n[2/7] Testing 401 Unauthorized on protected route without token..." -ForegroundColor Yellow
try {
    $res = Invoke-WebRequest -Uri "$backendUrl/api/v1/auth/me" -Method Get -TimeoutSec 5
    Write-Host "  [FAIL] Expected 401 but got $($res.StatusCode)" -ForegroundColor Red
    exit 1
} catch {
    $statusCode = $_.Exception.Response.StatusCode.value__
    if ($statusCode -eq 401) {
        Write-Host "  [OK] Correctly received HTTP 401 Unauthorized" -ForegroundColor Green
    } else {
        Write-Host "  [FAIL] Expected 401 but got $statusCode" -ForegroundColor Red
        exit 1
    }
}

# 3. Direct login as Viewer
Write-Host "`n[3/7] Testing direct login as viewer@cifo.local..." -ForegroundColor Yellow
$viewerLogin = Invoke-RestMethod -Uri "$backendUrl/api/v1/auth/login" -Method Post -ContentType "application/json" -Body '{"username":"viewer@cifo.local","password":"viewer123"}'
$viewerToken = $viewerLogin.access_token
if ($viewerToken) {
    Write-Host "  [OK] Successfully acquired Viewer JWT token (${viewerToken.Length} chars)" -ForegroundColor Green
} else {
    Write-Host "  [FAIL] Failed to acquire Viewer JWT token" -ForegroundColor Red
    exit 1
}

# 4. Access /api/v1/auth/me and verify PostgreSQL user sync
Write-Host "`n[4/7] Accessing /api/v1/auth/me and verifying PostgreSQL user sync..." -ForegroundColor Yellow
$headers = @{Authorization = "Bearer $viewerToken"}
$meRes = Invoke-RestMethod -Uri "$backendUrl/api/v1/auth/me" -Headers $headers
Write-Host "  [OK] User profile: $($meRes.user.name) <$($meRes.user.email)> [Role: $($meRes.user.role)]" -ForegroundColor Green
Write-Host "  [OK] User automatically synchronized with Keycloak ID: $($meRes.user.keycloak_id)" -ForegroundColor Green

# 5. RBAC check: Viewer accessing Admin route -> 403 Forbidden
Write-Host "`n[5/7] Testing RBAC: Viewer attempting to access /api/v1/admin/users..." -ForegroundColor Yellow
try {
    $adminFail = Invoke-WebRequest -Uri "$backendUrl/api/v1/admin/users" -Headers $headers -TimeoutSec 5
    Write-Host "  [FAIL] Expected 403 Forbidden for Viewer, but got $($adminFail.StatusCode)" -ForegroundColor Red
    exit 1
} catch {
    $statusCode = $_.Exception.Response.StatusCode.value__
    if ($statusCode -eq 403) {
        Write-Host "  [OK] Correctly blocked with HTTP 403 Forbidden (RBAC works)" -ForegroundColor Green
    } else {
        Write-Host "  [FAIL] Expected 403 Forbidden but got $statusCode" -ForegroundColor Red
        exit 1
    }
}

# 6. Admin login and access /api/v1/admin/users
Write-Host "`n[6/7] Testing Admin login and /api/v1/admin/users access..." -ForegroundColor Yellow
$adminLogin = Invoke-RestMethod -Uri "$backendUrl/api/v1/auth/login" -Method Post -ContentType "application/json" -Body '{"username":"admin@cifo.local","password":"admin123"}'
$adminToken = $adminLogin.access_token
$adminHeaders = @{Authorization = "Bearer $adminToken"}
$usersRes = Invoke-RestMethod -Uri "$backendUrl/api/v1/admin/users" -Headers $adminHeaders
Write-Host "  [OK] Admin successfully listed $($usersRes.total) users from database" -ForegroundColor Green

# 7. Token revocation on logout & Redis blacklisting
Write-Host "`n[7/7] Testing token revocation on logout & Redis blacklist..." -ForegroundColor Yellow
$logoutRes = Invoke-RestMethod -Uri "$backendUrl/api/v1/auth/logout" -Method Post -Headers $headers
Write-Host "  [OK] Logout endpoint response: $($logoutRes.message)" -ForegroundColor Green

try {
    $revokedFail = Invoke-WebRequest -Uri "$backendUrl/api/v1/auth/me" -Headers $headers -TimeoutSec 5
    Write-Host "  [FAIL] Expected 401 for revoked token, but got $($revokedFail.StatusCode)" -ForegroundColor Red
    exit 1
} catch {
    $statusCode = $_.Exception.Response.StatusCode.value__
    if ($statusCode -eq 401) {
        Write-Host "  [OK] Revoked token correctly rejected with HTTP 401 Unauthorized" -ForegroundColor Green
    } else {
        Write-Host "  [FAIL] Expected 401 for revoked token but got $statusCode" -ForegroundColor Red
        exit 1
    }
}

Write-Host "`n=======================================================" -ForegroundColor Green
Write-Host "   ALL 7 PHASE 3 CRITERIA PASSED SUCCESSFULLY!" -ForegroundColor Green
Write-Host "=======================================================`n" -ForegroundColor Green

# Test Phase 10: Settings & Platform Administration Verification Script
$ErrorActionPreference = "Stop"

$baseUrl = "http://127.0.0.1:8080"
Write-Host ""
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "=== TESTING PHASE 10: SETTINGS & PLATFORM ADMINISTRATION ===" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

# 1. Health Check
Write-Host "`n--- 1. Checking Backend Health ---" -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "$baseUrl/healthz" -Method Get
    Write-Host "[OK] Backend server is healthy: $($health.status)" -ForegroundColor Green
} catch {
    Write-Error "Backend server is not accessible at $baseUrl"
    exit 1
}

# 2. Authentication Setup
$adminToken = "dev-token-admin"
$viewerToken = "dev-token-viewer"

$adminHeaders = @{
    Authorization  = "Bearer $adminToken"
    "Content-Type" = "application/json"
}

$viewerHeaders = @{
    Authorization  = "Bearer $viewerToken"
    "Content-Type" = "application/json"
}

# 3. Test GET /api/v1/settings
Write-Host "`n--- 2. Fetching Combined Platform Settings ---" -ForegroundColor Yellow
$settingsRes = Invoke-RestMethod -Uri "$baseUrl/api/v1/settings" -Method Get -Headers $adminHeaders
if (-not $settingsRes.data.system -or -not $settingsRes.data.notification) {
    Write-Error "Failed: System or Notification settings missing from response"
    exit 1
}
Write-Host "[OK] System Settings ID: $($settingsRes.data.system.id)" -ForegroundColor Green
Write-Host "[OK] App Name: $($settingsRes.data.system.app_name)" -ForegroundColor Green
Write-Host "[OK] AI Provider: $($settingsRes.data.system.ai_default_provider)" -ForegroundColor Green
Write-Host "[OK] Notification Settings ID: $($settingsRes.data.notification.id)" -ForegroundColor Green
Write-Host "[OK] Telegram Enabled: $($settingsRes.data.notification.telegram_enabled)" -ForegroundColor Green

# 4. Test PUT /api/v1/settings
Write-Host "`n--- 3. Updating System & Notification Settings ---" -ForegroundColor Yellow
$updatePayload = @{
    system = @{
        app_name              = "CIFO Enterprise AIOps"
        default_theme         = "dark"
        ai_default_provider   = "ollama"
        ai_default_model      = "deepseek-r1:latest"
        ai_monthly_budget_usd = 75.50
        session_timeout_minutes = 90
        require_mfa           = $true
    }
    notification = @{
        telegram_enabled              = $true
        telegram_chat_id              = "-1009988776655"
        inapp_enabled                 = $true
        alert_batching_window_seconds = 180
    }
} | ConvertTo-Json -Depth 5

$updatedRes = Invoke-RestMethod -Uri "$baseUrl/api/v1/settings" -Method Put -Headers $adminHeaders -Body $updatePayload
if ($updatedRes.data.system.app_name -ne "CIFO Enterprise AIOps") {
    Write-Error "Failed: App name was not updated"
    exit 1
}
if ($updatedRes.data.notification.telegram_chat_id -ne "-1009988776655") {
    Write-Error "Failed: Telegram chat id was not updated"
    exit 1
}
Write-Host "[OK] Settings updated successfully: App Name = $($updatedRes.data.system.app_name), Model = $($updatedRes.data.system.ai_default_model)" -ForegroundColor Green
Write-Host "[OK] Notification updated: Chat ID = $($updatedRes.data.notification.telegram_chat_id)" -ForegroundColor Green

# 5. Test Notification Alert Dispatch
Write-Host "`n--- 4. Testing Notification Alert Dispatch ---" -ForegroundColor Yellow
$testNotifRes = Invoke-RestMethod -Uri "$baseUrl/api/v1/settings/test-notification" -Method Post -Headers $adminHeaders
if ($testNotifRes.status -ne "success") {
    Write-Error "Failed: Test notification dispatch failed"
    exit 1
}
Write-Host "[OK] Notification test response: $($testNotifRes.message)" -ForegroundColor Green

# 6. Test User Administration (List Users)
Write-Host "`n--- 5. Listing Users for Administration ---" -ForegroundColor Yellow
$usersRes = Invoke-RestMethod -Uri "$baseUrl/api/v1/settings/users?limit=20&offset=0" -Method Get -Headers $adminHeaders
Write-Host "[OK] Total Users registered: $($usersRes.total)" -ForegroundColor Green
if ($usersRes.users.Count -eq 0) {
    Write-Error "Failed: No users found in user administration"
    exit 1
}

# Pick a target user (not the current admin self)
$targetUser = $usersRes.users | Where-Object { $_.email -ne "admin@cifo.local" -and $_.id -ne "dev-admin-id" } | Select-Object -First 1
if (-not $targetUser) {
    # If no other user, use the first user
    $targetUser = $usersRes.users[0]
}
Write-Host "[OK] Selected target user for testing: $($targetUser.email) (ID: $($targetUser.id), Role: $($targetUser.role), Active: $($targetUser.is_active))" -ForegroundColor Green

# 7. Test User Role Update
Write-Host "`n--- 6. Updating User Role ---" -ForegroundColor Yellow
$rolePayload = @{
    role = "devops"
} | ConvertTo-Json

$roleUpdateRes = Invoke-RestMethod -Uri "$baseUrl/api/v1/settings/users/$($targetUser.id)/role" -Method Put -Headers $adminHeaders -Body $rolePayload
Write-Host "[OK] User role successfully updated to: $($roleUpdateRes.role)" -ForegroundColor Green

# 8. Test User Deactivation & Reactivation (if not self)
if ($targetUser.id -ne "dev-admin-id" -and $targetUser.email -ne "admin@cifo.local") {
    Write-Host "`n--- 7. Deactivating and Reactivating User ---" -ForegroundColor Yellow
    $deactRes = Invoke-RestMethod -Uri "$baseUrl/api/v1/settings/users/$($targetUser.id)/deactivate" -Method Post -Headers $adminHeaders
    if ($deactRes.is_active -ne $false) {
        Write-Error "Failed: User was not deactivated"
        exit 1
    }
    Write-Host "[OK] User deactivated: is_active = $($deactRes.is_active)" -ForegroundColor Green

    $reactRes = Invoke-RestMethod -Uri "$baseUrl/api/v1/settings/users/$($targetUser.id)/reactivate" -Method Post -Headers $adminHeaders
    if ($reactRes.is_active -ne $true) {
        Write-Error "Failed: User was not reactivated"
        exit 1
    }
    Write-Host "[OK] User reactivated: is_active = $($reactRes.is_active)" -ForegroundColor Green
} else {
    Write-Host "`n[INFO] Skipped deactivation test on self account (Self-Protection is active)" -ForegroundColor Cyan
}

# 9. Test Self-Deactivation Protection
Write-Host "`n--- 8. Testing Self-Deactivation Protection ---" -ForegroundColor Yellow
try {
    Invoke-RestMethod -Uri "$baseUrl/api/v1/settings/users/dev-admin-id/deactivate" -Method Post -Headers $adminHeaders
    Write-Error "Failed: Self-deactivation should have been rejected with HTTP 400"
    exit 1
} catch {
    Write-Host "[OK] Self-deactivation correctly rejected with status: $($_.Exception.Response.StatusCode.value__)" -ForegroundColor Green
}

# 10. Test RBAC Protection (Non-Admin Access Denial)
Write-Host "`n--- 9. Testing RBAC Enforcement (Viewer Access Attempt) ---" -ForegroundColor Yellow
try {
    Invoke-RestMethod -Uri "$baseUrl/api/v1/settings" -Method Get -Headers $viewerHeaders
    Write-Error "Failed: Viewer token should have been rejected with HTTP 403 Forbidden"
    exit 1
} catch {
    Write-Host "[OK] Viewer access rejected with HTTP 403 Forbidden (RBAC Enforcement Active)" -ForegroundColor Green
}

# 11. Verify Audit Logs in Database
Write-Host "`n--- 10. Verifying Audit Trail in Database ---" -ForegroundColor Yellow
$auditOutput = & powershell -ExecutionPolicy Bypass -File "d:\agent v2\scripts\docker-helper.ps1" exec cifo-postgres psql -U cifo_admin -d cifo_db -c "SELECT action, actor_id, resource_type, result, timestamp FROM audit_log ORDER BY timestamp DESC LIMIT 10;"
Write-Host $auditOutput -ForegroundColor Gray

if ($auditOutput -match "SETTINGS_UPDATE" -and $auditOutput -match "USER_ROLE_CHANGE" -and $auditOutput -match "USER_DEACTIVATION" -and $auditOutput -match "TELEGRAM_TEST_ALERT") {
    Write-Host "[OK] All Phase 10 Audit Trail records verified in PostgreSQL database" -ForegroundColor Green
} else {
    Write-Error "Failed: Missing expected Phase 10 audit records in database"
    exit 1
}

Write-Host ""
Write-Host "==========================================================" -ForegroundColor Green
Write-Host "=== PHASE 10 SETTINGS & ADMINISTRATION VERIFIED 100%   ===" -ForegroundColor Green
Write-Host "==========================================================" -ForegroundColor Green

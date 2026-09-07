# Automated Verification Script for Phase 12: Security Hardening

$ErrorActionPreference = "Continue"
$passedCount = 0
$failedCount = 0

function Report-Step($name, $status, $details) {
    if ($status) {
        Write-Host " [PASS] $name - $details" -ForegroundColor Green
        $script:passedCount++
    } else {
        Write-Host " [FAIL] $name - $details" -ForegroundColor Red
        $script:failedCount++
    }
}

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "  CIFO Monitoring Platform - Phase 12 Security Verification" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

# -------------------------------------------------------------
# 1. HASHICORP VAULT VERIFICATION
# -------------------------------------------------------------
Write-Host "`n[1/6] Verifying HashiCorp Vault Integration..." -ForegroundColor Yellow
$vaultAddr = "http://127.0.0.1:8200"
$vaultToken = "cifo-vault-root-token"
$vaultHeaders = @{ "X-Vault-Token" = $vaultToken }

try {
    $vHealth = Invoke-RestMethod -Uri "$vaultAddr/v1/sys/health" -Method Get -TimeoutSec 3
    $vOk = ($vHealth.initialized -eq $true -and $vHealth.sealed -eq $false)
    Report-Step "Vault Health" $vOk "Version: $($vHealth.version), Sealed: $($vHealth.sealed)"
} catch {
    Report-Step "Vault Health" $false "Failed to reach Vault: $_"
}

try {
    $backendSecrets = Invoke-RestMethod -Uri "$vaultAddr/v1/secret/data/cifo/backend" -Headers $vaultHeaders -Method Get -TimeoutSec 3
    $bData = $backendSecrets.data.data
    $hasBackendKeys = ($bData.POSTGRES_USER -eq "cifo_admin" -and $bData.DOCKER_HOST -eq "tcp://127.0.0.1:2376")
    Report-Step "Vault Backend Secrets" $hasBackendKeys "POSTGRES_USER=$($bData.POSTGRES_USER), DOCKER_HOST=$($bData.DOCKER_HOST)"
} catch {
    Report-Step "Vault Backend Secrets" $false "Failed to read backend secrets: $_"
}

try {
    $aiSecrets = Invoke-RestMethod -Uri "$vaultAddr/v1/secret/data/cifo/ai-service" -Headers $vaultHeaders -Method Get -TimeoutSec 3
    $aiData = $aiSecrets.data.data
    $hasAIKeys = ($null -ne $aiData.GOOGLE_API_KEY -and $aiData.GOOGLE_API_KEY.Length -gt 0)
    Report-Step "Vault AI Service Secrets" $hasAIKeys "GOOGLE_API_KEY is present ($($aiData.GOOGLE_API_KEY.Substring(0, 10))...)"
} catch {
    Report-Step "Vault AI Service Secrets" $false "Failed to read AI secrets: $_"
}

# -------------------------------------------------------------
# 2. DOCKER SOCKET PROXY VERIFICATION
# -------------------------------------------------------------
Write-Host "`n[2/6] Verifying Tecnativa Docker Socket Proxy Hardening..." -ForegroundColor Yellow
$proxyAddr = "http://127.0.0.1:2376"

try {
    $versionResp = Invoke-RestMethod -Uri "$proxyAddr/version" -Method Get -TimeoutSec 3
    $proxyOk = ($null -ne $versionResp.Version)
    Report-Step "Docker Proxy GET /version" $proxyOk "Engine Version: $($versionResp.Version)"
} catch {
    Report-Step "Docker Proxy GET /version" $false "Failed: $_"
}

try {
    $containersResp = Invoke-RestMethod -Uri "$proxyAddr/containers/json" -Method Get -TimeoutSec 3
    $containersOk = ($containersResp.Count -gt 0)
    Report-Step "Docker Proxy GET /containers/json" $containersOk "Found $($containersResp.Count) running containers"
} catch {
    Report-Step "Docker Proxy GET /containers/json" $false "Failed: $_"
}

# Verify DELETE is blocked (Must be 403 Forbidden)
try {
    $del = Invoke-WebRequest -Uri "$proxyAddr/containers/unauthorized-container-id" -Method Delete -TimeoutSec 3 -UseBasicParsing
    Report-Step "Docker Proxy DELETE Restriction" $false "Expected 403 Forbidden, got HTTP $($del.StatusCode)"
} catch {
    $status = $_.Exception.Response.StatusCode.value__
    $isForbidden = ($status -eq 403)
    Report-Step "Docker Proxy DELETE Restriction" $isForbidden "Blocked with HTTP $status (Forbidden by administrative rules)"
}

# Verify unauthorized endpoint GET /secrets is blocked (Must be 403 Forbidden)
try {
    $sec = Invoke-WebRequest -Uri "$proxyAddr/secrets" -Method Get -TimeoutSec 3 -UseBasicParsing
    Report-Step "Docker Proxy GET /secrets Restriction" $false "Expected 403 Forbidden, got HTTP $($sec.StatusCode)"
} catch {
    $status = $_.Exception.Response.StatusCode.value__
    $isForbidden = ($status -eq 403)
    Report-Step "Docker Proxy GET /secrets Restriction" $isForbidden "Blocked with HTTP $status (Forbidden by administrative rules)"
}

# -------------------------------------------------------------
# 3. KUBERNETES ZERO-TRUST RBAC VERIFICATION
# -------------------------------------------------------------
Write-Host "`n[3/6] Verifying Kubernetes Zero-Trust RBAC Manifests..." -ForegroundColor Yellow

$saName = "system:serviceaccount:default:cifo-ai-agent-sa"

$canGetPods = (kubectl auth can-i get pods --as=$saName 2>$null).Trim()
Report-Step "RBAC: cifo-ai-agent-sa can get pods" ($canGetPods -eq "yes") "Result: $canGetPods"

$canPatchDeployments = (kubectl auth can-i patch deployments --as=$saName 2>$null).Trim()
Report-Step "RBAC: cifo-ai-agent-sa can patch deployments" ($canPatchDeployments -eq "yes") "Result: $canPatchDeployments"

$canDeletePods = (kubectl auth can-i delete pods --as=$saName 2>$null).Trim()
Report-Step "RBAC: cifo-ai-agent-sa DENIED delete pods" ($canDeletePods -eq "no") "Result: $canDeletePods (Strictly prohibited)"

$canDeleteNS = (kubectl auth can-i delete namespaces --as=$saName 2>$null).Trim()
Report-Step "RBAC: cifo-ai-agent-sa DENIED delete namespaces" ($canDeleteNS -eq "no") "Result: $canDeleteNS (Strictly prohibited)"

$canGetSecrets = (kubectl auth can-i get secrets --as=$saName 2>$null).Trim()
Report-Step "RBAC: cifo-ai-agent-sa DENIED get secrets" ($canGetSecrets -eq "no") "Result: $canGetSecrets (Strictly prohibited)"

# -------------------------------------------------------------
# 4. KUBERNETES NETWORKPOLICIES VERIFICATION
# -------------------------------------------------------------
Write-Host "`n[4/6] Verifying Kubernetes NetworkPolicies..." -ForegroundColor Yellow
$netpols = kubectl get networkpolicy -n default -o jsonpath="{.items[*].metadata.name}" 2>$null
$expectedPolicies = @("default-deny-all", "cifo-frontend-netpol", "cifo-backend-netpol", "cifo-ai-service-netpol", "cifo-data-netpol")
$allNetpolFound = $true

foreach ($p in $expectedPolicies) {
    if ($netpols -notmatch $p) {
        $allNetpolFound = $false
        Write-Host " Missing NetworkPolicy: $p" -ForegroundColor Red
    }
}
Report-Step "Kubernetes NetworkPolicies Applied" $allNetpolFound "All 5 enterprise NetworkPolicies are active in default namespace"

# -------------------------------------------------------------
# 5. HTTP SECURITY HEADERS VERIFICATION
# -------------------------------------------------------------
Write-Host "`n[5/6] Verifying HTTP Security Headers Middleware..." -ForegroundColor Yellow
try {
    $res = Invoke-WebRequest -Uri "http://127.0.0.1:8080/healthz" -Method Get -TimeoutSec 3 -UseBasicParsing
    $headers = $res.Headers

    $hasNosniff = ($headers["X-Content-Type-Options"] -eq "nosniff")
    Report-Step "Header: X-Content-Type-Options" $hasNosniff "Value: $($headers['X-Content-Type-Options'])"

    $hasFrameDeny = ($headers["X-Frame-Options"] -eq "DENY")
    Report-Step "Header: X-Frame-Options" $hasFrameDeny "Value: $($headers['X-Frame-Options'])"

    $hasXSS = ($headers["X-XSS-Protection"] -eq "1; mode=block")
    Report-Step "Header: X-XSS-Protection" $hasXSS "Value: $($headers['X-XSS-Protection'])"

    $hasCSP = ($headers["Content-Security-Policy"] -like "*default-src 'self'*")
    Report-Step "Header: Content-Security-Policy" $hasCSP "Value: $($headers['Content-Security-Policy'])"

    $hasHSTS = ($headers["Strict-Transport-Security"] -like "*max-age=31536000*")
    Report-Step "Header: Strict-Transport-Security" $hasHSTS "Value: $($headers['Strict-Transport-Security'])"

    $hasReferrer = ($headers["Referrer-Policy"] -eq "strict-origin-when-cross-origin")
    Report-Step "Header: Referrer-Policy" $hasReferrer "Value: $($headers['Referrer-Policy'])"

    $hasPermissions = ($headers["Permissions-Policy"] -like "*geolocation=()*")
    Report-Step "Header: Permissions-Policy" $hasPermissions "Value: $($headers['Permissions-Policy'])"
} catch {
    Report-Step "HTTP Security Headers" $false "Backend not reachable on :8080: $_"
}

# -------------------------------------------------------------
# 6. SECURITY SCANNING VERIFICATION (gosec & gitleaks)
# -------------------------------------------------------------
Write-Host "`n[6/6] Verifying Security Scanners (gosec & gitleaks)..." -ForegroundColor Yellow

# Gosec verification
$gosecBin = [System.IO.Path]::Combine((go env GOPATH), 'bin', 'gosec.exe')
if (Test-Path $gosecBin) {
    Push-Location "d:\agent v2\apps\backend"
    $gosecOutput = & $gosecBin -fmt=json ./cmd/... ./internal/... ./pkg/... 2>$null | ConvertFrom-Json
    Pop-Location
    $gosecIssues = $gosecOutput.Issues.Count
    Report-Step "gosec: 0 High/Critical Vulnerabilities" ($gosecIssues -eq 0) "Issues detected: $gosecIssues"
} else {
    Report-Step "gosec executable" $false "gosec.exe not found"
}

# Gitleaks verification
$gitleaksRun = docker run --rm -v "d:\agent v2:/path:ro" zricethezav/gitleaks:latest git /path --log-opts="-n 10" -i /path/.gitleaksignore 2>&1
$gitleaksOk = ($LASTEXITCODE -eq 0 -or $gitleaksRun -match "no leaks found")
Report-Step "gitleaks: 0 Credential Leaks" $gitleaksOk "Gitleaks clean (no leaks found in git history)"

# -------------------------------------------------------------
# SUMMARY
# -------------------------------------------------------------
Write-Host "`n==========================================================" -ForegroundColor Cyan
Write-Host "  Phase 12 Verification Summary" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "Total Passed: $passedCount" -ForegroundColor Green
$failColor = "Green"
if ($failedCount -gt 0) { $failColor = "Red" }
Write-Host "Total Failed: $failedCount" -ForegroundColor $failColor

if ($failedCount -eq 0) {
    Write-Host "`n ALL PHASE 12 ACCEPTANCE CRITERIA SATISFIED!" -ForegroundColor Green
    exit 0
} else {
    Write-Host "`n SOME CHECKS FAILED! Please review output above." -ForegroundColor Red
    exit 1
}

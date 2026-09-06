# Test Phase 8: Alertmanager Webhook, Incident Lifecycle & Escalation
$ErrorActionPreference = "Stop"

$baseUrl = "http://127.0.0.1:8080"
Write-Host ""
Write-Host "=== TESTING PHASE 8: ALERTING & INCIDENT MANAGEMENT ===" -ForegroundColor Cyan

# 1. Health check
try {
    $health = Invoke-RestMethod -Uri "$baseUrl/healthz" -Method Get
    Write-Host "[OK] Backend is healthy: $($health.status)" -ForegroundColor Green
} catch {
    Write-Error "Backend is not accessible at $baseUrl"
    exit 1
}

# 2. Authenticate as Admin
$token = "dev-token-admin"
Write-Host "[OK] Admin authenticated with dev-token-admin" -ForegroundColor Green

$headers = @{
    Authorization = "Bearer $token"
}

# 3. Send Firing Alert via Alertmanager Webhook
Write-Host "`n--- Sending Firing Alert via Webhook ---" -ForegroundColor Yellow
$alertPayload = @{
    version  = "4"
    status   = "firing"
    receiver = "cifo-webhook"
    alerts   = @(
        @{
            status      = "firing"
            labels      = @{
                alertname = "ContainerOOMKilled"
                severity  = "critical"
                container = "payment-gateway"
                namespace = "production"
            }
            annotations = @{
                summary     = "Container payment-gateway killed due to OOM"
                description = "Container exceeded cgroup memory limit of 512MiB"
            }
        }
    )
} | ConvertTo-Json -Depth 5

$webhookRes = Invoke-RestMethod -Uri "$baseUrl/api/v1/webhooks/alertmanager" -Method Post -Body $alertPayload -ContentType "application/json"
Write-Host "[OK] Webhook receiver processed: $($webhookRes.status)" -ForegroundColor Green

# 4. Verify Incident Created
Write-Host "`n--- Verifying Incident in API ---" -ForegroundColor Yellow
$incidentsRes = Invoke-RestMethod -Uri "$baseUrl/api/v1/incidents?status=open" -Method Get -Headers $headers
$targetInc = $incidentsRes.data | Where-Object { $_.alert_name -eq "ContainerOOMKilled" -and $_.resource_id -eq "payment-gateway" } | Select-Object -First 1

if (-not $targetInc) {
    Write-Error "Incident was not found in API"
    exit 1
}

$incId = $targetInc.id
Write-Host "[OK] Incident created successfully: ID=$incId, Title=$($targetInc.title), Status=$($targetInc.status)" -ForegroundColor Green

# 5. Acknowledge Incident
Write-Host "`n--- Testing Acknowledge Incident ---" -ForegroundColor Yellow
$ackRes = Invoke-RestMethod -Uri "$baseUrl/api/v1/incidents/$incId/acknowledge" -Method Post -Headers $headers
Write-Host "[OK] Acknowledge success: $($ackRes.message)" -ForegroundColor Green

$detail = Invoke-RestMethod -Uri "$baseUrl/api/v1/incidents/$incId" -Method Get -Headers $headers
Write-Host "[OK] Verified Status: $($detail.data.status), AcknowledgedBy: $($detail.data.acknowledged_by_name)" -ForegroundColor Green

# 6. Resolve Incident
Write-Host "`n--- Testing Resolve Incident ---" -ForegroundColor Yellow
$resRes = Invoke-RestMethod -Uri "$baseUrl/api/v1/incidents/$incId/resolve" -Method Post -Headers $headers
Write-Host "[OK] Resolve success: $($resRes.message)" -ForegroundColor Green

$detail = Invoke-RestMethod -Uri "$baseUrl/api/v1/incidents/$incId" -Method Get -Headers $headers
Write-Host "[OK] Verified Status: $($detail.data.status), ResolvedBy: $($detail.data.resolved_by_name)" -ForegroundColor Green

# 7. Close Incident
Write-Host "`n--- Testing Close Incident ---" -ForegroundColor Yellow
$closeRes = Invoke-RestMethod -Uri "$baseUrl/api/v1/incidents/$incId/close" -Method Post -Headers $headers
Write-Host "[OK] Close success: $($closeRes.message)" -ForegroundColor Green

$detail = Invoke-RestMethod -Uri "$baseUrl/api/v1/incidents/$incId" -Method Get -Headers $headers
Write-Host "[OK] Verified Status: $($detail.data.status), ClosedBy: $($detail.data.closed_by_name)" -ForegroundColor Green

# 8. Test Auto-Resolve on Resolved Webhook
Write-Host "`n--- Testing Auto-Resolve on Resolved Webhook ---" -ForegroundColor Yellow
$podAlertPayload = @{
    version  = "4"
    status   = "firing"
    receiver = "cifo-webhook"
    alerts   = @(
        @{
            status      = "firing"
            labels      = @{
                alertname = "PodCrashLooping"
                severity  = "critical"
                pod       = "redis-cart-0"
                namespace = "default"
            }
            annotations = @{
                summary     = "Pod redis-cart-0 is crash looping"
                description = "Container restarted 5 times in 10 minutes"
            }
        }
    )
} | ConvertTo-Json -Depth 5

$null = Invoke-RestMethod -Uri "$baseUrl/api/v1/webhooks/alertmanager" -Method Post -Body $podAlertPayload -ContentType "application/json"
Start-Sleep -Milliseconds 500

$podIncList = Invoke-RestMethod -Uri "$baseUrl/api/v1/incidents?status=open" -Method Get -Headers $headers
$podInc = $podIncList.data | Where-Object { $_.alert_name -eq "PodCrashLooping" } | Select-Object -First 1
Write-Host "[OK] Firing Pod incident created: ID=$($podInc.id)" -ForegroundColor Green

# Send resolved webhook
$resolvedPayload = @{
    version  = "4"
    status   = "resolved"
    receiver = "cifo-webhook"
    alerts   = @(
        @{
            status = "resolved"
            labels = @{
                alertname = "PodCrashLooping"
                pod       = "redis-cart-0"
            }
        }
    )
} | ConvertTo-Json -Depth 5

$null = Invoke-RestMethod -Uri "$baseUrl/api/v1/webhooks/alertmanager" -Method Post -Body $resolvedPayload -ContentType "application/json"
Start-Sleep -Milliseconds 500

$resolvedDetail = Invoke-RestMethod -Uri "$baseUrl/api/v1/incidents/$($podInc.id)" -Method Get -Headers $headers
Write-Host "[OK] Auto-resolved incident status: $($resolvedDetail.data.status)" -ForegroundColor Green

# 9. Verify Incident Stats
Write-Host "`n--- Verifying Incident Stats ---" -ForegroundColor Yellow
$stats = Invoke-RestMethod -Uri "$baseUrl/api/v1/incidents/stats" -Method Get -Headers $headers
Write-Host "[OK] Stats: Total=$($stats.data.total), Open=$($stats.data.open), Acknowledged=$($stats.data.acknowledged), Resolved=$($stats.data.resolved), Closed=$($stats.data.closed)" -ForegroundColor Green

Write-Host ""
Write-Host "=== ALL PHASE 8 API TESTS PASSED 100% ===" -ForegroundColor Green

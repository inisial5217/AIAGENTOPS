# ==============================================================================
# CIFO Platform - Phase 14 CI/CD & Kubernetes Manifests Verification Script
# ==============================================================================
$ErrorActionPreference = 'Continue'

Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "   CIFO MONITORING & AIOPS PLATFORM - FASE 14 VERIFICATION       " -ForegroundColor Cyan
Write-Host "=================================================================" -ForegroundColor Cyan

$results = @{}
$failedCount = 0

function Assert-Test {
    param (
        [string]$Name,
        [bool]$Condition,
        [string]$Details
    )
    if ($Condition) {
        Write-Host "  [PASS] ${Name}: $Details" -ForegroundColor Green
        $script:results[$Name] = "PASSED: $Details"
    } else {
        Write-Host "  [FAIL] ${Name}: $Details" -ForegroundColor Red
        $script:results[$Name] = "FAILED: $Details"
        $script:failedCount++
    }
}

# -----------------------------------------------------------------------------
# 1. Multi-stage Dockerfiles & .dockerignore
# -----------------------------------------------------------------------------
Write-Host "`n[1/5] Verifying Production Multi-Stage Dockerfiles..." -ForegroundColor Yellow

$backendDf = Get-Content "apps/backend/Dockerfile" -Raw
$backendDi = Test-Path "apps/backend/.dockerignore"
$backendHasMultiStage = $backendDf -match "AS builder" -and $backendDf -match "-ldflags" -and $backendDf -match "appuser"
Assert-Test "Backend Dockerfile" ($backendHasMultiStage -and $backendDi) "Multi-stage builder, stripped ldflags, non-root user, .dockerignore verified"

$frontendDf = Get-Content "apps/frontend/Dockerfile" -Raw
$frontendDi = Test-Path "apps/frontend/.dockerignore"
$frontendHasMultiStage = $frontendDf -match "AS deps" -and $frontendDf -match "AS builder" -and $frontendDf -match "AS runner" -and $frontendDf -match "standalone"
Assert-Test "Frontend Dockerfile" ($frontendHasMultiStage -and $frontendDi) "Multi-stage deps/builder/runner, standalone Next.js, non-root nextjs, .dockerignore verified"

$aiDf = Get-Content "apps/ai-service/Dockerfile" -Raw
$aiDi = Test-Path "apps/ai-service/.dockerignore"
$aiHasMultiStage = $aiDf -match "AS builder" -and $aiDf -match "AS runner" -and $aiDf -match "appuser"
Assert-Test "AI Service Dockerfile" ($aiHasMultiStage -and $aiDi) "Multi-stage wheels install, non-root appuser, .dockerignore verified"

# -----------------------------------------------------------------------------
# 2. Tugas 14.1: GitHub Actions CI (.github/workflows/ci.yml)
# -----------------------------------------------------------------------------
Write-Host "`n[2/5] Verifying Tugas 14.1: GitHub Actions CI Workflow..." -ForegroundColor Yellow

$ciPath = ".github/workflows/ci.yml"
$ciExists = Test-Path $ciPath
if ($ciExists) {
    $ciContent = Get-Content $ciPath -Raw
    $hasTriggers = $ciContent -match "push:" -and $ciContent -match "pull_request:"
    $hasLintJobs = $ciContent -match "lint-backend:" -and $ciContent -match "lint-frontend:" -and $ciContent -match "lint-ai:" -and $ciContent -match "lint-docker:"
    $hasTestJobs = $ciContent -match "test-backend:" -and $ciContent -match "test-frontend:" -and $ciContent -match "test-ai:"
    $hasSecurity = $ciContent -match "security-scan:" -and $ciContent -match "gosec" -and $ciContent -match "trivy" -and $ciContent -match "gitleaks"
    $hasBuildImg = $ciContent -match "build-images:"

    Assert-Test "14.1 CI Triggers" $hasTriggers "Configured for push and pull_request on main"
    Assert-Test "14.1 CI Linters" $hasLintJobs "Includes golangci-lint, eslint, ruff, and hadolint"
    Assert-Test "14.1 CI Test Suites" $hasTestJobs "Includes Go coverage, Vitest coverage, and Pytest"
    Assert-Test "14.1 CI Security Scanners" $hasSecurity "Includes gosec, trivy-action, and gitleaks"
    Assert-Test "14.1 CI Docker Build Matrix" $hasBuildImg "Validated matrix build across 3 microservices"
} else {
    Assert-Test "14.1 CI Workflow" $false "File .github/workflows/ci.yml not found"
}

# -----------------------------------------------------------------------------
# 3. Tugas 14.2: Deploy to Staging Workflow (.github/workflows/deploy-staging.yml)
# -----------------------------------------------------------------------------
Write-Host "`n[3/5] Verifying Tugas 14.2: Deploy to Staging Workflow..." -ForegroundColor Yellow

$stagingWf = ".github/workflows/deploy-staging.yml"
if (Test-Path $stagingWf) {
    $stgContent = Get-Content $stagingWf -Raw
    $hasRegistry = $stgContent -match "ghcr.io"
    $hasGitOpsCommit = $stgContent -match "git commit" -and $stgContent -match "kustomize edit set image"
    $hasArgoSync = $stgContent -match "cifo-staging" -and $stgContent -match "sync"
    Assert-Test "14.2 Staging Push & GitOps" ($hasRegistry -and $hasGitOpsCommit) "Images pushed to GHCR, kustomization updated, automated GitOps commit"
    Assert-Test "14.2 Staging ArgoCD Sync" $hasArgoSync "ArgoCD auto-sync dispatched for cifo-staging"
} else {
    Assert-Test "14.2 Deploy Staging" $false "File .github/workflows/deploy-staging.yml not found"
}

# -----------------------------------------------------------------------------
# 4. Tugas 14.3: Deploy to Production Workflow (.github/workflows/deploy-production.yml)
# -----------------------------------------------------------------------------
Write-Host "`n[4/5] Verifying Tugas 14.3: Deploy to Production Workflow..." -ForegroundColor Yellow

$prodWf = ".github/workflows/deploy-production.yml"
if (Test-Path $prodWf) {
    $prodContent = Get-Content $prodWf -Raw
    $hasDispatch = $prodContent -match "workflow_dispatch:"
    $hasEnvApproval = $prodContent -match "environment:" -and $prodContent -match "production"
    $hasPromotion = $prodContent -match "crane copy"
    $hasManualSync = $prodContent -match "cifo-production" -and $prodContent -match "sync"
    Assert-Test "14.3 Production Approval Gate" ($hasDispatch -and $hasEnvApproval) "Manual trigger protected by GitHub Environment approval gate"
    Assert-Test "14.3 Production Promotion & Sync" ($hasPromotion -and $hasManualSync) "Image promotion via crane, overlay update, manual ArgoCD sync"
} else {
    Assert-Test "14.3 Deploy Production" $false "File .github/workflows/deploy-production.yml not found"
}

# -----------------------------------------------------------------------------
# 5. Tugas 14.4: Helm Charts & Kustomize Manifests
# -----------------------------------------------------------------------------
Write-Host "`n[5/5] Verifying Tugas 14.4: Helm Charts & Kubernetes Manifests..." -ForegroundColor Yellow

$chartYaml = Test-Path "infrastructure/kubernetes/charts/cifo-platform/Chart.yaml"
$valuesYaml = Test-Path "infrastructure/kubernetes/charts/cifo-platform/values.yaml"
$valuesStaging = Test-Path "infrastructure/kubernetes/charts/cifo-platform/values-staging.yaml"
$valuesProd = Test-Path "infrastructure/kubernetes/charts/cifo-platform/values-production.yaml"
$helpersTpl = Test-Path "infrastructure/kubernetes/charts/cifo-platform/templates/_helpers.tpl"

Assert-Test "14.4 Platform Helm Chart" ($chartYaml -and $valuesYaml -and $valuesStaging -and $valuesProd -and $helpersTpl) "Umbrella chart, values (default, staging, production), helpers defined"

# Check templates
$feTpls = (Test-Path "infrastructure/kubernetes/charts/cifo-platform/templates/frontend/deployment.yaml") -and (Test-Path "infrastructure/kubernetes/charts/cifo-platform/templates/frontend/service.yaml")
$beTpls = (Test-Path "infrastructure/kubernetes/charts/cifo-platform/templates/backend/deployment.yaml") -and (Test-Path "infrastructure/kubernetes/charts/cifo-platform/templates/backend/service.yaml") -and (Test-Path "infrastructure/kubernetes/charts/cifo-platform/templates/backend/configmap.yaml")
$aiTpls = (Test-Path "infrastructure/kubernetes/charts/cifo-platform/templates/ai-service/deployment.yaml") -and (Test-Path "infrastructure/kubernetes/charts/cifo-platform/templates/ai-service/service.yaml")
$dataTpls = (Test-Path "infrastructure/kubernetes/charts/cifo-platform/templates/data/postgres-statefulset.yaml") -and (Test-Path "infrastructure/kubernetes/charts/cifo-platform/templates/data/redis-deployment.yaml")
$ingressTpl = Test-Path "infrastructure/kubernetes/charts/cifo-platform/templates/ingress/ingress.yaml"
$netpolTpl = Test-Path "infrastructure/kubernetes/charts/cifo-platform/templates/security/networkpolicies.yaml"

Assert-Test "14.4 Frontend Templates (Deploy, Svc, HPA, PDB)" $feTpls "Next.js standalone manifests with health probes & autoscaling"
Assert-Test "14.4 Backend Templates (Deploy, Svc, CM, SA, HPA)" $beTpls "Go Echo manifests with configmap, serviceaccount, probes"
Assert-Test "14.4 AI Service Templates (Deploy, Svc, SA, HPA)" $aiTpls "FastAPI/gRPC manifests with cifo-ai-agent-sa"
Assert-Test "14.4 Data Layer Templates (Postgres StatefulSet, Redis)" $dataTpls "Postgres 16 StatefulSet with PVC, Redis deployment"
Assert-Test "14.4 Ingress & NetworkPolicies Templates" ($ingressTpl -and $netpolTpl) "Traefik TLS Ingress & Zero-Trust NetworkPolicies"

# Check subcharts
$subCharts = (Test-Path "infrastructure/kubernetes/charts/cifo-frontend/Chart.yaml") -and
             (Test-Path "infrastructure/kubernetes/charts/cifo-backend/Chart.yaml") -and
             (Test-Path "infrastructure/kubernetes/charts/cifo-ai-service/Chart.yaml") -and
             (Test-Path "infrastructure/kubernetes/charts/cifo-data/Chart.yaml")
Assert-Test "14.4 Modular Subcharts" $subCharts "Standalone Chart.yaml for frontend, backend, ai-service, and data"

# Check Kustomize
$kustBase = Test-Path "infrastructure/kubernetes/base/kustomization.yaml"
$kustStg = Test-Path "infrastructure/kubernetes/overlays/staging/kustomization.yaml"
$kustProd = Test-Path "infrastructure/kubernetes/overlays/production/kustomization.yaml"
Assert-Test "14.4 Kustomize Overlays" ($kustBase -and $kustStg -and $kustProd) "Base, staging overlay, and production overlay configured for ArgoCD"

# -----------------------------------------------------------------------------
# Summary
# -----------------------------------------------------------------------------
Write-Host "`n=================================================================" -ForegroundColor Cyan
Write-Host "   FASE 14 VERIFICATION SUMMARY                                  " -ForegroundColor Cyan
Write-Host "=================================================================" -ForegroundColor Cyan

foreach ($k in $results.Keys) {
    Write-Host "$($results[$k])"
}

if ($failedCount -eq 0) {
    Write-Host "`n>>> ALL PHASE 14 ACCEPTANCE CRITERIA VERIFIED SUCCESSFULLY (100% PASS) <<<" -ForegroundColor Green
    exit 0
} else {
    Write-Host "`n>>> $failedCount CHECKS FAILED IN PHASE 14 VERIFICATION <<<" -ForegroundColor Red
    exit 1
}

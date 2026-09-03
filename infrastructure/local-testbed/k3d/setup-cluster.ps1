# Setup K3d local cluster and deploy ArgoCD
$ErrorActionPreference = "Stop"

Write-Host "==> Checking prerequisites..." -ForegroundColor Cyan
if (-not (Get-Command k3d -ErrorAction SilentlyContinue)) {
    Write-Error "k3d CLI is not found on PATH. Run 'winget install k3d.k3d' first."
}
if (-not (Get-Command kubectl -ErrorAction SilentlyContinue)) {
    Write-Error "kubectl CLI is not found on PATH."
}

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $scriptDir

Write-Host "==> Checking for existing cifo-dev cluster..." -ForegroundColor Cyan
$existing = k3d cluster list --no-headers | Select-String "cifo-dev"
if ($existing) {
    Write-Host "Deleting existing cifo-dev cluster..." -ForegroundColor Yellow
    k3d cluster delete cifo-dev
}

Write-Host "==> Creating K3d cluster cifo-dev (1 server, 2 agents)..." -ForegroundColor Cyan
k3d cluster create --config cluster-config.yaml

# Fix Windows localhost connection to 127.0.0.1
kubectl config set-cluster k3d-cifo-dev --server=https://127.0.0.1:6443

Write-Host "==> Waiting for all nodes to be Ready..." -ForegroundColor Cyan
kubectl wait --for=condition=Ready nodes --all --timeout=180s

Write-Host "==> Creating argocd namespace..." -ForegroundColor Cyan
kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -

Write-Host "==> Installing ArgoCD into cluster..." -ForegroundColor Cyan
$localInstall = Join-Path $scriptDir "..\argocd\install.yaml"
if (Test-Path $localInstall) {
    kubectl apply -n argocd -f $localInstall
} else {
    kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/v2.10.4/manifests/install.yaml
}

Write-Host "==> Waiting for ArgoCD Server deployment to be available..." -ForegroundColor Cyan
kubectl wait --for=condition=Available deployment/argocd-server -n argocd --timeout=300s

Write-Host "==> Fetching ArgoCD initial admin password..." -ForegroundColor Green
$adminPass = kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}"
if ($adminPass) {
    $decoded = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($adminPass))
    Write-Host "ArgoCD URL: https://localhost:8443 (or port-forward 8080)" -ForegroundColor Green
    Write-Host "Username: admin" -ForegroundColor Green
    Write-Host "Password: $decoded" -ForegroundColor Green
}

Write-Host "==> Deploying sample applications..." -ForegroundColor Cyan
kubectl apply -f ../argocd/sample-apps/sample-nginx.yaml
kubectl apply -f ../argocd/sample-apps/sample-httpbin.yaml

Write-Host "==> Registering sample apps in ArgoCD..." -ForegroundColor Cyan
if (Test-Path "../argocd/sample-apps/argocd-app-nginx.yaml") {
    kubectl apply -f ../argocd/sample-apps/argocd-app-nginx.yaml
    kubectl apply -f ../argocd/sample-apps/argocd-app-httpbin.yaml
}

Write-Host "==> K3d cluster & ArgoCD testbed ready!" -ForegroundColor Green

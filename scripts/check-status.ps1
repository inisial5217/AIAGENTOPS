# Check status of local testbed
$ErrorActionPreference = "Continue"

# Refresh PATH from registry
$machinePath = [System.Environment]::GetEnvironmentVariable("Path", "Machine")
$userPath = [System.Environment]::GetEnvironmentVariable("Path", "User")
$env:Path = "$machinePath;$userPath"

Write-Host "=== 1. CLI Tools Status ===" -ForegroundColor Cyan
@("docker", "k3d", "kubectl", "curl") | ForEach-Object {
    $cmd = Get-Command $_ -ErrorAction SilentlyContinue
    if ($cmd) {
        Write-Host "  $_ : $($cmd.Source)" -ForegroundColor Green
    } else {
        Write-Host "  $_ : NOT FOUND" -ForegroundColor Red
    }
}

Write-Host "`n=== 2. Docker Containers Status ===" -ForegroundColor Cyan
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

Write-Host "`n=== 3. K3d Clusters Status ===" -ForegroundColor Cyan
k3d cluster list

Write-Host "`n=== 4. Kube Context & Nodes ===" -ForegroundColor Cyan
kubectl config current-context
kubectl config use-context k3d-cifo-dev 2>&1
kubectl get nodes -o wide

Write-Host "`n=== 5. Kubernetes Pods (All Namespaces) ===" -ForegroundColor Cyan
kubectl get pods -A -o wide

Write-Host "`n=== 6. ArgoCD Applications Status ===" -ForegroundColor Cyan
kubectl get applications.argoproj.io -n argocd 2>&1

Write-Host "`n=== 7. Observability Health Checks ===" -ForegroundColor Cyan
$endpoints = @{
    "VictoriaMetrics" = "http://127.0.0.1:8428/health"
    "Prometheus"      = "http://127.0.0.1:9090/-/healthy"
    "Loki"            = "http://127.0.0.1:3100/ready"
    "Tempo"           = "http://127.0.0.1:3200/status"
    "Alertmanager"    = "http://127.0.0.1:9093/-/healthy"
    "Grafana"         = "http://127.0.0.1:3000/api/health"
}

foreach ($name in $endpoints.Keys) {
    try {
        $res = Invoke-RestMethod -Uri $endpoints[$name] -TimeoutSec 3 -ErrorAction Stop
        Write-Host "  [OK] $name : healthy" -ForegroundColor Green
    } catch {
        Write-Host "  [FAIL] $name : $($_.Exception.Message)" -ForegroundColor Red
    }
}

Write-Host "`n=== End of Status Check ===" -ForegroundColor Cyan

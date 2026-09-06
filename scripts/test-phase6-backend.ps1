# Test Phase 6 Kubernetes & ArgoCD endpoints
$ErrorActionPreference = "Stop"

$login = Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/auth/login" -Method Post -ContentType "application/json" -Body '{"username":"admin@cifo.local","password":"admin123"}'
$token = $login.access_token
$h = @{ Authorization = "Bearer $token" }

Write-Host "=== KUBERNETES ENDPOINTS ==="
$pods = Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/kubernetes/pods" -Headers $h
Write-Host "Pods Count: $($pods.total)"
if ($pods.data.Count -gt 0) {
    Write-Host "First Pod: $($pods.data[0].name) | Phase: $($pods.data[0].phase) | Status: $($pods.data[0].status) | Age: $($pods.data[0].age)"
}

$deps = Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/kubernetes/deployments" -Headers $h
Write-Host "Deployments Count: $($deps.total)"
if ($deps.data.Count -gt 0) {
    Write-Host "First Deployment: $($deps.data[0].name) | Replicas: $($deps.data[0].replicas) | Ready: $($deps.data[0].ready_replicas)"
}

$nodes = Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/kubernetes/nodes" -Headers $h
Write-Host "Nodes Count: $($nodes.total)"
foreach ($n in $nodes.data) {
    Write-Host "Node: $($n.name) | Status: $($n.status) | Roles: $($n.roles) | Pods: $($n.pod_count)"
}

$services = Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/kubernetes/services" -Headers $h
Write-Host "Services Count: $($services.total)"

$k8sOverview = Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/kubernetes/overview" -Headers $h
Write-Host "K8s Overview: Nodes=$($k8sOverview.ready_nodes)/$($k8sOverview.total_nodes), Pods=$($k8sOverview.running_pods)/$($k8sOverview.total_pods), Deployments=$($k8sOverview.ready_deployments)/$($k8sOverview.total_deployments)"

Write-Host "`n=== ARGOCD ENDPOINTS ==="
$argo = Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/argocd/applications" -Headers $h
Write-Host "Argo Apps Count: $($argo.total)"
foreach ($a in $argo.data) {
    Write-Host "App: $($a.name) | Sync: $($a.sync_status) | Health: $($a.health_status) | Repo: $($a.repo_url)"
}

$argoOverview = Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/argocd/overview" -Headers $h
Write-Host "Argo Overview: Total=$($argoOverview.total), Synced=$($argoOverview.synced), OutOfSync=$($argoOverview.out_of_sync), Healthy=$($argoOverview.healthy), Degraded=$($argoOverview.degraded)"

Write-Host "`n=== MONITORING INTEGRATION ==="
$monStats = Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/monitoring/stats" -Headers $h
Write-Host "Dashboard Total Containers: $($monStats.total_containers)"
Write-Host "Dashboard Total Replicas (Docker + K8s): $($monStats.total_replicas)"
Write-Host "Dashboard Active Incidents: $($monStats.active_incidents)"
Write-Host "Dashboard RAM: $($monStats.overall_ram_percent)%"

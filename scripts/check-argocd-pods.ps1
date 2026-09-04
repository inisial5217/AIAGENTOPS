# Check argocd pods
$ErrorActionPreference = "Continue"

$machinePath = [System.Environment]::GetEnvironmentVariable("Path", "Machine")
$userPath = [System.Environment]::GetEnvironmentVariable("Path", "User")
$env:Path = "$machinePath;$userPath"

docker exec k3d-cifo-dev-server-0 kubectl get pods -n argocd -o wide
docker exec k3d-cifo-dev-server-0 kubectl get deployment -n argocd

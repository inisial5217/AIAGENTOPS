#!/usr/bin/env bash
# Setup K3d local cluster and deploy ArgoCD
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPT_DIR}"

echo "==> Checking prerequisites..."
command -v k3d >/dev/null 2>&1 || { echo "k3d CLI required but not found."; exit 1; }
command -v kubectl >/dev/null 2>&1 || { echo "kubectl CLI required but not found."; exit 1; }

echo "==> Checking for existing cifo-dev cluster..."
if k3d cluster list --no-headers 2>/dev/null | grep -q "cifo-dev"; then
    echo "Deleting existing cifo-dev cluster..."
    k3d cluster delete cifo-dev
fi

echo "==> Creating K3d cluster cifo-dev..."
k3d cluster create --config cluster-config.yaml

echo "==> Waiting for nodes to be Ready..."
kubectl wait --for=condition=Ready nodes --all --timeout=180s

echo "==> Creating argocd namespace..."
kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -

echo "==> Installing ArgoCD into cluster..."
if [ -f "../argocd/install.yaml" ]; then
    kubectl apply -n argocd -f ../argocd/install.yaml
else
    kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/v2.10.4/manifests/install.yaml
fi

echo "==> Waiting for ArgoCD Server to be Available..."
kubectl wait --for=condition=Available deployment/argocd-server -n argocd --timeout=300s

echo "==> ArgoCD initial admin password:"
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d && echo ""

echo "==> Deploying sample applications..."
kubectl apply -f ../argocd/sample-apps/sample-nginx.yaml
kubectl apply -f ../argocd/sample-apps/sample-httpbin.yaml

echo "==> Registering sample apps in ArgoCD..."
if [ -f "../argocd/sample-apps/argocd-app-nginx.yaml" ]; then
    kubectl apply -f ../argocd/sample-apps/argocd-app-nginx.yaml
    kubectl apply -f ../argocd/sample-apps/argocd-app-httpbin.yaml
fi

echo "==> Local K3d & ArgoCD setup complete!"

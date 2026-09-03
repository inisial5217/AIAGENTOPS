#!/usr/bin/env bash
# Local development bootstrap script for Linux/macOS
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "${SCRIPT_DIR}")"
cd "${ROOT_DIR}"

echo "===================================================="
echo "   CIFO Enterprise Monitoring Platform Bootstrap   "
echo "===================================================="

# 1. Environment config
echo "==> Preparing environment configuration..."
if [ ! -f ".env" ]; then
    cp .env.example .env
    echo "Created .env from .env.example"
fi

# 2. Start Data Services
echo "==> Starting Docker Compose data services..."
docker compose -f infrastructure/local-testbed/docker-compose.yml up -d

# 3. Wait for PostgreSQL
echo "==> Waiting for PostgreSQL readiness..."
for i in {1..30}; do
    if docker exec cifo-postgres pg_isready -U cifo_admin -d cifo_db >/dev/null 2>&1; then
        echo "PostgreSQL is ready!"
        break
    fi
    sleep 2
done

# 4. Service status
echo "==> Local Testbed running:"
echo "  - PostgreSQL:      localhost:5432"
echo "  - Redis:           localhost:6379"
echo "  - VictoriaMetrics: http://localhost:8428"
echo "  - Prometheus:      http://localhost:9090"
echo "  - Loki:            http://localhost:3100"
echo "  - Tempo:           http://localhost:3200"

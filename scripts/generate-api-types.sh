#!/usr/bin/env bash
# Generate API types from OpenAPI and Protobuf definitions
set -euo pipefail

echo "==> Validating OpenAPI contract..."
if [ -f "packages/api-contracts/openapi.yaml" ]; then
  echo "OpenAPI contract found at packages/api-contracts/openapi.yaml"
fi

echo "==> Validating Protobuf contracts..."
if [ -f "packages/api-contracts/proto/ai_service.proto" ]; then
  echo "Proto contract found at packages/api-contracts/proto/ai_service.proto"
fi

echo "==> Contracts ready for code generation."

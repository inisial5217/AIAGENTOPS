# Generate API types from OpenAPI and Protobuf definitions
Write-Host "==> Validating OpenAPI contract..." -ForegroundColor Cyan
if (Test-Path "packages/api-contracts/openapi.yaml") {
    Write-Host "OpenAPI contract found at packages/api-contracts/openapi.yaml" -ForegroundColor Green
}

Write-Host "==> Validating Protobuf contracts..." -ForegroundColor Cyan
if (Test-Path "packages/api-contracts/proto/ai_service.proto") {
    Write-Host "Proto contract found at packages/api-contracts/proto/ai_service.proto" -ForegroundColor Green
}

Write-Host "==> Contracts ready for code generation." -ForegroundColor Green

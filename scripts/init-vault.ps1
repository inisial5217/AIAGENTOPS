# HashiCorp Vault Initializer and Secrets Seeder
$VaultAddr = "http://127.0.0.1:8200"
$VaultToken = "cifo-vault-root-token"
$Headers = @{
    "X-Vault-Token" = $VaultToken
    "Content-Type"  = "application/json"
}

Write-Host "=== Initializing HashiCorp Vault Configuration ===" -ForegroundColor Cyan

# 1. Verify Vault Health
try {
    $health = Invoke-RestMethod -Uri "$VaultAddr/v1/sys/health" -Method Get
    Write-Host " Vault is healthy (Version: $($health.version), Sealed: $($health.sealed))" -ForegroundColor Green
} catch {
    Write-Error "Failed to connect to Vault at ${VaultAddr}: $_"
    exit 1
}

# 2. Upload Policies
Write-Host "`n--- Applying Vault Policies ---" -ForegroundColor Yellow
$backendPolicyPath = "infrastructure/security/vault/cifo-backend-policy.hcl"
$aiPolicyPath = "infrastructure/security/vault/cifo-ai-policy.hcl"

if (Test-Path $backendPolicyPath) {
    docker cp $backendPolicyPath cifo-vault:/tmp/cifo-backend-policy.hcl | Out-Null
    docker exec -e VAULT_ADDR=http://127.0.0.1:8200 -e VAULT_TOKEN=$VaultToken cifo-vault vault policy write cifo-backend /tmp/cifo-backend-policy.hcl | Out-Null
    Write-Host " Applied policy: cifo-backend" -ForegroundColor Green
}

if (Test-Path $aiPolicyPath) {
    docker cp $aiPolicyPath cifo-vault:/tmp/cifo-ai-policy.hcl | Out-Null
    docker exec -e VAULT_ADDR=http://127.0.0.1:8200 -e VAULT_TOKEN=$VaultToken cifo-vault vault policy write cifo-ai-service /tmp/cifo-ai-policy.hcl | Out-Null
    Write-Host " Applied policy: cifo-ai-service" -ForegroundColor Green
}

# 3. Seed Secrets for Backend (KV v2)
Write-Host "`n--- Seeding Secrets for Backend ---" -ForegroundColor Yellow
$backendSecrets = @{
    data = @{
        POSTGRES_DB             = "cifo_db"
        POSTGRES_USER           = "cifo_admin"
        POSTGRES_PASSWORD       = "cifo_secure_password"
        DATABASE_DSN            = "postgres://cifo_admin:cifo_secure_password@127.0.0.1:5432/cifo_db?sslmode=disable"
        REDIS_ADDR              = "127.0.0.1:6379"
        REDIS_PASSWORD          = "cifo_redis_secret"
        ARGOCD_URL              = "https://127.0.0.1:8443"
        ARGOCD_TOKEN            = "cifo-vault-seeded-token"
        TELEGRAM_BOT_TOKEN      = "cifo_vault_telegram_bot_token"
        TELEGRAM_CHAT_ID        = "12345678"
        KEYCLOAK_URL            = "http://127.0.0.1:8180"
        KEYCLOAK_ADMIN_PASSWORD = "admin"
        DOCKER_HOST             = "tcp://127.0.0.1:2376"
    }
}
$backendJson = $backendSecrets | ConvertTo-Json -Depth 5
Invoke-RestMethod -Uri "$VaultAddr/v1/secret/data/cifo/backend" -Method Post -Headers $Headers -Body $backendJson | Out-Null
Write-Host " Seeded secret/data/cifo/backend successfully" -ForegroundColor Green

# 4. Seed Secrets for AI Service (KV v2)
Write-Host "`n--- Seeding Secrets for AI Service ---" -ForegroundColor Yellow
$aiSecrets = @{
    data = @{
        GOOGLE_API_KEY    = "vault-google-gemini-key-live"
        OPENAI_API_KEY    = "vault-openai-gpt4-key-live"
        ANTHROPIC_API_KEY = "vault-anthropic-claude-key-live"
        OLLAMA_BASE_URL   = "http://127.0.0.1:11434"
    }
}
$aiJson = $aiSecrets | ConvertTo-Json -Depth 5
Invoke-RestMethod -Uri "$VaultAddr/v1/secret/data/cifo/ai-service" -Method Post -Headers $Headers -Body $aiJson | Out-Null
Write-Host " Seeded secret/data/cifo/ai-service successfully" -ForegroundColor Green

# 5. Verify Reading Back
Write-Host "`n--- Verifying Read from Vault ---" -ForegroundColor Yellow
$readBackend = Invoke-RestMethod -Uri "$VaultAddr/v1/secret/data/cifo/backend" -Method Get -Headers $Headers
Write-Host " Backend POSTGRES_USER from Vault:" $readBackend.data.data.POSTGRES_USER -ForegroundColor Cyan
Write-Host " Backend DOCKER_HOST from Vault:" $readBackend.data.data.DOCKER_HOST -ForegroundColor Cyan

$readAI = Invoke-RestMethod -Uri "$VaultAddr/v1/secret/data/cifo/ai-service" -Method Get -Headers $Headers
Write-Host " AI GOOGLE_API_KEY from Vault:" $readAI.data.data.GOOGLE_API_KEY -ForegroundColor Cyan

Write-Host "`n=== HashiCorp Vault Initialized and Ready ===" -ForegroundColor Green

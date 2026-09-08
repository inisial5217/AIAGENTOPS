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

# 3. Read overrides from root .env if available
$envMap = @{}
$rootEnv = Join-Path $PSScriptRoot "..\.env"
if (Test-Path $rootEnv) {
    Get-Content $rootEnv | ForEach-Object {
        $line = $_.Trim()
        if ($line -and -not $line.StartsWith("#") -and $line.Contains("=")) {
            $parts = $line.Split("=", 2)
            $k = $parts[0].Trim()
            $v = $parts[1].Trim()
            if ($v -and -not $v.StartsWith("your_")) {
                $envMap[$k] = $v
            }
        }
    }
}

# 4. Seed Secrets for Backend (KV v2)
Write-Host "`n--- Seeding Secrets for Backend ---" -ForegroundColor Yellow
$backendSecrets = @{
    data = @{
        POSTGRES_DB             = if ($envMap["POSTGRES_DB"]) { $envMap["POSTGRES_DB"] } else { "cifo_db" }
        POSTGRES_USER           = if ($envMap["POSTGRES_USER"]) { $envMap["POSTGRES_USER"] } else { "cifo_admin" }
        POSTGRES_PASSWORD       = if ($envMap["POSTGRES_PASSWORD"]) { $envMap["POSTGRES_PASSWORD"] } else { "cifo_secure_password" }
        DATABASE_DSN            = if ($envMap["DATABASE_DSN"]) { $envMap["DATABASE_DSN"] } else { "postgres://cifo_admin:cifo_secure_password@127.0.0.1:5432/cifo_db?sslmode=disable" }
        REDIS_ADDR              = if ($envMap["REDIS_ADDR"]) { $envMap["REDIS_ADDR"] } else { "127.0.0.1:6379" }
        REDIS_PASSWORD          = if ($envMap["REDIS_PASSWORD"]) { $envMap["REDIS_PASSWORD"] } else { "cifo_redis_secret" }
        ARGOCD_URL              = if ($envMap["ARGOCD_URL"]) { $envMap["ARGOCD_URL"] } else { "https://127.0.0.1:8443" }
        ARGOCD_TOKEN            = if ($envMap["ARGOCD_TOKEN"]) { $envMap["ARGOCD_TOKEN"] } else { "cifo-vault-seeded-token" }
        TELEGRAM_BOT_TOKEN      = if ($envMap["TELEGRAM_BOT_TOKEN"]) { $envMap["TELEGRAM_BOT_TOKEN"] } else { "cifo_vault_telegram_bot_token" }
        TELEGRAM_CHAT_ID        = if ($envMap["TELEGRAM_CHAT_ID"]) { $envMap["TELEGRAM_CHAT_ID"] } else { "12345678" }
        KEYCLOAK_URL            = if ($envMap["KEYCLOAK_URL"]) { $envMap["KEYCLOAK_URL"] } else { "http://127.0.0.1:8180" }
        KEYCLOAK_ADMIN_PASSWORD = if ($envMap["KEYCLOAK_ADMIN_PASSWORD"]) { $envMap["KEYCLOAK_ADMIN_PASSWORD"] } else { "admin" }
        DOCKER_HOST             = if ($envMap["DOCKER_HOST"]) { $envMap["DOCKER_HOST"] } else { "tcp://127.0.0.1:2376" }
    }
}
$backendJson = $backendSecrets | ConvertTo-Json -Depth 5
Invoke-RestMethod -Uri "$VaultAddr/v1/secret/data/cifo/backend" -Method Post -Headers $Headers -Body $backendJson | Out-Null
Write-Host " Seeded secret/data/cifo/backend successfully" -ForegroundColor Green

# 5. Seed Secrets for AI Service (KV v2)
Write-Host "`n--- Seeding Secrets for AI Service ---" -ForegroundColor Yellow
$aiSecrets = @{
    data = @{
        GOOGLE_API_KEY    = if ($envMap["GOOGLE_API_KEY"]) { $envMap["GOOGLE_API_KEY"] } elseif ($envMap["GEMINI_API_KEY"]) { $envMap["GEMINI_API_KEY"] } else { "vault-google-gemini-key-live" }
        OPENAI_API_KEY    = if ($envMap["OPENAI_API_KEY"]) { $envMap["OPENAI_API_KEY"] } else { "vault-openai-gpt4-key-live" }
        ANTHROPIC_API_KEY = if ($envMap["ANTHROPIC_API_KEY"]) { $envMap["ANTHROPIC_API_KEY"] } else { "vault-anthropic-claude-key-live" }
        OLLAMA_BASE_URL   = if ($envMap["OLLAMA_BASE_URL"]) { $envMap["OLLAMA_BASE_URL"] } else { "http://127.0.0.1:11434" }
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

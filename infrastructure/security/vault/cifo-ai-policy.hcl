# HashiCorp Vault Policy for CIFO AI Service
# Grants read access to AI service secrets (LLM API keys, model configs)

path "secret/data/cifo/ai-service" {
  capabilities = ["read"]
}

path "secret/data/cifo/shared" {
  capabilities = ["read"]
}

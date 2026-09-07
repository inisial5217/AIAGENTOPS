# HashiCorp Vault Policy for CIFO Backend
# Grants read access to backend secrets and shared infrastructure secrets

path "secret/data/cifo/backend" {
  capabilities = ["read"]
}

path "secret/data/cifo/shared" {
  capabilities = ["read"]
}

path "secret/metadata/cifo/*" {
  capabilities = ["list"]
}

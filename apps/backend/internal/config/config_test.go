package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigLoadDefaults(t *testing.T) {
	_ = os.Setenv("VAULT_ENABLED", "false")
	_ = os.Setenv("DATABASE_DSN", "postgres://test:test@localhost:5432/test_db")
	_ = os.Setenv("REDIS_ADDR", "localhost:6379")

	cfg, err := Load()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "development", cfg.Environment)
	assert.Equal(t, "http://localhost:8180", cfg.KeycloakURL)
	assert.Equal(t, "cifo", cfg.KeycloakRealm)
	assert.Equal(t, "http://localhost:8180/realms/cifo", cfg.KeycloakIssuer)
}

func TestConfigMissingRequired(t *testing.T) {
	_ = os.Setenv("VAULT_ENABLED", "false")
	_ = os.Setenv("DATABASE_DSN", "")
	_ = os.Setenv("REDIS_ADDR", "localhost:6379")

	cfg, err := Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)

	_ = os.Setenv("DATABASE_DSN", "postgres://test:test@localhost:5432/test_db")
	_ = os.Setenv("REDIS_ADDR", "")

	cfg, err = Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestConfigLoadVaultSecrets(t *testing.T) {
	// test vault loader with fallback
	_ = os.Setenv("VAULT_ENABLED", "true")
	_ = os.Setenv("VAULT_ADDR", "http://127.0.0.1:8200")
	_ = os.Setenv("VAULT_TOKEN", "cifo-vault-root-token")
	_ = os.Setenv("DATABASE_DSN", "postgres://fallback:fallback@localhost:5432/test_db")
	_ = os.Setenv("REDIS_ADDR", "localhost:6379")

	cfg, err := Load()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.True(t, cfg.VaultEnabled)
	assert.Equal(t, "http://127.0.0.1:8200", cfg.VaultAddr)
}

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigLoadDefaults(t *testing.T) {
	// set required envs
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

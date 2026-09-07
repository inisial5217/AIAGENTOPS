package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cifo-monitoring/backend/internal/security"
)

// Config application settings
type Config struct {
	Port           int
	Environment    string
	DatabaseDSN    string
	RedisAddr      string
	RedisPass      string
	DockerHost     string
	ArgoCDURL      string
	ArgoCDToken    string
	TelegramToken    string
	TelegramChatID   string
	AIServiceURL     string
	LogLevel         string
	AllowedOrigins   []string
	KeycloakURL      string
	KeycloakRealm    string
	KeycloakIssuer   string
	KeycloakJWKSURL  string
	KeycloakClientID string
	OTelEndpoint     string
	OTelServiceName  string
	OTelEnabled      bool
	VaultAddr        string
	VaultToken       string
	VaultEnabled     bool
}

// Load loads env config
func Load() (*Config, error) {
	port := 8080
	if p := os.Getenv("PORT"); p != "" {
		val, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid PORT: %w", err)
		}
		port = val
	}

	env := getEnv("APP_ENV", "development")
	dsn := os.Getenv("DATABASE_DSN")
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPass := getEnv("REDIS_PASSWORD", "cifo_redis_secret")
	dockerHost := getEnv("DOCKER_HOST", "")
	argocdURL := getEnv("ARGOCD_URL", "https://localhost:8443")
	argocdToken := getEnv("ARGOCD_TOKEN", "")
	telegramToken := getEnv("TELEGRAM_BOT_TOKEN", "")
	telegramChatID := getEnv("TELEGRAM_CHAT_ID", "")
	aiServiceURL := getEnv("AI_SERVICE_URL", "http://localhost:8000")
	logLevel := getEnv("LOG_LEVEL", "INFO")
	originsStr := getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://127.0.0.1:3000,http://localhost:3001,http://127.0.0.1:3001,http://localhost:5173")
	kcURL := getEnv("KEYCLOAK_URL", "http://localhost:8180")
	kcRealm := getEnv("KEYCLOAK_REALM", "cifo")
	kcIssuer := getEnv("KEYCLOAK_ISSUER", fmt.Sprintf("%s/realms/%s", kcURL, kcRealm))
	kcJWKS := getEnv("KEYCLOAK_JWKS_URL", fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", kcURL, kcRealm))
	kcClientID := getEnv("KEYCLOAK_CLIENT_ID", "cifo-frontend")

	vaultAddr := getEnv("VAULT_ADDR", "http://127.0.0.1:8200")
	vaultToken := getEnv("VAULT_TOKEN", "cifo-vault-root-token")
	vaultEnabledStr := getEnv("VAULT_ENABLED", "true")
	vaultEnabled := (vaultEnabledStr == "true" || vaultEnabledStr == "1") && vaultAddr != ""

	// load secrets from vault if available
	if vaultEnabled {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		vc := security.NewVaultClient(vaultAddr, vaultToken)
		if secrets, err := vc.GetSecret(ctx, "cifo/backend"); err == nil && secrets != nil {
			if v, ok := secrets["DATABASE_DSN"].(string); ok && v != "" {
				dsn = v
			}
			if v, ok := secrets["REDIS_ADDR"].(string); ok && v != "" {
				redisAddr = v
			}
			if v, ok := secrets["REDIS_PASSWORD"].(string); ok && v != "" {
				redisPass = v
			}
			if v, ok := secrets["DOCKER_HOST"].(string); ok && v != "" {
				dockerHost = v
			}
			if v, ok := secrets["ARGOCD_TOKEN"].(string); ok && v != "" {
				argocdToken = v
			}
			if v, ok := secrets["TELEGRAM_BOT_TOKEN"].(string); ok && v != "" {
				telegramToken = v
			}
			if v, ok := secrets["TELEGRAM_CHAT_ID"].(string); ok && v != "" {
				telegramChatID = v
			}
		}
	}

	// validate required fields
	if strings.TrimSpace(dsn) == "" {
		return nil, errors.New("DATABASE_DSN is required")
	}
	if strings.TrimSpace(redisAddr) == "" {
		return nil, errors.New("REDIS_ADDR is required")
	}

	var origins []string
	for _, o := range strings.Split(originsStr, ",") {
		if trimmed := strings.TrimSpace(o); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}

	otelEndpoint := getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318")
	otelServiceName := getEnv("OTEL_SERVICE_NAME", "cifo-backend")
	otelEnabledStr := getEnv("OTEL_ENABLED", "true")
	otelEnabled := otelEnabledStr == "true" || otelEnabledStr == "1"

	return &Config{
		Port:             port,
		Environment:      env,
		DatabaseDSN:      dsn,
		RedisAddr:        redisAddr,
		RedisPass:        redisPass,
		DockerHost:       dockerHost,
		ArgoCDURL:        argocdURL,
		ArgoCDToken:      argocdToken,
		TelegramToken:    telegramToken,
		TelegramChatID:   telegramChatID,
		AIServiceURL:     aiServiceURL,
		LogLevel:         logLevel,
		AllowedOrigins:   origins,
		KeycloakURL:      kcURL,
		KeycloakRealm:    kcRealm,
		KeycloakIssuer:   kcIssuer,
		KeycloakJWKSURL:  kcJWKS,
		KeycloakClientID: kcClientID,
		OTelEndpoint:     otelEndpoint,
		OTelServiceName:  otelServiceName,
		OTelEnabled:      otelEnabled,
		VaultAddr:        vaultAddr,
		VaultToken:       vaultToken,
		VaultEnabled:     vaultEnabled,
	}, nil
}

// getEnv gets environment variable
func getEnv(key string, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

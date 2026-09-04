package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
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
	TelegramToken  string
	AIServiceURL   string
	LogLevel       string
	AllowedOrigins []string
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
	aiServiceURL := getEnv("AI_SERVICE_URL", "localhost:50051")
	logLevel := getEnv("LOG_LEVEL", "INFO")
	originsStr := getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://127.0.0.1:3000")

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

	return &Config{
		Port:           port,
		Environment:    env,
		DatabaseDSN:    dsn,
		RedisAddr:      redisAddr,
		RedisPass:      redisPass,
		DockerHost:     dockerHost,
		ArgoCDURL:      argocdURL,
		ArgoCDToken:    argocdToken,
		TelegramToken:  telegramToken,
		AIServiceURL:   aiServiceURL,
		LogLevel:       logLevel,
		AllowedOrigins: origins,
	}, nil
}

// getEnv gets environment variable
func getEnv(key string, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

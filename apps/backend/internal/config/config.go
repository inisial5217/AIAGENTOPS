package config

import (
	"os"
	"strconv"
)

// Config holds application settings
type Config struct {
	Port         int
	Environment  string
	DatabaseDSN  string
	RedisAddr    string
	RedisPass    string
	DockerHost   string
	ArgoCDURL    string
	ArgoCDToken  string
	AIServiceURL string
	LogLevel     string
}

// Load loads env config
func Load() (*Config, error) {
	port := 8080
	if p := os.Getenv("PORT"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			port = val
		}
	}

	env := getEnv("APP_ENV", "development")
	dsn := getEnv("DATABASE_DSN", "postgres://cifo_admin:cifo_secret@localhost:5432/cifo_db?sslmode=disable")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPass := getEnv("REDIS_PASSWORD", "")
	dockerHost := getEnv("DOCKER_HOST", "")
	argocdURL := getEnv("ARGOCD_URL", "https://localhost:8443")
	argocdToken := getEnv("ARGOCD_TOKEN", "")
	aiServiceURL := getEnv("AI_SERVICE_URL", "localhost:50051")
	logLevel := getEnv("LOG_LEVEL", "INFO")

	return &Config{
		Port:         port,
		Environment:  env,
		DatabaseDSN:  dsn,
		RedisAddr:    redisAddr,
		RedisPass:    redisPass,
		DockerHost:   dockerHost,
		ArgoCDURL:    argocdURL,
		ArgoCDToken:  argocdToken,
		AIServiceURL: aiServiceURL,
		LogLevel:     logLevel,
	}, nil
}

// getEnv helper with default
func getEnv(key string, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

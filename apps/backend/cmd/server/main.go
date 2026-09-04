package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/cifo-monitoring/backend/internal/config"
	"github.com/cifo-monitoring/backend/internal/handler"
	"github.com/cifo-monitoring/backend/internal/middleware"
	"github.com/cifo-monitoring/backend/internal/repository"
	"github.com/cifo-monitoring/backend/pkg/logger"
	"github.com/cifo-monitoring/backend/pkg/validator"
	"github.com/labstack/echo/v4"
)

func main() {
	// load env config
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	// init structured logger
	appLogger := logger.New(cfg.LogLevel)
	appLogger.Info("starting cifo backend server",
		slog.String("env", cfg.Environment),
		slog.Int("port", cfg.Port),
	)

	// create base context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// init database pool
	dbPool, err := repository.NewPostgresPool(ctx, cfg.DatabaseDSN, appLogger)
	if err != nil {
		appLogger.Error("failed to connect database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer dbPool.Close()

	// init redis client
	redisClient, err := repository.NewRedisClient(ctx, cfg.RedisAddr, cfg.RedisPass, appLogger)
	if err != nil {
		appLogger.Error("failed to connect redis", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		_ = redisClient.Close()
	}()

	// run database migrations
	migrationsDir := findMigrationsDir()
	migrator := repository.NewMigrator(dbPool, migrationsDir, appLogger)
	if err := migrator.Up(ctx); err != nil {
		appLogger.Error("database migration failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	appLogger.Info("database migrations verified")

	// init echo router
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Validator = validator.New()

	// init metrics collector
	metrics := handler.NewMetrics(dbPool)

	// init rate limiter
	rateLimiter := middleware.NewRateLimiter(redisClient, appLogger)

	// register global middleware
	e.Use(middleware.RequestLogger(appLogger))
	e.Use(middleware.Recover(appLogger))
	e.Use(middleware.CORS(cfg.AllowedOrigins))
	e.Use(metrics.Middleware())
	e.Use(rateLimiter.LimitIP(100, time.Minute))
	e.Use(middleware.AuthStub())

	// register probe handlers
	healthHandler := handler.NewHealthHandler(dbPool, redisClient, cfg.DockerHost)
	e.GET("/healthz", healthHandler.Liveness)
	e.GET("/readyz", healthHandler.Readiness)
	e.GET("/metrics", metrics.Handler())

	// start http server
	serverAddr := fmt.Sprintf(":%d", cfg.Port)
	go func() {
		if err := e.Start(serverAddr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error("http server failed", slog.String("error", err.Error()))
			stop()
		}
	}()
	appLogger.Info("http server listening", slog.String("addr", serverAddr))

	// wait for termination signal
	<-ctx.Done()
	appLogger.Info("shutting down server gracefully")

	// timeout for active connections
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("error during server shutdown", slog.String("error", err.Error()))
	}

	appLogger.Info("server shutdown complete")
}

// findMigrationsDir locates migrations
func findMigrationsDir() string {
	candidates := []string{
		"migrations",
		"apps/backend/migrations",
		"../migrations",
		"../../migrations",
	}

	for _, path := range candidates {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			absPath, _ := filepath.Abs(path)
			return absPath
		}
	}

	return "migrations"
}

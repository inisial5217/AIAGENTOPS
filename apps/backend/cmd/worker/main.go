package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cifo-monitoring/backend/internal/config"
	"github.com/cifo-monitoring/backend/internal/integration"
	"github.com/cifo-monitoring/backend/internal/repository"
	"github.com/cifo-monitoring/backend/internal/service"
	"github.com/cifo-monitoring/backend/pkg/logger"
)

func main() {
	// load env config
	cfg, err := config.Load()
	if err != nil {
		os.Exit(1)
	}

	appLogger := logger.New(cfg.LogLevel, "cifo-worker")
	appLogger.Info("starting cifo background worker")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// init database pool
	dbPool, err := repository.NewPostgresPool(ctx, cfg.DatabaseDSN, appLogger)
	if err != nil {
		appLogger.Error("worker db connection failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer dbPool.Close()

	// init redis client
	redisClient, err := repository.NewRedisClient(ctx, cfg.RedisAddr, cfg.RedisPass, appLogger)
	if err != nil {
		appLogger.Warn("worker redis connection failed", slog.String("error", err.Error()))
	} else {
		defer redisClient.Close()
	}

	// init services
	telegramClient := integration.NewTelegramClient(cfg.TelegramToken, cfg.TelegramChatID, appLogger)
	telegramService := service.NewTelegramService(telegramClient, redisClient, appLogger)
	incidentRepo := repository.NewIncidentRepository(dbPool)
	auditRepo := repository.NewAuditRepository(dbPool)
	notifService := service.NewNotificationService(telegramService, nil, incidentRepo, appLogger)
	incidentService := service.NewIncidentService(incidentRepo, auditRepo, notifService, appLogger)

	// run escalation ticker
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	appLogger.Info("worker escalation scheduler started")

	for {
		select {
		case <-ctx.Done():
			appLogger.Info("cifo worker stopping")
			return
		case <-ticker.C:
			_ = incidentService.CheckEscalations(ctx)
			_ = telegramService.ProcessRetryQueue(ctx)
		}
	}
}

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/cifo-monitoring/backend/internal/config"
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

	<-ctx.Done()
	appLogger.Info("cifo worker stopped")
}

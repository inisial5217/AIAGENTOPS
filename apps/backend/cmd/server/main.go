package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cifo-monitoring/backend/internal/config"
	"github.com/cifo-monitoring/backend/internal/handler"
	"github.com/cifo-monitoring/backend/internal/middleware"
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
	appLogger := logger.New(cfg.LogLevel, "cifo-backend")
	appLogger.Info("starting cifo backend server", "env", cfg.Environment, "port", cfg.Port)

	// init echo web framework
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Validator = validator.New()

	// register middleware
	e.Use(middleware.Recover(appLogger))

	// register health routes
	healthHdl := handler.NewHealthHandler()
	e.GET("/healthz", healthHdl.Liveness)
	e.GET("/readyz", healthHdl.Readiness)

	// graceful shutdown handler
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		addr := fmt.Sprintf(":%d", cfg.Port)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			appLogger.Error("server fatal error", "err", err)
			stop()
		}
	}()

	// wait for interrupt
	<-ctx.Done()
	appLogger.Info("shutting down backend server gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("server shutdown error", "err", err)
	}

	appLogger.Info("backend server stopped")
}

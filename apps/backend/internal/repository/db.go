package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/cifo-monitoring/backend/pkg/telemetry"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPostgresPool inits db pool
func NewPostgresPool(ctx context.Context, dsn string, logger *slog.Logger) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	// attach otel query tracer
	cfg.ConnConfig.Tracer = telemetry.NewDBQueryTracer()
	cfg.ConnConfig.ConnectTimeout = 5 * time.Second

	// configure connection pool
	cfg.MinConns = 1
	cfg.MaxConns = 25
	cfg.MaxConnLifetime = 1 * time.Hour
	cfg.MaxConnIdleTime = 15 * time.Minute
	cfg.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	// ping database with retries
	var pingErr error
	maxRetries := 15
	retryInterval := 2 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		pingErr = pool.Ping(pingCtx)
		cancel()

		if pingErr == nil {
			break
		}

		if logger != nil {
			logger.Warn("waiting for database ready",
				slog.Int("attempt", attempt),
				slog.Int("max_attempts", maxRetries),
				slog.String("error", pingErr.Error()),
			)
		}

		select {
		case <-ctx.Done():
			pool.Close()
			return nil, ctx.Err()
		case <-time.After(retryInterval):
		}
	}

	if pingErr != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", pingErr)
	}

	if logger != nil {
		logger.Info("database pool connected", slog.Int("min_conns", int(cfg.MinConns)), slog.Int("max_conns", int(cfg.MaxConns)))
	}

	return pool, nil
}

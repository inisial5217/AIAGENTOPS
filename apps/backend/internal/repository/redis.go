package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient inits redis client
func NewRedisClient(ctx context.Context, addr string, password string, logger *slog.Logger) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     20,
		MinIdleConns: 5,
	})

	// ping redis with retries
	var pingErr error
	maxRetries := 10
	retryInterval := 2 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		pingErr = client.Ping(pingCtx).Err()
		cancel()

		if pingErr == nil {
			break
		}

		if logger != nil {
			logger.Warn("waiting for redis ready",
				slog.Int("attempt", attempt),
				slog.Int("max_attempts", maxRetries),
				slog.String("error", pingErr.Error()),
			)
		}

		select {
		case <-ctx.Done():
			_ = client.Close()
			return nil, ctx.Err()
		case <-time.After(retryInterval):
		}
	}

	if pingErr != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", pingErr)
	}

	if logger != nil {
		logger.Info("redis client connected", slog.String("addr", addr))
	}

	return client, nil
}

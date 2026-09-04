package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/cifo-monitoring/backend/pkg/apperror"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

// RateLimiter struct
type RateLimiter struct {
	redis  *redis.Client
	logger *slog.Logger
}

// NewRateLimiter creates limiter
func NewRateLimiter(r *redis.Client, l *slog.Logger) *RateLimiter {
	return &RateLimiter{
		redis:  r,
		logger: l,
	}
}

// LimitIP limits by ip
func (rl *RateLimiter) LimitIP(limit int, window time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if rl.redis == nil {
				return next(c)
			}

			ip := c.RealIP()
			key := fmt.Sprintf("ratelimit:ip:%s", ip)
			return rl.applyLimit(c, next, key, limit, window)
		}
	}
}

// LimitUser limits by user
func (rl *RateLimiter) LimitUser(limit int, window time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if rl.redis == nil {
				return next(c)
			}

			userID := c.Get("user_id")
			identifier := c.RealIP()
			if userID != nil {
				if uid, ok := userID.(string); ok && uid != "" {
					identifier = uid
				}
			}

			key := fmt.Sprintf("ratelimit:user:%s", identifier)
			return rl.applyLimit(c, next, key, limit, window)
		}
	}
}

// applyLimit sliding window logic
func (rl *RateLimiter) applyLimit(c echo.Context, next echo.HandlerFunc, key string, limit int, window time.Duration) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	now := time.Now().UnixNano()
	windowStart := now - window.Nanoseconds()
	nowMember := fmt.Sprintf("%d", now)

	// execute pipeline
	pipe := rl.redis.Pipeline()
	pipe.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%d", windowStart))
	countCmd := pipe.ZCard(ctx, key)
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: nowMember})
	pipe.Expire(ctx, key, window+time.Minute)

	_, err := pipe.Exec(ctx)
	if err != nil {
		if rl.logger != nil {
			rl.logger.Warn("ratelimit redis error", slog.String("error", err.Error()))
		}
		// fail open on error
		return next(c)
	}

	count := int(countCmd.Val())
	remaining := limit - (count + 1)
	if remaining < 0 {
		remaining = 0
	}

	resetTime := time.Now().Add(window).Unix()
	c.Response().Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
	c.Response().Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
	c.Response().Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))

	if count >= limit {
		return c.JSON(http.StatusTooManyRequests, apperror.ErrRateLimit)
	}

	return next(c)
}

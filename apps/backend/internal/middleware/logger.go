package middleware

import (
	"log/slog"
	"time"

	"github.com/cifo-monitoring/backend/pkg/logger"
	"github.com/labstack/echo/v4"
)

// RequestLogger logs requests
func RequestLogger(l *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			duration := time.Since(start)
			status := c.Response().Status
			reqCtx := c.Request().Context()

			reqLogger := logger.WithContext(reqCtx, l)
			reqLogger.InfoContext(reqCtx, "http request",
				slog.String("method", c.Request().Method),
				slog.String("path", c.Path()),
				slog.Int("status", status),
				slog.Float64("duration_ms", float64(duration.Microseconds())/1000.0),
				slog.String("ip", c.RealIP()),
				slog.String("user_agent", c.Request().UserAgent()),
			)

			return nil
		}
	}
}

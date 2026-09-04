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

			// get request id
			reqID := c.Request().Header.Get(echo.HeaderXRequestID)
			if reqID == "" {
				reqID = c.Response().Header().Get(echo.HeaderXRequestID)
			}

			// inject trace context
			ctx := logger.ContextWithTrace(c.Request().Context(), reqID, "")
			c.SetRequest(c.Request().WithContext(ctx))

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			duration := time.Since(start)
			status := c.Response().Status

			reqLogger := logger.WithContext(c.Request().Context(), l)
			reqLogger.Info("http request",
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

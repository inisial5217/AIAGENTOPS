package middleware

import (
	"log/slog"
	"net/http"

	"github.com/cifo-monitoring/backend/pkg/apperror"
	"github.com/labstack/echo/v4"
)

// Recover handles server panics
func Recover(l *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					l.Error("panic recovered", slog.Any("error", r))
					_ = c.JSON(http.StatusInternalServerError, apperror.ErrInternal)
				}
			}()
			return next(c)
		}
	}
}

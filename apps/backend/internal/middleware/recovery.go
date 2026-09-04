package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/cifo-monitoring/backend/pkg/apperror"
	"github.com/labstack/echo/v4"
)

// Recover handles server panics
func Recover(l *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					stack := string(debug.Stack())
					if l != nil {
						l.Error("panic recovered",
							slog.Any("error", r),
							slog.String("stack", stack),
						)
					}
					_ = c.JSON(http.StatusInternalServerError, apperror.ErrInternal)
				}
			}()
			return next(c)
		}
	}
}

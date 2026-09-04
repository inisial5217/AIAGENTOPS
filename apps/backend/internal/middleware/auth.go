package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
)

// AuthStub stub authentication middleware
func AuthStub() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				// set stub context claims
				c.Set("user_id", "00000000-0000-0000-0000-000000000001")
				c.Set("role", "admin")
				c.Set("token", token)
			}
			return next(c)
		}
	}
}

package middleware

import (
	"github.com/labstack/echo/v4"
)

// SecurityHeaders sets secure headers
func SecurityHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			res := c.Response()
			res.Header().Set("X-Content-Type-Options", "nosniff")
			res.Header().Set("X-Frame-Options", "DENY")
			res.Header().Set("X-XSS-Protection", "1; mode=block")
			res.Header().Set("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'; object-src 'none'; base-uri 'self';")
			res.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
			res.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			res.Header().Set("Permissions-Policy", "geolocation=(), camera=(), microphone=(), payment=()")

			return next(c)
		}
	}
}

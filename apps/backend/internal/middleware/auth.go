package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cifo-monitoring/backend/internal/service"
	"github.com/labstack/echo/v4"
)

// ProblemDetail represents RFC 7807 error
type ProblemDetail struct {
	Type     string `json:"type,omitempty"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	Instance string `json:"instance,omitempty"`
}

// writeProblem response helper
func writeProblem(c echo.Context, status int, title, detail string) error {
	c.Response().Header().Set(echo.HeaderContentType, "application/problem+json")
	return c.JSON(status, ProblemDetail{
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: c.Request().RequestURI,
	})
}

// RequireAuth authenticates JWT token
func RequireAuth(authService *service.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return writeProblem(c, http.StatusUnauthorized, "Unauthorized", "Missing or invalid authorization header")
			}

			tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			if tokenString == "" {
				return writeProblem(c, http.StatusUnauthorized, "Unauthorized", "Empty bearer token")
			}

			ctx := c.Request().Context()
			claims, err := authService.ValidateToken(ctx, tokenString)
			if err != nil {
				return writeProblem(c, http.StatusUnauthorized, "Unauthorized", fmt.Sprintf("Invalid token: %v", err))
			}

			// sync user in postgres
			user, err := authService.SyncUser(ctx, claims)
			if err != nil {
				// fallback to claims if sync errors
				c.Set("claims", claims)
				c.Set("user_id", claims.Subject)
				c.Set("user_email", claims.Email)
				c.Set("user_role", authService.ExtractRole(claims))
				c.Set("token", tokenString)
				return next(c)
			}

			c.Set("claims", claims)
			c.Set("user", user)
			c.Set("user_id", user.ID)
			c.Set("user_email", user.Email)
			c.Set("user_role", user.Role)
			c.Set("token", tokenString)

			return next(c)
		}
	}
}

// AuthStub stub authentication middleware
func AuthStub() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				c.Set("user_id", "00000000-0000-0000-0000-000000000001")
				c.Set("role", "admin")
				c.Set("user_role", "admin")
				c.Set("token", token)
			}
			return next(c)
		}
	}
}

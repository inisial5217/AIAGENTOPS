package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/cifo-monitoring/backend/internal/service"
	"github.com/labstack/echo/v4"
)

// roleHierarchy defines role levels
var roleHierarchy = map[string]int{
	"admin":  3,
	"devops": 2,
	"viewer": 1,
}

// RequireRole enforces role hierarchy and logs violations
func RequireRole(authService *service.AuthService, minRole string) echo.MiddlewareFunc {
	minLevel := roleHierarchy[strings.ToLower(minRole)]

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userRole, ok := c.Get("user_role").(string)
			if !ok || userRole == "" {
				userRole = "viewer"
			}

			userLevel := roleHierarchy[strings.ToLower(userRole)]

			if userLevel < minLevel {
				// extract actor id
				actorID, _ := c.Get("user_id").(string)
				if actorID == "" {
					actorID, _ = c.Get("user_email").(string)
				}
				if actorID == "" {
					actorID = "unknown"
				}

				ip := c.RealIP()
				ua := c.Request().UserAgent()
				uri := c.Request().RequestURI
				details, _ := json.Marshal(map[string]interface{}{
					"user_role":     userRole,
					"required_role": minRole,
					"path":          uri,
					"method":        c.Request().Method,
				})

				// log access denied audit event
				if authService != nil {
					_ = authService.LogAuditEvent(c.Request().Context(), &model.AuditLog{
						Timestamp:    time.Now(),
						ActorType:    "user",
						ActorID:      actorID,
						Action:       "access_denied",
						ResourceType: "route",
						ResourceID:   &uri,
						Details:      details,
						IPAddress:    &ip,
						UserAgent:    &ua,
						Result:       "failure",
					})
				}

				return writeProblem(c, http.StatusForbidden, "Forbidden", fmt.Sprintf("Insufficient permissions: role '%s' required", minRole))
			}

			return next(c)
		}
	}
}

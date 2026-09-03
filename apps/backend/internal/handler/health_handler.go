package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// HealthHandler handles probes
type HealthHandler struct{}

// NewHealthHandler creates handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Liveness handles liveness check
func (h *HealthHandler) Liveness(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// Readiness handles readiness check
func (h *HealthHandler) Readiness(c echo.Context) error {
	// checks will query dependencies
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "ready",
		"checks": map[string]string{
			"database": "stub",
			"redis":    "stub",
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

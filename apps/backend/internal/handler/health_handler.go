package handler

import (
	"context"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

// HealthHandler handles probes
type HealthHandler struct {
	db         *pgxpool.Pool
	redis      *redis.Client
	dockerHost string
}

// NewHealthHandler creates handler
func NewHealthHandler(db *pgxpool.Pool, r *redis.Client, dockerHost string) *HealthHandler {
	return &HealthHandler{
		db:         db,
		redis:      r,
		dockerHost: dockerHost,
	}
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
	ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
	defer cancel()

	checks := make(map[string]string)
	allHealthy := true

	// check postgres database
	if h.db != nil {
		if err := h.db.Ping(ctx); err != nil {
			checks["database"] = "unhealthy: " + err.Error()
			allHealthy = false
		} else {
			checks["database"] = "connected"
		}
	} else {
		checks["database"] = "not_configured"
		allHealthy = false
	}

	// check redis cache
	if h.redis != nil {
		if err := h.redis.Ping(ctx).Err(); err != nil {
			checks["redis"] = "unhealthy: " + err.Error()
			allHealthy = false
		} else {
			checks["redis"] = "connected"
		}
	} else {
		checks["redis"] = "not_configured"
		allHealthy = false
	}

	// check docker daemon
	dockerStatus := h.checkDocker(ctx)
	checks["docker"] = dockerStatus
	if dockerStatus != "connected" {
		allHealthy = false
	}

	statusCode := http.StatusOK
	statusText := "ready"
	if !allHealthy {
		statusCode = http.StatusServiceUnavailable
		statusText = "unhealthy"
	}

	return c.JSON(statusCode, map[string]interface{}{
		"status":    statusText,
		"checks":    checks,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// checkDocker tests docker daemon
func (h *HealthHandler) checkDocker(ctx context.Context) string {
	if h.dockerHost != "" {
		client := &http.Client{Timeout: 2 * time.Second}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.dockerHost+"/_ping", nil)
		if err != nil {
			return "error: " + err.Error()
		}
		resp, err := client.Do(req)
		if err != nil {
			return "unreachable: " + err.Error()
		}
		_ = resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return "connected"
		}
		return "invalid_status"
	}

	// check local socket or pipe
	if runtime.GOOS == "windows" {
		if _, err := os.Stat(`\\.\pipe\docker_engine`); err == nil {
			return "connected"
		}
		// fallback to TCP 2375
		client := &http.Client{Timeout: 1 * time.Second}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:2375/_ping", nil)
		if err == nil {
			if resp, err := client.Do(req); err == nil {
				_ = resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					return "connected"
				}
			}
		}
		return "pipe_not_found"
	}

	conn, err := net.DialTimeout("unix", "/var/run/docker.sock", 2*time.Second)
	if err != nil {
		return "socket_unreachable: " + err.Error()
	}
	_ = conn.Close()
	return "connected"
}

package handler

import (
	"net/http"

	"github.com/cifo-monitoring/backend/internal/middleware"
	"github.com/cifo-monitoring/backend/internal/service"
	"github.com/labstack/echo/v4"
)

// MonitoringHandler handles dashboard telemetry endpoints
type MonitoringHandler struct {
	monitoringService service.MonitoringService
}

// NewMonitoringHandler creates monitoring handler
func NewMonitoringHandler(monitoringService service.MonitoringService) *MonitoringHandler {
	return &MonitoringHandler{monitoringService: monitoringService}
}

// GetStats returns aggregated dashboard statistics
func (h *MonitoringHandler) GetStats(c echo.Context) error {
	ctx := c.Request().Context()
	stats, err := h.monitoringService.GetDashboardStats(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, stats)
}

// GetCPUMetrics returns time-series CPU utilization
func (h *MonitoringHandler) GetCPUMetrics(c echo.Context) error {
	ctx := c.Request().Context()
	timeRange := c.QueryParam("range")
	metrics, err := h.monitoringService.GetCPUMetrics(ctx, timeRange)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"range": timeRange,
		"data":  metrics,
	})
}

// GetMemoryMetrics returns time-series memory utilization
func (h *MonitoringHandler) GetMemoryMetrics(c echo.Context) error {
	ctx := c.Request().Context()
	timeRange := c.QueryParam("range")
	metrics, err := h.monitoringService.GetMemoryMetrics(ctx, timeRange)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"range": timeRange,
		"data":  metrics,
	})
}

// GetNetworkMetrics returns time-series network I/O
func (h *MonitoringHandler) GetNetworkMetrics(c echo.Context) error {
	ctx := c.Request().Context()
	timeRange := c.QueryParam("range")
	metrics, err := h.monitoringService.GetNetworkMetrics(ctx, timeRange)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"range": timeRange,
		"data":  metrics,
	})
}

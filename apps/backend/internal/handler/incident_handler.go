package handler

import (
	"net/http"
	"strconv"

	"github.com/cifo-monitoring/backend/internal/middleware"
	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/cifo-monitoring/backend/internal/service"
	"github.com/labstack/echo/v4"
)

// IncidentHandler handles incident endpoints
type IncidentHandler struct {
	incidentService service.IncidentService
}

// NewIncidentHandler creates handler
func NewIncidentHandler(svc service.IncidentService) *IncidentHandler {
	return &IncidentHandler{incidentService: svc}
}

// HandleAlertmanagerWebhook receives alertmanager webhook
func (h *IncidentHandler) HandleAlertmanagerWebhook(c echo.Context) error {
	var payload model.AlertmanagerWebhookPayload
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, middleware.ProblemDetail{
			Title:    "Bad Request",
			Status:   http.StatusBadRequest,
			Detail:   "invalid alertmanager webhook payload",
			Instance: c.Request().RequestURI,
		})
	}

	ctx := c.Request().Context()
	if err := h.incidentService.ProcessAlertmanagerWebhook(ctx, &payload); err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": "alertmanager webhook processed",
	})
}

// ListIncidents lists incidents
func (h *IncidentHandler) ListIncidents(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	filter := model.IncidentFilter{
		Status:   c.QueryParam("status"),
		Severity: c.QueryParam("severity"),
		Source:   c.QueryParam("source"),
		Search:   c.QueryParam("search"),
		Page:     page,
		Limit:    limit,
	}

	ctx := c.Request().Context()
	items, total, err := h.incidentService.ListIncidents(ctx, filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  items,
		"total": total,
		"page":  filter.Page,
		"limit": filter.Limit,
	})
}

// GetIncidentStats returns stats
func (h *IncidentHandler) GetIncidentStats(c echo.Context) error {
	ctx := c.Request().Context()
	stats, err := h.incidentService.GetIncidentStats(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": stats,
	})
}

// GetIncident returns incident detail
func (h *IncidentHandler) GetIncident(c echo.Context) error {
	id := c.Param("id")
	ctx := c.Request().Context()

	detail, err := h.incidentService.GetIncident(ctx, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, middleware.ProblemDetail{
			Title:    "Not Found",
			Status:   http.StatusNotFound,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": detail,
	})
}

// AcknowledgeIncident acknowledges incident
func (h *IncidentHandler) AcknowledgeIncident(c echo.Context) error {
	id := c.Param("id")
	actor := extractActor(c)
	ip := c.RealIP()
	ctx := c.Request().Context()

	if err := h.incidentService.AcknowledgeIncident(ctx, id, actor, ip); err != nil {
		return c.JSON(http.StatusBadRequest, middleware.ProblemDetail{
			Title:    "Bad Request",
			Status:   http.StatusBadRequest,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": "incident acknowledged successfully",
	})
}

// ResolveIncident resolves incident
func (h *IncidentHandler) ResolveIncident(c echo.Context) error {
	id := c.Param("id")
	actor := extractActor(c)
	ip := c.RealIP()
	ctx := c.Request().Context()

	if err := h.incidentService.ResolveIncident(ctx, id, actor, ip); err != nil {
		return c.JSON(http.StatusBadRequest, middleware.ProblemDetail{
			Title:    "Bad Request",
			Status:   http.StatusBadRequest,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": "incident resolved successfully",
	})
}

// CloseIncident closes incident
func (h *IncidentHandler) CloseIncident(c echo.Context) error {
	id := c.Param("id")
	actor := extractActor(c)
	ip := c.RealIP()
	ctx := c.Request().Context()

	if err := h.incidentService.CloseIncident(ctx, id, actor, ip); err != nil {
		return c.JSON(http.StatusBadRequest, middleware.ProblemDetail{
			Title:    "Bad Request",
			Status:   http.StatusBadRequest,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": "incident closed successfully",
	})
}

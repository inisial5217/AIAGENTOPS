package handler

import (
	"net/http"

	"github.com/cifo-monitoring/backend/internal/middleware"
	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/cifo-monitoring/backend/internal/service"
	"github.com/labstack/echo/v4"
)

// ArgoCDHandler handles argocd endpoints
type ArgoCDHandler struct {
	argoService service.ArgoCDService
}

// NewArgoCDHandler creates argocd handler
func NewArgoCDHandler(argoService service.ArgoCDService) *ArgoCDHandler {
	return &ArgoCDHandler{argoService: argoService}
}

// ListApplications returns argocd applications
func (h *ArgoCDHandler) ListApplications(c echo.Context) error {
	namespace := c.QueryParam("namespace")
	ctx := c.Request().Context()

	apps, err := h.argoService.ListApplications(ctx, namespace)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  apps,
		"total": len(apps),
	})
}

// GetApplication returns application detail
func (h *ArgoCDHandler) GetApplication(c echo.Context) error {
	name := c.Param("name")
	namespace := c.QueryParam("namespace")
	ctx := c.Request().Context()

	app, err := h.argoService.GetApplication(ctx, namespace, name)
	if err != nil {
		return c.JSON(http.StatusNotFound, middleware.ProblemDetail{
			Title:    "Not Found",
			Status:   http.StatusNotFound,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, app)
}

// SyncApplication triggers application sync
func (h *ArgoCDHandler) SyncApplication(c echo.Context) error {
	name := c.Param("name")
	namespace := c.QueryParam("namespace")
	ctx := c.Request().Context()
	actor := extractActor(c)
	ip := c.RealIP()

	var req model.ArgoSyncRequest
	if err := c.Bind(&req); err != nil {
		// default options if empty body
		req = model.ArgoSyncRequest{
			Prune:  false,
			DryRun: false,
		}
	}

	if err := h.argoService.SyncApplication(ctx, namespace, name, req, actor, ip); err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": "application sync triggered successfully",
	})
}

// GetOverview returns application status overview
func (h *ArgoCDHandler) GetOverview(c echo.Context) error {
	namespace := c.QueryParam("namespace")
	ctx := c.Request().Context()

	overview, err := h.argoService.GetOverview(ctx, namespace)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, overview)
}

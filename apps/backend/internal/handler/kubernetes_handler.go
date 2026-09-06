package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/cifo-monitoring/backend/internal/middleware"
	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/cifo-monitoring/backend/internal/service"
	"github.com/labstack/echo/v4"
)

// KubernetesHandler handles kubernetes endpoints
type KubernetesHandler struct {
	k8sService service.KubernetesService
}

// NewKubernetesHandler creates kubernetes handler
func NewKubernetesHandler(k8sService service.KubernetesService) *KubernetesHandler {
	return &KubernetesHandler{k8sService: k8sService}
}

// ListPods returns pod summaries
func (h *KubernetesHandler) ListPods(c echo.Context) error {
	namespace := c.QueryParam("namespace")
	ctx := c.Request().Context()

	pods, err := h.k8sService.ListPods(ctx, namespace)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  pods,
		"total": len(pods),
	})
}

// GetPod returns pod detail
func (h *KubernetesHandler) GetPod(c echo.Context) error {
	namespace := c.Param("namespace")
	name := c.Param("name")
	ctx := c.Request().Context()

	detail, err := h.k8sService.GetPod(ctx, namespace, name)
	if err != nil {
		return c.JSON(http.StatusNotFound, middleware.ProblemDetail{
			Title:    "Not Found",
			Status:   http.StatusNotFound,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, detail)
}

// GetPodLogs returns container logs
func (h *KubernetesHandler) GetPodLogs(c echo.Context) error {
	namespace := c.Param("namespace")
	name := c.Param("name")
	container := c.QueryParam("container")
	tailStr := c.QueryParam("tail")

	tail := int64(200)
	if tailStr != "" {
		if parsed, err := strconv.ParseInt(tailStr, 10, 64); err == nil && parsed > 0 {
			tail = parsed
		}
	}

	ctx := c.Request().Context()
	stream, err := h.k8sService.GetPodLogs(ctx, namespace, name, container, tail)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}
	defer stream.Close()

	logBytes, err := io.ReadAll(stream)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"namespace": namespace,
		"pod":       name,
		"container": container,
		"tail":      tail,
		"logs":      string(logBytes),
	})
}

// ListDeployments returns deployment summaries
func (h *KubernetesHandler) ListDeployments(c echo.Context) error {
	namespace := c.QueryParam("namespace")
	ctx := c.Request().Context()

	deployments, err := h.k8sService.ListDeployments(ctx, namespace)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  deployments,
		"total": len(deployments),
	})
}

// GetDeployment returns deployment detail
func (h *KubernetesHandler) GetDeployment(c echo.Context) error {
	namespace := c.Param("namespace")
	name := c.Param("name")
	ctx := c.Request().Context()

	detail, err := h.k8sService.GetDeployment(ctx, namespace, name)
	if err != nil {
		return c.JSON(http.StatusNotFound, middleware.ProblemDetail{
			Title:    "Not Found",
			Status:   http.StatusNotFound,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, detail)
}

// RestartDeployment handles rollout restart
func (h *KubernetesHandler) RestartDeployment(c echo.Context) error {
	namespace := c.Param("namespace")
	name := c.Param("name")
	ctx := c.Request().Context()
	actor := extractActor(c)
	ip := c.RealIP()

	if err := h.k8sService.RestartDeployment(ctx, namespace, name, actor, ip); err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": "deployment rollout restarted successfully",
	})
}

// ScaleDeployment handles replica scaling
func (h *KubernetesHandler) ScaleDeployment(c echo.Context) error {
	namespace := c.Param("namespace")
	name := c.Param("name")
	ctx := c.Request().Context()
	actor := extractActor(c)
	ip := c.RealIP()

	var req model.ScaleDeploymentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, middleware.ProblemDetail{
			Title:    "Bad Request",
			Status:   http.StatusBadRequest,
			Detail:   "invalid scale request body",
			Instance: c.Request().RequestURI,
		})
	}

	if req.Replicas < 0 || req.Replicas > 100 {
		return c.JSON(http.StatusBadRequest, middleware.ProblemDetail{
			Title:    "Bad Request",
			Status:   http.StatusBadRequest,
			Detail:   "replicas must be between 0 and 100",
			Instance: c.Request().RequestURI,
		})
	}

	if err := h.k8sService.ScaleDeployment(ctx, namespace, name, req.Replicas, actor, ip); err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":   "success",
		"message":  fmt.Sprintf("deployment scaled to %d replicas", req.Replicas),
		"replicas": req.Replicas,
	})
}

// ListNodes returns cluster nodes
func (h *KubernetesHandler) ListNodes(c echo.Context) error {
	ctx := c.Request().Context()
	nodes, err := h.k8sService.ListNodes(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  nodes,
		"total": len(nodes),
	})
}

// ListServices returns cluster services
func (h *KubernetesHandler) ListServices(c echo.Context) error {
	namespace := c.QueryParam("namespace")
	ctx := c.Request().Context()

	services, err := h.k8sService.ListServices(ctx, namespace)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  services,
		"total": len(services),
	})
}

// GetClusterOverview returns cluster overview
func (h *KubernetesHandler) GetClusterOverview(c echo.Context) error {
	ctx := c.Request().Context()
	overview, err := h.k8sService.GetClusterOverview(ctx)
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

func extractActor(c echo.Context) string {
	if actor, ok := c.Get("user_id").(string); ok && actor != "" {
		return actor
	}
	if email, ok := c.Get("user_email").(string); ok && email != "" {
		return email
	}
	return "anonymous"
}

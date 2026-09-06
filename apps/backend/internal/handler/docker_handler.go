package handler

import (
	"net/http"
	"strconv"

	"github.com/cifo-monitoring/backend/internal/middleware"
	"github.com/cifo-monitoring/backend/internal/service"
	"github.com/labstack/echo/v4"
)

// DockerHandler handles docker endpoints
type DockerHandler struct {
	dockerService service.DockerService
}

// NewDockerHandler creates docker handler
func NewDockerHandler(dockerService service.DockerService) *DockerHandler {
	return &DockerHandler{dockerService: dockerService}
}

// ListContainers returns container list
func (h *DockerHandler) ListContainers(c echo.Context) error {
	status := c.QueryParam("status")
	ctx := c.Request().Context()

	containers, err := h.dockerService.ListContainers(ctx, status)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  containers,
		"total": len(containers),
	})
}

// GetContainer returns container detail
func (h *DockerHandler) GetContainer(c echo.Context) error {
	id := c.Param("id")
	ctx := c.Request().Context()

	detail, err := h.dockerService.GetContainer(ctx, id)
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

// GetContainerStats returns realtime resource stats
func (h *DockerHandler) GetContainerStats(c echo.Context) error {
	id := c.Param("id")
	ctx := c.Request().Context()

	stats, err := h.dockerService.GetContainerStats(ctx, id)
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

// GetContainerLogs returns recent log lines
func (h *DockerHandler) GetContainerLogs(c echo.Context) error {
	id := c.Param("id")
	tailStr := c.QueryParam("tail")
	tail := 200
	if tailStr != "" {
		if t, err := strconv.Atoi(tailStr); err == nil && t > 0 {
			tail = t
		}
	}

	ctx := c.Request().Context()
	logs, err := h.dockerService.GetContainerLogs(ctx, id, tail)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"container_id": id,
		"tail":         tail,
		"logs":         logs,
	})
}

// RestartContainer handles container restart
func (h *DockerHandler) RestartContainer(c echo.Context) error {
	id := c.Param("id")
	ctx := c.Request().Context()
	actor, _ := c.Get("user_id").(string)
	ip := c.RealIP()

	if err := h.dockerService.RestartContainer(ctx, id, actor, ip); err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": "container restarted successfully",
	})
}

// StopContainer handles container stop
func (h *DockerHandler) StopContainer(c echo.Context) error {
	id := c.Param("id")
	ctx := c.Request().Context()
	actor, _ := c.Get("user_id").(string)
	ip := c.RealIP()

	if err := h.dockerService.StopContainer(ctx, id, actor, ip); err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": "container stopped successfully",
	})
}

// ListImages returns images
func (h *DockerHandler) ListImages(c echo.Context) error {
	ctx := c.Request().Context()
	images, err := h.dockerService.ListImages(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  images,
		"total": len(images),
	})
}

// ListVolumes returns volumes
func (h *DockerHandler) ListVolumes(c echo.Context) error {
	ctx := c.Request().Context()
	vols, err := h.dockerService.ListVolumes(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  vols,
		"total": len(vols),
	})
}

// ListNetworks returns networks
func (h *DockerHandler) ListNetworks(c echo.Context) error {
	ctx := c.Request().Context()
	nets, err := h.dockerService.ListNetworks(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  nets,
		"total": len(nets),
	})
}

// GetSystemInfo returns system information
func (h *DockerHandler) GetSystemInfo(c echo.Context) error {
	ctx := c.Request().Context()
	info, err := h.dockerService.GetSystemInfo(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, info)
}

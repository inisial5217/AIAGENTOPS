package handler

import (
	"net/http"
	"strconv"

	"github.com/cifo-monitoring/backend/internal/middleware"
	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/cifo-monitoring/backend/internal/service"
	"github.com/cifo-monitoring/backend/pkg/apperror"
	"github.com/labstack/echo/v4"
)

// SettingsHandler handles settings endpoints
type SettingsHandler struct {
	settingsService service.SettingsService
}

// NewSettingsHandler constructor
func NewSettingsHandler(settingsService service.SettingsService) *SettingsHandler {
	return &SettingsHandler{settingsService: settingsService}
}

// GetSettings retrieves system settings
func (h *SettingsHandler) GetSettings(c echo.Context) error {
	settings, err := h.settingsService.GetSettings(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": settings,
	})
}

// UpdateSettings updates system settings
func (h *SettingsHandler) UpdateSettings(c echo.Context) error {
	var req model.UpdateSettingsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, middleware.ProblemDetail{
			Title:    "Bad Request",
			Status:   http.StatusBadRequest,
			Detail:   "invalid request payload",
			Instance: c.Request().RequestURI,
		})
	}

	actor := getActorID(c)
	ip := c.RealIP()

	updated, err := h.settingsService.UpdateSettings(c.Request().Context(), &req, actor, ip)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": updated,
	})
}

// TestNotification sends test alert
func (h *SettingsHandler) TestNotification(c echo.Context) error {
	actor := getActorID(c)
	ip := c.RealIP()

	if err := h.settingsService.TestTelegramNotification(c.Request().Context(), actor, ip); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return c.JSON(appErr.HTTPStatus, middleware.ProblemDetail{
				Title:    "Bad Request",
				Status:   appErr.HTTPStatus,
				Detail:   appErr.Message,
				Instance: c.Request().RequestURI,
			})
		}
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Test notification dispatched successfully",
	})
}

// ListUsers lists system users
func (h *SettingsHandler) ListUsers(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 50
	}
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	users, total, err := h.settingsService.ListUsers(c.Request().Context(), limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"users": users,
		"total": total,
	})
}

// UpdateUserRole updates user role
func (h *SettingsHandler) UpdateUserRole(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, middleware.ProblemDetail{
			Title:    "Bad Request",
			Status:   http.StatusBadRequest,
			Detail:   "user id required",
			Instance: c.Request().RequestURI,
		})
	}

	var req model.UpdateUserRoleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, middleware.ProblemDetail{
			Title:    "Bad Request",
			Status:   http.StatusBadRequest,
			Detail:   "invalid request payload",
			Instance: c.Request().RequestURI,
		})
	}

	actor := getActorID(c)
	ip := c.RealIP()

	user, err := h.settingsService.UpdateUserRole(c.Request().Context(), id, req.Role, actor, ip)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return c.JSON(appErr.HTTPStatus, middleware.ProblemDetail{
				Title:    "Error",
				Status:   appErr.HTTPStatus,
				Detail:   appErr.Message,
				Instance: c.Request().RequestURI,
			})
		}
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, user)
}

// DeactivateUser disables user account
func (h *SettingsHandler) DeactivateUser(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, middleware.ProblemDetail{
			Title:    "Bad Request",
			Status:   http.StatusBadRequest,
			Detail:   "user id required",
			Instance: c.Request().RequestURI,
		})
	}

	actor := getActorID(c)
	ip := c.RealIP()

	user, err := h.settingsService.DeactivateUser(c.Request().Context(), id, actor, ip)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return c.JSON(appErr.HTTPStatus, middleware.ProblemDetail{
				Title:    "Error",
				Status:   appErr.HTTPStatus,
				Detail:   appErr.Message,
				Instance: c.Request().RequestURI,
			})
		}
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, user)
}

// ReactivateUser enables user account
func (h *SettingsHandler) ReactivateUser(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, middleware.ProblemDetail{
			Title:    "Bad Request",
			Status:   http.StatusBadRequest,
			Detail:   "user id required",
			Instance: c.Request().RequestURI,
		})
	}

	actor := getActorID(c)
	ip := c.RealIP()

	user, err := h.settingsService.ReactivateUser(c.Request().Context(), id, actor, ip)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return c.JSON(appErr.HTTPStatus, middleware.ProblemDetail{
				Title:    "Error",
				Status:   appErr.HTTPStatus,
				Detail:   appErr.Message,
				Instance: c.Request().RequestURI,
			})
		}
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, user)
}

func getActorID(c echo.Context) string {
	// extract actor identifier
	if email, ok := c.Get("email").(string); ok && email != "" {
		return email
	}
	if uid, ok := c.Get("user_id").(string); ok && uid != "" {
		return uid
	}
	return "admin@cifo.local"
}

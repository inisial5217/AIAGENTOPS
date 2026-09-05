package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/cifo-monitoring/backend/internal/middleware"
	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/cifo-monitoring/backend/internal/repository"
	"github.com/cifo-monitoring/backend/internal/service"
	"github.com/labstack/echo/v4"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *service.AuthService
	userRepo    repository.UserRepository
	auditRepo   repository.AuditRepository
}

// NewAuthHandler creates auth handler
func NewAuthHandler(
	authService *service.AuthService,
	userRepo repository.UserRepository,
	auditRepo repository.AuditRepository,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userRepo:    userRepo,
		auditRepo:   auditRepo,
	}
}

// LoginRequest credentials request body
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// Login authenticates with credentials
func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, middleware.ProblemDetail{
			Title:    "Bad Request",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid request body",
			Instance: c.Request().RequestURI,
		})
	}

	if req.Username == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, middleware.ProblemDetail{
			Title:    "Bad Request",
			Status:   http.StatusBadRequest,
			Detail:   "Username and password are required",
			Instance: c.Request().RequestURI,
		})
	}

	ctx := c.Request().Context()
	ip := c.RealIP()
	ua := c.Request().UserAgent()

	tokenResp, err := h.authService.DirectLogin(ctx, req.Username, req.Password)
	if err != nil {
		details, _ := json.Marshal(map[string]string{"username": req.Username, "reason": err.Error()})
		_ = h.authService.LogAuditEvent(ctx, &model.AuditLog{
			Timestamp:    time.Now(),
			ActorType:    "user",
			ActorID:      req.Username,
			Action:       "login",
			ResourceType: "auth",
			Details:      details,
			IPAddress:    &ip,
			UserAgent:    &ua,
			Result:       "failure",
		})

		return c.JSON(http.StatusUnauthorized, middleware.ProblemDetail{
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	details, _ := json.Marshal(map[string]string{"username": req.Username})
	_ = h.authService.LogAuditEvent(ctx, &model.AuditLog{
		Timestamp:    time.Now(),
		ActorType:    "user",
		ActorID:      req.Username,
		Action:       "login",
		ResourceType: "auth",
		Details:      details,
		IPAddress:    &ip,
		UserAgent:    &ua,
		Result:       "success",
	})

	return c.JSON(http.StatusOK, tokenResp)
}

// Me returns current user profile
func (h *AuthHandler) Me(c echo.Context) error {
	user, ok := c.Get("user").(*model.User)
	if ok && user != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"user": user,
		})
	}

	claims, ok := c.Get("claims").(*service.AuthClaims)
	if ok && claims != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"user": map[string]interface{}{
				"id":          claims.Subject,
				"email":       claims.Email,
				"name":        claims.Name,
				"role":        c.Get("user_role"),
				"keycloak_id": claims.Subject,
			},
		})
	}

	return c.JSON(http.StatusUnauthorized, middleware.ProblemDetail{
		Title:    "Unauthorized",
		Status:   http.StatusUnauthorized,
		Detail:   "User not authenticated",
		Instance: c.Request().RequestURI,
	})
}

// Logout revokes active token
func (h *AuthHandler) Logout(c echo.Context) error {
	token, _ := c.Get("token").(string)
	claims, _ := c.Get("claims").(*service.AuthClaims)
	actorID, _ := c.Get("user_id").(string)
	if actorID == "" {
		actorID, _ = c.Get("user_email").(string)
	}

	ctx := c.Request().Context()
	ip := c.RealIP()
	ua := c.Request().UserAgent()

	if token != "" && claims != nil && claims.ExpiresAt != nil {
		_ = h.authService.BlacklistToken(ctx, token, claims.ExpiresAt.Time)
	}

	_ = h.authService.LogAuditEvent(ctx, &model.AuditLog{
		Timestamp:    time.Now(),
		ActorType:    "user",
		ActorID:      actorID,
		Action:       "logout",
		ResourceType: "auth",
		IPAddress:    &ip,
		UserAgent:    &ua,
		Result:       "success",
	})

	return c.JSON(http.StatusOK, map[string]string{
		"message": "successfully logged out",
	})
}

// ListUsers admin lists all users
func (h *AuthHandler) ListUsers(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	users, total, err := h.userRepo.List(c.Request().Context(), limit, offset)
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

// ListAuditLogs admin lists audit logs
func (h *AuthHandler) ListAuditLogs(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	logs, total, err := h.auditRepo.List(c.Request().Context(), limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"audit_logs": logs,
		"total":      total,
	})
}

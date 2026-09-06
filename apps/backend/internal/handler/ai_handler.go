package handler

import (
	"fmt"
	"net/http"

	"github.com/cifo-monitoring/backend/internal/middleware"
	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/cifo-monitoring/backend/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// AIHandler handles ai endpoints
type AIHandler struct {
	aiService service.AIService
}

// NewAIHandler constructor
func NewAIHandler(aiService service.AIService) *AIHandler {
	return &AIHandler{aiService: aiService}
}

// HandleChat process chat message
func (h *AIHandler) HandleChat(c echo.Context) error {
	// extract user context
	userID, ok := getUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, middleware.ProblemDetail{
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "user unauthenticated",
			Instance: c.Request().RequestURI,
		})
	}
	role := getUserRole(c)

	var req model.AIChatRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, middleware.ProblemDetail{
			Title:    "Bad Request",
			Status:   http.StatusBadRequest,
			Detail:   "invalid request payload",
			Instance: c.Request().RequestURI,
		})
	}
	if req.Message == "" {
		return c.JSON(http.StatusBadRequest, middleware.ProblemDetail{
			Title:    "Bad Request",
			Status:   http.StatusBadRequest,
			Detail:   "message is required",
			Instance: c.Request().RequestURI,
		})
	}

	resp, err := h.aiService.ProcessChat(c.Request().Context(), userID, role, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, resp)
}

// HandleListSessions get user sessions
func (h *AIHandler) HandleListSessions(c echo.Context) error {
	// extract user id
	userID, ok := getUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, middleware.ProblemDetail{
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "user unauthenticated",
			Instance: c.Request().RequestURI,
		})
	}

	sessions, err := h.aiService.ListSessions(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"sessions": sessions})
}

// HandleGetSessionMessages get messages
func (h *AIHandler) HandleGetSessionMessages(c echo.Context) error {
	// extract user and id
	userID, ok := getUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, middleware.ProblemDetail{
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "user unauthenticated",
			Instance: c.Request().RequestURI,
		})
	}

	idStr := c.Param("id")
	sessionID, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, middleware.ProblemDetail{
			Title:    "Bad Request",
			Status:   http.StatusBadRequest,
			Detail:   "invalid session id",
			Instance: c.Request().RequestURI,
		})
	}

	messages, err := h.aiService.GetSessionMessages(c.Request().Context(), sessionID, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"messages": messages})
}

// HandleApproveTool approve write tool
func (h *AIHandler) HandleApproveTool(c echo.Context) error {
	// extract auth context
	userID, ok := getUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, middleware.ProblemDetail{
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "user unauthenticated",
			Instance: c.Request().RequestURI,
		})
	}
	role := getUserRole(c)

	idStr := c.Param("id")
	approvalID, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, middleware.ProblemDetail{
			Title:    "Bad Request",
			Status:   http.StatusBadRequest,
			Detail:   "invalid approval id",
			Instance: c.Request().RequestURI,
		})
	}

	result, err := h.aiService.ApproveTool(c.Request().Context(), approvalID, userID, role)
	if err != nil {
		return c.JSON(http.StatusBadRequest, middleware.ProblemDetail{
			Title:    "Bad Request",
			Status:   http.StatusBadRequest,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, result)
}

// HandleRejectTool reject write tool
func (h *AIHandler) HandleRejectTool(c echo.Context) error {
	// extract auth context
	userID, ok := getUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, middleware.ProblemDetail{
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "user unauthenticated",
			Instance: c.Request().RequestURI,
		})
	}
	role := getUserRole(c)

	idStr := c.Param("id")
	approvalID, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, middleware.ProblemDetail{
			Title:    "Bad Request",
			Status:   http.StatusBadRequest,
			Detail:   "invalid approval id",
			Instance: c.Request().RequestURI,
		})
	}

	result, err := h.aiService.RejectTool(c.Request().Context(), approvalID, userID, role)
	if err != nil {
		return c.JSON(http.StatusBadRequest, middleware.ProblemDetail{
			Title:    "Bad Request",
			Status:   http.StatusBadRequest,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, result)
}

// HandleGetUsage get usage metrics
func (h *AIHandler) HandleGetUsage(c echo.Context) error {
	// extract auth context
	userID, ok := getUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, middleware.ProblemDetail{
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "user unauthenticated",
			Instance: c.Request().RequestURI,
		})
	}
	role := getUserRole(c)

	stats, err := h.aiService.GetUsage(c.Request().Context(), userID, role)
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

// HandleListModels returns available models
func (h *AIHandler) HandleListModels(c echo.Context) error {
	models, err := h.aiService.ListModels(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   err.Error(),
			Instance: c.Request().RequestURI,
		})
	}
	return c.JSON(http.StatusOK, models)
}

// HandleGenerateRCA trigger incident rca
func (h *AIHandler) HandleGenerateRCA(c echo.Context) error {
	// extract incident id
	idStr := c.Param("id")
	incidentID, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, middleware.ProblemDetail{
			Title:    "Bad Request",
			Status:   http.StatusBadRequest,
			Detail:   "invalid incident id",
			Instance: c.Request().RequestURI,
		})
	}

	rcaResp, err := h.aiService.GenerateRCAForIncident(c.Request().Context(), incidentID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, middleware.ProblemDetail{
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   fmt.Sprintf("generate rca: %v", err),
			Instance: c.Request().RequestURI,
		})
	}

	return c.JSON(http.StatusOK, rcaResp)
}

func getUserID(c echo.Context) (uuid.UUID, bool) {
	// parse user id
	if u, ok := c.Get("user_id").(uuid.UUID); ok && u != uuid.Nil {
		return u, true
	}
	if s, ok := c.Get("user_id").(string); ok {
		if u, err := uuid.Parse(s); err == nil && u != uuid.Nil {
			return u, true
		}
	}
	return uuid.Nil, false
}

func getUserRole(c echo.Context) string {
	// extract user role
	if r, ok := c.Get("user_role").(string); ok && r != "" {
		return r
	}
	if r, ok := c.Get("role").(string); ok && r != "" {
		return r
	}
	return "viewer"
}

package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAIService struct {
	mock.Mock
}

func (m *mockAIService) ProcessChat(ctx context.Context, userID uuid.UUID, role string, req *model.AIChatRequest) (*model.AIChatResponse, error) {
	args := m.Called(ctx, userID, role, req)
	if resp, ok := args.Get(0).(*model.AIChatResponse); ok {
		return resp, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAIService) ListSessions(ctx context.Context, userID uuid.UUID) ([]model.AISession, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]model.AISession), args.Error(1)
}

func (m *mockAIService) GetSessionMessages(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) ([]model.AIMessage, error) {
	args := m.Called(ctx, sessionID, userID)
	return args.Get(0).([]model.AIMessage), args.Error(1)
}

func (m *mockAIService) ApproveTool(ctx context.Context, approvalID uuid.UUID, userID uuid.UUID, role string) (*model.AIActionAuditLog, error) {
	args := m.Called(ctx, approvalID, userID, role)
	if audit, ok := args.Get(0).(*model.AIActionAuditLog); ok {
		return audit, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAIService) RejectTool(ctx context.Context, approvalID uuid.UUID, userID uuid.UUID, role string) (*model.AIActionAuditLog, error) {
	args := m.Called(ctx, approvalID, userID, role)
	if audit, ok := args.Get(0).(*model.AIActionAuditLog); ok {
		return audit, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAIService) GetUsage(ctx context.Context, userID uuid.UUID, role string) (*model.AIUsageStats, error) {
	args := m.Called(ctx, userID, role)
	if stats, ok := args.Get(0).(*model.AIUsageStats); ok {
		return stats, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAIService) GenerateRCAForIncident(ctx context.Context, incidentID uuid.UUID) (*model.RCAResponse, error) {
	args := m.Called(ctx, incidentID)
	if resp, ok := args.Get(0).(*model.RCAResponse); ok {
		return resp, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAIService) ListModels(ctx context.Context) (map[string]interface{}, error) {
	args := m.Called(ctx)
	if resp, ok := args.Get(0).(map[string]interface{}); ok {
		return resp, args.Error(1)
	}
	return nil, args.Error(1)
}

func TestHandleChat_Unauthorized(t *testing.T) {
	e := echo.New()
	svc := new(mockAIService)
	h := NewAIHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/chat", bytes.NewBufferString(`{"message":"hello"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.HandleChat(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandleChat_Success(t *testing.T) {
	e := echo.New()
	svc := new(mockAIService)
	h := NewAIHandler(svc)

	userID := uuid.New()
	chatReq := model.AIChatRequest{Message: "hello"}
	chatResp := &model.AIChatResponse{
		SessionID: uuid.New(),
		Content:   "Hello from AI",
		ModelUsed: "gemini-2.0-flash",
	}

	svc.On("ProcessChat", mock.Anything, userID, "viewer", &chatReq).Return(chatResp, nil)

	body, _ := json.Marshal(chatReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/chat", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", userID)
	c.Set("user_role", "viewer")

	err := h.HandleChat(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var out model.AIChatResponse
	err = json.NewDecoder(rec.Body).Decode(&out)
	assert.NoError(t, err)
	assert.Equal(t, "Hello from AI", out.Content)
}

func TestHandleGetUsage_Success(t *testing.T) {
	e := echo.New()
	svc := new(mockAIService)
	h := NewAIHandler(svc)

	userID := uuid.New()
	stats := &model.AIUsageStats{
		TotalTokens:  150,
		TotalCostUSD: 0.005,
	}

	svc.On("GetUsage", mock.Anything, userID, "admin").Return(stats, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai/usage", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", userID)
	c.Set("user_role", "admin")

	err := h.HandleGetUsage(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

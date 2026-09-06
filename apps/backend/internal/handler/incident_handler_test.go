package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockIncidentService mock
type mockIncidentService struct {
	mock.Mock
}

func (m *mockIncidentService) ProcessAlertmanagerWebhook(ctx context.Context, payload *model.AlertmanagerWebhookPayload) error {
	args := m.Called(ctx, payload)
	return args.Error(0)
}

func (m *mockIncidentService) CreateIncident(ctx context.Context, req *model.CreateIncidentRequest) (*model.Incident, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*model.Incident), args.Error(1)
}

func (m *mockIncidentService) ListIncidents(ctx context.Context, filter model.IncidentFilter) ([]*model.IncidentSummary, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*model.IncidentSummary), args.Int(1), args.Error(2)
}

func (m *mockIncidentService) GetIncident(ctx context.Context, id string) (*model.IncidentDetail, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.IncidentDetail), args.Error(1)
}

func (m *mockIncidentService) GetIncidentStats(ctx context.Context) (*model.IncidentStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(*model.IncidentStats), args.Error(1)
}

func (m *mockIncidentService) AcknowledgeIncident(ctx context.Context, id string, userID string, ipAddress string) error {
	args := m.Called(ctx, id, userID, ipAddress)
	return args.Error(0)
}

func (m *mockIncidentService) ResolveIncident(ctx context.Context, id string, userID string, ipAddress string) error {
	args := m.Called(ctx, id, userID, ipAddress)
	return args.Error(0)
}

func (m *mockIncidentService) CloseIncident(ctx context.Context, id string, userID string, ipAddress string) error {
	args := m.Called(ctx, id, userID, ipAddress)
	return args.Error(0)
}

func (m *mockIncidentService) EscalateIncident(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockIncidentService) CheckEscalations(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestIncidentHandler_HandleAlertmanagerWebhook(t *testing.T) {
	e := echo.New()
	mockSvc := new(mockIncidentService)
	h := NewIncidentHandler(mockSvc)

	mockSvc.On("ProcessAlertmanagerWebhook", mock.Anything, mock.AnythingOfType("*model.AlertmanagerWebhookPayload")).Return(nil)

	payload := map[string]interface{}{
		"status": "firing",
		"alerts": []map[string]interface{}{
			{
				"status": "firing",
				"labels": map[string]string{
					"alertname": "PodCrashLooping",
					"severity":  "critical",
				},
			},
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/alertmanager", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.HandleAlertmanagerWebhook(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestIncidentHandler_ListIncidents(t *testing.T) {
	e := echo.New()
	mockSvc := new(mockIncidentService)
	h := NewIncidentHandler(mockSvc)

	expected := []*model.IncidentSummary{
		{ID: "inc-1", Title: "CrashLoopBackOff", Status: "open", Severity: "critical"},
	}
	mockSvc.On("ListIncidents", mock.Anything, mock.MatchedBy(func(f model.IncidentFilter) bool {
		return f.Status == "open"
	})).Return(expected, 1, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents?status=open", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListIncidents(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestIncidentHandler_AcknowledgeIncident(t *testing.T) {
	e := echo.New()
	mockSvc := new(mockIncidentService)
	h := NewIncidentHandler(mockSvc)

	mockSvc.On("AcknowledgeIncident", mock.Anything, "inc-1", "user-uuid", mock.Anything).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/incidents/inc-1/acknowledge", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("inc-1")
	c.Set("user_id", "user-uuid")

	err := h.AcknowledgeIncident(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

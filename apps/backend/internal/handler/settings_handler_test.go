package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cifo-monitoring/backend/internal/handler"
	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSettingsService struct {
	mock.Mock
}

func (m *mockSettingsService) GetSettings(ctx context.Context) (*model.CombinedSettings, error) {
	args := m.Called(ctx)
	if s, ok := args.Get(0).(*model.CombinedSettings); ok {
		return s, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockSettingsService) UpdateSettings(ctx context.Context, req *model.UpdateSettingsRequest, actor string, ip string) (*model.CombinedSettings, error) {
	args := m.Called(ctx, req, actor, ip)
	if s, ok := args.Get(0).(*model.CombinedSettings); ok {
		return s, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockSettingsService) ListUsers(ctx context.Context, limit, offset int) ([]*model.User, int, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*model.User), args.Int(1), args.Error(2)
}

func (m *mockSettingsService) UpdateUserRole(ctx context.Context, userID string, role string, actor string, ip string) (*model.User, error) {
	args := m.Called(ctx, userID, role, actor, ip)
	if u, ok := args.Get(0).(*model.User); ok {
		return u, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockSettingsService) DeactivateUser(ctx context.Context, userID string, actor string, ip string) (*model.User, error) {
	args := m.Called(ctx, userID, actor, ip)
	if u, ok := args.Get(0).(*model.User); ok {
		return u, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockSettingsService) ReactivateUser(ctx context.Context, userID string, actor string, ip string) (*model.User, error) {
	args := m.Called(ctx, userID, actor, ip)
	if u, ok := args.Get(0).(*model.User); ok {
		return u, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockSettingsService) TestTelegramNotification(ctx context.Context, actor string, ip string) error {
	args := m.Called(ctx, actor, ip)
	return args.Error(0)
}

func TestSettingsHandler_GetSettings(t *testing.T) {
	e := echo.New()
	mockSvc := new(mockSettingsService)
	h := handler.NewSettingsHandler(mockSvc)

	mockSvc.On("GetSettings", mock.Anything).Return(&model.CombinedSettings{
		System:       &model.SystemSettings{AppName: "CIFO Platform"},
		Notification: &model.NotificationSettings{TelegramEnabled: true},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetSettings(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "CIFO Platform")
}

func TestSettingsHandler_UpdateUserRole(t *testing.T) {
	e := echo.New()
	mockSvc := new(mockSettingsService)
	h := handler.NewSettingsHandler(mockSvc)

	mockSvc.On("UpdateUserRole", mock.Anything, "user-123", "admin", mock.Anything, mock.Anything).Return(&model.User{
		ID:   "user-123",
		Role: "admin",
	}, nil)

	body := `{"role":"admin"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/settings/users/user-123/role", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("user-123")

	err := h.UpdateUserRole(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "admin")
}

package service_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/cifo-monitoring/backend/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSettingsRepo struct {
	mock.Mock
}

func (m *mockSettingsRepo) GetSystemSettings(ctx context.Context) (*model.SystemSettings, error) {
	args := m.Called(ctx)
	if s, ok := args.Get(0).(*model.SystemSettings); ok {
		return s, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockSettingsRepo) UpdateSystemSettings(ctx context.Context, s *model.SystemSettings) (*model.SystemSettings, error) {
	args := m.Called(ctx, s)
	if res, ok := args.Get(0).(*model.SystemSettings); ok {
		return res, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockSettingsRepo) GetNotificationSettings(ctx context.Context) (*model.NotificationSettings, error) {
	args := m.Called(ctx)
	if n, ok := args.Get(0).(*model.NotificationSettings); ok {
		return n, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockSettingsRepo) UpdateNotificationSettings(ctx context.Context, n *model.NotificationSettings) (*model.NotificationSettings, error) {
	args := m.Called(ctx, n)
	if res, ok := args.Get(0).(*model.NotificationSettings); ok {
		return res, args.Error(1)
	}
	return nil, args.Error(1)
}

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*model.User, error) {
	args := m.Called(ctx, id)
	if u, ok := args.Get(0).(*model.User); ok {
		return u, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if u, ok := args.Get(0).(*model.User); ok {
		return u, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserRepo) FindByKeycloakID(ctx context.Context, keycloakID string) (*model.User, error) {
	args := m.Called(ctx, keycloakID)
	if u, ok := args.Get(0).(*model.User); ok {
		return u, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserRepo) UpsertKeycloakUser(ctx context.Context, user *model.User) (*model.User, error) {
	args := m.Called(ctx, user)
	if u, ok := args.Get(0).(*model.User); ok {
		return u, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserRepo) List(ctx context.Context, limit, offset int) ([]*model.User, int, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*model.User), args.Int(1), args.Error(2)
}

func (m *mockUserRepo) UpdateRole(ctx context.Context, id string, role string) (*model.User, error) {
	args := m.Called(ctx, id, role)
	if u, ok := args.Get(0).(*model.User); ok {
		return u, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserRepo) SetActive(ctx context.Context, id string, isActive bool) (*model.User, error) {
	args := m.Called(ctx, id, isActive)
	if u, ok := args.Get(0).(*model.User); ok {
		return u, args.Error(1)
	}
	return nil, args.Error(1)
}

type mockAuditRepo struct {
	mock.Mock
}

func (m *mockAuditRepo) Create(ctx context.Context, log *model.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *mockAuditRepo) List(ctx context.Context, limit, offset int) ([]*model.AuditLog, int, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*model.AuditLog), args.Int(1), args.Error(2)
}

func TestGetSettings_Success(t *testing.T) {
	settingsRepo := new(mockSettingsRepo)
	userRepo := new(mockUserRepo)
	auditRepo := new(mockAuditRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := service.NewSettingsService(settingsRepo, userRepo, auditRepo, nil, logger)

	expectedSys := &model.SystemSettings{ID: "sys-1", AppName: "CIFO Platform"}
	expectedNotif := &model.NotificationSettings{ID: "notif-1", TelegramEnabled: true}

	settingsRepo.On("GetSystemSettings", mock.Anything).Return(expectedSys, nil)
	settingsRepo.On("GetNotificationSettings", mock.Anything).Return(expectedNotif, nil)

	res, err := svc.GetSettings(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "CIFO Platform", res.System.AppName)
	assert.True(t, res.Notification.TelegramEnabled)
}

func TestUpdateUserRole_InvalidRole(t *testing.T) {
	settingsRepo := new(mockSettingsRepo)
	userRepo := new(mockUserRepo)
	auditRepo := new(mockAuditRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := service.NewSettingsService(settingsRepo, userRepo, auditRepo, nil, logger)

	_, err := svc.UpdateUserRole(context.Background(), "user-1", "superhero", "admin@cifo.local", "127.0.0.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid role")
}

func TestUpdateUserRole_Success(t *testing.T) {
	settingsRepo := new(mockSettingsRepo)
	userRepo := new(mockUserRepo)
	auditRepo := new(mockAuditRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := service.NewSettingsService(settingsRepo, userRepo, auditRepo, nil, logger)

	existingUser := &model.User{ID: "user-1", Email: "dev@cifo.local", Role: "viewer"}
	updatedUser := &model.User{ID: "user-1", Email: "dev@cifo.local", Role: "devops"}

	userRepo.On("FindByID", mock.Anything, "user-1").Return(existingUser, nil)
	userRepo.On("UpdateRole", mock.Anything, "user-1", "devops").Return(updatedUser, nil)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	res, err := svc.UpdateUserRole(context.Background(), "user-1", "devops", "admin@cifo.local", "127.0.0.1")
	assert.NoError(t, err)
	assert.Equal(t, "devops", res.Role)
}

func TestDeactivateUser_PreventSelfDeactivation(t *testing.T) {
	settingsRepo := new(mockSettingsRepo)
	userRepo := new(mockUserRepo)
	auditRepo := new(mockAuditRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := service.NewSettingsService(settingsRepo, userRepo, auditRepo, nil, logger)

	adminUser := &model.User{ID: "admin-id", Email: "admin@cifo.local"}
	userRepo.On("FindByID", mock.Anything, "admin-id").Return(adminUser, nil)

	_, err := svc.DeactivateUser(context.Background(), "admin-id", "admin-id", "127.0.0.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot deactivate your own account")
}

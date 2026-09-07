package service

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/cifo-monitoring/backend/internal/model"
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

type mockSettingsUserRepo struct {
	mock.Mock
}

func (m *mockSettingsUserRepo) FindByID(ctx context.Context, id string) (*model.User, error) {
	args := m.Called(ctx, id)
	if u, ok := args.Get(0).(*model.User); ok {
		return u, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockSettingsUserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if u, ok := args.Get(0).(*model.User); ok {
		return u, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockSettingsUserRepo) FindByKeycloakID(ctx context.Context, kid string) (*model.User, error) {
	args := m.Called(ctx, kid)
	if u, ok := args.Get(0).(*model.User); ok {
		return u, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockSettingsUserRepo) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockSettingsUserRepo) Update(ctx context.Context, user *model.User) (*model.User, error) {
	args := m.Called(ctx, user)
	if u, ok := args.Get(0).(*model.User); ok {
		return u, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockSettingsUserRepo) List(ctx context.Context, limit, offset int) ([]*model.User, int, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*model.User), args.Int(1), args.Error(2)
}

func (m *mockSettingsUserRepo) UpdateRole(ctx context.Context, id string, role string) (*model.User, error) {
	args := m.Called(ctx, id, role)
	if u, ok := args.Get(0).(*model.User); ok {
		return u, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockSettingsUserRepo) UpsertKeycloakUser(ctx context.Context, user *model.User) (*model.User, error) {
	args := m.Called(ctx, user)
	if u, ok := args.Get(0).(*model.User); ok {
		return u, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockSettingsUserRepo) SetActive(ctx context.Context, id string, isActive bool) (*model.User, error) {
	args := m.Called(ctx, id, isActive)
	if u, ok := args.Get(0).(*model.User); ok {
		return u, args.Error(1)
	}
	return nil, args.Error(1)
}

type mockSettingsAuditRepo struct {
	mock.Mock
}

func (m *mockSettingsAuditRepo) Create(ctx context.Context, log *model.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *mockSettingsAuditRepo) List(ctx context.Context, limit, offset int) ([]*model.AuditLog, int, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*model.AuditLog), args.Int(1), args.Error(2)
}

func TestGetSettings_Success(t *testing.T) {
	settingsRepo := new(mockSettingsRepo)
	userRepo := new(mockSettingsUserRepo)
	auditRepo := new(mockSettingsAuditRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := NewSettingsService(settingsRepo, userRepo, auditRepo, nil, logger)

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

func TestUpdateSettings_Success(t *testing.T) {
	settingsRepo := new(mockSettingsRepo)
	auditRepo := new(mockSettingsAuditRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := NewSettingsService(settingsRepo, nil, auditRepo, nil, logger)

	currentSys := &model.SystemSettings{ID: "sys-1", AppName: "CIFO Platform"}
	currentNotif := &model.NotificationSettings{ID: "notif-1", TelegramEnabled: false}

	settingsRepo.On("GetSystemSettings", mock.Anything).Return(currentSys, nil)
	settingsRepo.On("GetNotificationSettings", mock.Anything).Return(currentNotif, nil)
	settingsRepo.On("UpdateSystemSettings", mock.Anything, mock.Anything).Return(currentSys, nil)
	settingsRepo.On("UpdateNotificationSettings", mock.Anything, mock.Anything).Return(currentNotif, nil)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	appName := "New CIFO"
	tgEnabled := true
	req := &model.UpdateSettingsRequest{
		AppName:         &appName,
		TelegramEnabled: &tgEnabled,
	}

	res, err := svc.UpdateSettings(context.Background(), req, "admin@cifo.local", "127.0.0.1")
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestListUsers(t *testing.T) {
	userRepo := new(mockSettingsUserRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := NewSettingsService(nil, userRepo, nil, nil, logger)

	users := []*model.User{{ID: "u1", Email: "test@cifo.local"}}
	userRepo.On("List", mock.Anything, 10, 0).Return(users, 1, nil)

	list, total, err := svc.ListUsers(context.Background(), 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, list, 1)
}

func TestUpdateUserRole_InvalidRole(t *testing.T) {
	userRepo := new(mockSettingsUserRepo)
	auditRepo := new(mockSettingsAuditRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := NewSettingsService(nil, userRepo, auditRepo, nil, logger)

	_, err := svc.UpdateUserRole(context.Background(), "user-1", "superhero", "admin@cifo.local", "127.0.0.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid role")
}

func TestUpdateUserRole_Success(t *testing.T) {
	userRepo := new(mockSettingsUserRepo)
	auditRepo := new(mockSettingsAuditRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := NewSettingsService(nil, userRepo, auditRepo, nil, logger)

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
	userRepo := new(mockSettingsUserRepo)
	auditRepo := new(mockSettingsAuditRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := NewSettingsService(nil, userRepo, auditRepo, nil, logger)

	adminUser := &model.User{ID: "admin-id", Email: "admin@cifo.local"}
	userRepo.On("FindByID", mock.Anything, "admin-id").Return(adminUser, nil)

	_, err := svc.DeactivateUser(context.Background(), "admin-id", "admin-id", "127.0.0.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot deactivate your own account")
}

func TestDeactivateAndReactivateUser_Success(t *testing.T) {
	userRepo := new(mockSettingsUserRepo)
	auditRepo := new(mockSettingsAuditRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := NewSettingsService(nil, userRepo, auditRepo, nil, logger)

	targetUser := &model.User{ID: "user-2", Email: "user2@cifo.local", IsActive: true}
	deactivatedUser := &model.User{ID: "user-2", Email: "user2@cifo.local", IsActive: false}

	userRepo.On("FindByID", mock.Anything, "user-2").Return(targetUser, nil)
	userRepo.On("SetActive", mock.Anything, "user-2", false).Return(deactivatedUser, nil)
	userRepo.On("SetActive", mock.Anything, "user-2", true).Return(targetUser, nil)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	res, err := svc.DeactivateUser(context.Background(), "user-2", "admin-id", "127.0.0.1")
	assert.NoError(t, err)
	assert.False(t, res.IsActive)

	res2, err := svc.ReactivateUser(context.Background(), "user-2", "admin-id", "127.0.0.1")
	assert.NoError(t, err)
	assert.True(t, res2.IsActive)
}

func TestTelegramNotification_Disabled(t *testing.T) {
	settingsRepo := new(mockSettingsRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := NewSettingsService(settingsRepo, nil, nil, nil, logger)

	notif := &model.NotificationSettings{TelegramEnabled: false}
	settingsRepo.On("GetNotificationSettings", mock.Anything).Return(notif, nil)

	err := svc.TestTelegramNotification(context.Background(), "admin@cifo.local", "127.0.0.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "disabled")
}

func TestTelegramNotification_Success(t *testing.T) {
	settingsRepo := new(mockSettingsRepo)
	auditRepo := new(mockSettingsAuditRepo)
	tgSvc := &mockTelegramService{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := NewSettingsService(settingsRepo, nil, auditRepo, tgSvc, logger)

	notif := &model.NotificationSettings{ID: "notif-1", TelegramEnabled: true, TelegramChatID: "998877"}
	settingsRepo.On("GetNotificationSettings", mock.Anything).Return(notif, nil)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	err := svc.TestTelegramNotification(context.Background(), "admin@cifo.local", "127.0.0.1")
	assert.NoError(t, err)
}

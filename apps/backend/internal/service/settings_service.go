package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/cifo-monitoring/backend/internal/repository"
	"github.com/cifo-monitoring/backend/pkg/apperror"
)

// SettingsService interface
type SettingsService interface {
	GetSettings(ctx context.Context) (*model.CombinedSettings, error)
	UpdateSettings(ctx context.Context, req *model.UpdateSettingsRequest, actor string, ip string) (*model.CombinedSettings, error)
	ListUsers(ctx context.Context, limit, offset int) ([]*model.User, int, error)
	UpdateUserRole(ctx context.Context, userID string, role string, actor string, ip string) (*model.User, error)
	DeactivateUser(ctx context.Context, userID string, actor string, ip string) (*model.User, error)
	ReactivateUser(ctx context.Context, userID string, actor string, ip string) (*model.User, error)
	TestTelegramNotification(ctx context.Context, actor string, ip string) error
}

// DefaultSettingsService implementation
type DefaultSettingsService struct {
	settingsRepo repository.SettingsRepository
	userRepo     repository.UserRepository
	auditRepo    repository.AuditRepository
	telegramSvc  TelegramService
	logger       *slog.Logger
}

// NewSettingsService constructor
func NewSettingsService(
	settingsRepo repository.SettingsRepository,
	userRepo repository.UserRepository,
	auditRepo repository.AuditRepository,
	telegramSvc TelegramService,
	logger *slog.Logger,
) *DefaultSettingsService {
	return &DefaultSettingsService{
		settingsRepo: settingsRepo,
		userRepo:     userRepo,
		auditRepo:    auditRepo,
		telegramSvc:  telegramSvc,
		logger:       logger,
	}
}

// GetSettings retrieves combined settings
func (s *DefaultSettingsService) GetSettings(ctx context.Context) (*model.CombinedSettings, error) {
	// fetch system settings
	sys, err := s.settingsRepo.GetSystemSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("get system settings: %w", err)
	}

	// fetch notification settings
	notif, err := s.settingsRepo.GetNotificationSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("get notification settings: %w", err)
	}

	return &model.CombinedSettings{
		System:       sys,
		Notification: notif,
	}, nil
}

// UpdateSettings updates system and notification settings
func (s *DefaultSettingsService) UpdateSettings(ctx context.Context, req *model.UpdateSettingsRequest, actor string, ip string) (*model.CombinedSettings, error) {
	// fetch current state
	current, err := s.GetSettings(ctx)
	if err != nil {
		return nil, err
	}

	sys := current.System
	notif := current.Notification

	// merge nested system partial if present
	if req.System != nil {
		if req.System.AppName != nil { req.AppName = req.System.AppName }
		if req.System.DefaultTheme != nil { req.DefaultTheme = req.System.DefaultTheme }
		if req.System.Language != nil { req.Language = req.System.Language }
		if req.System.Timezone != nil { req.Timezone = req.System.Timezone }
		if req.System.AIDefaultModel != nil { req.AIDefaultModel = req.System.AIDefaultModel }
		if req.System.AIDefaultProvider != nil { req.AIDefaultProvider = req.System.AIDefaultProvider }
		if req.System.AIMonthlyBudgetUSD != nil { req.AIMonthlyBudgetUSD = req.System.AIMonthlyBudgetUSD }
		if req.System.AIMaxTokensPerRequest != nil { req.AIMaxTokensPerRequest = req.System.AIMaxTokensPerRequest }
		if len(req.System.AIModelPreferenceOrder) > 0 { req.AIModelPreferenceOrder = req.System.AIModelPreferenceOrder }
		if req.System.SessionTimeoutMinutes != nil { req.SessionTimeoutMinutes = req.System.SessionTimeoutMinutes }
		if req.System.MaxLoginAttempts != nil { req.MaxLoginAttempts = req.System.MaxLoginAttempts }
		if req.System.RequireMFA != nil { req.RequireMFA = req.System.RequireMFA }
		if req.System.MaintenanceMode != nil { req.MaintenanceMode = req.System.MaintenanceMode }
	}

	// merge nested notification partial if present
	if req.Notification != nil {
		if req.Notification.TelegramBotTokenRef != nil { req.TelegramBotTokenRef = req.Notification.TelegramBotTokenRef }
		if req.Notification.TelegramChatID != nil { req.TelegramChatID = req.Notification.TelegramChatID }
		if req.Notification.TelegramEnabled != nil { req.TelegramEnabled = req.Notification.TelegramEnabled }
		if req.Notification.InAppEnabled != nil { req.InAppEnabled = req.Notification.InAppEnabled }
		if req.Notification.AlertBatchingWindowSeconds != nil { req.AlertBatchingWindowSeconds = req.Notification.AlertBatchingWindowSeconds }
	}

	// apply system updates
	if req.AppName != nil {
		sys.AppName = *req.AppName
	}
	if req.DefaultTheme != nil {
		sys.DefaultTheme = *req.DefaultTheme
	}
	if req.Language != nil {
		sys.Language = *req.Language
	}
	if req.Timezone != nil {
		sys.Timezone = *req.Timezone
	}
	if req.AIDefaultModel != nil {
		sys.AIDefaultModel = *req.AIDefaultModel
	}
	if req.AIDefaultProvider != nil {
		sys.AIDefaultProvider = *req.AIDefaultProvider
	}
	if req.AIMonthlyBudgetUSD != nil {
		sys.AIMonthlyBudgetUSD = *req.AIMonthlyBudgetUSD
	}
	if req.AIMaxTokensPerRequest != nil {
		sys.AIMaxTokensPerRequest = *req.AIMaxTokensPerRequest
	}
	if len(req.AIModelPreferenceOrder) > 0 {
		sys.AIModelPreferenceOrder = req.AIModelPreferenceOrder
	}
	if req.SessionTimeoutMinutes != nil {
		sys.SessionTimeoutMinutes = *req.SessionTimeoutMinutes
	}
	if req.MaxLoginAttempts != nil {
		sys.MaxLoginAttempts = *req.MaxLoginAttempts
	}
	if req.RequireMFA != nil {
		sys.RequireMFA = *req.RequireMFA
	}
	if req.MaintenanceMode != nil {
		sys.MaintenanceMode = *req.MaintenanceMode
	}

	// apply notification updates
	if req.TelegramBotTokenRef != nil {
		notif.TelegramBotTokenRef = *req.TelegramBotTokenRef
	}
	if req.TelegramChatID != nil {
		notif.TelegramChatID = *req.TelegramChatID
	}
	if req.TelegramEnabled != nil {
		notif.TelegramEnabled = *req.TelegramEnabled
	}
	if req.InAppEnabled != nil {
		notif.InAppEnabled = *req.InAppEnabled
	}
	if req.AlertBatchingWindowSeconds != nil {
		notif.AlertBatchingWindowSeconds = *req.AlertBatchingWindowSeconds
	}

	// persist system settings
	updatedSys, err := s.settingsRepo.UpdateSystemSettings(ctx, sys)
	if err != nil {
		return nil, fmt.Errorf("update system settings: %w", err)
	}

	// persist notification settings
	updatedNotif, err := s.settingsRepo.UpdateNotificationSettings(ctx, notif)
	if err != nil {
		return nil, fmt.Errorf("update notification settings: %w", err)
	}

	// record audit log
	if s.auditRepo != nil {
		details, _ := json.Marshal(map[string]interface{}{
			"updated_by": actor,
			"request":    req,
		})
		_ = s.auditRepo.Create(ctx, &model.AuditLog{
			ActorType:    "user",
			ActorID:      actor,
			Action:       "SETTINGS_UPDATE",
			ResourceType: "system_settings",
			ResourceID:   &updatedSys.ID,
			Details:      details,
			IPAddress:    &ip,
			Result:       "success",
			Timestamp:    time.Now(),
		})
	}

	return &model.CombinedSettings{
		System:       updatedSys,
		Notification: updatedNotif,
	}, nil
}

// ListUsers lists users
func (s *DefaultSettingsService) ListUsers(ctx context.Context, limit, offset int) ([]*model.User, int, error) {
	return s.userRepo.List(ctx, limit, offset)
}

// UpdateUserRole changes user role
func (s *DefaultSettingsService) UpdateUserRole(ctx context.Context, userID string, role string, actor string, ip string) (*model.User, error) {
	// validate target role
	validRoles := map[string]bool{"admin": true, "devops": true, "viewer": true}
	cleanRole := strings.ToLower(strings.TrimSpace(role))
	if !validRoles[cleanRole] {
		return nil, apperror.NewBadRequest("invalid role; must be admin, devops, or viewer")
	}

	// verify user exists
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if user == nil {
		return nil, apperror.NewNotFound("user not found")
	}

	oldRole := user.Role
	updated, err := s.userRepo.UpdateRole(ctx, userID, cleanRole)
	if err != nil {
		return nil, fmt.Errorf("update role: %w", err)
	}

	// record audit log
	if s.auditRepo != nil {
		details, _ := json.Marshal(map[string]interface{}{
			"previous_role": oldRole,
			"new_role":      cleanRole,
			"target_email":  user.Email,
		})
		_ = s.auditRepo.Create(ctx, &model.AuditLog{
			ActorType:    "user",
			ActorID:      actor,
			Action:       "USER_ROLE_CHANGE",
			ResourceType: "user",
			ResourceID:   &userID,
			Details:      details,
			IPAddress:    &ip,
			Result:       "success",
			Timestamp:    time.Now(),
		})
	}

	return updated, nil
}

// DeactivateUser disables user
func (s *DefaultSettingsService) DeactivateUser(ctx context.Context, userID string, actor string, ip string) (*model.User, error) {
	// check user existence
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if user == nil {
		return nil, apperror.NewNotFound("user not found")
	}

	// prevent self-deactivation
	if user.ID == actor || user.Email == actor {
		return nil, apperror.NewBadRequest("cannot deactivate your own account")
	}

	updated, err := s.userRepo.SetActive(ctx, userID, false)
	if err != nil {
		return nil, fmt.Errorf("deactivate user: %w", err)
	}

	// record audit log
	if s.auditRepo != nil {
		details, _ := json.Marshal(map[string]interface{}{
			"target_email": user.Email,
			"is_active":    false,
		})
		_ = s.auditRepo.Create(ctx, &model.AuditLog{
			ActorType:    "user",
			ActorID:      actor,
			Action:       "USER_DEACTIVATION",
			ResourceType: "user",
			ResourceID:   &userID,
			Details:      details,
			IPAddress:    &ip,
			Result:       "success",
			Timestamp:    time.Now(),
		})
	}

	return updated, nil
}

// ReactivateUser enables user
func (s *DefaultSettingsService) ReactivateUser(ctx context.Context, userID string, actor string, ip string) (*model.User, error) {
	// check user existence
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if user == nil {
		return nil, apperror.NewNotFound("user not found")
	}

	updated, err := s.userRepo.SetActive(ctx, userID, true)
	if err != nil {
		return nil, fmt.Errorf("reactivate user: %w", err)
	}

	// record audit log
	if s.auditRepo != nil {
		details, _ := json.Marshal(map[string]interface{}{
			"target_email": user.Email,
			"is_active":    true,
		})
		_ = s.auditRepo.Create(ctx, &model.AuditLog{
			ActorType:    "user",
			ActorID:      actor,
			Action:       "USER_ACTIVATION",
			ResourceType: "user",
			ResourceID:   &userID,
			Details:      details,
			IPAddress:    &ip,
			Result:       "success",
			Timestamp:    time.Now(),
		})
	}

	return updated, nil
}

// TestTelegramNotification sends verification test
func (s *DefaultSettingsService) TestTelegramNotification(ctx context.Context, actor string, ip string) error {
	// check notification settings
	notif, err := s.settingsRepo.GetNotificationSettings(ctx)
	if err != nil {
		return fmt.Errorf("get notification settings: %w", err)
	}

	if !notif.TelegramEnabled {
		return apperror.NewBadRequest("telegram notifications are disabled in settings")
	}

	if s.telegramSvc != nil {
		testIncident := &model.Incident{
			Title:        "Telegram Integration Verification Test",
			Description:  "This is a test notification generated from CIFO Platform settings.",
			Severity:     "info",
			Status:       "open",
			Source:       "system_settings",
			ResourceType: "settings_test",
			ResourceID:   "cifo-settings",
			CreatedAt:    time.Now(),
		}
		_ = s.telegramSvc.SendIncidentAlert(ctx, testIncident, false)
	}

	// record audit log
	if s.auditRepo != nil {
		details, _ := json.Marshal(map[string]interface{}{
			"triggered_by": actor,
			"chat_id":      notif.TelegramChatID,
		})
		_ = s.auditRepo.Create(ctx, &model.AuditLog{
			ActorType:    "user",
			ActorID:      actor,
			Action:       "TELEGRAM_TEST_ALERT",
			ResourceType: "notification_settings",
			ResourceID:   &notif.ID,
			Details:      details,
			IPAddress:    &ip,
			Result:       "success",
			Timestamp:    time.Now(),
		})
	}

	return nil
}

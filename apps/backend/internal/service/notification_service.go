package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/cifo-monitoring/backend/internal/repository"
	"github.com/cifo-monitoring/backend/internal/ws"
)

// NotificationService interface
type NotificationService interface {
	Notify(ctx context.Context, inc *model.Incident, isEscalation bool)
	ListHistory(ctx context.Context, incidentID string) ([]model.NotificationRecord, error)
}

// DefaultNotificationService implementation
type DefaultNotificationService struct {
	telegramService TelegramService
	wsHub           *ws.Hub
	incidentRepo    repository.IncidentRepository
	logger          *slog.Logger
}

// NewNotificationService creates service
func NewNotificationService(
	telegram TelegramService,
	hub *ws.Hub,
	repo repository.IncidentRepository,
	logger *slog.Logger,
) *DefaultNotificationService {
	return &DefaultNotificationService{
		telegramService: telegram,
		wsHub:           hub,
		incidentRepo:    repo,
		logger:          logger,
	}
}

// Notify dispatches notifications
func (s *DefaultNotificationService) Notify(ctx context.Context, inc *model.Incident, isEscalation bool) {
	// in-app websocket notification
	s.sendInAppNotification(inc, isEscalation)

	// telegram notification async
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		var status = "sent"
		var errMsg = ""
		err := s.telegramService.SendIncidentAlert(bgCtx, inc, isEscalation)
		if err != nil {
			status = "failed"
			errMsg = err.Error()
		}

		title := inc.Title
		if isEscalation {
			title = "[ESCALATED] " + title
		}

		// record notification history
		rec := &model.NotificationRecord{
			IncidentID:   &inc.ID,
			Channel:      "telegram",
			Recipient:    "telegram_group",
			Title:        title,
			Message:      s.telegramService.FormatIncidentMessage(inc, isEscalation),
			Severity:     inc.Severity,
			Status:       status,
			ErrorMessage: errMsg,
		}
		if s.incidentRepo != nil {
			_ = s.incidentRepo.SaveNotification(bgCtx, rec)
		}
	}()
}

// sendInAppNotification pushes websocket
func (s *DefaultNotificationService) sendInAppNotification(inc *model.Incident, isEscalation bool) {
	if s.wsHub == nil {
		return
	}

	title := inc.Title
	if isEscalation {
		title = "[ESCALATED] " + title
	}

	payload := ws.NotificationPayload{
		ID:        inc.ID,
		Title:     title,
		Message:   inc.Description,
		Severity:  inc.Severity,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Source:    inc.Source,
	}

	msg := ws.NewWSMessage(ws.TypeNotification, "notifications", payload)
	s.wsHub.Broadcast("notifications", msg)

	// save in-app record
	if s.incidentRepo != nil {
		rec := &model.NotificationRecord{
			IncidentID: &inc.ID,
			Channel:    "inapp",
			Recipient:  "dashboard",
			Title:      title,
			Message:    inc.Description,
			Severity:   inc.Severity,
			Status:     "sent",
		}
		_ = s.incidentRepo.SaveNotification(context.Background(), rec)
	}
}

// ListHistory gets notifications
func (s *DefaultNotificationService) ListHistory(ctx context.Context, incidentID string) ([]model.NotificationRecord, error) {
	if s.incidentRepo == nil {
		return nil, nil
	}
	return s.incidentRepo.ListNotificationsByIncidentID(ctx, incidentID)
}

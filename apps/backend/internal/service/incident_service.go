package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/cifo-monitoring/backend/internal/repository"
	"github.com/cifo-monitoring/backend/pkg/apperror"
)

// IncidentService interface
type IncidentService interface {
	ProcessAlertmanagerWebhook(ctx context.Context, payload *model.AlertmanagerWebhookPayload) error
	CreateIncident(ctx context.Context, req *model.CreateIncidentRequest) (*model.Incident, error)
	ListIncidents(ctx context.Context, filter model.IncidentFilter) ([]*model.IncidentSummary, int, error)
	GetIncident(ctx context.Context, id string) (*model.IncidentDetail, error)
	GetIncidentStats(ctx context.Context) (*model.IncidentStats, error)
	AcknowledgeIncident(ctx context.Context, id string, userID string, ipAddress string) error
	ResolveIncident(ctx context.Context, id string, userID string, ipAddress string) error
	CloseIncident(ctx context.Context, id string, userID string, ipAddress string) error
	EscalateIncident(ctx context.Context, id string) error
	CheckEscalations(ctx context.Context) error
}

// DefaultIncidentService implementation
type DefaultIncidentService struct {
	repo         repository.IncidentRepository
	auditRepo    repository.AuditRepository
	notification NotificationService
	logger       *slog.Logger
}

// NewIncidentService creates service
func NewIncidentService(
	repo repository.IncidentRepository,
	audit repository.AuditRepository,
	notif NotificationService,
	logger *slog.Logger,
) *DefaultIncidentService {
	return &DefaultIncidentService{
		repo:         repo,
		auditRepo:    audit,
		notification: notif,
		logger:       logger,
	}
}

// ProcessAlertmanagerWebhook handles alerts
func (s *DefaultIncidentService) ProcessAlertmanagerWebhook(ctx context.Context, payload *model.AlertmanagerWebhookPayload) error {
	if payload == nil || len(payload.Alerts) == 0 {
		return nil
	}

	for _, alert := range payload.Alerts {
		alertName := alert.Labels["alertname"]
		if alertName == "" {
			alertName = "SystemAlert"
		}

		severity := alert.Labels["severity"]
		if severity == "" {
			severity = model.SeverityWarning
		}

		// determine resource id
		resourceID := alert.Labels["container"]
		resourceType := "container"
		if resourceID == "" {
			resourceID = alert.Labels["pod"]
			resourceType = "pod"
		}
		if resourceID == "" {
			resourceID = alert.Labels["app"]
			resourceType = "argocd_app"
		}
		if resourceID == "" {
			resourceID = alert.Labels["instance"]
			resourceType = "host"
		}

		namespace := alert.Labels["namespace"]

		title := alert.Annotations["summary"]
		if title == "" {
			title = fmt.Sprintf("%s on %s", alertName, resourceID)
		}

		desc := alert.Annotations["description"]
		if desc == "" {
			desc = title
		}

		if alert.Status == "resolved" {
			// auto-resolve open incident
			existing, err := s.repo.FindOpenByAlertAndResource(ctx, alertName, resourceID)
			if err == nil && existing != nil {
				s.logger.Info("auto-resolving incident from alertmanager", slog.String("id", existing.ID))
				_ = s.repo.UpdateStatus(ctx, existing.ID, model.IncidentStatusResolved, nil)
			}
			continue
		}

		// check existing firing
		existing, err := s.repo.FindOpenByAlertAndResource(ctx, alertName, resourceID)
		if err == nil && existing != nil {
			s.logger.Debug("alert already firing", slog.String("incident_id", existing.ID))
			continue
		}

		// create new incident
		inc := &model.Incident{
			Title:        title,
			Description:  desc,
			Severity:     severity,
			Status:       model.IncidentStatusOpen,
			Source:       model.SourceAlertmanager,
			AlertName:    alertName,
			ResourceType: resourceType,
			ResourceID:   resourceID,
			Namespace:    namespace,
		}

		if err := s.repo.Create(ctx, inc); err != nil {
			s.logger.Error("failed to create incident from alert", slog.String("error", err.Error()))
			continue
		}

		s.logger.Info("new incident created from alertmanager", slog.String("id", inc.ID), slog.String("title", inc.Title))

		// trigger notification
		if s.notification != nil {
			s.notification.Notify(ctx, inc, false)
		}
	}

	return nil
}

// CreateIncident creates incident
func (s *DefaultIncidentService) CreateIncident(ctx context.Context, req *model.CreateIncidentRequest) (*model.Incident, error) {
	inc := &model.Incident{
		Title:        req.Title,
		Description:  req.Description,
		Severity:     req.Severity,
		Status:       model.IncidentStatusOpen,
		Source:       req.Source,
		AlertName:    req.AlertName,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Namespace:    req.Namespace,
	}

	if err := s.repo.Create(ctx, inc); err != nil {
		return nil, fmt.Errorf("create incident: %w", err)
	}

	// trigger notification
	if s.notification != nil {
		s.notification.Notify(ctx, inc, false)
	}

	return inc, nil
}

// ListIncidents lists incidents
func (s *DefaultIncidentService) ListIncidents(ctx context.Context, filter model.IncidentFilter) ([]*model.IncidentSummary, int, error) {
	return s.repo.List(ctx, filter)
}

// GetIncident gets detail
func (s *DefaultIncidentService) GetIncident(ctx context.Context, id string) (*model.IncidentDetail, error) {
	inc, err := s.repo.GetDetailByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get incident: %w", err)
	}
	if inc == nil {
		return nil, apperror.NewNotFound("incident not found")
	}
	return inc, nil
}

// GetIncidentStats returns stats
func (s *DefaultIncidentService) GetIncidentStats(ctx context.Context) (*model.IncidentStats, error) {
	return s.repo.GetStats(ctx)
}

// AcknowledgeIncident acknowledges incident
func (s *DefaultIncidentService) AcknowledgeIncident(ctx context.Context, id string, userID string, ipAddress string) error {
	inc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get incident: %w", err)
	}
	if inc == nil {
		return apperror.NewNotFound("incident not found")
	}

	if inc.Status != model.IncidentStatusOpen {
		return apperror.NewBadRequest("only open incidents can be acknowledged")
	}

	actorID := &userID
	if err := s.repo.UpdateStatus(ctx, id, model.IncidentStatusAcknowledged, actorID); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	// write audit log
	if s.auditRepo != nil {
		details, _ := json.Marshal(map[string]string{
			"previous_status": inc.Status,
			"new_status":      model.IncidentStatusAcknowledged,
			"incident_title":  inc.Title,
		})
		_ = s.auditRepo.Create(ctx, &model.AuditLog{
			ActorType:    "user",
			ActorID:      userID,
			Action:       "acknowledge_incident",
			ResourceType: "incident",
			ResourceID:   &id,
			Details:      details,
			IPAddress:    &ipAddress,
			Result:       "success",
		})
	}

	s.logger.Info("incident acknowledged", slog.String("id", id), slog.String("user_id", userID))
	return nil
}

// ResolveIncident resolves incident
func (s *DefaultIncidentService) ResolveIncident(ctx context.Context, id string, userID string, ipAddress string) error {
	inc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get incident: %w", err)
	}
	if inc == nil {
		return apperror.NewNotFound("incident not found")
	}

	if inc.Status == model.IncidentStatusClosed {
		return apperror.NewBadRequest("closed incidents cannot be resolved")
	}

	actorID := &userID
	if err := s.repo.UpdateStatus(ctx, id, model.IncidentStatusResolved, actorID); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	// write audit log
	if s.auditRepo != nil {
		details, _ := json.Marshal(map[string]string{
			"previous_status": inc.Status,
			"new_status":      model.IncidentStatusResolved,
			"incident_title":  inc.Title,
		})
		_ = s.auditRepo.Create(ctx, &model.AuditLog{
			ActorType:    "user",
			ActorID:      userID,
			Action:       "resolve_incident",
			ResourceType: "incident",
			ResourceID:   &id,
			Details:      details,
			IPAddress:    &ipAddress,
			Result:       "success",
		})
	}

	s.logger.Info("incident resolved", slog.String("id", id), slog.String("user_id", userID))
	return nil
}

// CloseIncident closes incident
func (s *DefaultIncidentService) CloseIncident(ctx context.Context, id string, userID string, ipAddress string) error {
	inc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get incident: %w", err)
	}
	if inc == nil {
		return apperror.NewNotFound("incident not found")
	}

	actorID := &userID
	if err := s.repo.UpdateStatus(ctx, id, model.IncidentStatusClosed, actorID); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	// write audit log
	if s.auditRepo != nil {
		details, _ := json.Marshal(map[string]string{
			"previous_status": inc.Status,
			"new_status":      model.IncidentStatusClosed,
			"incident_title":  inc.Title,
		})
		_ = s.auditRepo.Create(ctx, &model.AuditLog{
			ActorType:    "user",
			ActorID:      userID,
			Action:       "close_incident",
			ResourceType: "incident",
			ResourceID:   &id,
			Details:      details,
			IPAddress:    &ipAddress,
			Result:       "success",
		})
	}

	s.logger.Info("incident closed", slog.String("id", id), slog.String("user_id", userID))
	return nil
}

// EscalateIncident triggers escalation
func (s *DefaultIncidentService) EscalateIncident(ctx context.Context, id string) error {
	inc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get incident: %w", err)
	}
	if inc == nil {
		return apperror.NewNotFound("incident not found")
	}

	if inc.Status != model.IncidentStatusOpen {
		return nil
	}

	if s.notification != nil {
		s.notification.Notify(ctx, inc, true)
	}

	s.logger.Warn("incident escalated", slog.String("id", id), slog.String("title", inc.Title))
	return nil
}

// CheckEscalations checks unacknowledged
func (s *DefaultIncidentService) CheckEscalations(ctx context.Context) error {
	// 15 minutes unacknowledged
	unacknowledged, err := s.repo.GetUnacknowledgedOlderThan(ctx, 15*time.Minute)
	if err != nil {
		return fmt.Errorf("get unacknowledged incidents: %w", err)
	}

	for _, inc := range unacknowledged {
		s.logger.Warn("auto-escalating unacknowledged incident", slog.String("id", inc.ID), slog.String("title", inc.Title))
		if s.notification != nil {
			s.notification.Notify(ctx, inc, true)
		}
	}

	return nil
}

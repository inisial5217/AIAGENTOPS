package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/cifo-monitoring/backend/internal/integration"
	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/cifo-monitoring/backend/internal/repository"
)

// ArgoCDService business logic for argocd
type ArgoCDService interface {
	ListApplications(ctx context.Context, namespace string) ([]model.ArgoApplicationSummary, error)
	GetApplication(ctx context.Context, namespace string, name string) (*model.ArgoApplicationDetail, error)
	SyncApplication(ctx context.Context, namespace string, name string, req model.ArgoSyncRequest, actor string, ip string) error
	GetOverview(ctx context.Context, namespace string) (map[string]int, error)
}

type argoCDServiceImpl struct {
	client    integration.ArgoCDClient
	auditRepo repository.AuditRepository
	logger    *slog.Logger
}

// NewArgoCDService creates argocd service
func NewArgoCDService(
	client integration.ArgoCDClient,
	auditRepo repository.AuditRepository,
	logger *slog.Logger,
) ArgoCDService {
	return &argoCDServiceImpl{
		client:    client,
		auditRepo: auditRepo,
		logger:    logger,
	}
}

// ListApplications retrieves application summaries
func (s *argoCDServiceImpl) ListApplications(ctx context.Context, namespace string) ([]model.ArgoApplicationSummary, error) {
	return s.client.ListApplications(ctx, namespace)
}

// GetApplication retrieves application detail
func (s *argoCDServiceImpl) GetApplication(ctx context.Context, namespace string, name string) (*model.ArgoApplicationDetail, error) {
	return s.client.GetApplication(ctx, namespace, name)
}

// SyncApplication triggers sync and logs audit
func (s *argoCDServiceImpl) SyncApplication(ctx context.Context, namespace string, name string, req model.ArgoSyncRequest, actor string, ip string) error {
	if err := s.client.SyncApplication(ctx, namespace, name, req); err != nil {
		return err
	}

	resourceID := fmt.Sprintf("%s/%s", namespace, name)
	details, _ := json.Marshal(map[string]interface{}{
		"namespace": namespace,
		"name":      name,
		"prune":     req.Prune,
		"dry_run":   req.DryRun,
	})
	_ = s.auditRepo.Create(ctx, &model.AuditLog{
		Timestamp:    time.Now(),
		ActorType:    "user",
		ActorID:      actor,
		Action:       "sync_argocd_application",
		ResourceType: "argocd",
		ResourceID:   &resourceID,
		Details:      details,
		IPAddress:    &ip,
		Result:       "success",
	})
	return nil
}

// GetOverview summarizes application statuses
func (s *argoCDServiceImpl) GetOverview(ctx context.Context, namespace string) (map[string]int, error) {
	apps, err := s.client.ListApplications(ctx, namespace)
	if err != nil {
		return nil, err
	}

	overview := map[string]int{
		"total":       len(apps),
		"synced":      0,
		"out_of_sync": 0,
		"healthy":     0,
		"degraded":    0,
		"progressing": 0,
		"unknown":     0,
	}

	for _, app := range apps {
		switch app.SyncStatus {
		case "Synced":
			overview["synced"]++
		case "OutOfSync":
			overview["out_of_sync"]++
		default:
			overview["unknown"]++
		}

		switch app.HealthStatus {
		case "Healthy":
			overview["healthy"]++
		case "Degraded":
			overview["degraded"]++
		case "Progressing":
			overview["progressing"]++
		}
	}

	return overview, nil
}

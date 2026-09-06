package service

import (
	"context"
	"testing"
	"time"

	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/cifo-monitoring/backend/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockIncidentRepo mock
type mockIncidentRepo struct {
	mock.Mock
}

func (m *mockIncidentRepo) Create(ctx context.Context, inc *model.Incident) error {
	args := m.Called(ctx, inc)
	inc.ID = "test-incident-uuid"
	return args.Error(0)
}

func (m *mockIncidentRepo) GetByID(ctx context.Context, id string) (*model.Incident, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Incident), args.Error(1)
}

func (m *mockIncidentRepo) GetDetailByID(ctx context.Context, id string) (*model.IncidentDetail, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.IncidentDetail), args.Error(1)
}

func (m *mockIncidentRepo) List(ctx context.Context, filter model.IncidentFilter) ([]*model.IncidentSummary, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*model.IncidentSummary), args.Int(1), args.Error(2)
}

func (m *mockIncidentRepo) GetStats(ctx context.Context) (*model.IncidentStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(*model.IncidentStats), args.Error(1)
}

func (m *mockIncidentRepo) FindOpenByAlertAndResource(ctx context.Context, alertName, resourceID string) (*model.Incident, error) {
	args := m.Called(ctx, alertName, resourceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Incident), args.Error(1)
}

func (m *mockIncidentRepo) UpdateStatus(ctx context.Context, id string, status string, actorID *string) error {
	args := m.Called(ctx, id, status, actorID)
	return args.Error(0)
}

func (m *mockIncidentRepo) GetUnacknowledgedOlderThan(ctx context.Context, d time.Duration) ([]*model.Incident, error) {
	args := m.Called(ctx, d)
	return args.Get(0).([]*model.Incident), args.Error(1)
}

func (m *mockIncidentRepo) SaveNotification(ctx context.Context, notif *model.NotificationRecord) error {
	args := m.Called(ctx, notif)
	return args.Error(0)
}

func (m *mockIncidentRepo) ListNotificationsByIncidentID(ctx context.Context, incidentID string) ([]model.NotificationRecord, error) {
	args := m.Called(ctx, incidentID)
	return args.Get(0).([]model.NotificationRecord), args.Error(1)
}

// mockAuditRepo mock
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

// mockNotificationSvc mock
type mockNotificationSvc struct {
	mock.Mock
}

func (m *mockNotificationSvc) Notify(ctx context.Context, inc *model.Incident, isEscalation bool) {
	m.Called(ctx, inc, isEscalation)
}

func (m *mockNotificationSvc) ListHistory(ctx context.Context, incidentID string) ([]model.NotificationRecord, error) {
	args := m.Called(ctx, incidentID)
	return args.Get(0).([]model.NotificationRecord), args.Error(1)
}

func TestIncidentService_ProcessAlertmanagerWebhook(t *testing.T) {
	log := logger.New("DEBUG", "test")
	repo := new(mockIncidentRepo)
	audit := new(mockAuditRepo)
	notif := new(mockNotificationSvc)
	svc := NewIncidentService(repo, audit, notif, log)

	ctx := context.Background()

	// firing alert scenario
	repo.On("FindOpenByAlertAndResource", ctx, "ContainerOOMKilled", "payment-gateway").Return(nil, nil).Once()
	repo.On("Create", ctx, mock.AnythingOfType("*model.Incident")).Return(nil).Once()
	notif.On("Notify", ctx, mock.AnythingOfType("*model.Incident"), false).Return().Once()

	payload := &model.AlertmanagerWebhookPayload{
		Status: "firing",
		Alerts: []model.AlertmanagerAlert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "ContainerOOMKilled",
					"severity":  "critical",
					"container": "payment-gateway",
				},
				Annotations: map[string]string{
					"summary":     "Container payment-gateway OOM killed",
					"description": "Container exceeded memory limit",
				},
			},
		},
	}

	err := svc.ProcessAlertmanagerWebhook(ctx, payload)
	assert.NoError(t, err)

	// resolved alert scenario
	existingInc := &model.Incident{
		ID:         "test-incident-uuid",
		Status:     model.IncidentStatusOpen,
		AlertName:  "ContainerOOMKilled",
		ResourceID: "payment-gateway",
	}
	repo.On("FindOpenByAlertAndResource", ctx, "ContainerOOMKilled", "payment-gateway").Return(existingInc, nil).Once()
	repo.On("UpdateStatus", ctx, "test-incident-uuid", model.IncidentStatusResolved, (*string)(nil)).Return(nil).Once()

	resolvedPayload := &model.AlertmanagerWebhookPayload{
		Status: "resolved",
		Alerts: []model.AlertmanagerAlert{
			{
				Status: "resolved",
				Labels: map[string]string{
					"alertname": "ContainerOOMKilled",
					"container": "payment-gateway",
				},
			},
		},
	}

	err = svc.ProcessAlertmanagerWebhook(ctx, resolvedPayload)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestIncidentService_LifecycleTransitions(t *testing.T) {
	log := logger.New("DEBUG", "test")
	repo := new(mockIncidentRepo)
	audit := new(mockAuditRepo)
	notif := new(mockNotificationSvc)
	svc := NewIncidentService(repo, audit, notif, log)

	ctx := context.Background()
	userUUID := "user-uuid-123"

	// 1. Acknowledge
	openInc := &model.Incident{
		ID:     "inc-1",
		Status: model.IncidentStatusOpen,
		Title:  "CrashLoopBackOff",
	}
	repo.On("GetByID", ctx, "inc-1").Return(openInc, nil)
	repo.On("UpdateStatus", ctx, "inc-1", model.IncidentStatusAcknowledged, &userUUID).Return(nil)
	audit.On("Create", ctx, mock.MatchedBy(func(log *model.AuditLog) bool {
		return log.Action == "acknowledge_incident" && log.ResourceID != nil && *log.ResourceID == "inc-1"
	})).Return(nil)

	err := svc.AcknowledgeIncident(ctx, "inc-1", userUUID, "127.0.0.1")
	assert.NoError(t, err)

	// 2. Resolve
	ackInc := &model.Incident{
		ID:     "inc-1",
		Status: model.IncidentStatusAcknowledged,
		Title:  "CrashLoopBackOff",
	}
	repo.On("GetByID", ctx, "inc-1").Return(ackInc, nil)
	repo.On("UpdateStatus", ctx, "inc-1", model.IncidentStatusResolved, &userUUID).Return(nil)
	audit.On("Create", ctx, mock.MatchedBy(func(log *model.AuditLog) bool {
		return log.Action == "resolve_incident" && log.ResourceID != nil && *log.ResourceID == "inc-1"
	})).Return(nil)

	err = svc.ResolveIncident(ctx, "inc-1", userUUID, "127.0.0.1")
	assert.NoError(t, err)

	// 3. Close
	resInc := &model.Incident{
		ID:     "inc-1",
		Status: model.IncidentStatusResolved,
		Title:  "CrashLoopBackOff",
	}
	repo.On("GetByID", ctx, "inc-1").Return(resInc, nil)
	repo.On("UpdateStatus", ctx, "inc-1", model.IncidentStatusClosed, &userUUID).Return(nil)
	audit.On("Create", ctx, mock.MatchedBy(func(log *model.AuditLog) bool {
		return log.Action == "close_incident" && log.ResourceID != nil && *log.ResourceID == "inc-1"
	})).Return(nil)

	err = svc.CloseIncident(ctx, "inc-1", userUUID, "127.0.0.1")
	assert.NoError(t, err)

	repo.AssertExpectations(t)
	audit.AssertExpectations(t)
}

func TestIncidentService_CheckEscalations(t *testing.T) {
	log := logger.New("DEBUG", "test")
	repo := new(mockIncidentRepo)
	audit := new(mockAuditRepo)
	notif := new(mockNotificationSvc)
	svc := NewIncidentService(repo, audit, notif, log)

	ctx := context.Background()
	oldInc := &model.Incident{
		ID:        "inc-old",
		Title:     "OldUnacknowledgedAlert",
		Status:    model.IncidentStatusOpen,
		CreatedAt: time.Now().Add(-20 * time.Minute),
	}

	repo.On("GetUnacknowledgedOlderThan", ctx, 15*time.Minute).Return([]*model.Incident{oldInc}, nil)
	notif.On("Notify", ctx, oldInc, true).Return()

	err := svc.CheckEscalations(ctx)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	notif.AssertExpectations(t)
}

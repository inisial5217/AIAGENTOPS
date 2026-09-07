package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/cifo-monitoring/backend/internal/ws"
)

// mockTelegramService for test
type mockTelegramService struct {
	mu           sync.Mutex
	alertCalls   int
	lastInc      *model.Incident
	isEscalation bool
	sendErr      error
}

func (m *mockTelegramService) SendIncidentAlert(ctx context.Context, inc *model.Incident, isEscalation bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alertCalls++
	m.lastInc = inc
	m.isEscalation = isEscalation
	return m.sendErr
}

func (m *mockTelegramService) ProcessRetryQueue(ctx context.Context) error {
	return nil
}

func (m *mockTelegramService) FormatIncidentMessage(inc *model.Incident, isEscalation bool) string {
	return "formatted alert message"
}

func (m *mockTelegramService) FormatBatchSummary(incidents []*model.Incident) string {
	return "batch summary"
}

// mockIncidentRepo for test
type mockNotifIncidentRepo struct {
	mu            sync.Mutex
	savedRecords  []*model.NotificationRecord
	listRecords   []model.NotificationRecord
	listErr       error
	saveErr       error
	saveCompleted chan struct{}
}

func (m *mockNotifIncidentRepo) Create(ctx context.Context, inc *model.Incident) error {
	return nil
}

func (m *mockNotifIncidentRepo) GetByID(ctx context.Context, id string) (*model.Incident, error) {
	return nil, nil
}

func (m *mockNotifIncidentRepo) GetDetailByID(ctx context.Context, id string) (*model.IncidentDetail, error) {
	return nil, nil
}

func (m *mockNotifIncidentRepo) List(ctx context.Context, filter model.IncidentFilter) ([]*model.IncidentSummary, int, error) {
	return nil, 0, nil
}

func (m *mockNotifIncidentRepo) GetStats(ctx context.Context) (*model.IncidentStats, error) {
	return nil, nil
}

func (m *mockNotifIncidentRepo) FindOpenByAlertAndResource(ctx context.Context, alertName, resourceID string) (*model.Incident, error) {
	return nil, nil
}

func (m *mockNotifIncidentRepo) UpdateStatus(ctx context.Context, id string, status string, actorID *string) error {
	return nil
}

func (m *mockNotifIncidentRepo) GetUnacknowledgedOlderThan(ctx context.Context, duration time.Duration) ([]*model.Incident, error) {
	return nil, nil
}

func (m *mockNotifIncidentRepo) SaveNotification(ctx context.Context, record *model.NotificationRecord) error {
	m.mu.Lock()
	m.savedRecords = append(m.savedRecords, record)
	m.mu.Unlock()
	if m.saveCompleted != nil {
		select {
		case m.saveCompleted <- struct{}{}:
		default:
		}
	}
	return m.saveErr
}

func (m *mockNotifIncidentRepo) ListNotificationsByIncidentID(ctx context.Context, incidentID string) ([]model.NotificationRecord, error) {
	return m.listRecords, m.listErr
}

func TestNotificationService_Notify_Normal(t *testing.T) {
	// init test components
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tg := &mockTelegramService{}
	repo := &mockNotifIncidentRepo{
		saveCompleted: make(chan struct{}, 5),
	}
	hub := ws.NewHub(logger)
	svc := NewNotificationService(tg, hub, repo, logger)

	inc := &model.Incident{
		ID:          "inc-1",
		Title:       "High CPU Usage",
		Description: "CPU exceeded 90%",
		Severity:    "warning",
		Source:      "prometheus",
	}

	// trigger notification
	svc.Notify(context.Background(), inc, false)

	// wait for async save
	select {
	case <-repo.saveCompleted:
	case <-time.After(2 * time.Second):
		t.Fatal("save notification timeout")
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()
	if len(repo.savedRecords) == 0 {
		t.Fatal("expected notification records saved")
	}
	if repo.savedRecords[0].Severity != "warning" {
		t.Errorf("expected warning, got %s", repo.savedRecords[0].Severity)
	}
}

func TestNotificationService_Notify_Escalation(t *testing.T) {
	// init escalation test
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tg := &mockTelegramService{}
	repo := &mockNotifIncidentRepo{
		saveCompleted: make(chan struct{}, 5),
	}
	hub := ws.NewHub(logger)
	svc := NewNotificationService(tg, hub, repo, logger)

	inc := &model.Incident{
		ID:          "inc-2",
		Title:       "Service Outage",
		Description: "Database unreachable",
		Severity:    "critical",
		Source:      "k3d",
	}

	// notify escalation
	svc.Notify(context.Background(), inc, true)

	for i := 0; i < 2; i++ {
		select {
		case <-repo.saveCompleted:
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting escalation save")
		}
	}

	tg.mu.Lock()
	defer tg.mu.Unlock()
	if !tg.isEscalation {
		t.Errorf("expected isEscalation true")
	}
}

func TestNotificationService_Notify_TelegramFail(t *testing.T) {
	// init failure test
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tg := &mockTelegramService{
		sendErr: errors.New("network timeout"),
	}
	repo := &mockNotifIncidentRepo{
		saveCompleted: make(chan struct{}, 5),
	}
	hub := ws.NewHub(logger)
	svc := NewNotificationService(tg, hub, repo, logger)

	inc := &model.Incident{
		ID:          "inc-3",
		Title:       "Memory Leak",
		Description: "OOM imminent",
		Severity:    "critical",
		Source:      "docker",
	}

	// trigger with failed tg
	svc.Notify(context.Background(), inc, false)

	for i := 0; i < 2; i++ {
		select {
		case <-repo.saveCompleted:
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting failed telegram save")
		}
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()
	// verify record has failed status
	var foundFailed bool
	for _, r := range repo.savedRecords {
		if r.Status == "failed" && r.ErrorMessage == "network timeout" {
			foundFailed = true
		}
	}
	if !foundFailed {
		t.Errorf("expected failed status record saved")
	}
}

func TestNotificationService_ListHistory(t *testing.T) {
	// init list history test
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := &mockNotifIncidentRepo{
		listRecords: []model.NotificationRecord{
			{Title: "Alert 1", Status: "sent"},
			{Title: "Alert 2", Status: "sent"},
		},
	}
	svc := NewNotificationService(nil, nil, repo, logger)

	recs, err := svc.ListHistory(context.Background(), "inc-10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recs) != 2 {
		t.Errorf("expected 2 records, got %d", len(recs))
	}
}

func TestNotificationService_NilHubRepo(t *testing.T) {
	// test nil dependencies
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tg := &mockTelegramService{}
	svc := NewNotificationService(tg, nil, nil, logger)

	inc := &model.Incident{
		ID:    "inc-nil",
		Title: "Nil Test",
	}

	// should not panic
	svc.Notify(context.Background(), inc, false)

	recs, err := svc.ListHistory(context.Background(), "inc-nil")
	if err != nil || recs != nil {
		t.Errorf("expected nil nil, got %v, %v", recs, err)
	}
}

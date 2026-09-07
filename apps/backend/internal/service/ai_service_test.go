package service

import (
	"context"
	"testing"
	"time"

	"github.com/cifo-monitoring/backend/internal/integration"
	"github.com/cifo-monitoring/backend/internal/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAIRepo struct {
	mock.Mock
}

func (m *mockAIRepo) CreateSession(ctx context.Context, session *model.AISession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *mockAIRepo) GetSession(ctx context.Context, id uuid.UUID) (*model.AISession, error) {
	args := m.Called(ctx, id)
	if s, ok := args.Get(0).(*model.AISession); ok {
		return s, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAIRepo) ListSessionsByUser(ctx context.Context, userID uuid.UUID) ([]model.AISession, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]model.AISession), args.Error(1)
}

func (m *mockAIRepo) UpdateSessionActivity(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockAIRepo) CreateMessage(ctx context.Context, message *model.AIMessage) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *mockAIRepo) GetMessagesBySession(ctx context.Context, sessionID uuid.UUID, limit int) ([]model.AIMessage, error) {
	args := m.Called(ctx, sessionID, limit)
	return args.Get(0).([]model.AIMessage), args.Error(1)
}

func (m *mockAIRepo) RecordUsage(ctx context.Context, usage *model.AIUsageTracking) error {
	args := m.Called(ctx, usage)
	return args.Error(0)
}

func (m *mockAIRepo) GetUsageStats(ctx context.Context, userID *uuid.UUID) (*model.AIUsageStats, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*model.AIUsageStats), args.Error(1)
}

func (m *mockAIRepo) CreateActionAudit(ctx context.Context, audit *model.AIActionAuditLog) error {
	args := m.Called(ctx, audit)
	return args.Error(0)
}

func (m *mockAIRepo) GetActionAudit(ctx context.Context, id uuid.UUID) (*model.AIActionAuditLog, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.AIActionAuditLog), args.Error(1)
}

func (m *mockAIRepo) UpdateActionAuditStatus(ctx context.Context, id uuid.UUID, status string, result *string) error {
	args := m.Called(ctx, id, status, result)
	return args.Error(0)
}

func (m *mockAIRepo) UpdateIncidentRCA(ctx context.Context, incidentID uuid.UUID, rcaSummary string) error {
	args := m.Called(ctx, incidentID, rcaSummary)
	return args.Error(0)
}

type mockAIClient struct {
	mock.Mock
}

func (m *mockAIClient) Chat(ctx context.Context, sessionID string, userID string, message string, role string, history []map[string]string) (*integration.AIChatClientResponse, error) {
	args := m.Called(ctx, sessionID, userID, message, role, history)
	return args.Get(0).(*integration.AIChatClientResponse), args.Error(1)
}

func (m *mockAIClient) Diagnose(ctx context.Context, incidentID string, alertName string, severity string, resource string, namespace string, logs string, metrics map[string]interface{}) (*integration.AIDiagnoseClientResponse, error) {
	args := m.Called(ctx, incidentID, alertName, severity, resource, namespace, logs, metrics)
	return args.Get(0).(*integration.AIDiagnoseClientResponse), args.Error(1)
}

func (m *mockAIClient) GetModels(ctx context.Context) (map[string]interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func TestProcessChat_NewSession(t *testing.T) {
	repo := new(mockAIRepo)
	client := new(mockAIClient)
	svc := NewDefaultAIService(repo, nil, client, nil, nil, nil)

	userID := uuid.New()
	req := &model.AIChatRequest{
		Message: "How is the cluster?",
	}

	repo.On("CreateSession", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		s := args.Get(1).(*model.AISession)
		s.ID = uuid.New()
	})
	repo.On("GetMessagesBySession", mock.Anything, mock.Anything, 20).Return([]model.AIMessage{}, nil)
	repo.On("CreateMessage", mock.Anything, mock.Anything).Return(nil)
	repo.On("RecordUsage", mock.Anything, mock.Anything).Return(nil)

	client.On("Chat", mock.Anything, mock.Anything, userID.String(), "How is the cluster?", "viewer", mock.Anything).Return(&integration.AIChatClientResponse{
		Content:          "Cluster is nominal",
		ModelUsed:        "gemini-2.0-flash",
		ProviderName:     "google",
		InputTokens:      10,
		OutputTokens:     5,
		EstimatedCostUSD: 0.0001,
		ToolCalls: []map[string]interface{}{
			{
				"name":              "get_pod_status",
				"parameters":        map[string]interface{}{"namespace": "default"},
				"requires_approval": false,
				"required_role":     "viewer",
			},
		},
	}, nil)

	repo.On("CreateActionAudit", mock.Anything, mock.Anything).Return(nil)

	resp, err := svc.ProcessChat(context.Background(), userID, "viewer", req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Cluster is nominal", resp.Content)
	assert.Equal(t, "gemini-2.0-flash", resp.ModelUsed)
	assert.Equal(t, 1, len(resp.ToolCalls))
	assert.Equal(t, "executed", resp.ToolCalls[0].Status)
}

func TestApproveTool_Success(t *testing.T) {
	repo := new(mockAIRepo)
	svc := NewDefaultAIService(repo, nil, nil, nil, nil, nil)

	approvalID := uuid.New()
	userID := uuid.New()
	toolName := "restart_deployment"

	audit := &model.AIActionAuditLog{
		ID:             approvalID,
		UserID:         userID,
		ToolName:       &toolName,
		ToolParameters: []byte(`{"namespace":"default","deployment_name":"api"}`),
		ApprovalStatus: "pending",
		Timestamp:      time.Now(),
	}

	repo.On("GetActionAudit", mock.Anything, approvalID).Return(audit, nil)
	repo.On("UpdateActionAuditStatus", mock.Anything, approvalID, "approved", mock.Anything).Return(nil)

	res, err := svc.ApproveTool(context.Background(), approvalID, userID, "devops")
	assert.NoError(t, err)
	assert.Equal(t, "approved", res.ApprovalStatus)
}

func TestRejectTool_Success(t *testing.T) {
	repo := new(mockAIRepo)
	svc := NewDefaultAIService(repo, nil, nil, nil, nil, nil)

	approvalID := uuid.New()
	userID := uuid.New()

	audit := &model.AIActionAuditLog{
		ID:             approvalID,
		UserID:         userID,
		ApprovalStatus: "pending",
	}

	repo.On("GetActionAudit", mock.Anything, approvalID).Return(audit, nil)
	repo.On("UpdateActionAuditStatus", mock.Anything, approvalID, "rejected", mock.Anything).Return(nil)

	res, err := svc.RejectTool(context.Background(), approvalID, userID, "viewer")
	assert.NoError(t, err)
	assert.Equal(t, "rejected", res.ApprovalStatus)
}

func TestListSessions(t *testing.T) {
	repo := new(mockAIRepo)
	svc := NewDefaultAIService(repo, nil, nil, nil, nil, nil)
	userID := uuid.New()

	sessions := []model.AISession{
		{ID: uuid.New(), UserID: userID, Status: "active"},
	}
	repo.On("ListSessionsByUser", mock.Anything, userID).Return(sessions, nil)

	res, err := svc.ListSessions(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, res, 1)
}

func TestGetSessionMessages(t *testing.T) {
	repo := new(mockAIRepo)
	svc := NewDefaultAIService(repo, nil, nil, nil, nil, nil)
	userID := uuid.New()
	sessionID := uuid.New()

	sess := &model.AISession{ID: sessionID, UserID: userID}
	repo.On("GetSession", mock.Anything, sessionID).Return(sess, nil)

	msgs := []model.AIMessage{
		{ID: uuid.New(), SessionID: sessionID, Role: "user", Content: "hello"},
	}
	repo.On("GetMessagesBySession", mock.Anything, sessionID, 50).Return(msgs, nil)

	res, err := svc.GetSessionMessages(context.Background(), sessionID, userID)
	assert.NoError(t, err)
	assert.Len(t, res, 1)

	// forbidden for different user
	otherUser := uuid.New()
	_, err = svc.GetSessionMessages(context.Background(), sessionID, otherUser)
	assert.Error(t, err)
}

func TestGetUsage(t *testing.T) {
	repo := new(mockAIRepo)
	svc := NewDefaultAIService(repo, nil, nil, nil, nil, nil)
	userID := uuid.New()

	stats := &model.AIUsageStats{TotalTokens: 100, TotalCostUSD: 0.05, RequestCount: 10}
	// admin gets system-wide stats (userID filter is nil)
	repo.On("GetUsageStats", mock.Anything, (*uuid.UUID)(nil)).Return(stats, nil)
	res, err := svc.GetUsage(context.Background(), userID, "admin")
	assert.NoError(t, err)
	assert.Equal(t, 100, res.TotalTokens)

	// viewer gets filtered stats
	repo.On("GetUsageStats", mock.Anything, &userID).Return(stats, nil)
	res, err = svc.GetUsage(context.Background(), userID, "viewer")
	assert.NoError(t, err)
	assert.Equal(t, 10, res.RequestCount)
}

func TestListModels(t *testing.T) {
	client := new(mockAIClient)
	svc := NewDefaultAIService(nil, nil, client, nil, nil, nil)

	models := map[string]interface{}{"primary": "gemini-2.5"}
	client.On("GetModels", mock.Anything).Return(models, nil)

	res, err := svc.ListModels(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "gemini-2.5", res["primary"])
}

func TestApproveTool_EdgeCases(t *testing.T) {
	repo := new(mockAIRepo)
	svc := NewDefaultAIService(repo, nil, nil, nil, nil, nil)
	userID := uuid.New()
	approvalID := uuid.New()

	// not pending
	auditNotPending := &model.AIActionAuditLog{ID: approvalID, ApprovalStatus: "already_approved"}
	repo.On("GetActionAudit", mock.Anything, approvalID).Return(auditNotPending, nil)
	_, err := svc.ApproveTool(context.Background(), approvalID, userID, "devops")
	assert.Error(t, err)

	// viewer role forbidden
	toolName := "restart_deployment"
	auditPending := &model.AIActionAuditLog{
		ID:             uuid.New(),
		ApprovalStatus: "pending",
		ToolName:       &toolName,
	}
	repo.On("GetActionAudit", mock.Anything, auditPending.ID).Return(auditPending, nil)
	_, err = svc.ApproveTool(context.Background(), auditPending.ID, userID, "viewer")
	assert.Error(t, err)
}

type mockIncidentRepoForAI struct {
	mock.Mock
}

func (m *mockIncidentRepoForAI) Create(ctx context.Context, incident *model.Incident) error {
	return nil
}
func (m *mockIncidentRepoForAI) GetByID(ctx context.Context, id string) (*model.Incident, error) {
	args := m.Called(ctx, id)
	if inc, ok := args.Get(0).(*model.Incident); ok {
		return inc, args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockIncidentRepoForAI) GetDetailByID(ctx context.Context, id string) (*model.IncidentDetail, error) {
	return nil, nil
}
func (m *mockIncidentRepoForAI) List(ctx context.Context, filter model.IncidentFilter) ([]*model.IncidentSummary, int, error) {
	return nil, 0, nil
}
func (m *mockIncidentRepoForAI) GetStats(ctx context.Context) (*model.IncidentStats, error) {
	return nil, nil
}
func (m *mockIncidentRepoForAI) FindOpenByAlertAndResource(ctx context.Context, alertName, resourceID string) (*model.Incident, error) {
	return nil, nil
}
func (m *mockIncidentRepoForAI) UpdateStatus(ctx context.Context, id string, status string, actorID *string) error {
	return nil
}
func (m *mockIncidentRepoForAI) GetUnacknowledgedOlderThan(ctx context.Context, duration time.Duration) ([]*model.Incident, error) {
	return nil, nil
}
func (m *mockIncidentRepoForAI) SaveNotification(ctx context.Context, notif *model.NotificationRecord) error {
	return nil
}
func (m *mockIncidentRepoForAI) ListNotificationsByIncidentID(ctx context.Context, incidentID string) ([]model.NotificationRecord, error) {
	return nil, nil
}

func TestGenerateRCAForIncident(t *testing.T) {
	repo := new(mockAIRepo)
	incRepo := new(mockIncidentRepoForAI)
	client := new(mockAIClient)
	svc := NewDefaultAIService(repo, incRepo, client, nil, nil, nil)

	incID := uuid.New()
	inc := &model.Incident{
		ID:          incID.String(),
		Title:       "High latency",
		Severity:    "critical",
		Namespace:   "prod",
		ResourceID:  "api-gateway",
		Description: "Timeout spike",
	}

	incRepo.On("GetByID", mock.Anything, incID.String()).Return(inc, nil)

	diag := &integration.AIDiagnoseClientResponse{
		RCASummary:       "Memory leak in pod",
		ModelUsed:        "gemini-2.5",
		ProviderName:     "google",
		EstimatedCostUSD: 0.002,
	}
	client.On("Diagnose", mock.Anything, incID.String(), "High latency", "critical", "api-gateway", "prod", "Timeout spike", mock.Anything).Return(diag, nil)
	repo.On("UpdateIncidentRCA", mock.Anything, incID, "Memory leak in pod").Return(nil)

	res, err := svc.GenerateRCAForIncident(context.Background(), incID)
	assert.NoError(t, err)
	assert.Equal(t, "Memory leak in pod", res.RCASummary)
}

func TestExecuteReadOnlyTool_AllBranches(t *testing.T) {
	svc := NewDefaultAIService(nil, nil, nil, nil, nil, nil)
	ctx := context.Background()

	res1 := svc.executeReadOnlyTool(ctx, "get_pod_status", map[string]interface{}{"namespace": "test-ns"})
	assert.Contains(t, res1, "test-ns")

	res2 := svc.executeReadOnlyTool(ctx, "get_container_logs", map[string]interface{}{"container_id": "c123"})
	assert.Contains(t, res2, "c123")

	res3 := svc.executeReadOnlyTool(ctx, "list_docker_containers", nil)
	assert.Contains(t, res3, "Docker")

	res4 := svc.executeReadOnlyTool(ctx, "get_argocd_app_status", map[string]interface{}{"app_name": "payment"})
	assert.Contains(t, res4, "payment")

	resDefault := svc.executeReadOnlyTool(ctx, "custom_read_tool", nil)
	assert.Contains(t, resDefault, "custom_read_tool")
}

func TestExecuteWriteTool_AllBranches(t *testing.T) {
	svc := NewDefaultAIService(nil, nil, nil, nil, nil, nil)
	ctx := context.Background()

	res1 := svc.executeWriteTool(ctx, "restart_deployment", map[string]interface{}{"namespace": "kube", "deployment_name": "core"}, "admin")
	assert.Contains(t, res1, "core")

	res2 := svc.executeWriteTool(ctx, "scale_deployment", map[string]interface{}{"namespace": "kube", "deployment_name": "core", "replicas": float64(3)}, "admin")
	assert.Contains(t, res2, "3 replicas")

	res3 := svc.executeWriteTool(ctx, "restart_container", map[string]interface{}{"container_id": "cont-1"}, "admin")
	assert.Contains(t, res3, "cont-1")

	res4 := svc.executeWriteTool(ctx, "stop_container", map[string]interface{}{"container_id": "cont-2"}, "admin")
	assert.Contains(t, res4, "cont-2")

	res5 := svc.executeWriteTool(ctx, "sync_argocd_app", map[string]interface{}{"app_name": "app-1", "prune": true}, "admin")
	assert.Contains(t, res5, "app-1")

	resDefault := svc.executeWriteTool(ctx, "unknown_write", nil, "admin")
	assert.Contains(t, resDefault, "unknown_write")
}

func TestSha256Hex(t *testing.T) {
	h := sha256Hex("hello world")
	assert.NotEmpty(t, h)
}

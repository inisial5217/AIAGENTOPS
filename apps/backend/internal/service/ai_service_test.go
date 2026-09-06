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

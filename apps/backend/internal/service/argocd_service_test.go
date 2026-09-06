package service

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockArgoCDClient mock argocd client
type MockArgoCDClient struct {
	mock.Mock
}

func (m *MockArgoCDClient) ListApplications(ctx context.Context, namespace string) ([]model.ArgoApplicationSummary, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).([]model.ArgoApplicationSummary), args.Error(1)
}

func (m *MockArgoCDClient) GetApplication(ctx context.Context, namespace string, name string) (*model.ArgoApplicationDetail, error) {
	args := m.Called(ctx, namespace, name)
	return args.Get(0).(*model.ArgoApplicationDetail), args.Error(1)
}

func (m *MockArgoCDClient) SyncApplication(ctx context.Context, namespace string, name string, req model.ArgoSyncRequest) error {
	args := m.Called(ctx, namespace, name, req)
	return args.Error(0)
}

func TestListApplications(t *testing.T) {
	mockClient := new(MockArgoCDClient)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewArgoCDService(mockClient, nil, logger)

	mockApps := []model.ArgoApplicationSummary{
		{
			Name:         "sample-nginx-app",
			Project:      "default",
			RepoURL:      "https://github.com/argoproj/argocd-example-apps.git",
			SyncStatus:   "Synced",
			HealthStatus: "Healthy",
		},
		{
			Name:         "sample-httpbin-app",
			Project:      "default",
			RepoURL:      "https://github.com/argoproj/argocd-example-apps.git",
			SyncStatus:   "OutOfSync",
			HealthStatus: "Degraded",
		},
	}

	mockClient.On("ListApplications", mock.Anything, "argocd").Return(mockApps, nil)

	ctx := context.Background()
	apps, err := svc.ListApplications(ctx, "argocd")
	assert.NoError(t, err)
	assert.Len(t, apps, 2)
	assert.Equal(t, "sample-nginx-app", apps[0].Name)
	assert.Equal(t, "Synced", apps[0].SyncStatus)
}

func TestGetApplication(t *testing.T) {
	mockClient := new(MockArgoCDClient)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewArgoCDService(mockClient, nil, logger)

	mockDetail := &model.ArgoApplicationDetail{
		ArgoApplicationSummary: model.ArgoApplicationSummary{
			Name:         "sample-nginx-app",
			SyncStatus:   "Synced",
			HealthStatus: "Healthy",
		},
		AutomatedSync: true,
		Prune:         true,
		SelfHeal:      true,
		Resources: []model.ArgoResourceStatus{
			{
				Kind:   "Deployment",
				Name:   "nginx-deployment",
				Status: "Synced",
				Health: "Healthy",
			},
		},
		History: []model.ArgoDeploymentRevision{
			{
				ID:       1,
				Revision: "8088f4c0",
			},
		},
	}

	mockClient.On("GetApplication", mock.Anything, "argocd", "sample-nginx-app").Return(mockDetail, nil)

	ctx := context.Background()
	detail, err := svc.GetApplication(ctx, "argocd", "sample-nginx-app")
	assert.NoError(t, err)
	assert.Equal(t, "sample-nginx-app", detail.Name)
	assert.True(t, detail.AutomatedSync)
	assert.Len(t, detail.Resources, 1)
	assert.Len(t, detail.History, 1)
}

func TestSyncApplication(t *testing.T) {
	mockClient := new(MockArgoCDClient)
	mockAudit := new(MockAuditRepo)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewArgoCDService(mockClient, mockAudit, logger)

	req := model.ArgoSyncRequest{
		Prune:  true,
		DryRun: false,
	}

	mockClient.On("SyncApplication", mock.Anything, "argocd", "sample-nginx-app", req).Return(nil)
	mockAudit.On("Create", mock.Anything, mock.MatchedBy(func(l *model.AuditLog) bool {
		return l.Action == "sync_argocd_application" && l.ResourceType == "argocd"
	})).Return(nil)

	ctx := context.Background()
	err := svc.SyncApplication(ctx, "argocd", "sample-nginx-app", req, "devops-user", "127.0.0.1")
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
	mockAudit.AssertExpectations(t)
}

func TestGetOverview(t *testing.T) {
	mockClient := new(MockArgoCDClient)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewArgoCDService(mockClient, nil, logger)

	mockApps := []model.ArgoApplicationSummary{
		{
			Name:         "app-1",
			SyncStatus:   "Synced",
			HealthStatus: "Healthy",
		},
		{
			Name:         "app-2",
			SyncStatus:   "OutOfSync",
			HealthStatus: "Degraded",
		},
		{
			Name:         "app-3",
			SyncStatus:   "Synced",
			HealthStatus: "Progressing",
		},
	}

	mockClient.On("ListApplications", mock.Anything, "argocd").Return(mockApps, nil)

	ctx := context.Background()
	overview, err := svc.GetOverview(ctx, "argocd")
	assert.NoError(t, err)
	assert.Equal(t, 3, overview["total"])
	assert.Equal(t, 2, overview["synced"])
	assert.Equal(t, 1, overview["out_of_sync"])
	assert.Equal(t, 1, overview["healthy"])
	assert.Equal(t, 1, overview["degraded"])
	assert.Equal(t, 1, overview["progressing"])
}

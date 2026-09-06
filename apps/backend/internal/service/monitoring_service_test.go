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

type MockDockerService struct {
	mock.Mock
}

func (m *MockDockerService) ListContainers(ctx context.Context, statusFilter string) ([]model.ContainerSummary, error) {
	args := m.Called(ctx, statusFilter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.ContainerSummary), args.Error(1)
}

func (m *MockDockerService) GetContainer(ctx context.Context, id string) (model.ContainerDetail, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.ContainerDetail), args.Error(1)
}

func (m *MockDockerService) GetContainerStats(ctx context.Context, id string) (model.ContainerStats, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.ContainerStats), args.Error(1)
}

func (m *MockDockerService) GetContainerLogs(ctx context.Context, id string, tail int) (string, error) {
	args := m.Called(ctx, id, tail)
	return args.String(0), args.Error(1)
}

func (m *MockDockerService) RestartContainer(ctx context.Context, id string, actor string, ip string) error {
	args := m.Called(ctx, id, actor, ip)
	return args.Error(0)
}

func (m *MockDockerService) StopContainer(ctx context.Context, id string, actor string, ip string) error {
	args := m.Called(ctx, id, actor, ip)
	return args.Error(0)
}

func (m *MockDockerService) ListImages(ctx context.Context) ([]model.DockerImageInfo, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.DockerImageInfo), args.Error(1)
}

func (m *MockDockerService) ListVolumes(ctx context.Context) ([]model.DockerVolumeInfo, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.DockerVolumeInfo), args.Error(1)
}

func (m *MockDockerService) ListNetworks(ctx context.Context) ([]model.DockerNetworkInfo, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.DockerNetworkInfo), args.Error(1)
}

func (m *MockDockerService) GetSystemInfo(ctx context.Context) (model.DockerSystemInfo, error) {
	args := m.Called(ctx)
	return args.Get(0).(model.DockerSystemInfo), args.Error(1)
}

func TestGetDashboardStats(t *testing.T) {
	mockDocker := new(MockDockerService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewMonitoringService(mockDocker, nil, logger)

	mockDocker.On("ListContainers", mock.Anything, "").Return([]model.ContainerSummary{
		{ID: "c1", State: "running"},
		{ID: "c2", State: "running"},
		{ID: "c3", State: "stopped"},
	}, nil)

	mockDocker.On("GetSystemInfo", mock.Anything).Return(model.DockerSystemInfo{
		ContainersTotal:   10,
		ContainersRunning: 8,
		ContainersStopped: 2,
		TotalMemory:       16 * 1024 * 1024 * 1024,
	}, nil)

	ctx := context.Background()
	stats, err := svc.GetDashboardStats(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 10, stats.TotalContainers)
	assert.Equal(t, 8, stats.ContainersOn)
	assert.Equal(t, 2, stats.ContainersOff)
}

func TestGetCPUMetrics(t *testing.T) {
	mockDocker := new(MockDockerService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewMonitoringService(mockDocker, nil, logger)

	ctx := context.Background()
	points, err := svc.GetCPUMetrics(ctx, "1h")
	assert.NoError(t, err)
	assert.Len(t, points, 8)
	assert.NotEmpty(t, points[0].Timestamp)
}

func TestGetMemoryMetrics(t *testing.T) {
	mockDocker := new(MockDockerService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewMonitoringService(mockDocker, nil, logger)

	ctx := context.Background()
	points, err := svc.GetMemoryMetrics(ctx, "1h")
	assert.NoError(t, err)
	assert.Len(t, points, 8)
	assert.NotEmpty(t, points[0].Timestamp)
	assert.Greater(t, points[0].Value, 0.0)
}

func TestGetNetworkMetrics(t *testing.T) {
	mockDocker := new(MockDockerService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewMonitoringService(mockDocker, nil, logger)

	ctx := context.Background()
	points, err := svc.GetNetworkMetrics(ctx, "1h")
	assert.NoError(t, err)
	assert.Len(t, points, 8)
	assert.NotEmpty(t, points[0].Timestamp)
}

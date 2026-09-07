package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/api/types/volume"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/cifo-monitoring/backend/internal/model"
)

// MockDockerClient mock implementation of integration.DockerClient
type MockDockerClient struct {
	mock.Mock
}

func (m *MockDockerClient) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockDockerClient) ListContainers(ctx context.Context, all bool) ([]types.Container, error) {
	args := m.Called(ctx, all)
	return args.Get(0).([]types.Container), args.Error(1)
}

func (m *MockDockerClient) GetContainer(ctx context.Context, id string) (types.ContainerJSON, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(types.ContainerJSON), args.Error(1)
}

func (m *MockDockerClient) GetContainerStats(ctx context.Context, id string) (types.ContainerStats, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(types.ContainerStats), args.Error(1)
}

func (m *MockDockerClient) GetContainerLogs(ctx context.Context, id string, tail int) (io.ReadCloser, error) {
	args := m.Called(ctx, id, tail)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockDockerClient) StreamContainerLogs(ctx context.Context, id string, tail int, follow bool) (io.ReadCloser, error) {
	args := m.Called(ctx, id, tail, follow)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockDockerClient) ListenEvents(ctx context.Context) (<-chan events.Message, <-chan error) {
	args := m.Called(ctx)
	return args.Get(0).(<-chan events.Message), args.Get(1).(<-chan error)
}

func (m *MockDockerClient) RestartContainer(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDockerClient) StopContainer(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDockerClient) ListImages(ctx context.Context) ([]image.Summary, error) {
	args := m.Called(ctx)
	return args.Get(0).([]image.Summary), args.Error(1)
}

func (m *MockDockerClient) ListVolumes(ctx context.Context) (volume.ListResponse, error) {
	args := m.Called(ctx)
	return args.Get(0).(volume.ListResponse), args.Error(1)
}

func (m *MockDockerClient) ListNetworks(ctx context.Context) ([]network.Inspect, error) {
	args := m.Called(ctx)
	return args.Get(0).([]network.Inspect), args.Error(1)
}

func (m *MockDockerClient) GetSystemInfo(ctx context.Context) (system.Info, error) {
	args := m.Called(ctx)
	return args.Get(0).(system.Info), args.Error(1)
}

func (m *MockDockerClient) Close() error {
	return nil
}

// mockDockerAuditRepo mock
type mockDockerAuditRepo struct {
	mock.Mock
}

func (m *mockDockerAuditRepo) Create(ctx context.Context, log *model.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *mockDockerAuditRepo) List(ctx context.Context, limit, offset int) ([]*model.AuditLog, int, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*model.AuditLog), args.Int(1), args.Error(2)
}

func TestListContainers_Filter(t *testing.T) {
	mockClient := new(MockDockerClient)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewDockerService(mockClient, nil, nil, logger)

	mockClient.On("ListContainers", mock.Anything, true).Return([]types.Container{
		{
			ID:    "c1",
			Names: []string{"/web"},
			State: "running",
			Ports: []types.Port{{IP: "0.0.0.0", PrivatePort: 80, PublicPort: 8080, Type: "tcp"}},
		},
		{ID: "c2", Names: []string{"/db"}, State: "exited"},
	}, nil)

	ctx := context.Background()

	// test all
	all, err := svc.ListContainers(ctx, "")
	assert.NoError(t, err)
	assert.Len(t, all, 2)

	// test running filter
	running, err := svc.ListContainers(ctx, "running")
	assert.NoError(t, err)
	assert.Len(t, running, 1)
	assert.Equal(t, "c1", running[0].ID)

	// test stopped filter
	stopped, err := svc.ListContainers(ctx, "stopped")
	assert.NoError(t, err)
	assert.Len(t, stopped, 1)
	assert.Equal(t, "c2", stopped[0].ID)
}

func TestGetContainer_Success(t *testing.T) {
	mockClient := new(MockDockerClient)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewDockerService(mockClient, nil, nil, logger)

	jsonResp := types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			ID:   "c123",
			Name: "/cifo-app",
			State: &types.ContainerState{
				Status:  "running",
				Running: true,
			},
		},
		Config: &container.Config{
			Image:  "cifo:latest",
			Labels: map[string]string{"env": "test"},
		},
		NetworkSettings: &types.NetworkSettings{
			DefaultNetworkSettings: types.DefaultNetworkSettings{
				IPAddress: "172.18.0.2",
			},
		},
	}

	mockClient.On("GetContainer", mock.Anything, "c123").Return(jsonResp, nil)

	ctx := context.Background()
	detail, err := svc.GetContainer(ctx, "c123")
	assert.NoError(t, err)
	assert.Equal(t, "cifo-app", detail.Name)
	assert.Equal(t, "172.18.0.2", detail.IPAddress)
	assert.True(t, detail.State.Running)
}

func TestGetContainer_Error(t *testing.T) {
	mockClient := new(MockDockerClient)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewDockerService(mockClient, nil, nil, logger)

	mockClient.On("GetContainer", mock.Anything, "c-notfound").Return(types.ContainerJSON{}, errors.New("not found"))

	ctx := context.Background()
	_, err := svc.GetContainer(ctx, "c-notfound")
	assert.Error(t, err)
}

func TestGetContainerStats_Success(t *testing.T) {
	mockClient := new(MockDockerClient)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewDockerService(mockClient, nil, nil, logger)

	statsJSON := `{
		"name": "/cifo-web",
		"cpu_stats": {
			"cpu_usage": {"total_usage": 200000000, "percpu_usage": [100000000, 100000000]},
			"system_cpu_usage": 2000000000,
			"online_cpus": 2
		},
		"precpu_stats": {
			"cpu_usage": {"total_usage": 100000000},
			"system_cpu_usage": 1000000000
		},
		"memory_stats": {
			"usage": 52428800,
			"limit": 104857600
		},
		"networks": {
			"eth0": {"rx_bytes": 1024, "tx_bytes": 2048}
		},
		"pids_stats": {
			"current": 5
		}
	}`

	mockClient.On("GetContainerStats", mock.Anything, "c1").Return(types.ContainerStats{
		Body: io.NopCloser(bytes.NewBufferString(statsJSON)),
	}, nil)

	ctx := context.Background()
	stats, err := svc.GetContainerStats(ctx, "c1")
	assert.NoError(t, err)
	assert.Equal(t, "cifo-web", stats.ContainerName)
	assert.Equal(t, uint64(52428800), stats.MemoryUsageBytes)
	assert.Equal(t, 50.0, stats.MemoryPercentage)
	assert.Equal(t, uint64(5), stats.PidsCurrent)
}

func TestRestartContainer_Success(t *testing.T) {
	mockClient := new(MockDockerClient)
	mockAudit := new(mockDockerAuditRepo)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewDockerService(mockClient, mockAudit, nil, logger)

	mockClient.On("RestartContainer", mock.Anything, "c1").Return(nil)
	mockAudit.On("Create", mock.Anything, mock.Anything).Return(nil)

	ctx := context.Background()
	err := svc.RestartContainer(ctx, "c1", "admin@cifo.local", "127.0.0.1")
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
	mockAudit.AssertExpectations(t)
}

func TestRestartContainer_Error(t *testing.T) {
	mockClient := new(MockDockerClient)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewDockerService(mockClient, nil, nil, logger)

	mockClient.On("RestartContainer", mock.Anything, "c1").Return(errors.New("restart blocked"))

	ctx := context.Background()
	err := svc.RestartContainer(ctx, "c1", "admin@cifo.local", "127.0.0.1")
	assert.Error(t, err)
}

func TestStopContainer_Success(t *testing.T) {
	mockClient := new(MockDockerClient)
	mockAudit := new(mockDockerAuditRepo)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewDockerService(mockClient, mockAudit, nil, logger)

	mockClient.On("StopContainer", mock.Anything, "c1").Return(nil)
	mockAudit.On("Create", mock.Anything, mock.Anything).Return(nil)

	ctx := context.Background()
	err := svc.StopContainer(ctx, "c1", "admin@cifo.local", "127.0.0.1")
	assert.NoError(t, err)
}

func TestStopContainer_Error(t *testing.T) {
	mockClient := new(MockDockerClient)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewDockerService(mockClient, nil, nil, logger)

	mockClient.On("StopContainer", mock.Anything, "c1").Return(errors.New("stop failed"))

	ctx := context.Background()
	err := svc.StopContainer(ctx, "c1", "admin@cifo.local", "127.0.0.1")
	assert.Error(t, err)
}

func TestListImages_Success(t *testing.T) {
	mockClient := new(MockDockerClient)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewDockerService(mockClient, nil, nil, logger)

	mockClient.On("ListImages", mock.Anything).Return([]image.Summary{
		{ID: "img-1", RepoTags: []string{"alpine:latest"}, Size: 5000000},
	}, nil)

	ctx := context.Background()
	imgs, err := svc.ListImages(ctx)
	assert.NoError(t, err)
	assert.Len(t, imgs, 1)
	assert.Equal(t, "alpine:latest", imgs[0].RepoTags[0])
}

func TestListVolumes_Success(t *testing.T) {
	mockClient := new(MockDockerClient)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewDockerService(mockClient, nil, nil, logger)

	mockClient.On("ListVolumes", mock.Anything).Return(volume.ListResponse{
		Volumes: []*volume.Volume{
			{Name: "cifo_data", Driver: "local", Mountpoint: "/var/lib/docker/volumes/cifo_data"},
		},
	}, nil)

	ctx := context.Background()
	vols, err := svc.ListVolumes(ctx)
	assert.NoError(t, err)
	assert.Len(t, vols, 1)
	assert.Equal(t, "cifo_data", vols[0].Name)
}

func TestListNetworks_Success(t *testing.T) {
	mockClient := new(MockDockerClient)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewDockerService(mockClient, nil, nil, logger)

	mockClient.On("ListNetworks", mock.Anything).Return([]network.Inspect{
		{ID: "net-1", Name: "cifo-net", Driver: "bridge", Scope: "local"},
	}, nil)

	ctx := context.Background()
	nets, err := svc.ListNetworks(ctx)
	assert.NoError(t, err)
	assert.Len(t, nets, 1)
	assert.Equal(t, "cifo-net", nets[0].Name)
}

func TestGetSystemInfo_Success(t *testing.T) {
	mockClient := new(MockDockerClient)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewDockerService(mockClient, nil, nil, logger)

	mockClient.On("GetSystemInfo", mock.Anything).Return(system.Info{
		Containers:        10,
		ContainersRunning: 8,
		ContainersPaused:  0,
		ContainersStopped: 2,
		Images:            15,
		ServerVersion:     "24.0.7",
		OperatingSystem:   "Linux",
		NCPU:              4,
		MemTotal:          16000000000,
	}, nil)

	ctx := context.Background()
	sys, err := svc.GetSystemInfo(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 10, sys.ContainersTotal)
	assert.Equal(t, 8, sys.ContainersRunning)
	assert.Equal(t, "24.0.7", sys.DockerVersion)
}

func TestGetContainerLogs_Demux(t *testing.T) {
	mockClient := new(MockDockerClient)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewDockerService(mockClient, nil, nil, logger)

	// mock raw logs
	logContent := "2026-09-05 12:00:00 [INFO] container starting up\n"
	mockReader := io.NopCloser(strings.NewReader(logContent))

	mockClient.On("GetContainerLogs", mock.Anything, "c1", 200).Return(mockReader, nil)

	ctx := context.Background()
	logs, err := svc.GetContainerLogs(ctx, "c1", 200)
	assert.NoError(t, err)
	assert.Contains(t, logs, "container starting up")
}

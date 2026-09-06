package service

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/api/types/volume"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestListContainers_Filter(t *testing.T) {
	mockClient := new(MockDockerClient)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewDockerService(mockClient, nil, nil, logger)

	mockClient.On("ListContainers", mock.Anything, true).Return([]types.Container{
		{ID: "c1", Names: []string{"/web"}, State: "running"},
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

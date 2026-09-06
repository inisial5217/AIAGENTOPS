package integration

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/cifo-monitoring/backend/pkg/apperror"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

// DockerClient interface for engine interactions
type DockerClient interface {
	Ping(ctx context.Context) error
	ListContainers(ctx context.Context, all bool) ([]types.Container, error)
	GetContainer(ctx context.Context, id string) (types.ContainerJSON, error)
	GetContainerStats(ctx context.Context, id string) (types.ContainerStats, error)
	GetContainerLogs(ctx context.Context, id string, tail int) (io.ReadCloser, error)
	StreamContainerLogs(ctx context.Context, id string, tail int, follow bool) (io.ReadCloser, error)
	ListenEvents(ctx context.Context) (<-chan events.Message, <-chan error)
	RestartContainer(ctx context.Context, id string) error
	StopContainer(ctx context.Context, id string) error
	ListImages(ctx context.Context) ([]image.Summary, error)
	ListVolumes(ctx context.Context) (volume.ListResponse, error)
	ListNetworks(ctx context.Context) ([]network.Inspect, error)
	GetSystemInfo(ctx context.Context) (system.Info, error)
	Close() error
}

// dockerClientImpl wrapper for official docker client
type dockerClientImpl struct {
	cli *client.Client
}

// NewDockerClient creates negotiated docker client
func NewDockerClient(host string) (DockerClient, error) {
	opts := []client.Opt{
		client.WithAPIVersionNegotiation(),
	}

	if host != "" {
		opts = append(opts, client.WithHost(host))
	} else {
		opts = append(opts, client.FromEnv)
	}

	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, apperror.NewInternal(fmt.Sprintf("failed to init docker client: %v", err))
	}

	return &dockerClientImpl{cli: cli}, nil
}

// Ping checks daemon connectivity
func (d *dockerClientImpl) Ping(ctx context.Context) error {
	_, err := d.cli.Ping(ctx)
	if err != nil {
		return apperror.NewInternal(fmt.Sprintf("docker daemon ping failed: %v", err))
	}
	return nil
}

// ListContainers returns active or all containers
func (d *dockerClientImpl) ListContainers(ctx context.Context, all bool) ([]types.Container, error) {
	containers, err := d.cli.ContainerList(ctx, container.ListOptions{All: all})
	if err != nil {
		return nil, apperror.NewInternal(fmt.Sprintf("failed to list containers: %v", err))
	}
	return containers, nil
}

// GetContainer inspects container details
func (d *dockerClientImpl) GetContainer(ctx context.Context, id string) (types.ContainerJSON, error) {
	detail, err := d.cli.ContainerInspect(ctx, id)
	if err != nil {
		if client.IsErrNotFound(err) {
			return types.ContainerJSON{}, apperror.NewNotFound(fmt.Sprintf("container %s not found", id))
		}
		return types.ContainerJSON{}, apperror.NewInternal(fmt.Sprintf("failed to inspect container: %v", err))
	}
	return detail, nil
}

// GetContainerStats retrieves raw stream stats
func (d *dockerClientImpl) GetContainerStats(ctx context.Context, id string) (types.ContainerStats, error) {
	stats, err := d.cli.ContainerStats(ctx, id, false)
	if err != nil {
		return types.ContainerStats{}, apperror.NewInternal(fmt.Sprintf("failed to get stats: %v", err))
	}
	return stats, nil
}

// GetContainerLogs returns log reader
func (d *dockerClientImpl) GetContainerLogs(ctx context.Context, id string, tail int) (io.ReadCloser, error) {
	tailStr := fmt.Sprintf("%d", tail)
	if tail <= 0 {
		tailStr = "200"
	}

	opts := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       tailStr,
		Timestamps: true,
	}

	logs, err := d.cli.ContainerLogs(ctx, id, opts)
	if err != nil {
		return nil, apperror.NewInternal(fmt.Sprintf("failed to fetch container logs: %v", err))
	}
	return logs, nil
}

// StreamContainerLogs returns continuous log reader
func (d *dockerClientImpl) StreamContainerLogs(ctx context.Context, id string, tail int, follow bool) (io.ReadCloser, error) {
	tailStr := fmt.Sprintf("%d", tail)
	if tail <= 0 {
		tailStr = "100"
	}

	opts := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Tail:       tailStr,
		Timestamps: true,
	}

	logs, err := d.cli.ContainerLogs(ctx, id, opts)
	if err != nil {
		return nil, apperror.NewInternal(fmt.Sprintf("failed to stream container logs: %v", err))
	}
	return logs, nil
}

// ListenEvents listens to daemon events
func (d *dockerClientImpl) ListenEvents(ctx context.Context) (<-chan events.Message, <-chan error) {
	return d.cli.Events(ctx, events.ListOptions{})
}

// RestartContainer restarts targeted container
func (d *dockerClientImpl) RestartContainer(ctx context.Context, id string) error {
	timeout := 15
	opts := container.StopOptions{Timeout: &timeout}
	if err := d.cli.ContainerRestart(ctx, id, opts); err != nil {
		return apperror.NewInternal(fmt.Sprintf("failed to restart container %s: %v", id, err))
	}
	return nil
}

// StopContainer stops running container
func (d *dockerClientImpl) StopContainer(ctx context.Context, id string) error {
	timeout := 15
	opts := container.StopOptions{Timeout: &timeout}
	if err := d.cli.ContainerStop(ctx, id, opts); err != nil {
		return apperror.NewInternal(fmt.Sprintf("failed to stop container %s: %v", id, err))
	}
	return nil
}

// ListImages retrieves cached images
func (d *dockerClientImpl) ListImages(ctx context.Context) ([]image.Summary, error) {
	images, err := d.cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return nil, apperror.NewInternal(fmt.Sprintf("failed to list images: %v", err))
	}
	return images, nil
}

// ListVolumes retrieves volumes
func (d *dockerClientImpl) ListVolumes(ctx context.Context) (volume.ListResponse, error) {
	vols, err := d.cli.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return volume.ListResponse{}, apperror.NewInternal(fmt.Sprintf("failed to list volumes: %v", err))
	}
	return vols, nil
}

// ListNetworks retrieves networks
func (d *dockerClientImpl) ListNetworks(ctx context.Context) ([]network.Inspect, error) {
	nets, err := d.cli.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return nil, apperror.NewInternal(fmt.Sprintf("failed to list networks: %v", err))
	}
	return nets, nil
}

// GetSystemInfo fetches host information
func (d *dockerClientImpl) GetSystemInfo(ctx context.Context) (system.Info, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	info, err := d.cli.Info(ctxTimeout)
	if err != nil {
		return system.Info{}, apperror.NewInternal(fmt.Sprintf("failed to fetch system info: %v", err))
	}
	return info, nil
}

// Close releases client connections
func (d *dockerClientImpl) Close() error {
	return d.cli.Close()
}

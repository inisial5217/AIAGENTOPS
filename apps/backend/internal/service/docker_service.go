package service

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/cifo-monitoring/backend/internal/integration"
	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/cifo-monitoring/backend/internal/repository"
	"github.com/cifo-monitoring/backend/pkg/apperror"
	"github.com/redis/go-redis/v9"
)

// DockerService business logic for Docker operations
type DockerService interface {
	ListContainers(ctx context.Context, statusFilter string) ([]model.ContainerSummary, error)
	GetContainer(ctx context.Context, id string) (model.ContainerDetail, error)
	GetContainerStats(ctx context.Context, id string) (model.ContainerStats, error)
	GetContainerLogs(ctx context.Context, id string, tail int) (string, error)
	RestartContainer(ctx context.Context, id string, actor string, ip string) error
	StopContainer(ctx context.Context, id string, actor string, ip string) error
	ListImages(ctx context.Context) ([]model.DockerImageInfo, error)
	ListVolumes(ctx context.Context) ([]model.DockerVolumeInfo, error)
	ListNetworks(ctx context.Context) ([]model.DockerNetworkInfo, error)
	GetSystemInfo(ctx context.Context) (model.DockerSystemInfo, error)
}

type dockerServiceImpl struct {
	client    integration.DockerClient
	auditRepo repository.AuditRepository
	redis     *redis.Client
	logger    *slog.Logger
}

// NewDockerService creates docker service instance
func NewDockerService(
	client integration.DockerClient,
	auditRepo repository.AuditRepository,
	redis *redis.Client,
	logger *slog.Logger,
) DockerService {
	return &dockerServiceImpl{
		client:    client,
		auditRepo: auditRepo,
		redis:     redis,
		logger:    logger,
	}
}

// ListContainers maps and filters containers
func (s *dockerServiceImpl) ListContainers(ctx context.Context, statusFilter string) ([]model.ContainerSummary, error) {
	raw, err := s.client.ListContainers(ctx, true)
	if err != nil {
		return nil, err
	}

	result := make([]model.ContainerSummary, 0, len(raw))
	for _, c := range raw {
		state := strings.ToLower(c.State)
		if statusFilter == "running" && state != "running" {
			continue
		}
		if statusFilter == "stopped" && state == "running" {
			continue
		}

		ports := make([]model.PortMapping, 0, len(c.Ports))
		for _, p := range c.Ports {
			ports = append(ports, model.PortMapping{
				IP:          p.IP,
				PrivatePort: p.PrivatePort,
				PublicPort:  p.PublicPort,
				Type:        p.Type,
			})
		}

		result = append(result, model.ContainerSummary{
			ID:      c.ID,
			Names:   c.Names,
			Image:   c.Image,
			ImageID: c.ImageID,
			Command: c.Command,
			Created: c.Created,
			State:   c.State,
			Status:  c.Status,
			Ports:   ports,
			Labels:  c.Labels,
		})
	}
	return result, nil
}

// GetContainer maps full container inspect
func (s *dockerServiceImpl) GetContainer(ctx context.Context, id string) (model.ContainerDetail, error) {
	c, err := s.client.GetContainer(ctx, id)
	if err != nil {
		return model.ContainerDetail{}, err
	}

	mounts := make([]model.MountInfo, 0, len(c.Mounts))
	for _, m := range c.Mounts {
		mounts = append(mounts, model.MountInfo{
			Type:        string(m.Type),
			Name:        m.Name,
			Source:      m.Source,
			Destination: m.Destination,
			Mode:        m.Mode,
			RW:          m.RW,
		})
	}

	ip := ""
	gw := ""
	mac := ""
	if c.NetworkSettings != nil {
		ip = c.NetworkSettings.IPAddress
		gw = c.NetworkSettings.Gateway
		mac = c.NetworkSettings.MacAddress
	}

	state := model.ContainerState{}
	if c.State != nil {
		state = model.ContainerState{
			Status:     c.State.Status,
			Running:    c.State.Running,
			Paused:     c.State.Paused,
			Restarting: c.State.Restarting,
			OOMKilled:  c.State.OOMKilled,
			Dead:       c.State.Dead,
			Pid:        c.State.Pid,
			ExitCode:   c.State.ExitCode,
			Error:      c.State.Error,
			StartedAt:  c.State.StartedAt,
			FinishedAt: c.State.FinishedAt,
		}
	}

	return model.ContainerDetail{
		ID:           c.ID,
		Name:         strings.TrimPrefix(c.Name, "/"),
		Image:        c.Config.Image,
		Created:      c.Created,
		State:        state,
		RestartCount: c.RestartCount,
		Platform:     c.Platform,
		Mounts:       mounts,
		IPAddress:    ip,
		Gateway:      gw,
		MacAddress:   mac,
		Labels:       c.Config.Labels,
	}, nil
}

// GetContainerStats parses raw docker JSON stats
func (s *dockerServiceImpl) GetContainerStats(ctx context.Context, id string) (model.ContainerStats, error) {
	statsResp, err := s.client.GetContainerStats(ctx, id)
	if err != nil {
		return model.ContainerStats{}, err
	}
	defer statsResp.Body.Close()

	var raw struct {
		Name     string `json:"name"`
		CPUStats struct {
			CPUUsage struct {
				TotalUsage  uint64   `json:"total_usage"`
				PercpuUsage []uint64 `json:"percpu_usage"`
			} `json:"cpu_usage"`
			SystemCPUUsage uint64 `json:"system_cpu_usage"`
			OnlineCPUs     uint32 `json:"online_cpus"`
		} `json:"cpu_stats"`
		PreCPUStats struct {
			CPUUsage struct {
				TotalUsage uint64 `json:"total_usage"`
			} `json:"cpu_usage"`
			SystemCPUUsage uint64 `json:"system_cpu_usage"`
		} `json:"precpu_stats"`
		MemoryStats struct {
			Usage uint64 `json:"usage"`
			Limit uint64 `json:"limit"`
		} `json:"memory_stats"`
		Networks map[string]struct {
			RxBytes uint64 `json:"rx_bytes"`
			TxBytes uint64 `json:"tx_bytes"`
		} `json:"networks"`
		PidsStats struct {
			Current uint64 `json:"current"`
		} `json:"pids_stats"`
	}

	if err := json.NewDecoder(statsResp.Body).Decode(&raw); err != nil {
		return model.ContainerStats{}, apperror.NewInternal(fmt.Sprintf("failed to parse stats JSON: %v", err))
	}

	// calculate CPU percentage
	cpuDelta := float64(raw.CPUStats.CPUUsage.TotalUsage) - float64(raw.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(raw.CPUStats.SystemCPUUsage) - float64(raw.PreCPUStats.SystemCPUUsage)
	cpuPercent := 0.0
	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpus := float64(len(raw.CPUStats.CPUUsage.PercpuUsage))
		if cpus == 0 && raw.CPUStats.OnlineCPUs > 0 {
			cpus = float64(raw.CPUStats.OnlineCPUs)
		}
		if cpus == 0 {
			cpus = 1.0
		}
		cpuPercent = (cpuDelta / systemDelta) * cpus * 100.0
	}

	// calculate memory percentage
	memPercent := 0.0
	if raw.MemoryStats.Limit > 0 {
		memPercent = (float64(raw.MemoryStats.Usage) / float64(raw.MemoryStats.Limit)) * 100.0
	}

	var rx, tx uint64
	for _, net := range raw.Networks {
		rx += net.RxBytes
		tx += net.TxBytes
	}

	return model.ContainerStats{
		ContainerID:      id,
		ContainerName:    strings.TrimPrefix(raw.Name, "/"),
		CPUPercentage:    cpuPercent,
		MemoryUsageBytes: raw.MemoryStats.Usage,
		MemoryLimitBytes: raw.MemoryStats.Limit,
		MemoryPercentage: memPercent,
		NetworkRxBytes:   rx,
		NetworkTxBytes:   tx,
		PidsCurrent:      raw.PidsStats.Current,
	}, nil
}

// GetContainerLogs demultiplexes docker log stream
func (s *dockerServiceImpl) GetContainerLogs(ctx context.Context, id string, tail int) (string, error) {
	reader, err := s.client.GetContainerLogs(ctx, id, tail)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	buf := new(bytes.Buffer)
	header := make([]byte, 8)

	for {
		n, err := io.ReadFull(reader, header)
		if err != nil {
			if n > 0 {
				buf.Write(header[:n])
			}
			break
		}

		// check for standard docker multiplexed frame
		if (header[0] == 1 || header[0] == 2) && header[1] == 0 && header[2] == 0 && header[3] == 0 {
			size := binary.BigEndian.Uint32(header[4:8])
			if size <= 1024*1024 {
				frame := make([]byte, size)
				_, err = io.ReadFull(reader, frame)
				if err != nil {
					break
				}
				buf.Write(frame)
				continue
			}
		}

		// raw non-multiplexed stream (TTY or mock)
		buf.Write(header)
		_, _ = io.Copy(buf, reader)
		break
	}

	return buf.String(), nil
}

// RestartContainer restarts and logs audit
func (s *dockerServiceImpl) RestartContainer(ctx context.Context, id string, actor string, ip string) error {
	if err := s.client.RestartContainer(ctx, id); err != nil {
		return err
	}

	details, _ := json.Marshal(map[string]string{"container_id": id})
	_ = s.auditRepo.Create(ctx, &model.AuditLog{
		Timestamp:    time.Now(),
		ActorType:    "user",
		ActorID:      actor,
		Action:       "restart_container",
		ResourceType: "docker",
		ResourceID:   &id,
		Details:      details,
		IPAddress:    &ip,
		Result:       "success",
	})
	return nil
}

// StopContainer stops and logs audit
func (s *dockerServiceImpl) StopContainer(ctx context.Context, id string, actor string, ip string) error {
	if err := s.client.StopContainer(ctx, id); err != nil {
		return err
	}

	details, _ := json.Marshal(map[string]string{"container_id": id})
	_ = s.auditRepo.Create(ctx, &model.AuditLog{
		Timestamp:    time.Now(),
		ActorType:    "user",
		ActorID:      actor,
		Action:       "stop_container",
		ResourceType: "docker",
		ResourceID:   &id,
		Details:      details,
		IPAddress:    &ip,
		Result:       "success",
	})
	return nil
}

// ListImages caches images in Redis
func (s *dockerServiceImpl) ListImages(ctx context.Context) ([]model.DockerImageInfo, error) {
	cacheKey := "cifo:docker:images"
	if s.redis != nil {
		val, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil && val != "" {
			var cached []model.DockerImageInfo
			if json.Unmarshal([]byte(val), &cached) == nil {
				return cached, nil
			}
		}
	}

	raw, err := s.client.ListImages(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.DockerImageInfo, 0, len(raw))
	for _, img := range raw {
		result = append(result, model.DockerImageInfo{
			ID:       img.ID,
			RepoTags: img.RepoTags,
			Size:     img.Size,
			Created:  img.Created,
		})
	}

	if s.redis != nil && len(result) > 0 {
		data, _ := json.Marshal(result)
		_ = s.redis.Set(ctx, cacheKey, data, 60*time.Second).Err()
	}

	return result, nil
}

// ListVolumes caches volumes in Redis
func (s *dockerServiceImpl) ListVolumes(ctx context.Context) ([]model.DockerVolumeInfo, error) {
	cacheKey := "cifo:docker:volumes"
	if s.redis != nil {
		val, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil && val != "" {
			var cached []model.DockerVolumeInfo
			if json.Unmarshal([]byte(val), &cached) == nil {
				return cached, nil
			}
		}
	}

	raw, err := s.client.ListVolumes(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.DockerVolumeInfo, 0, len(raw.Volumes))
	for _, v := range raw.Volumes {
		result = append(result, model.DockerVolumeInfo{
			Name:       v.Name,
			Driver:     v.Driver,
			Scope:      v.Scope,
			Mountpoint: v.Mountpoint,
			CreatedAt:  v.CreatedAt,
		})
	}

	if s.redis != nil && len(result) > 0 {
		data, _ := json.Marshal(result)
		_ = s.redis.Set(ctx, cacheKey, data, 60*time.Second).Err()
	}

	return result, nil
}

// ListNetworks maps network metadata
func (s *dockerServiceImpl) ListNetworks(ctx context.Context) ([]model.DockerNetworkInfo, error) {
	raw, err := s.client.ListNetworks(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.DockerNetworkInfo, 0, len(raw))
	for _, n := range raw {
		subnet := ""
		gw := ""
		if len(n.IPAM.Config) > 0 {
			subnet = n.IPAM.Config[0].Subnet
			gw = n.IPAM.Config[0].Gateway
		}

		result = append(result, model.DockerNetworkInfo{
			ID:       n.ID,
			Name:     n.Name,
			Driver:   n.Driver,
			Scope:    n.Scope,
			Internal: n.Internal,
			Subnet:   subnet,
			Gateway:  gw,
		})
	}
	return result, nil
}

// GetSystemInfo maps daemon system telemetry
func (s *dockerServiceImpl) GetSystemInfo(ctx context.Context) (model.DockerSystemInfo, error) {
	info, err := s.client.GetSystemInfo(ctx)
	if err != nil {
		return model.DockerSystemInfo{}, err
	}

	return model.DockerSystemInfo{
		ContainersTotal:   info.Containers,
		ContainersRunning: info.ContainersRunning,
		ContainersPaused:  info.ContainersPaused,
		ContainersStopped: info.ContainersStopped,
		ImagesTotal:       info.Images,
		NCPU:              info.NCPU,
		TotalMemory:       info.MemTotal,
		DockerVersion:     info.ServerVersion,
		KernelVersion:     info.KernelVersion,
		OperatingSystem:   info.OperatingSystem,
		OSType:            info.OSType,
		Architecture:      info.Architecture,
	}, nil
}

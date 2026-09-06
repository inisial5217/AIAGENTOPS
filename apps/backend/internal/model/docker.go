package model

// PortMapping container port binding
type PortMapping struct {
	IP          string `json:"ip,omitempty"`
	PrivatePort uint16 `json:"private_port"`
	PublicPort  uint16 `json:"public_port,omitempty"`
	Type        string `json:"type"`
}

// ContainerSummary high-level container overview
type ContainerSummary struct {
	ID      string            `json:"id"`
	Names   []string          `json:"names"`
	Image   string            `json:"image"`
	ImageID string            `json:"image_id"`
	Command string            `json:"command"`
	Created int64             `json:"created"`
	State   string            `json:"state"`
	Status  string            `json:"status"`
	Ports   []PortMapping     `json:"ports"`
	Labels  map[string]string `json:"labels"`
}

// ContainerState runtime state flags
type ContainerState struct {
	Status     string `json:"status"`
	Running    bool   `json:"running"`
	Paused     bool   `json:"paused"`
	Restarting bool   `json:"restarting"`
	OOMKilled  bool   `json:"oom_killed"`
	Dead       bool   `json:"dead"`
	Pid        int    `json:"pid"`
	ExitCode   int    `json:"exit_code"`
	Error      string `json:"error"`
	StartedAt  string `json:"started_at"`
	FinishedAt string `json:"finished_at"`
}

// MountInfo volume mount details
type MountInfo struct {
	Type        string `json:"type"`
	Name        string `json:"name,omitempty"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Mode        string `json:"mode"`
	RW          bool   `json:"rw"`
}

// ContainerDetail full inspect metadata
type ContainerDetail struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Image        string            `json:"image"`
	Created      string            `json:"created"`
	State        ContainerState    `json:"state"`
	RestartCount int               `json:"restart_count"`
	Platform     string            `json:"platform"`
	Mounts       []MountInfo       `json:"mounts"`
	IPAddress    string            `json:"ip_address"`
	Gateway      string            `json:"gateway"`
	MacAddress   string            `json:"mac_address"`
	Labels       map[string]string `json:"labels"`
	Env          []string          `json:"env,omitempty"`
}

// ContainerStats realtime resource metrics
type ContainerStats struct {
	ContainerID      string  `json:"container_id"`
	ContainerName    string  `json:"container_name"`
	CPUPercentage    float64 `json:"cpu_percentage"`
	MemoryUsageBytes uint64  `json:"memory_usage_bytes"`
	MemoryLimitBytes uint64  `json:"memory_limit_bytes"`
	MemoryPercentage float64 `json:"memory_percentage"`
	NetworkRxBytes   uint64  `json:"network_rx_bytes"`
	NetworkTxBytes   uint64  `json:"network_tx_bytes"`
	BlockReadBytes   uint64  `json:"block_read_bytes"`
	BlockWriteBytes  uint64  `json:"block_write_bytes"`
	PidsCurrent      uint64  `json:"pids_current"`
}

// DockerSystemInfo host and daemon telemetry
type DockerSystemInfo struct {
	ContainersTotal   int    `json:"containers_total"`
	ContainersRunning int    `json:"containers_running"`
	ContainersPaused  int    `json:"containers_paused"`
	ContainersStopped int    `json:"containers_stopped"`
	ImagesTotal       int    `json:"images_total"`
	NCPU              int    `json:"ncpu"`
	TotalMemory       int64  `json:"total_memory"`
	DockerVersion     string `json:"docker_version"`
	KernelVersion     string `json:"kernel_version"`
	OperatingSystem   string `json:"operating_system"`
	OSType            string `json:"os_type"`
	Architecture      string `json:"architecture"`
}

// DockerImageInfo cached image metadata
type DockerImageInfo struct {
	ID       string   `json:"id"`
	RepoTags []string `json:"repo_tags"`
	Size     int64    `json:"size"`
	Created  int64    `json:"created"`
}

// DockerVolumeInfo volume item metadata
type DockerVolumeInfo struct {
	Name       string `json:"name"`
	Driver     string `json:"driver"`
	Scope      string `json:"scope"`
	Mountpoint string `json:"mountpoint"`
	CreatedAt  string `json:"created_at,omitempty"`
}

// DockerNetworkInfo network bridge/overlay metadata
type DockerNetworkInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Driver   string `json:"driver"`
	Scope    string `json:"scope"`
	Internal bool   `json:"internal"`
	Subnet   string `json:"subnet,omitempty"`
	Gateway  string `json:"gateway,omitempty"`
}

// DashboardStats aggregated Tier 1 KPI telemetry
type DashboardStats struct {
	TotalContainers   int     `json:"total_containers"`
	TotalReplicas     int     `json:"total_replicas"`
	OverallRAMPercent float64 `json:"overall_ram_percent"`
	UsedRAMBytes      uint64  `json:"used_ram_bytes"`
	TotalRAMBytes     uint64  `json:"total_ram_bytes"`
	ContainersOn      int     `json:"containers_on"`
	ContainersOff     int     `json:"containers_off"`
	ActiveIncidents   int     `json:"active_incidents"`
}

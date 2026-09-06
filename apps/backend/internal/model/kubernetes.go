package model

// ContainerInfo pod container status
type ContainerInfo struct {
	Name         string   `json:"name"`
	Image        string   `json:"image"`
	Ready        bool     `json:"ready"`
	RestartCount int32    `json:"restart_count"`
	State        string   `json:"state"`
	Ports        []string `json:"ports,omitempty"`
}

// PodSummary high-level pod overview
type PodSummary struct {
	Name         string            `json:"name"`
	Namespace    string            `json:"namespace"`
	Status       string            `json:"status"`
	Phase        string            `json:"phase"`
	Restarts     int32             `json:"restarts"`
	CPURequest   string            `json:"cpu_request,omitempty"`
	MemoryRequest string           `json:"memory_request,omitempty"`
	CPULimit     string            `json:"cpu_limit,omitempty"`
	MemoryLimit  string            `json:"memory_limit,omitempty"`
	Node         string            `json:"node"`
	IP           string            `json:"ip"`
	Age          string            `json:"age"`
	CreatedAt    int64             `json:"created_at"`
	Labels       map[string]string `json:"labels,omitempty"`
}

// PodDetail full pod inspect
type PodDetail struct {
	PodSummary
	QoSClass   string          `json:"qos_class"`
	StartTime  string          `json:"start_time,omitempty"`
	Containers []ContainerInfo `json:"containers"`
	Conditions []string        `json:"conditions,omitempty"`
	NodeIP     string          `json:"node_ip,omitempty"`
}

// DeploymentSummary high-level deployment overview
type DeploymentSummary struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Replicas          int32             `json:"replicas"`
	ReadyReplicas     int32             `json:"ready_replicas"`
	AvailableReplicas int32             `json:"available_replicas"`
	UpdatedReplicas   int32             `json:"updated_replicas"`
	Images            []string          `json:"images"`
	Age               string            `json:"age"`
	CreatedAt         int64             `json:"created_at"`
	Labels            map[string]string `json:"labels,omitempty"`
}

// DeploymentDetail full deployment inspect
type DeploymentDetail struct {
	DeploymentSummary
	Strategy   string            `json:"strategy"`
	Selector   map[string]string `json:"selector,omitempty"`
	Conditions []string          `json:"conditions,omitempty"`
}

// NodeSummary high-level node overview
type NodeSummary struct {
	Name                string `json:"name"`
	Status              string `json:"status"`
	Roles               string `json:"roles"`
	Version             string `json:"version"`
	InternalIP          string `json:"internal_ip"`
	CPUCapacity         string `json:"cpu_capacity"`
	MemoryCapacityBytes int64  `json:"memory_capacity_bytes"`
	PodCount            int    `json:"pod_count"`
	OSImage             string `json:"os_image"`
	KernelVersion       string `json:"kernel_version"`
	ContainerRuntime    string `json:"container_runtime"`
}

// ServicePortMapping service port definition
type ServicePortMapping struct {
	Name       string `json:"name,omitempty"`
	Port       int32  `json:"port"`
	TargetPort string `json:"target_port"`
	Protocol   string `json:"protocol"`
	NodePort   int32  `json:"node_port,omitempty"`
}

// ServiceSummary high-level service overview
type ServiceSummary struct {
	Name       string               `json:"name"`
	Namespace  string               `json:"namespace"`
	Type       string               `json:"type"`
	ClusterIP  string               `json:"cluster_ip"`
	ExternalIP string               `json:"external_ip,omitempty"`
	Ports      []ServicePortMapping `json:"ports"`
	Selector   map[string]string    `json:"selector,omitempty"`
	Age        string               `json:"age"`
	CreatedAt  int64                `json:"created_at"`
}

// ScaleDeploymentRequest scale payload
type ScaleDeploymentRequest struct {
	Replicas int32 `json:"replicas" validate:"required,min=0,max=100"`
}

// K8sClusterOverview cluster level stats
type K8sClusterOverview struct {
	TotalNodes       int `json:"total_nodes"`
	ReadyNodes       int `json:"ready_nodes"`
	TotalPods        int `json:"total_pods"`
	RunningPods      int `json:"running_pods"`
	TotalDeployments int `json:"total_deployments"`
	ReadyDeployments int `json:"ready_deployments"`
	TotalServices    int `json:"total_services"`
}


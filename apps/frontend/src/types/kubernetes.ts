// Kubernetes domain types matching backend API

export interface ContainerInfo {
  name: string;
  image: string;
  ready: boolean;
  restart_count: number;
  state: string;
  ports?: string[];
}

export interface PodSummary {
  name: string;
  namespace: string;
  status: string;
  phase: string;
  restarts: number;
  cpu_request?: string;
  memory_request?: string;
  cpu_limit?: string;
  memory_limit?: string;
  node: string;
  ip: string;
  age: string;
  created_at: number;
  labels?: Record<string, string>;
}

export interface PodDetail extends PodSummary {
  qos_class: string;
  start_time?: string;
  containers: ContainerInfo[];
  conditions?: string[];
  node_ip?: string;
}

export interface DeploymentSummary {
  name: string;
  namespace: string;
  replicas: number;
  ready_replicas: number;
  available_replicas: number;
  updated_replicas: number;
  images: string[];
  age: string;
  created_at: number;
  labels?: Record<string, string>;
}

export interface DeploymentDetail extends DeploymentSummary {
  strategy: string;
  selector?: Record<string, string>;
  conditions?: string[];
}

export interface NodeSummary {
  name: string;
  status: string;
  roles: string;
  version: string;
  internal_ip: string;
  cpu_capacity: string;
  memory_capacity_bytes: number;
  pod_count: number;
  os_image: string;
  kernel_version: string;
  container_runtime: string;
}

export interface ServicePortMapping {
  name?: string;
  port: number;
  target_port: string;
  protocol: string;
  node_port?: number;
}

export interface ServiceSummary {
  name: string;
  namespace: string;
  type: string;
  cluster_ip: string;
  external_ip?: string;
  ports: ServicePortMapping[];
  selector?: Record<string, string>;
  age: string;
  created_at: number;
}

export interface K8sClusterOverview {
  total_nodes: number;
  ready_nodes: number;
  total_pods: number;
  running_pods: number;
  total_deployments: number;
  ready_deployments: number;
  total_services: number;
}

export interface ScaleDeploymentRequest {
  replicas: number;
}

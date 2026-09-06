export interface PortMapping {
  ip?: string;
  private_port: number;
  public_port?: number;
  type: string;
}

export interface ContainerSummary {
  id: string;
  names: string[];
  image: string;
  image_id: string;
  command: string;
  created: number;
  state: "running" | "exited" | "paused" | "restarting" | string;
  status: string;
  ports: PortMapping[];
  labels: Record<string, string>;
}

export interface ContainerState {
  status: string;
  running: boolean;
  paused: boolean;
  restarting: boolean;
  oom_killed: boolean;
  dead: boolean;
  pid: number;
  exit_code: number;
  error: string;
  started_at: string;
  finished_at: string;
}

export interface MountInfo {
  type: string;
  name?: string;
  source: string;
  destination: string;
  mode: string;
  rw: boolean;
}

export interface ContainerDetail {
  id: string;
  name: string;
  image: string;
  created: string;
  state: ContainerState;
  restart_count: number;
  platform: string;
  mounts: MountInfo[];
  ip_address: string;
  gateway: string;
  mac_address: string;
  labels: Record<string, string>;
  env?: string[];
}

export interface ContainerStats {
  container_id: string;
  container_name: string;
  cpu_percentage: number;
  memory_usage_bytes: number;
  memory_limit_bytes: number;
  memory_percentage: number;
  network_rx_bytes: number;
  network_tx_bytes: number;
  block_read_bytes: number;
  block_write_bytes: number;
  pids_current: number;
}

export interface DockerSystemInfo {
  containers_total: number;
  containers_running: number;
  containers_paused: number;
  containers_stopped: number;
  images_total: number;
  ncpu: number;
  total_memory: number;
  docker_version: string;
  kernel_version: string;
  operating_system: string;
  os_type: string;
  architecture: string;
}

export interface DockerImageInfo {
  id: string;
  repo_tags: string[];
  size: number;
  created: number;
}

export interface DockerVolumeInfo {
  name: string;
  driver: string;
  scope: string;
  mountpoint: string;
  created_at?: string;
}

export interface DockerNetworkInfo {
  id: string;
  name: string;
  driver: string;
  scope: string;
  internal: boolean;
  subnet?: string;
  gateway?: string;
}

export interface DashboardStats {
  total_containers: number;
  total_replicas: number;
  overall_ram_percent: number;
  used_ram_bytes: number;
  total_ram_bytes: number;
  containers_on: number;
  containers_off: number;
  active_incidents: number;
}

export interface TimeSeriesPoint {
  timestamp: string;
  value: number;
  secondary?: number;
}

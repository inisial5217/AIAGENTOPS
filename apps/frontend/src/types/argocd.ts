// ArgoCD domain types matching backend API

export interface ArgoResourceStatus {
  group?: string;
  version: string;
  kind: string;
  namespace: string;
  name: string;
  status: string;
  health?: string;
  hook?: string;
}

export interface ArgoDeploymentRevision {
  id: number;
  revision: string;
  source: string;
  deployed_at: string;
}

export interface ArgoApplicationSummary {
  name: string;
  project: string;
  repo_url: string;
  path: string;
  target_revision: string;
  destination_server: string;
  destination_namespace: string;
  sync_status: string;
  health_status: string;
  sync_message?: string;
  health_message?: string;
  images?: string[];
  created_at?: string;
}

export interface ArgoApplicationDetail extends ArgoApplicationSummary {
  automated_sync: boolean;
  prune: boolean;
  self_heal: boolean;
  resources: ArgoResourceStatus[];
  history: ArgoDeploymentRevision[];
}

export interface ArgoSyncRequest {
  prune: boolean;
  dry_run: boolean;
}

export interface ArgoOverview {
  total: number;
  synced: number;
  out_of_sync: number;
  healthy: number;
  degraded: number;
  progressing: number;
  unknown: number;
}

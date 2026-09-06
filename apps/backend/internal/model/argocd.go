package model

// ArgoResourceStatus managed resource status
type ArgoResourceStatus struct {
	Group     string `json:"group,omitempty"`
	Version   string `json:"version"`
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Health    string `json:"health,omitempty"`
	Hook      string `json:"hook,omitempty"`
}

// ArgoDeploymentRevision deployment history revision
type ArgoDeploymentRevision struct {
	ID         int64  `json:"id"`
	Revision   string `json:"revision"`
	Source     string `json:"source"`
	DeployedAt string `json:"deployed_at"`
}

// ArgoApplicationSummary overview of ArgoCD application
type ArgoApplicationSummary struct {
	Name                 string   `json:"name"`
	Project              string   `json:"project"`
	RepoURL              string   `json:"repo_url"`
	Path                 string   `json:"path"`
	TargetRevision       string   `json:"target_revision"`
	DestinationServer    string   `json:"destination_server"`
	DestinationNamespace string   `json:"destination_namespace"`
	SyncStatus           string   `json:"sync_status"`
	HealthStatus         string   `json:"health_status"`
	SyncMessage          string   `json:"sync_message,omitempty"`
	HealthMessage        string   `json:"health_message,omitempty"`
	Images               []string `json:"images,omitempty"`
	CreatedAt            string   `json:"created_at,omitempty"`
}

// ArgoApplicationDetail full application status
type ArgoApplicationDetail struct {
	ArgoApplicationSummary
	AutomatedSync bool                     `json:"automated_sync"`
	Prune         bool                     `json:"prune"`
	SelfHeal      bool                     `json:"self_heal"`
	Resources     []ArgoResourceStatus     `json:"resources"`
	History       []ArgoDeploymentRevision `json:"history"`
}

// ArgoSyncRequest sync trigger options
type ArgoSyncRequest struct {
	Prune  bool `json:"prune"`
	DryRun bool `json:"dry_run"`
}

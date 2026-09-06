package model

import (
	"time"
)

// Incident status constants
const (
	IncidentStatusOpen          = "open"
	IncidentStatusAcknowledged  = "acknowledged"
	IncidentStatusInvestigating = "investigating"
	IncidentStatusResolved      = "resolved"
	IncidentStatusClosed        = "closed"
)

// Incident severity constants
const (
	SeverityCritical = "critical"
	SeverityWarning  = "warning"
	SeverityInfo     = "info"
)

// Incident source constants
const (
	SourceAlertmanager = "alertmanager"
	SourceDocker       = "docker"
	SourceKubernetes   = "kubernetes"
	SourceManual       = "manual"
)

// Incident database entity
type Incident struct {
	ID             string     `json:"id" db:"id"`
	Title          string     `json:"title" db:"title" validate:"required"`
	Description    string     `json:"description" db:"description"`
	Severity       string     `json:"severity" db:"severity" validate:"required"`
	Status         string     `json:"status" db:"status"`
	Source         string     `json:"source" db:"source" validate:"required"`
	AlertName      string     `json:"alert_name" db:"alert_name"`
	ResourceType   string     `json:"resource_type" db:"resource_type"`
	ResourceID     string     `json:"resource_id" db:"resource_id"`
	Namespace      string     `json:"namespace" db:"namespace"`
	RCASummary     string     `json:"rca_summary" db:"rca_summary"`
	AcknowledgedBy *string    `json:"acknowledged_by" db:"acknowledged_by"`
	AcknowledgedAt *time.Time `json:"acknowledged_at" db:"acknowledged_at"`
	ResolvedBy     *string    `json:"resolved_by" db:"resolved_by"`
	ResolvedAt     *time.Time `json:"resolved_at" db:"resolved_at"`
	ClosedBy       *string    `json:"closed_by" db:"closed_by"`
	ClosedAt       *time.Time `json:"closed_at" db:"closed_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// IncidentSummary list item
type IncidentSummary struct {
	ID                 string     `json:"id"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	Severity           string     `json:"severity"`
	Status             string     `json:"status"`
	Source             string     `json:"source"`
	AlertName          string     `json:"alert_name"`
	ResourceType       string     `json:"resource_type"`
	ResourceID         string     `json:"resource_id"`
	Namespace          string     `json:"namespace"`
	AcknowledgedBy     *string    `json:"acknowledged_by"`
	AcknowledgedByName *string    `json:"acknowledged_by_name"`
	AcknowledgedAt     *time.Time `json:"acknowledged_at"`
	ResolvedBy         *string    `json:"resolved_by"`
	ResolvedByName     *string    `json:"resolved_by_name"`
	ResolvedAt         *time.Time `json:"resolved_at"`
	ClosedBy           *string    `json:"closed_by"`
	ClosedByName       *string    `json:"closed_by_name"`
	ClosedAt           *time.Time `json:"closed_at"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// IncidentDetail detailed view
type IncidentDetail struct {
	IncidentSummary
	RCASummary    string               `json:"rca_summary"`
	Labels        map[string]string    `json:"labels,omitempty"`
	Annotations   map[string]string    `json:"annotations,omitempty"`
	Notifications []NotificationRecord `json:"notifications,omitempty"`
}

// IncidentFilter query filter
type IncidentFilter struct {
	Status   string `json:"status" query:"status"`
	Severity string `json:"severity" query:"severity"`
	Source   string `json:"source" query:"source"`
	Search   string `json:"search" query:"search"`
	Page     int    `json:"page" query:"page"`
	Limit    int    `json:"limit" query:"limit"`
}

// IncidentStats aggregated counts
type IncidentStats struct {
	Total         int `json:"total"`
	Open          int `json:"open"`
	Acknowledged  int `json:"acknowledged"`
	Investigating int `json:"investigating"`
	Resolved      int `json:"resolved"`
	Closed        int `json:"closed"`
	CriticalCount int `json:"critical_count"`
}

// CreateIncidentRequest payload
type CreateIncidentRequest struct {
	Title        string `json:"title" validate:"required"`
	Description  string `json:"description"`
	Severity     string `json:"severity" validate:"required"`
	Source       string `json:"source" validate:"required"`
	AlertName    string `json:"alert_name"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	Namespace    string `json:"namespace"`
}

// AlertmanagerWebhookPayload webhook format
type AlertmanagerWebhookPayload struct {
	Version           string                 `json:"version"`
	GroupKey          string                 `json:"groupKey"`
	TruncatedAlerts   int                    `json:"truncatedAlerts"`
	Status            string                 `json:"status"` // "firing" or "resolved"
	Receiver          string                 `json:"receiver"`
	GroupLabels       map[string]string      `json:"groupLabels"`
	CommonLabels      map[string]string      `json:"commonLabels"`
	CommonAnnotations map[string]string      `json:"commonAnnotations"`
	ExternalURL       string                 `json:"externalURL"`
	Alerts            []AlertmanagerAlert    `json:"alerts"`
}

// AlertmanagerAlert alert item
type AlertmanagerAlert struct {
	Status       string            `json:"status"` // "firing" or "resolved"
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

// NotificationRecord database entity
type NotificationRecord struct {
	ID           string    `json:"id" db:"id"`
	IncidentID   *string   `json:"incident_id" db:"incident_id"`
	Channel      string    `json:"channel" db:"channel"` // "telegram", "inapp"
	Recipient    string    `json:"recipient" db:"recipient"`
	Title        string    `json:"title" db:"title"`
	Message      string    `json:"message" db:"message"`
	Severity     string    `json:"severity" db:"severity"`
	Status       string    `json:"status" db:"status"` // "sent", "failed", "queued"
	ErrorMessage string    `json:"error_message,omitempty" db:"error_message"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

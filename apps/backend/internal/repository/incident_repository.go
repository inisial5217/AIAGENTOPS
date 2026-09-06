package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// IncidentRepository interface definition
type IncidentRepository interface {
	Create(ctx context.Context, incident *model.Incident) error
	GetByID(ctx context.Context, id string) (*model.Incident, error)
	GetDetailByID(ctx context.Context, id string) (*model.IncidentDetail, error)
	List(ctx context.Context, filter model.IncidentFilter) ([]*model.IncidentSummary, int, error)
	GetStats(ctx context.Context) (*model.IncidentStats, error)
	FindOpenByAlertAndResource(ctx context.Context, alertName, resourceID string) (*model.Incident, error)
	UpdateStatus(ctx context.Context, id string, status string, actorID *string) error
	GetUnacknowledgedOlderThan(ctx context.Context, duration time.Duration) ([]*model.Incident, error)
	SaveNotification(ctx context.Context, notif *model.NotificationRecord) error
	ListNotificationsByIncidentID(ctx context.Context, incidentID string) ([]model.NotificationRecord, error)
}

// PostgresIncidentRepository pgx implementation
type PostgresIncidentRepository struct {
	pool *pgxpool.Pool
}

// NewIncidentRepository creates repository
func NewIncidentRepository(pool *pgxpool.Pool) *PostgresIncidentRepository {
	return &PostgresIncidentRepository{pool: pool}
}

// Create inserts new incident
func (r *PostgresIncidentRepository) Create(ctx context.Context, inc *model.Incident) error {
	query := `
		INSERT INTO incidents (
			title, description, severity, status, source,
			alert_name, resource_type, resource_id, namespace, rca_summary,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	if inc.Status == "" {
		inc.Status = model.IncidentStatusOpen
	}

	err := r.pool.QueryRow(ctx, query,
		inc.Title,
		inc.Description,
		inc.Severity,
		inc.Status,
		inc.Source,
		inc.AlertName,
		inc.ResourceType,
		inc.ResourceID,
		inc.Namespace,
		inc.RCASummary,
	).Scan(&inc.ID, &inc.CreatedAt, &inc.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create incident: %w", err)
	}

	return nil
}

// GetByID gets incident by id
func (r *PostgresIncidentRepository) GetByID(ctx context.Context, id string) (*model.Incident, error) {
	query := `
		SELECT id, title, description, severity, status, source,
		       alert_name, resource_type, resource_id, namespace, rca_summary,
		       acknowledged_by, acknowledged_at, resolved_by, resolved_at,
		       closed_by, closed_at, created_at, updated_at
		FROM incidents
		WHERE id = $1`

	var inc model.Incident
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&inc.ID, &inc.Title, &inc.Description, &inc.Severity, &inc.Status, &inc.Source,
		&inc.AlertName, &inc.ResourceType, &inc.ResourceID, &inc.Namespace, &inc.RCASummary,
		&inc.AcknowledgedBy, &inc.AcknowledgedAt, &inc.ResolvedBy, &inc.ResolvedAt,
		&inc.ClosedBy, &inc.ClosedAt, &inc.CreatedAt, &inc.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get incident: %w", err)
	}

	return &inc, nil
}

// GetDetailByID gets detailed incident
func (r *PostgresIncidentRepository) GetDetailByID(ctx context.Context, id string) (*model.IncidentDetail, error) {
	query := `
		SELECT i.id, i.title, i.description, i.severity, i.status, i.source,
		       i.alert_name, i.resource_type, i.resource_id, i.namespace, i.rca_summary,
		       i.acknowledged_by, u1.name AS acknowledged_by_name, i.acknowledged_at,
		       i.resolved_by, u2.name AS resolved_by_name, i.resolved_at,
		       i.closed_by, u3.name AS closed_by_name, i.closed_at,
		       i.created_at, i.updated_at
		FROM incidents i
		LEFT JOIN users u1 ON i.acknowledged_by = u1.id
		LEFT JOIN users u2 ON i.resolved_by = u2.id
		LEFT JOIN users u3 ON i.closed_by = u3.id
		WHERE i.id = $1`

	var detail model.IncidentDetail
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&detail.ID, &detail.Title, &detail.Description, &detail.Severity, &detail.Status, &detail.Source,
		&detail.AlertName, &detail.ResourceType, &detail.ResourceID, &detail.Namespace, &detail.RCASummary,
		&detail.AcknowledgedBy, &detail.AcknowledgedByName, &detail.AcknowledgedAt,
		&detail.ResolvedBy, &detail.ResolvedByName, &detail.ResolvedAt,
		&detail.ClosedBy, &detail.ClosedByName, &detail.ClosedAt,
		&detail.CreatedAt, &detail.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get incident detail: %w", err)
	}

	// fetch related notifications
	notifs, err := r.ListNotificationsByIncidentID(ctx, id)
	if err == nil {
		detail.Notifications = notifs
	}

	return &detail, nil
}

// List lists incidents with filter
func (r *PostgresIncidentRepository) List(ctx context.Context, f model.IncidentFilter) ([]*model.IncidentSummary, int, error) {
	whereClauses := []string{"1=1"}
	args := []interface{}{}
	argIdx := 1

	if f.Status != "" && f.Status != "all" {
		whereClauses = append(whereClauses, fmt.Sprintf("i.status = $%d", argIdx))
		args = append(args, f.Status)
		argIdx++
	}

	if f.Severity != "" && f.Severity != "all" {
		whereClauses = append(whereClauses, fmt.Sprintf("i.severity = $%d", argIdx))
		args = append(args, f.Severity)
		argIdx++
	}

	if f.Source != "" && f.Source != "all" {
		whereClauses = append(whereClauses, fmt.Sprintf("i.source = $%d", argIdx))
		args = append(args, f.Source)
		argIdx++
	}

	if f.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(i.title ILIKE $%d OR i.resource_id ILIKE $%d OR i.alert_name ILIKE $%d)", argIdx, argIdx, argIdx))
		args = append(args, "%"+f.Search+"%")
		argIdx++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	// count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM incidents i WHERE %s", whereSQL)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count incidents: %w", err)
	}

	// pagination
	page := f.Page
	if page < 1 {
		page = 1
	}
	limit := f.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	query := fmt.Sprintf(`
		SELECT i.id, i.title, i.description, i.severity, i.status, i.source,
		       i.alert_name, i.resource_type, i.resource_id, i.namespace,
		       i.acknowledged_by, u1.name AS acknowledged_by_name, i.acknowledged_at,
		       i.resolved_by, u2.name AS resolved_by_name, i.resolved_at,
		       i.closed_by, u3.name AS closed_by_name, i.closed_at,
		       i.created_at, i.updated_at
		FROM incidents i
		LEFT JOIN users u1 ON i.acknowledged_by = u1.id
		LEFT JOIN users u2 ON i.resolved_by = u2.id
		LEFT JOIN users u3 ON i.closed_by = u3.id
		WHERE %s
		ORDER BY i.created_at DESC
		LIMIT $%d OFFSET $%d`, whereSQL, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list incidents: %w", err)
	}
	defer rows.Close()

	var list []*model.IncidentSummary
	for rows.Next() {
		var s model.IncidentSummary
		err := rows.Scan(
			&s.ID, &s.Title, &s.Description, &s.Severity, &s.Status, &s.Source,
			&s.AlertName, &s.ResourceType, &s.ResourceID, &s.Namespace,
			&s.AcknowledgedBy, &s.AcknowledgedByName, &s.AcknowledgedAt,
			&s.ResolvedBy, &s.ResolvedByName, &s.ResolvedAt,
			&s.ClosedBy, &s.ClosedByName, &s.ClosedAt,
			&s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan incident: %w", err)
		}
		list = append(list, &s)
	}

	return list, total, nil
}

// GetStats returns aggregated stats
func (r *PostgresIncidentRepository) GetStats(ctx context.Context) (*model.IncidentStats, error) {
	query := `
		SELECT 
			COUNT(*) AS total,
			COUNT(CASE WHEN status = 'open' THEN 1 END) AS open,
			COUNT(CASE WHEN status = 'acknowledged' THEN 1 END) AS acknowledged,
			COUNT(CASE WHEN status = 'investigating' THEN 1 END) AS investigating,
			COUNT(CASE WHEN status = 'resolved' THEN 1 END) AS resolved,
			COUNT(CASE WHEN status = 'closed' THEN 1 END) AS closed,
			COUNT(CASE WHEN severity = 'critical' AND status != 'closed' THEN 1 END) AS critical_count
		FROM incidents`

	var s model.IncidentStats
	err := r.pool.QueryRow(ctx, query).Scan(
		&s.Total,
		&s.Open,
		&s.Acknowledged,
		&s.Investigating,
		&s.Resolved,
		&s.Closed,
		&s.CriticalCount,
	)
	if err != nil {
		return nil, fmt.Errorf("get incident stats: %w", err)
	}

	return &s, nil
}

// FindOpenByAlertAndResource finds active incident
func (r *PostgresIncidentRepository) FindOpenByAlertAndResource(ctx context.Context, alertName, resourceID string) (*model.Incident, error) {
	query := `
		SELECT id, title, description, severity, status, source,
		       alert_name, resource_type, resource_id, namespace, rca_summary,
		       acknowledged_by, acknowledged_at, resolved_by, resolved_at,
		       closed_by, closed_at, created_at, updated_at
		FROM incidents
		WHERE alert_name = $1 
		  AND resource_id = $2 
		  AND status IN ('open', 'acknowledged', 'investigating')
		ORDER BY created_at DESC
		LIMIT 1`

	var inc model.Incident
	err := r.pool.QueryRow(ctx, query, alertName, resourceID).Scan(
		&inc.ID, &inc.Title, &inc.Description, &inc.Severity, &inc.Status, &inc.Source,
		&inc.AlertName, &inc.ResourceType, &inc.ResourceID, &inc.Namespace, &inc.RCASummary,
		&inc.AcknowledgedBy, &inc.AcknowledgedAt, &inc.ResolvedBy, &inc.ResolvedAt,
		&inc.ClosedBy, &inc.ClosedAt, &inc.CreatedAt, &inc.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find open incident: %w", err)
	}

	return &inc, nil
}

// UpdateStatus updates lifecycle status
func (r *PostgresIncidentRepository) UpdateStatus(ctx context.Context, id string, status string, actorID *string) error {
	var query string
	now := time.Now()

	switch status {
	case model.IncidentStatusAcknowledged:
		query = `
			UPDATE incidents 
			SET status = $2, acknowledged_by = $3, acknowledged_at = $4, updated_at = NOW()
			WHERE id = $1`
		_, err := r.pool.Exec(ctx, query, id, status, actorID, now)
		return err
	case model.IncidentStatusResolved:
		query = `
			UPDATE incidents 
			SET status = $2, resolved_by = $3, resolved_at = $4, updated_at = NOW()
			WHERE id = $1`
		_, err := r.pool.Exec(ctx, query, id, status, actorID, now)
		return err
	case model.IncidentStatusClosed:
		query = `
			UPDATE incidents 
			SET status = $2, closed_by = $3, closed_at = $4, updated_at = NOW()
			WHERE id = $1`
		_, err := r.pool.Exec(ctx, query, id, status, actorID, now)
		return err
	default:
		query = `UPDATE incidents SET status = $2, updated_at = NOW() WHERE id = $1`
		_, err := r.pool.Exec(ctx, query, id, status)
		return err
	}
}

// GetUnacknowledgedOlderThan finds unacknowledged incidents
func (r *PostgresIncidentRepository) GetUnacknowledgedOlderThan(ctx context.Context, d time.Duration) ([]*model.Incident, error) {
	cutoff := time.Now().Add(-d)
	query := `
		SELECT id, title, description, severity, status, source,
		       alert_name, resource_type, resource_id, namespace, rca_summary,
		       acknowledged_by, acknowledged_at, resolved_by, resolved_at,
		       closed_by, closed_at, created_at, updated_at
		FROM incidents
		WHERE status = 'open' 
		  AND created_at <= $1
		ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, query, cutoff)
	if err != nil {
		return nil, fmt.Errorf("get unacknowledged incidents: %w", err)
	}
	defer rows.Close()

	var list []*model.Incident
	for rows.Next() {
		var inc model.Incident
		err := rows.Scan(
			&inc.ID, &inc.Title, &inc.Description, &inc.Severity, &inc.Status, &inc.Source,
			&inc.AlertName, &inc.ResourceType, &inc.ResourceID, &inc.Namespace, &inc.RCASummary,
			&inc.AcknowledgedBy, &inc.AcknowledgedAt, &inc.ResolvedBy, &inc.ResolvedAt,
			&inc.ClosedBy, &inc.ClosedAt, &inc.CreatedAt, &inc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan incident: %w", err)
		}
		list = append(list, &inc)
	}

	return list, nil
}

// SaveNotification stores notification record
func (r *PostgresIncidentRepository) SaveNotification(ctx context.Context, n *model.NotificationRecord) error {
	query := `
		INSERT INTO notifications (
			incident_id, channel, recipient, title, message, severity, status, error_message, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, query,
		n.IncidentID,
		n.Channel,
		n.Recipient,
		n.Title,
		n.Message,
		n.Severity,
		n.Status,
		n.ErrorMessage,
	).Scan(&n.ID, &n.CreatedAt)
	if err != nil {
		return fmt.Errorf("save notification: %w", err)
	}

	return nil
}

// ListNotificationsByIncidentID lists incident notifications
func (r *PostgresIncidentRepository) ListNotificationsByIncidentID(ctx context.Context, incidentID string) ([]model.NotificationRecord, error) {
	query := `
		SELECT id, incident_id, channel, recipient, title, message, severity, status, COALESCE(error_message, ''), created_at
		FROM notifications
		WHERE incident_id = $1
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, incidentID)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	var notifs []model.NotificationRecord
	for rows.Next() {
		var n model.NotificationRecord
		if err := rows.Scan(&n.ID, &n.IncidentID, &n.Channel, &n.Recipient, &n.Title, &n.Message, &n.Severity, &n.Status, &n.ErrorMessage, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		notifs = append(notifs, n)
	}

	return notifs, nil
}

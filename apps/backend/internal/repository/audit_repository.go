package repository

import (
	"context"
	"fmt"

	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AuditRepository audit log interface
type AuditRepository interface {
	Create(ctx context.Context, log *model.AuditLog) error
	List(ctx context.Context, limit, offset int) ([]*model.AuditLog, int, error)
}

// PostgresAuditRepository pgx implementation
type PostgresAuditRepository struct {
	pool *pgxpool.Pool
}

// NewAuditRepository creates audit repository
func NewAuditRepository(pool *pgxpool.Pool) *PostgresAuditRepository {
	return &PostgresAuditRepository{pool: pool}
}

// Create inserts audit log entry
func (r *PostgresAuditRepository) Create(ctx context.Context, log *model.AuditLog) error {
	query := `
		INSERT INTO audit_log (
			actor_type, actor_id, action, resource_type, resource_id,
			details, ip_address, user_agent, result, timestamp
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7::inet, $8, $9, COALESCE(NULLIF($10, '0001-01-01T00:00:00Z'::timestamptz), NOW()))
		RETURNING id, timestamp`

	var detailsArg []byte
	if len(log.Details) > 0 {
		detailsArg = log.Details
	}

	// handle ip_address formatting
	var ipArg interface{}
	if log.IPAddress != nil && *log.IPAddress != "" && *log.IPAddress != "::1" && *log.IPAddress != "127.0.0.1" {
		ipArg = *log.IPAddress
	} else if log.IPAddress != nil && (*log.IPAddress == "127.0.0.1" || *log.IPAddress == "::1") {
		ipArg = "127.0.0.1"
	}

	queryFallback := `
		INSERT INTO audit_log (
			actor_type, actor_id, action, resource_type, resource_id,
			details, user_agent, result
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, timestamp`

	if ipArg != nil {
		err := r.pool.QueryRow(ctx, query,
			log.ActorType, log.ActorID, log.Action, log.ResourceType, log.ResourceID,
			detailsArg, ipArg, log.UserAgent, log.Result, log.Timestamp,
		).Scan(&log.ID, &log.Timestamp)
		if err != nil {
			return fmt.Errorf("insert audit log with ip: %w", err)
		}
		return nil
	}

	err := r.pool.QueryRow(ctx, queryFallback,
		log.ActorType, log.ActorID, log.Action, log.ResourceType, log.ResourceID,
		detailsArg, log.UserAgent, log.Result,
	).Scan(&log.ID, &log.Timestamp)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}

// List retrieves audit log entries
func (r *PostgresAuditRepository) List(ctx context.Context, limit, offset int) ([]*model.AuditLog, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM audit_log`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count audit log: %w", err)
	}

	query := `
		SELECT id, timestamp, actor_type, actor_id, action, resource_type,
		       resource_id, details, host(ip_address), user_agent, result
		FROM audit_log
		ORDER BY timestamp DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list audit log: %w", err)
	}
	defer rows.Close()

	var logs []*model.AuditLog
	for rows.Next() {
		var l model.AuditLog
		var ip *string
		if err := rows.Scan(
			&l.ID, &l.Timestamp, &l.ActorType, &l.ActorID, &l.Action,
			&l.ResourceType, &l.ResourceID, &l.Details, &ip, &l.UserAgent, &l.Result,
		); err != nil {
			return nil, 0, fmt.Errorf("scan audit log: %w", err)
		}
		l.IPAddress = ip
		logs = append(logs, &l)
	}

	return logs, total, nil
}

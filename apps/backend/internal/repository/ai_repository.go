package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/cifo-monitoring/backend/pkg/apperror"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AIRepository data access interface
type AIRepository interface {
	CreateSession(ctx context.Context, session *model.AISession) error
	GetSession(ctx context.Context, id uuid.UUID) (*model.AISession, error)
	ListSessionsByUser(ctx context.Context, userID uuid.UUID) ([]model.AISession, error)
	UpdateSessionActivity(ctx context.Context, id uuid.UUID) error
	CreateMessage(ctx context.Context, message *model.AIMessage) error
	GetMessagesBySession(ctx context.Context, sessionID uuid.UUID, limit int) ([]model.AIMessage, error)
	RecordUsage(ctx context.Context, usage *model.AIUsageTracking) error
	GetUsageStats(ctx context.Context, userID *uuid.UUID) (*model.AIUsageStats, error)
	CreateActionAudit(ctx context.Context, audit *model.AIActionAuditLog) error
	GetActionAudit(ctx context.Context, id uuid.UUID) (*model.AIActionAuditLog, error)
	UpdateActionAuditStatus(ctx context.Context, id uuid.UUID, status string, result *string) error
	UpdateIncidentRCA(ctx context.Context, incidentID uuid.UUID, rcaSummary string) error
}

// PostgresAIRepository postgres implementation
type PostgresAIRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresAIRepository constructor
func NewPostgresAIRepository(pool *pgxpool.Pool) *PostgresAIRepository {
	return &PostgresAIRepository{pool: pool}
}

// CreateSession insert new session
func (r *PostgresAIRepository) CreateSession(ctx context.Context, s *model.AISession) error {
	// insert session record
	query := `
		INSERT INTO ai_sessions (id, user_id, status, model_preference, created_at, last_activity_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	now := time.Now().UTC()
	s.CreatedAt = now
	s.LastActivityAt = now

	_, err := r.pool.Exec(ctx, query, s.ID, s.UserID, s.Status, s.ModelPreference, s.CreatedAt, s.LastActivityAt)
	if err != nil {
		return fmt.Errorf("create ai session: %w", err)
	}
	return nil
}

// GetSession find session by id
func (r *PostgresAIRepository) GetSession(ctx context.Context, id uuid.UUID) (*model.AISession, error) {
	// query session by id
	query := `
		SELECT id, user_id, status, model_preference, created_at, last_activity_at
		FROM ai_sessions
		WHERE id = $1
	`
	s := &model.AISession{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.UserID, &s.Status, &s.ModelPreference, &s.CreatedAt, &s.LastActivityAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.NewNotFound("AI session not found")
		}
		return nil, fmt.Errorf("get ai session: %w", err)
	}
	return s, nil
}

// ListSessionsByUser list sessions
func (r *PostgresAIRepository) ListSessionsByUser(ctx context.Context, userID uuid.UUID) ([]model.AISession, error) {
	// query sessions by user
	query := `
		SELECT id, user_id, status, model_preference, created_at, last_activity_at
		FROM ai_sessions
		WHERE user_id = $1
		ORDER BY last_activity_at DESC
		LIMIT 50
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list ai sessions: %w", err)
	}
	defer rows.Close()

	var sessions []model.AISession
	for rows.Next() {
		var s model.AISession
		if err := rows.Scan(&s.ID, &s.UserID, &s.Status, &s.ModelPreference, &s.CreatedAt, &s.LastActivityAt); err != nil {
			return nil, fmt.Errorf("scan session row: %w", err)
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

// UpdateSessionActivity touch session timestamp
func (r *PostgresAIRepository) UpdateSessionActivity(ctx context.Context, id uuid.UUID) error {
	// update activity timestamp
	query := `UPDATE ai_sessions SET last_activity_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("update session activity: %w", err)
	}
	return nil
}

// CreateMessage save chat message
func (r *PostgresAIRepository) CreateMessage(ctx context.Context, m *model.AIMessage) error {
	// insert chat message
	query := `
		INSERT INTO ai_messages (id, session_id, role, content, model_used, input_tokens, output_tokens, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now().UTC()
	}

	_, err := r.pool.Exec(ctx, query, m.ID, m.SessionID, m.Role, m.Content, m.ModelUsed, m.InputTokens, m.OutputTokens, m.CreatedAt)
	if err != nil {
		return fmt.Errorf("create ai message: %w", err)
	}
	return nil
}

// GetMessagesBySession load session messages
func (r *PostgresAIRepository) GetMessagesBySession(ctx context.Context, sessionID uuid.UUID, limit int) ([]model.AIMessage, error) {
	// select messages in order
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	query := `
		SELECT id, session_id, role, content, model_used, input_tokens, output_tokens, created_at
		FROM ai_messages
		WHERE session_id = $1
		ORDER BY created_at ASC
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, query, sessionID, limit)
	if err != nil {
		return nil, fmt.Errorf("get ai messages: %w", err)
	}
	defer rows.Close()

	var messages []model.AIMessage
	for rows.Next() {
		var m model.AIMessage
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &m.ModelUsed, &m.InputTokens, &m.OutputTokens, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan ai message: %w", err)
		}
		messages = append(messages, m)
	}
	return messages, nil
}

// RecordUsage save tracking record
func (r *PostgresAIRepository) RecordUsage(ctx context.Context, u *model.AIUsageTracking) error {
	// insert usage record
	query := `
		INSERT INTO ai_usage_tracking (id, user_id, session_id, model_provider, model_name, input_tokens, output_tokens, estimated_cost_usd, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	if u.Timestamp.IsZero() {
		u.Timestamp = time.Now().UTC()
	}

	_, err := r.pool.Exec(ctx, query, u.ID, u.UserID, u.SessionID, u.ModelProvider, u.ModelName, u.InputTokens, u.OutputTokens, u.EstimatedCostUSD, u.Timestamp)
	if err != nil {
		return fmt.Errorf("record ai usage: %w", err)
	}
	return nil
}

// GetUsageStats calculate stats
func (r *PostgresAIRepository) GetUsageStats(ctx context.Context, userID *uuid.UUID) (*model.AIUsageStats, error) {
	// aggregate usage query
	query := `
		SELECT 
			COALESCE(SUM(input_tokens + output_tokens), 0) as total_tokens,
			COALESCE(SUM(estimated_cost_usd), 0.0) as total_cost,
			COUNT(id) as req_count
		FROM ai_usage_tracking
		WHERE ($1::uuid IS NULL OR user_id = $1)
	`
	stats := &model.AIUsageStats{
		ProviderBreakdown: make(map[string]int),
		ModelBreakdown:    make(map[string]int),
	}

	err := r.pool.QueryRow(ctx, query, userID).Scan(&stats.TotalTokens, &stats.TotalCostUSD, &stats.RequestCount)
	if err != nil {
		return nil, fmt.Errorf("get ai usage summary: %w", err)
	}

	// provider breakdown
	pQuery := `
		SELECT model_provider, COUNT(id)
		FROM ai_usage_tracking
		WHERE ($1::uuid IS NULL OR user_id = $1)
		GROUP BY model_provider
	`
	pRows, err := r.pool.Query(ctx, pQuery, userID)
	if err == nil {
		defer pRows.Close()
		for pRows.Next() {
			var p string
			var c int
			if err := pRows.Scan(&p, &c); err == nil {
				stats.ProviderBreakdown[p] = c
			}
		}
	}

	// model breakdown
	mQuery := `
		SELECT model_name, COUNT(id)
		FROM ai_usage_tracking
		WHERE ($1::uuid IS NULL OR user_id = $1)
		GROUP BY model_name
	`
	mRows, err := r.pool.Query(ctx, mQuery, userID)
	if err == nil {
		defer mRows.Close()
		for mRows.Next() {
			var m string
			var c int
			if err := mRows.Scan(&m, &c); err == nil {
				stats.ModelBreakdown[m] = c
			}
		}
	}

	return stats, nil
}

// CreateActionAudit record action request
func (r *PostgresAIRepository) CreateActionAudit(ctx context.Context, a *model.AIActionAuditLog) error {
	// insert audit action
	query := `
		INSERT INTO ai_action_audit_log (
			id, user_id, session_id, prompt_input_hash, ai_output_summary, tool_name,
			tool_parameters, approval_status, execution_result, model_used, timestamp
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.Timestamp.IsZero() {
		a.Timestamp = time.Now().UTC()
	}

	_, err := r.pool.Exec(
		ctx, query, a.ID, a.UserID, a.SessionID, a.PromptInputHash, a.AIOutputSummary,
		a.ToolName, a.ToolParameters, a.ApprovalStatus, a.ExecutionResult, a.ModelUsed, a.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("create ai action audit: %w", err)
	}
	return nil
}

// GetActionAudit query action audit
func (r *PostgresAIRepository) GetActionAudit(ctx context.Context, id uuid.UUID) (*model.AIActionAuditLog, error) {
	// get audit by id
	query := `
		SELECT id, user_id, session_id, prompt_input_hash, ai_output_summary,
		       tool_name, tool_parameters, approval_status, execution_result, model_used, timestamp
		FROM ai_action_audit_log
		WHERE id = $1
	`
	a := &model.AIActionAuditLog{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&a.ID, &a.UserID, &a.SessionID, &a.PromptInputHash, &a.AIOutputSummary,
		&a.ToolName, &a.ToolParameters, &a.ApprovalStatus, &a.ExecutionResult, &a.ModelUsed, &a.Timestamp,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.NewNotFound("AI action approval request not found")
		}
		return nil, fmt.Errorf("get ai action audit: %w", err)
	}
	return a, nil
}

// UpdateActionAuditStatus update status
func (r *PostgresAIRepository) UpdateActionAuditStatus(ctx context.Context, id uuid.UUID, status string, result *string) error {
	// update approval status
	query := `UPDATE ai_action_audit_log SET approval_status = $1, execution_result = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, status, result, id)
	if err != nil {
		return fmt.Errorf("update ai action audit status: %w", err)
	}
	return nil
}

// UpdateIncidentRCA save incident rca
func (r *PostgresAIRepository) UpdateIncidentRCA(ctx context.Context, incidentID uuid.UUID, rcaSummary string) error {
	// update incident rca
	query := `UPDATE incidents SET rca_summary = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, rcaSummary, incidentID)
	if err != nil {
		return fmt.Errorf("update incident rca: %w", err)
	}
	return nil
}

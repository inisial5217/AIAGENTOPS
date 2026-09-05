package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository interface for users
type UserRepository interface {
	FindByID(ctx context.Context, id string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByKeycloakID(ctx context.Context, keycloakID string) (*model.User, error)
	UpsertKeycloakUser(ctx context.Context, user *model.User) (*model.User, error)
	List(ctx context.Context, limit, offset int) ([]*model.User, int, error)
}

// PostgresUserRepository pgx pool implementation
type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates user repository
func NewUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

// FindByID finds user by id
func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	query := `
		SELECT id, email, name, role, COALESCE(keycloak_id, ''), is_active, created_at, updated_at
		FROM users
		WHERE id = $1`

	var u model.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.Name, &u.Role, &u.KeycloakID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &u, nil
}

// FindByEmail finds user by email
func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, email, name, role, COALESCE(keycloak_id, ''), is_active, created_at, updated_at
		FROM users
		WHERE email = $1`

	var u model.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.Name, &u.Role, &u.KeycloakID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return &u, nil
}

// FindByKeycloakID finds user by keycloak id
func (r *PostgresUserRepository) FindByKeycloakID(ctx context.Context, keycloakID string) (*model.User, error) {
	query := `
		SELECT id, email, name, role, COALESCE(keycloak_id, ''), is_active, created_at, updated_at
		FROM users
		WHERE keycloak_id = $1`

	var u model.User
	err := r.pool.QueryRow(ctx, query, keycloakID).Scan(
		&u.ID, &u.Email, &u.Name, &u.Role, &u.KeycloakID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by keycloak id: %w", err)
	}
	return &u, nil
}

// UpsertKeycloakUser inserts or updates user from keycloak
func (r *PostgresUserRepository) UpsertKeycloakUser(ctx context.Context, user *model.User) (*model.User, error) {
	// check existing user
	existing, err := r.FindByKeycloakID(ctx, user.KeycloakID)
	if err != nil {
		return nil, err
	}

	if existing == nil && user.Email != "" {
		existing, err = r.FindByEmail(ctx, user.Email)
		if err != nil {
			return nil, err
		}
	}

	if existing != nil {
		updateQuery := `
			UPDATE users
			SET name = $1, role = $2, keycloak_id = $3, updated_at = $4
			WHERE id = $5
			RETURNING id, email, name, role, COALESCE(keycloak_id, ''), is_active, created_at, updated_at`

		var u model.User
		err = r.pool.QueryRow(ctx, updateQuery, user.Name, user.Role, user.KeycloakID, time.Now(), existing.ID).Scan(
			&u.ID, &u.Email, &u.Name, &u.Role, &u.KeycloakID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("update keycloak user: %w", err)
		}
		return &u, nil
	}

	insertQuery := `
		INSERT INTO users (email, name, role, keycloak_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, true, NOW(), NOW())
		RETURNING id, email, name, role, COALESCE(keycloak_id, ''), is_active, created_at, updated_at`

	var u model.User
	err = r.pool.QueryRow(ctx, insertQuery, user.Email, user.Name, user.Role, user.KeycloakID).Scan(
		&u.ID, &u.Email, &u.Name, &u.Role, &u.KeycloakID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert keycloak user: %w", err)
	}
	return &u, nil
}

// List lists all users
func (r *PostgresUserRepository) List(ctx context.Context, limit, offset int) ([]*model.User, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	query := `
		SELECT id, email, name, role, COALESCE(keycloak_id, ''), is_active, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.KeycloakID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, &u)
	}

	return users, total, nil
}

package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

func getTestPool(t *testing.T) *pgxpool.Pool {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "postgres://cifo_admin:cifo_secure_password@127.0.0.1:5432/cifo_db?sslmode=disable"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skipf("skipping db test: postgres pool failed: %v", err)
		return nil
	}
	if err := pool.Ping(ctx); err != nil {
		t.Skipf("skipping db test: postgres ping failed: %v", err)
		return nil
	}
	return pool
}

func TestUserRepository_CRUD(t *testing.T) {
	pool := getTestPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := NewUserRepository(pool)
	ctx := context.Background()

	// 1. upsert keycloak user
	randomID := uuid.New().String()[:8]
	testEmail := "test-" + randomID + "@cifo.local"
	user := &model.User{
		Email:      testEmail,
		Name:       "Test User " + randomID,
		Role:       "viewer",
		KeycloakID: "kc-" + randomID,
		IsActive:   true,
	}

	created, err := repo.UpsertKeycloakUser(ctx, user)
	assert.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, testEmail, created.Email)

	// 2. find by id
	found, err := repo.FindByID(ctx, created.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, created.Name, found.Name)

	// 3. find by email
	byEmail, err := repo.FindByEmail(ctx, testEmail)
	assert.NoError(t, err)
	assert.NotNil(t, byEmail)
	assert.Equal(t, created.ID, byEmail.ID)

	// 4. find by keycloak id
	byKC, err := repo.FindByKeycloakID(ctx, "kc-"+randomID)
	assert.NoError(t, err)
	assert.NotNil(t, byKC)
	assert.Equal(t, created.ID, byKC.ID)

	// 5. update role
	updated, err := repo.UpdateRole(ctx, created.ID, "devops")
	assert.NoError(t, err)
	assert.Equal(t, "devops", updated.Role)

	// 6. set active
	deactivated, err := repo.SetActive(ctx, created.ID, false)
	assert.NoError(t, err)
	assert.False(t, deactivated.IsActive)

	// 7. list users
	users, total, err := repo.List(ctx, 10, 0)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, total, 1)
	assert.NotEmpty(t, users)

	// 8. upsert existing user (update branch)
	created.Name = "Updated User Name"
	created.Role = "admin"
	upserted, err := repo.UpsertKeycloakUser(ctx, created)
	assert.NoError(t, err)
	assert.Equal(t, "Updated User Name", upserted.Name)
	assert.Equal(t, "admin", upserted.Role)

	// cleanup
	_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id = $1", created.ID)
}

func TestUserRepository_NotFound(t *testing.T) {
	pool := getTestPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := NewUserRepository(pool)
	ctx := context.Background()

	// non-existent uuid
	missingID := uuid.New().String()
	found, err := repo.FindByID(ctx, missingID)
	assert.NoError(t, err)
	assert.Nil(t, found)

	byEmail, err := repo.FindByEmail(ctx, "nonexistent-user@cifo.local")
	assert.NoError(t, err)
	assert.Nil(t, byEmail)

	byKC, err := repo.FindByKeycloakID(ctx, "nonexistent-kc-id")
	assert.NoError(t, err)
	assert.Nil(t, byKC)

	upRole, err := repo.UpdateRole(ctx, missingID, "admin")
	assert.NoError(t, err)
	assert.Nil(t, upRole)

	setActive, err := repo.SetActive(ctx, missingID, false)
	assert.NoError(t, err)
	assert.Nil(t, setActive)
}

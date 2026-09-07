package repository

import (
	"context"
	"testing"
	"time"

	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuditRepository_CreateAndList(t *testing.T) {
	pool := getTestPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := NewAuditRepository(pool)
	ctx := context.Background()

	resID := "res-" + uuid.New().String()[:6]
	ip := "127.0.0.1"
	ua := "AuditTestRunner/1.0"
	logEntry := &model.AuditLog{
		Timestamp:    time.Now().UTC(),
		ActorType:    "user",
		ActorID:      "test-actor",
		Action:       "unit_test_action",
		ResourceType: "test_resource",
		ResourceID:   &resID,
		Details:      []byte(`{"key":"value"}`),
		IPAddress:    &ip,
		UserAgent:    &ua,
		Result:       "success",
	}

	// 1. create audit
	err := repo.Create(ctx, logEntry)
	assert.NoError(t, err)
	assert.NotEmpty(t, logEntry.ID)

	// 2. list audit
	logs, total, err := repo.List(ctx, 10, 0)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, total, 1)
	assert.NotEmpty(t, logs)

	// cleanup
	_, _ = pool.Exec(ctx, "DELETE FROM audit_log WHERE id = $1", logEntry.ID)
}

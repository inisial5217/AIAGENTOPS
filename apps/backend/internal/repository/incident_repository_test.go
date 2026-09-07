package repository

import (
	"context"
	"testing"
	"time"

	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIncidentRepository_Lifecycle(t *testing.T) {
	pool := getTestPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := NewIncidentRepository(pool)
	ctx := context.Background()

	resID := "cont-" + uuid.New().String()[:8]
	inc := &model.Incident{
		Title:        "High Memory Alert",
		Description:  "Memory consumption is 95%",
		Severity:     model.SeverityCritical,
		Status:       model.IncidentStatusOpen,
		Source:       "docker",
		AlertName:    "ContainerHighMemory",
		ResourceType: "container",
		ResourceID:   resID,
	}

	// 1. create
	err := repo.Create(ctx, inc)
	assert.NoError(t, err)
	assert.NotEmpty(t, inc.ID)

	// 2. get by id
	found, err := repo.GetByID(ctx, inc.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, inc.Title, found.Title)

	// 3. find open by alert and resource
	open, err := repo.FindOpenByAlertAndResource(ctx, "ContainerHighMemory", resID)
	assert.NoError(t, err)
	assert.NotNil(t, open)
	assert.Equal(t, inc.ID, open.ID)

	// 4. update status to acknowledged
	var existingUserID *string
	var uid string
	if err := pool.QueryRow(ctx, "SELECT id FROM users LIMIT 1").Scan(&uid); err == nil {
		existingUserID = &uid
	}
	err = repo.UpdateStatus(ctx, inc.ID, model.IncidentStatusAcknowledged, existingUserID)
	assert.NoError(t, err)

	// 5. update status to resolved
	err = repo.UpdateStatus(ctx, inc.ID, model.IncidentStatusResolved, existingUserID)
	assert.NoError(t, err)

	// 6. update status to closed
	err = repo.UpdateStatus(ctx, inc.ID, model.IncidentStatusClosed, existingUserID)
	assert.NoError(t, err)

	// 7. get detail by id
	detail, err := repo.GetDetailByID(ctx, inc.ID)
	assert.NoError(t, err)
	assert.NotNil(t, detail)
	assert.Equal(t, model.IncidentStatusClosed, detail.Status)

	// 8. list with filter
	list, total, err := repo.List(ctx, model.IncidentFilter{
		Status:   model.IncidentStatusClosed,
		Severity: model.SeverityCritical,
		Limit:    10,
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, total, 1)
	assert.NotEmpty(t, list)

	// 9. stats
	stats, err := repo.GetStats(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.GreaterOrEqual(t, stats.Total, 1)

	// 10. save notification record
	notif := &model.NotificationRecord{
		IncidentID: &inc.ID,
		Channel:    "telegram",
		Recipient:  "devops_chat",
		Title:      inc.Title,
		Message:    "Critical alert triggered",
		Severity:   inc.Severity,
		Status:     "sent",
	}
	err = repo.SaveNotification(ctx, notif)
	assert.NoError(t, err)

	// 11. list notifications by incident id
	notifs, err := repo.ListNotificationsByIncidentID(ctx, inc.ID)
	assert.NoError(t, err)
	assert.Len(t, notifs, 1)

	// cleanup
	_, _ = pool.Exec(ctx, "DELETE FROM notification_history WHERE incident_id = $1", inc.ID)
	_, _ = pool.Exec(ctx, "DELETE FROM incidents WHERE id = $1", inc.ID)
}

func TestIncidentRepository_GetUnacknowledged(t *testing.T) {
	pool := getTestPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := NewIncidentRepository(pool)
	ctx := context.Background()

	// create unacknowledged incident with older created_at
	inc := &model.Incident{
		Title:        "Old Unack Alert",
		Description:  "Unacknowledged alert test",
		Severity:     model.SeverityWarning,
		Status:       model.IncidentStatusOpen,
		Source:       "system",
		AlertName:    "OldUnackAlert",
		ResourceType: "node",
		ResourceID:   "node-1",
	}
	err := repo.Create(ctx, inc)
	assert.NoError(t, err)

	// update created_at to 2 hours ago
	_, _ = pool.Exec(ctx, "UPDATE incidents SET created_at = NOW() - INTERVAL '2 hours' WHERE id = $1", inc.ID)

	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM incidents WHERE id = $1", inc.ID)
	}()

	// check unacknowledged
	old, err := repo.GetUnacknowledgedOlderThan(ctx, 1*time.Hour)
	assert.NoError(t, err)
	assert.NotEmpty(t, old)
}

func TestIncidentRepository_NotFound(t *testing.T) {
	pool := getTestPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := NewIncidentRepository(pool)
	ctx := context.Background()

	inc, err := repo.GetByID(ctx, "00000000-0000-0000-0000-000000000000")
	assert.NoError(t, err)
	assert.Nil(t, inc)

	detail, err := repo.GetDetailByID(ctx, "00000000-0000-0000-0000-000000000000")
	assert.NoError(t, err)
	assert.Nil(t, detail)

	open, err := repo.FindOpenByAlertAndResource(ctx, "NonExistentAlert", "non-existent-res")
	assert.NoError(t, err)
	assert.Nil(t, open)
}

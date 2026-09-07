package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSettingsRepository_SystemSettings(t *testing.T) {
	pool := getTestPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := NewSettingsRepository(pool)
	ctx := context.Background()

	// 1. get settings
	settings, err := repo.GetSystemSettings(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, settings)
	assert.NotEmpty(t, settings.AppName)

	// 2. update settings
	origName := settings.AppName
	settings.AppName = "CIFO Enterprise Test"
	updated, err := repo.UpdateSystemSettings(ctx, settings)
	assert.NoError(t, err)
	assert.Equal(t, "CIFO Enterprise Test", updated.AppName)

	// restore
	settings.AppName = origName
	_, _ = repo.UpdateSystemSettings(ctx, settings)
}

func TestSettingsRepository_NotificationSettings(t *testing.T) {
	pool := getTestPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := NewSettingsRepository(pool)
	ctx := context.Background()

	notif, err := repo.GetNotificationSettings(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, notif)

	// update
	notif.TelegramEnabled = true
	updated, err := repo.UpdateNotificationSettings(ctx, notif)
	assert.NoError(t, err)
	assert.True(t, updated.TelegramEnabled)
}

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAIRepository_SessionAndMessages(t *testing.T) {
	pool := getTestPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	userRepo := NewUserRepository(pool)
	aiRepo := NewPostgresAIRepository(pool)
	ctx := context.Background()

	// 1. create test user for FK
	testEmail := "ai-test-" + uuid.New().String()[:6] + "@cifo.local"
	u, err := userRepo.UpsertKeycloakUser(ctx, &model.User{
		Email:      testEmail,
		Name:       "AI Test User",
		Role:       "devops",
		KeycloakID: "kc-ai-" + testEmail,
		IsActive:   true,
	})
	assert.NoError(t, err)
	userID, _ := uuid.Parse(u.ID)

	modelName := "gemini-2.0-flash"
	// 2. create ai session
	sess := &model.AISession{
		ID:              uuid.New(),
		UserID:          userID,
		Status:          "active",
		ModelPreference: &modelName,
	}
	err = aiRepo.CreateSession(ctx, sess)
	assert.NoError(t, err)

	// 3. get session
	gotSess, err := aiRepo.GetSession(ctx, sess.ID)
	assert.NoError(t, err)
	assert.NotNil(t, gotSess)
	assert.Equal(t, sess.ID, gotSess.ID)

	// 4. list sessions by user
	userSessions, err := aiRepo.ListSessionsByUser(ctx, userID)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(userSessions), 1)

	// 5. update activity
	err = aiRepo.UpdateSessionActivity(ctx, sess.ID)
	assert.NoError(t, err)

	// 6. create message
	msg := &model.AIMessage{
		ID:           uuid.New(),
		SessionID:    sess.ID,
		Role:         "user",
		Content:      "How to scale deployment?",
		ModelUsed:    &modelName,
		InputTokens:  10,
		OutputTokens: 25,
		CreatedAt:    time.Now().UTC(),
	}
	err = aiRepo.CreateMessage(ctx, msg)
	assert.NoError(t, err)

	// 7. get messages
	msgs, err := aiRepo.GetMessagesBySession(ctx, sess.ID, 10)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(msgs), 1)

	// 8. record usage
	usage := &model.AIUsageTracking{
		ID:               uuid.New(),
		UserID:           userID,
		SessionID:        &sess.ID,
		ModelProvider:    "google",
		ModelName:        "gemini-2.0-flash",
		InputTokens:      10,
		OutputTokens:     25,
		EstimatedCostUSD: 0.0001,
		Timestamp:        time.Now().UTC(),
	}
	err = aiRepo.RecordUsage(ctx, usage)
	assert.NoError(t, err)

	// 9. usage stats
	stats, err := aiRepo.GetUsageStats(ctx, &userID)
	assert.NoError(t, err)
	assert.NotNil(t, stats)

	// cleanup
	_, _ = pool.Exec(ctx, "DELETE FROM ai_usage_tracking WHERE user_id = $1", userID)
	_, _ = pool.Exec(ctx, "DELETE FROM ai_messages WHERE session_id = $1", sess.ID)
	_, _ = pool.Exec(ctx, "DELETE FROM ai_sessions WHERE id = $1", sess.ID)
	_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id = $1", u.ID)
}

func TestAIRepository_ActionAuditAndRCA(t *testing.T) {
	pool := getTestPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	userRepo := NewUserRepository(pool)
	aiRepo := NewPostgresAIRepository(pool)
	incRepo := NewIncidentRepository(pool)
	ctx := context.Background()

	// 1. user
	testEmail := "audit-test-" + uuid.New().String()[:6] + "@cifo.local"
	u, err := userRepo.UpsertKeycloakUser(ctx, &model.User{
		Email:      testEmail,
		Name:       "Audit User",
		Role:       "admin",
		KeycloakID: "kc-audit-" + testEmail,
		IsActive:   true,
	})
	assert.NoError(t, err)
	userID, _ := uuid.Parse(u.ID)

	// 2. create action audit
	toolName := "restart_deployment"
	audit := &model.AIActionAuditLog{
		ID:              uuid.New(),
		UserID:          userID,
		PromptInputHash: "hash123",
		ToolName:        &toolName,
		ToolParameters:  []byte(`{"namespace":"default"}`),
		ApprovalStatus:  "pending",
		Timestamp:       time.Now().UTC(),
	}
	err = aiRepo.CreateActionAudit(ctx, audit)
	assert.NoError(t, err)

	// 3. get action audit
	gotAudit, err := aiRepo.GetActionAudit(ctx, audit.ID)
	assert.NoError(t, err)
	assert.NotNil(t, gotAudit)
	assert.Equal(t, "pending", gotAudit.ApprovalStatus)

	// 4. update action audit status
	execResult := "Restarted deployment successfully"
	err = aiRepo.UpdateActionAuditStatus(ctx, audit.ID, "approved", &execResult)
	assert.NoError(t, err)

	updatedAudit, err := aiRepo.GetActionAudit(ctx, audit.ID)
	assert.NoError(t, err)
	assert.Equal(t, "approved", updatedAudit.ApprovalStatus)

	// 5. create incident and update rca
	inc := &model.Incident{
		Title:       "RCA Test Incident",
		Description: "Diagnostic test",
		Severity:    model.SeverityCritical,
		Status:      model.IncidentStatusOpen,
		Source:      "system",
	}
	err = incRepo.Create(ctx, inc)
	assert.NoError(t, err)
	incUUID, _ := uuid.Parse(inc.ID)

	err = aiRepo.UpdateIncidentRCA(ctx, incUUID, "Root cause: Out of memory")
	assert.NoError(t, err)

	// verify incident rca
	foundInc, err := incRepo.GetByID(ctx, inc.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Root cause: Out of memory", foundInc.RCASummary)

	// cleanup
	_, _ = pool.Exec(ctx, "DELETE FROM ai_action_audit_logs WHERE id = $1", audit.ID)
	_, _ = pool.Exec(ctx, "DELETE FROM incidents WHERE id = $1", inc.ID)
	_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id = $1", u.ID)
}

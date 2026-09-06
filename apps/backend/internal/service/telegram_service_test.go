package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/cifo-monitoring/backend/pkg/logger"
	"github.com/stretchr/testify/assert"
)

// mockTelegramClient mock client
type mockTelegramClient struct {
	messages []string
}

func (m *mockTelegramClient) SendMessage(ctx context.Context, text string) error {
	m.messages = append(m.messages, text)
	return nil
}

func (m *mockTelegramClient) IsConfigured() bool {
	return true
}

func TestTelegramService_FormatIncidentMessage(t *testing.T) {
	log := logger.New("DEBUG", "test")
	mockClient := &mockTelegramClient{}
	svc := NewTelegramService(mockClient, nil, log)

	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	inc := &model.Incident{
		Title:       "ContainerOOMKilled",
		Description: "Container killed due to OOM",
		Severity:    "critical",
		Source:      "Docker",
		ResourceID:  "payment-gateway",
		CreatedAt:   now,
	}

	// standard format test
	msg := svc.FormatIncidentMessage(inc, false)
	assert.Contains(t, msg, "[CRITICAL] ContainerOOMKilled")
	assert.Contains(t, msg, "Resource: payment-gateway (Docker)")
	assert.Contains(t, msg, "Action Required: Acknowledge in dashboard")
	// verify no emoji
	assert.False(t, strings.ContainsAny(msg, "🚨⚠️ℹ️🔴"))

	// escalated format test
	escMsg := svc.FormatIncidentMessage(inc, true)
	assert.Contains(t, escMsg, "[ESCALATED] [CRITICAL] ContainerOOMKilled")
	assert.Contains(t, escMsg, "Status: UNACKNOWLEDGED (> 15m)")
	assert.Contains(t, escMsg, "Action Required: Immediate acknowledgement required")
}

func TestTelegramService_FormatBatchSummary(t *testing.T) {
	log := logger.New("DEBUG", "test")
	mockClient := &mockTelegramClient{}
	svc := NewTelegramService(mockClient, nil, log)

	alerts := []*model.Incident{
		{Title: "HighCPU", Severity: "warning", ResourceID: "web-01"},
		{Title: "DiskFull", Severity: "critical", ResourceID: "db-01"},
		{Title: "NetworkDrop", Severity: "warning", ResourceID: "gw-01"},
		{Title: "PodRestart", Severity: "critical", ResourceID: "cart-01"},
	}

	summary := svc.FormatBatchSummary(alerts)
	assert.Contains(t, summary, "[SUMMARY] 4 Alerts Detected (2m window)")
	assert.Contains(t, summary, "- [WARNING] HighCPU on web-01")
	assert.Contains(t, summary, "- [CRITICAL] DiskFull on db-01")
	assert.Contains(t, summary, "Action Required: Review active incidents in dashboard")
}

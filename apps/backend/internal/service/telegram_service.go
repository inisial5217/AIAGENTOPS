package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/cifo-monitoring/backend/internal/integration"
	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/redis/go-redis/v9"
)

const (
	telegramRetryQueueKey = "cifo:telegram:retry_queue"
	batchWindowSeconds    = 120
	batchThresholdCount   = 3
)

// TelegramService interface
type TelegramService interface {
	SendIncidentAlert(ctx context.Context, inc *model.Incident, isEscalation bool) error
	ProcessRetryQueue(ctx context.Context) error
	FormatIncidentMessage(inc *model.Incident, isEscalation bool) string
	FormatBatchSummary(incidents []*model.Incident) string
}

// DefaultTelegramService implementation
type DefaultTelegramService struct {
	client       integration.TelegramClient
	redisClient  *redis.Client
	logger       *slog.Logger
	batchMutex   sync.Mutex
	recentAlerts []*model.Incident
	windowStart  time.Time
}

// NewTelegramService creates service
func NewTelegramService(client integration.TelegramClient, rdb *redis.Client, logger *slog.Logger) *DefaultTelegramService {
	return &DefaultTelegramService{
		client:      client,
		redisClient: rdb,
		logger:      logger,
		windowStart: time.Now(),
	}
}

// FormatIncidentMessage formats markdown
func (s *DefaultTelegramService) FormatIncidentMessage(inc *model.Incident, isEscalation bool) string {
	sev := strings.ToUpper(inc.Severity)
	if sev == "" {
		sev = "CRITICAL"
	}

	res := inc.ResourceID
	if res == "" {
		res = "unknown"
	}
	src := inc.Source
	if src == "" {
		src = "System"
	}

	if isEscalation {
		return fmt.Sprintf(
			"[ESCALATED] [%s] %s\nResource: %s (%s)\nStatus: UNACKNOWLEDGED (> 15m)\nCreated: %s\nAction Required: Immediate acknowledgement required",
			sev, inc.Title, res, src, inc.CreatedAt.UTC().Format("2006-01-02 15:04:05 UTC"),
		)
	}

	desc := inc.Description
	if desc == "" {
		desc = inc.Title
	}

	return fmt.Sprintf(
		"[%s] %s\nResource: %s (%s)\nTime: %s\nDescription: %s\nAction Required: Acknowledge in dashboard",
		sev, inc.Title, res, src, inc.CreatedAt.UTC().Format("2006-01-02 15:04:05 UTC"), desc,
	)
}

// FormatBatchSummary formats batch
func (s *DefaultTelegramService) FormatBatchSummary(incidents []*model.Incident) string {
	lines := []string{
		fmt.Sprintf("[SUMMARY] %d Alerts Detected (2m window)", len(incidents)),
	}

	for _, inc := range incidents {
		res := inc.ResourceID
		if res == "" {
			res = "system"
		}
		lines = append(lines, fmt.Sprintf("- [%s] %s on %s", strings.ToUpper(inc.Severity), inc.Title, res))
	}

	lines = append(lines, "Action Required: Review active incidents in dashboard")
	return strings.Join(lines, "\n")
}

// SendIncidentAlert sends alert
func (s *DefaultTelegramService) SendIncidentAlert(ctx context.Context, inc *model.Incident, isEscalation bool) error {
	s.batchMutex.Lock()
	now := time.Now()

	// check batch window
	if now.Sub(s.windowStart) > time.Duration(batchWindowSeconds)*time.Second {
		s.recentAlerts = nil
		s.windowStart = now
	}

	s.recentAlerts = append(s.recentAlerts, inc)
	alertCount := len(s.recentAlerts)
	alertsCopy := make([]*model.Incident, len(s.recentAlerts))
	copy(alertsCopy, s.recentAlerts)
	s.batchMutex.Unlock()

	var msg string
	// batch storm condition
	if alertCount == batchThresholdCount+1 {
		msg = s.FormatBatchSummary(alertsCopy)
	} else if alertCount > batchThresholdCount+1 {
		// already batched
		s.logger.Info("alert batched, skipping individual telegram message", slog.String("title", inc.Title))
		return nil
	} else {
		msg = s.FormatIncidentMessage(inc, isEscalation)
	}

	// attempt send
	err := s.client.SendMessage(ctx, msg)
	if err != nil {
		s.logger.Warn("telegram send failed, enqueuing retry", slog.String("error", err.Error()))
		// enqueue to redis
		s.enqueueRetry(ctx, msg)
		return err
	}

	return nil
}

// enqueueRetry pushes message
func (s *DefaultTelegramService) enqueueRetry(ctx context.Context, text string) {
	if s.redisClient == nil {
		return
	}

	payload := map[string]interface{}{
		"text":       text,
		"created_at": time.Now().Unix(),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	s.redisClient.LPush(ctx, telegramRetryQueueKey, data)
}

// ProcessRetryQueue drains queue
func (s *DefaultTelegramService) ProcessRetryQueue(ctx context.Context) error {
	if s.redisClient == nil || !s.client.IsConfigured() {
		return nil
	}

	// pop up to 5 items
	for i := 0; i < 5; i++ {
		data, err := s.redisClient.RPop(ctx, telegramRetryQueueKey).Bytes()
		if err != nil {
			if err == redis.Nil {
				return nil
			}
			return fmt.Errorf("redis rpop: %w", err)
		}

		var payload struct {
			Text      string `json:"text"`
			CreatedAt int64  `json:"created_at"`
		}
		if err := json.Unmarshal(data, &payload); err != nil {
			continue
		}

		// discard older than 24h
		if time.Since(time.Unix(payload.CreatedAt, 0)) > 24*time.Hour {
			continue
		}

		if err := s.client.SendMessage(ctx, payload.Text); err != nil {
			// re-enqueue item
			s.redisClient.LPush(ctx, telegramRetryQueueKey, data)
			return fmt.Errorf("retry send telegram: %w", err)
		}

		s.logger.Info("retried telegram message delivered successfully")
	}

	return nil
}

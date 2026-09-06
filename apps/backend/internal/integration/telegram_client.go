package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// TelegramClient interface
type TelegramClient interface {
	SendMessage(ctx context.Context, text string) error
	IsConfigured() bool
}

// HTTPTelegramClient implementation
type HTTPTelegramClient struct {
	token      string
	chatID     string
	client     *http.Client
	logger     *slog.Logger
	rateMutex  sync.Mutex
	timestamps []time.Time
}

// NewTelegramClient creates client
func NewTelegramClient(token, chatID string, logger *slog.Logger) *HTTPTelegramClient {
	return &HTTPTelegramClient{
		token:  token,
		chatID: chatID,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// IsConfigured checks credentials
func (c *HTTPTelegramClient) IsConfigured() bool {
	return c.token != "" && c.chatID != ""
}

// checkRateLimit enforces limit
func (c *HTTPTelegramClient) checkRateLimit() bool {
	c.rateMutex.Lock()
	defer c.rateMutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-1 * time.Minute)

	// prune old timestamps
	valid := []time.Time{}
	for _, t := range c.timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	c.timestamps = valid

	// max 30 per minute
	if len(c.timestamps) >= 30 {
		return false
	}

	c.timestamps = append(c.timestamps, now)
	return true
}

// SendMessage sends message
func (c *HTTPTelegramClient) SendMessage(ctx context.Context, text string) error {
	// graceful unconfigured check
	if !c.IsConfigured() {
		c.logger.Debug("telegram client unconfigured", slog.String("message_preview", text[:min(len(text), 60)]))
		return nil
	}

	// check rate limit
	if !c.checkRateLimit() {
		return fmt.Errorf("telegram rate limit exceeded (max 30/min)")
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.token)

	reqBody := map[string]interface{}{
		"chat_id":                  c.chatID,
		"text":                     text,
		"parse_mode":               "Markdown",
		"disable_web_page_preview": true,
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal telegram body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return fmt.Errorf("create telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("do telegram request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram API error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

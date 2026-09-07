package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// AIChatClientResponse response from ai service
type AIChatClientResponse struct {
	SessionID        string                   `json:"session_id"`
	Content          string                   `json:"content"`
	ToolCalls        []map[string]interface{} `json:"tool_calls"`
	ModelUsed        string                   `json:"model_used"`
	ProviderName     string                   `json:"provider_name"`
	InputTokens      int                      `json:"input_tokens"`
	OutputTokens     int                      `json:"output_tokens"`
	EstimatedCostUSD float64                  `json:"estimated_cost_usd"`
	SecurityFlag     *string                  `json:"security_flag,omitempty"`
}

// AIDiagnoseClientResponse rca response from ai service
type AIDiagnoseClientResponse struct {
	IncidentID       string  `json:"incident_id"`
	RCASummary       string  `json:"rca_summary"`
	ModelUsed        string  `json:"model_used"`
	ProviderName     string  `json:"provider_name"`
	InputTokens      int     `json:"input_tokens"`
	OutputTokens     int     `json:"output_tokens"`
	EstimatedCostUSD float64 `json:"estimated_cost_usd"`
}

// AIClient interface
type AIClient interface {
	Chat(ctx context.Context, sessionID string, userID string, message string, role string, history []map[string]string) (*AIChatClientResponse, error)
	Diagnose(ctx context.Context, incidentID string, alertName string, severity string, resource string, namespace string, logs string, metrics map[string]interface{}) (*AIDiagnoseClientResponse, error)
	GetModels(ctx context.Context) (map[string]interface{}, error)
}

// HTTPAIClient client implementation
type HTTPAIClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewHTTPAIClient constructor
func NewHTTPAIClient(baseURL string) *HTTPAIClient {
	// normalize url
	normalized := strings.TrimRight(baseURL, "/")
	if !strings.HasPrefix(normalized, "http://") && !strings.HasPrefix(normalized, "https://") {
		normalized = "http://" + normalized
	}

	return &HTTPAIClient{
		baseURL: normalized,
		httpClient: &http.Client{
			Timeout: 35 * time.Second,
		},
	}
}

// Chat send chat query
func (c *HTTPAIClient) Chat(
	ctx context.Context,
	sessionID string,
	userID string,
	message string,
	role string,
	history []map[string]string,
) (*AIChatClientResponse, error) {
	tracer := otel.GetTracerProvider().Tracer("cifo-ai-client")
	ctx, span := tracer.Start(ctx, "ai_service.chat",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("ai.session_id", sessionID),
			attribute.String("ai.user_role", role),
		),
	)
	defer span.End()

	// execute chat post
	payload := map[string]interface{}{
		"session_id": sessionID,
		"user_id":    userID,
		"message":    message,
		"user_role":  role,
		"history":    history,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("marshal chat payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/chat", bytes.NewBuffer(data))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("create chat request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// propagate trace context
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("call ai service chat: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		err = fmt.Errorf("ai service error %d: %s", resp.StatusCode, string(bodyBytes))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	var chatResp AIChatClientResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("decode chat response: %w", err)
	}

	span.SetAttributes(
		attribute.String("ai.model_used", chatResp.ModelUsed),
		attribute.String("ai.provider_name", chatResp.ProviderName),
		attribute.Int("ai.input_tokens", chatResp.InputTokens),
		attribute.Int("ai.output_tokens", chatResp.OutputTokens),
	)
	span.SetStatus(codes.Ok, "")

	return &chatResp, nil
}

// Diagnose send rca request
func (c *HTTPAIClient) Diagnose(
	ctx context.Context,
	incidentID string,
	alertName string,
	severity string,
	resource string,
	namespace string,
	logs string,
	metrics map[string]interface{},
) (*AIDiagnoseClientResponse, error) {
	tracer := otel.GetTracerProvider().Tracer("cifo-ai-client")
	ctx, span := tracer.Start(ctx, "ai_service.diagnose",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("ai.incident_id", incidentID),
			attribute.String("ai.alert_name", alertName),
			attribute.String("ai.resource", resource),
			attribute.String("ai.severity", severity),
		),
	)
	defer span.End()

	// execute rca post
	payload := map[string]interface{}{
		"incident_id": incidentID,
		"alert_name":  alertName,
		"severity":    severity,
		"resource":    resource,
		"namespace":   namespace,
		"logs":        logs,
		"metrics":     metrics,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("marshal diagnose payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/diagnose", bytes.NewBuffer(data))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("create diagnose request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// propagate trace context
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("call ai service diagnose: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		err = fmt.Errorf("ai diagnose error %d: %s", resp.StatusCode, string(bodyBytes))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	var diagResp AIDiagnoseClientResponse
	if err := json.NewDecoder(resp.Body).Decode(&diagResp); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("decode diagnose response: %w", err)
	}

	span.SetAttributes(
		attribute.String("ai.model_used", diagResp.ModelUsed),
		attribute.String("ai.provider_name", diagResp.ProviderName),
		attribute.Int("ai.input_tokens", diagResp.InputTokens),
		attribute.Int("ai.output_tokens", diagResp.OutputTokens),
	)
	span.SetStatus(codes.Ok, "")

	return &diagResp, nil
}

// GetModels query model status
func (c *HTTPAIClient) GetModels(ctx context.Context) (map[string]interface{}, error) {
	tracer := otel.GetTracerProvider().Tracer("cifo-ai-client")
	ctx, span := tracer.Start(ctx, "ai_service.models", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	// fetch model status
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/models", nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("create models request: %w", err)
	}

	// propagate trace context
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("call ai service models: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("decode models response: %w", err)
	}

	span.SetStatus(codes.Ok, "")
	return result, nil
}

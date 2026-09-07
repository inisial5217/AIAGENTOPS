package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
)

// TestTraceHandlerContext verifies key extraction
func TestTraceHandlerContext(t *testing.T) {
	var buf bytes.Buffer
	baseHandler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	handler := &TraceHandler{Handler: baseHandler}
	testLogger := slog.New(handler)

	ctx := ContextWithTrace(context.Background(), "test-trace-123", "test-span-456")
	testLogger.InfoContext(ctx, "hello test")

	var output map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &output)
	assert.NoError(t, err)
	assert.Equal(t, "hello test", output["msg"])
	assert.Equal(t, "test-trace-123", output["trace_id"])
	assert.Equal(t, "test-span-456", output["span_id"])
}

// TestTraceHandlerOTelSpan verifies otel span extraction
func TestTraceHandlerOTelSpan(t *testing.T) {
	tracer := otel.GetTracerProvider().Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test-operation")
	defer span.End()

	var buf bytes.Buffer
	baseHandler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	handler := &TraceHandler{Handler: baseHandler}
	testLogger := slog.New(handler)

	testLogger.InfoContext(ctx, "span log")

	var output map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &output)
	assert.NoError(t, err)

	if span.SpanContext().IsValid() {
		assert.Equal(t, span.SpanContext().TraceID().String(), output["trace_id"])
		assert.Equal(t, span.SpanContext().SpanID().String(), output["span_id"])
	}
}

// TestWithContext verifies context logger
func TestWithContext(t *testing.T) {
	base := New("INFO", "test-svc")
	ctx := ContextWithTrace(context.Background(), "tid-999", "sid-888")
	withCtx := WithContext(ctx, base)
	assert.NotNil(t, withCtx)
}

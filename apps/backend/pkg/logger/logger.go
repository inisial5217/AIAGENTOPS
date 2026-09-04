package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type contextKey string

const (
	// TraceIDKey context key
	TraceIDKey contextKey = "trace_id"
	// SpanIDKey context key
	SpanIDKey contextKey = "span_id"
)

// New creates slog logger
func New(levelStr string, serviceName ...string) *slog.Logger {
	service := "cifo-backend"
	if len(serviceName) > 0 && serviceName[0] != "" {
		service = serviceName[0]
	}

	var level slog.Level
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		level = slog.LevelDebug
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts).WithAttrs([]slog.Attr{
		slog.String("service", service),
	})

	return slog.New(handler)
}

// WithContext appends trace context
func WithContext(ctx context.Context, l *slog.Logger) *slog.Logger {
	if l == nil {
		l = slog.Default()
	}

	if traceID, ok := ctx.Value(TraceIDKey).(string); ok && traceID != "" {
		l = l.With(slog.String("trace_id", traceID))
	}
	if spanID, ok := ctx.Value(SpanIDKey).(string); ok && spanID != "" {
		l = l.With(slog.String("span_id", spanID))
	}

	return l
}

// ContextWithTrace injects trace ids
func ContextWithTrace(ctx context.Context, traceID string, spanID string) context.Context {
	if traceID != "" {
		ctx = context.WithValue(ctx, TraceIDKey, traceID)
	}
	if spanID != "" {
		ctx = context.WithValue(ctx, SpanIDKey, spanID)
	}
	return ctx
}

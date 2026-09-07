package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/otel/trace"
)

type contextKey string

const (
	// TraceIDKey context key
	TraceIDKey contextKey = "trace_id"
	// SpanIDKey context key
	SpanIDKey contextKey = "span_id"
)

// TraceHandler wraps slog handler
type TraceHandler struct {
	slog.Handler
}

// Handle appends trace attributes
func (h *TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	traceID := ""
	spanID := ""

	// extract otel span
	span := trace.SpanFromContext(ctx)
	if span != nil && span.SpanContext().IsValid() {
		traceID = span.SpanContext().TraceID().String()
		spanID = span.SpanContext().SpanID().String()
	} else {
		if tid, ok := ctx.Value(TraceIDKey).(string); ok {
			traceID = tid
		}
		if sid, ok := ctx.Value(SpanIDKey).(string); ok {
			spanID = sid
		}
	}

	if traceID != "" {
		r.AddAttrs(slog.String("trace_id", traceID))
	}
	if spanID != "" {
		r.AddAttrs(slog.String("span_id", spanID))
	}

	return h.Handler.Handle(ctx, r)
}

// Enabled returns handler enabled
func (h *TraceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Handler.Enabled(ctx, level)
}

// WithAttrs returns handler attributes
func (h *TraceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TraceHandler{Handler: h.Handler.WithAttrs(attrs)}
}

// WithGroup returns handler group
func (h *TraceHandler) WithGroup(name string) slog.Handler {
	return &TraceHandler{Handler: h.Handler.WithGroup(name)}
}

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

	jsonHandler := slog.NewJSONHandler(os.Stdout, opts).WithAttrs([]slog.Attr{
		slog.String("service", service),
	})

	handler := &TraceHandler{
		Handler: jsonHandler,
	}

	return slog.New(handler)
}

// WithContext appends trace context
func WithContext(ctx context.Context, l *slog.Logger) *slog.Logger {
	if l == nil {
		l = slog.Default()
	}

	span := trace.SpanFromContext(ctx)
	if span != nil && span.SpanContext().IsValid() {
		return l.With(
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
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

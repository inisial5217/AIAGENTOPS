package telemetry

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

type dbSpanKey struct{}

// DBQueryTracer pgx query tracer
type DBQueryTracer struct {
	tracer trace.Tracer
}

// NewDBQueryTracer creates tracer
func NewDBQueryTracer() *DBQueryTracer {
	return &DBQueryTracer{
		tracer: otel.GetTracerProvider().Tracer("cifo-postgres"),
	}
}

// TraceQueryStart starts db span
func (t *DBQueryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	opName := extractOperation(data.SQL)
	spanName := "db.query:" + opName

	ctx, span := t.tracer.Start(ctx, spanName,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.DBSystemPostgreSQL,
			attribute.String("db.operation", opName),
			attribute.String("db.statement", truncateQuery(data.SQL, 256)),
			attribute.String("db.name", "cifo_db"),
		),
	)

	return context.WithValue(ctx, dbSpanKey{}, span)
}

// TraceQueryEnd ends db span
func (t *DBQueryTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	spanVal := ctx.Value(dbSpanKey{})
	if spanVal == nil {
		return
	}

	span, ok := spanVal.(trace.Span)
	if !ok || span == nil {
		return
	}

	if data.Err != nil && data.Err != pgx.ErrNoRows {
		span.RecordError(data.Err)
		span.SetStatus(codes.Error, data.Err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}

	span.End()
}

// extractOperation gets first word
func extractOperation(sql string) string {
	trimmed := strings.TrimSpace(sql)
	if trimmed == "" {
		return "QUERY"
	}
	parts := strings.Fields(trimmed)
	if len(parts) > 0 {
		return strings.ToUpper(parts[0])
	}
	return "QUERY"
}

// truncateQuery limits sql length
func truncateQuery(sql string, maxLen int) string {
	trimmed := strings.TrimSpace(sql)
	if len(trimmed) <= maxLen {
		return trimmed
	}
	return trimmed[:maxLen] + "..."
}

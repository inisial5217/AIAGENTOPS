package telemetry

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
)

// TestInitTracerDisabled verifies disabled state
func TestInitTracerDisabled(t *testing.T) {
	ctx := context.Background()
	shutdown, err := InitTracer(ctx, TracerConfig{
		Enabled: false,
	})

	assert.NoError(t, err)
	assert.NotNil(t, shutdown)
	assert.NoError(t, shutdown(ctx))
}

// TestGetTracer verifies tracer retrieval
func TestGetTracer(t *testing.T) {
	tracer := GetTracer("test-tracer")
	assert.NotNil(t, tracer)
}

// TestDBQueryTracer verifies query tracing
func TestDBQueryTracer(t *testing.T) {
	tracer := NewDBQueryTracer()
	assert.NotNil(t, tracer)

	ctx := context.Background()
	startData := pgx.TraceQueryStartData{
		SQL: "SELECT * FROM users WHERE id = $1",
	}

	newCtx := tracer.TraceQueryStart(ctx, nil, startData)
	assert.NotNil(t, newCtx)

	span := trace.SpanFromContext(newCtx)
	assert.NotNil(t, span)

	endData := pgx.TraceQueryEndData{
		Err: nil,
	}
	tracer.TraceQueryEnd(newCtx, nil, endData)
}

// TestExtractOperation tests op extraction
func TestExtractOperation(t *testing.T) {
	assert.Equal(t, "SELECT", extractOperation("SELECT * FROM table"))
	assert.Equal(t, "INSERT", extractOperation("  insert into table values (1)"))
	assert.Equal(t, "UPDATE", extractOperation("UPDATE table SET x = 1"))
	assert.Equal(t, "QUERY", extractOperation(""))
}

// TestTruncateQuery tests sql truncation
func TestTruncateQuery(t *testing.T) {
	assert.Equal(t, "short query", truncateQuery("short query", 50))
	assert.Equal(t, "long...", truncateQuery("long query here", 4))
}

package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cifo-monitoring/backend/pkg/logger"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	// HeaderXTraceID response header name
	HeaderXTraceID = "X-Trace-Id"
	// HeaderTraceParent W3C header name
	HeaderTraceParent = "traceparent"
)

// TracerMiddleware starts echo span
func TracerMiddleware(serviceName string) echo.MiddlewareFunc {
	tracer := otel.GetTracerProvider().Tracer(serviceName)
	propagator := otel.GetTextMapPropagator()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			ctx := propagator.Extract(req.Context(), propagation.HeaderCarrier(req.Header))

			path := c.Path()
			if path == "" {
				path = req.URL.Path
			}
			spanName := req.Method + " " + path

			ctx, span := tracer.Start(ctx, spanName,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					semconv.HTTPRequestMethodKey.String(req.Method),
					semconv.HTTPRoute(path),
					semconv.URLPath(req.URL.Path),
					semconv.ClientAddress(c.RealIP()),
					semconv.UserAgentOriginal(req.UserAgent()),
				),
			)
			defer span.End()

			sc := span.SpanContext()
			traceID := ""
			spanID := ""
			if sc.IsValid() {
				traceID = sc.TraceID().String()
				spanID = sc.SpanID().String()
			} else {
				// extract or fallback
				if tp := req.Header.Get(HeaderTraceParent); tp != "" {
					parts := strings.Split(tp, "-")
					if len(parts) >= 3 {
						traceID = parts[1]
						spanID = parts[2]
					}
				}
				if traceID == "" {
					traceID = strings.ReplaceAll(uuid.New().String(), "-", "")
					spanID = traceID[:16]
				}
			}

			c.Response().Header().Set(HeaderXTraceID, traceID)
			c.Response().Header().Set(HeaderTraceParent, fmt.Sprintf("00-%s-%s-01", traceID, spanID))

			// pass trace context
			ctx = logger.ContextWithTrace(ctx, traceID, spanID)
			c.SetRequest(req.WithContext(ctx))
			c.Set("trace_id", traceID)
			c.Set("span_id", spanID)

			err := next(c)
			status := c.Response().Status

			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				span.SetAttributes(semconv.HTTPResponseStatusCode(status))
				return err
			}

			if status >= http.StatusInternalServerError {
				span.SetStatus(codes.Error, fmt.Sprintf("http status %d", status))
			} else {
				span.SetStatus(codes.Ok, "")
			}
			span.SetAttributes(semconv.HTTPResponseStatusCode(status))

			return nil
		}
	}
}

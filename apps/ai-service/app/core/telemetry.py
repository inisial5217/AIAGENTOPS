import logging
from typing import Any
from app.config.settings import settings

logger = logging.getLogger("cifo-ai-service")

try:
    from opentelemetry import trace, propagate
    from opentelemetry.trace.propagation.tracecontext import TraceContextTextMapPropagator
    from opentelemetry.sdk.trace import TracerProvider
    from opentelemetry.sdk.trace.export import BatchSpanProcessor
    from opentelemetry.sdk.resources import Resource, SERVICE_NAME, SERVICE_VERSION
    from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
    OTEL_AVAILABLE = True
except ImportError:
    OTEL_AVAILABLE = False


_tracer: Any = None


def init_telemetry() -> None:
    # initialize otel tracer
    global _tracer
    if not OTEL_AVAILABLE:
        logger.warning("OpenTelemetry packages not available; running in no-op tracing mode")
        return

    if not settings.otel_enabled:
        return

    try:
        resource = Resource.create({
            SERVICE_NAME: "cifo-ai-service",
            SERVICE_VERSION: "1.0.0",
            "deployment.environment": settings.environment,
        })

        provider = TracerProvider(resource=resource)
        endpoint = settings.otel_endpoint
        if not endpoint.endswith("/v1/traces"):
            endpoint = endpoint.rstrip("/") + "/v1/traces"

        exporter = OTLPSpanExporter(endpoint=endpoint)
        processor = BatchSpanProcessor(exporter)
        provider.add_span_processor(processor)

        trace.set_tracer_provider(provider)
        propagate.set_global_textmap(TraceContextTextMapPropagator())

        _tracer = trace.get_tracer("cifo-ai-service", "1.0.0")
        logger.info("OpenTelemetry tracer initialized successfully (endpoint: %s)", endpoint)
    except Exception as exc:
        logger.warning("Failed to initialize OpenTelemetry tracer: %s", exc)


def get_tracer() -> Any:
    # return active tracer
    global _tracer
    if _tracer is not None:
        return _tracer
    if OTEL_AVAILABLE:
        return trace.get_tracer("cifo-ai-service")
    return None


def extract_trace_context(headers: dict[str, str]) -> Any:
    # extract w3c trace context
    if not OTEL_AVAILABLE:
        return None
    try:
        propagator = TraceContextTextMapPropagator()
        return propagator.extract(carrier=headers)
    except Exception:
        return None


def get_current_trace_ids() -> tuple[str | None, str | None]:
    # get active trace and span ids
    if not OTEL_AVAILABLE:
        return None, None
    try:
        current_span = trace.get_current_span()
        if not current_span:
            return None, None
        ctx = current_span.get_span_context()
        if not ctx or not ctx.is_valid:
            return None, None
        trace_id = format(ctx.trace_id, "032x")
        span_id = format(ctx.span_id, "016x")
        return trace_id, span_id
    except Exception:
        return None, None

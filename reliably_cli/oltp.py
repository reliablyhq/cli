from contextlib import contextmanager

import pkg_resources

try:
    from opentelemetry import trace  # type: ignore
    from opentelemetry.exporter.otlp.proto.http.trace_exporter import (
        OTLPSpanExporter,
    )
    from opentelemetry.instrumentation.httpx import (
        HTTPXClientInstrumentor,
        RequestInfo,
    )
    from opentelemetry.instrumentation.logging import LoggingInstrumentor
    from opentelemetry.sdk.resources import Resource
    from opentelemetry.sdk.trace import TracerProvider
    from opentelemetry.sdk.trace.export import BatchSpanProcessor
    from opentelemetry.trace.span import NonRecordingSpan, Span
except pkg_resources.DistributionNotFound:
    pass

from . import is_executable
from .config import Settings

__all__ = ["configure_instrumentation", "oltp_span"]


def configure_instrumentation(settings: Settings) -> None:  # pragma: no cover
    if is_executable():
        return

    collector_endpoint = settings.otel.endpoint

    headers = {}
    if settings.otel.headers:
        for s in settings.otel.headers.split(","):
            k, v = s.split("=", 1)
            headers[k] = v

    resource = Resource(attributes={"service.name": settings.otel.service_name})

    provider = TracerProvider(resource=resource)
    exporter = OTLPSpanExporter(endpoint=collector_endpoint, headers=headers)
    processor = BatchSpanProcessor(exporter)
    provider.add_span_processor(processor)
    trace.set_tracer_provider(provider)
    trace.get_tracer(__name__)

    LoggingInstrumentor().instrument(
        tracer_provider=provider, set_logging_format=False
    )

    def request_oltp_hook(span: Span, request: RequestInfo) -> None:
        if span and span.is_recording():
            _, _, headers, _, _ = request
            org_id = headers.get("X-Reliably-Org-Id")
            if org_id:
                span.set_attribute("reliably.org_id", org_id)

    HTTPXClientInstrumentor().instrument(request_hook=request_oltp_hook)


@contextmanager
def oltp_span(
    name: str, settings: Settings, attrs: dict[str, str] = None
) -> Span:
    if is_executable() or not settings.otel or not settings.otel.enabled:
        yield NonRecordingSpan(None)
        return

    attrs = attrs or {}
    attrs["reliably.org_id"] = str(settings.organization.id)
    attrs["reliably.agent_id"] = str(settings.agent.id)

    tracer = trace.get_tracer(__name__)
    with tracer.start_as_current_span(
        name, attributes=attrs, record_exception=True, end_on_exit=True
    ) as span:
        yield span

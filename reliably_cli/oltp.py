from contextlib import contextmanager

import pkg_resources

try:
    from opentelemetry import metrics, trace  # type: ignore
    from opentelemetry.exporter.otlp.proto.http.metric_exporter import (
        OTLPMetricExporter,
    )
    from opentelemetry.exporter.otlp.proto.http.trace_exporter import (
        OTLPSpanExporter,
    )
    from opentelemetry.exporter.prometheus import PrometheusMetricReader
    from opentelemetry.instrumentation.httpx import HTTPXClientInstrumentor
    from opentelemetry.instrumentation.logging import LoggingInstrumentor
    from opentelemetry.instrumentation.system_metrics import (
        SystemMetricsInstrumentor,
    )
    from opentelemetry.sdk.metrics import MeterProvider
    from opentelemetry.sdk.metrics.export import PeriodicExportingMetricReader
    from opentelemetry.sdk.resources import Resource
    from opentelemetry.sdk.trace import TracerProvider
    from opentelemetry.sdk.trace.export import BatchSpanProcessor
    from opentelemetry.trace.span import NonRecordingSpan, Span
except pkg_resources.DistributionNotFound:
    NonRecordingSpan = Span = None

from . import is_executable
from .config.types import Settings
from .log import logger

__all__ = [
    "configure_traces",
    "oltp_span",
    "configure_metrics",
    "inc_metric_value",
]
METRICS = {}


def configure_traces(settings: Settings) -> None:  # pragma: no cover
    if is_executable():
        return

    if not settings.otel.traces.enabled:
        logger.debug("Open Telemetry traces not enabled")
        return

    collector_endpoint = settings.otel.traces.endpoint

    headers = {}
    if settings.otel.traces.headers:
        for s in settings.otel.traces.headers.split(","):
            k, v = s.split("=", 1)
            headers[k] = v

    resource = Resource(
        attributes={"service.name": settings.otel.traces.service_name}
    )

    provider = TracerProvider(resource=resource)
    exporter = OTLPSpanExporter(endpoint=collector_endpoint, headers=headers)
    processor = BatchSpanProcessor(exporter)
    provider.add_span_processor(processor)
    trace.set_tracer_provider(provider)
    trace.get_tracer(__name__)

    LoggingInstrumentor().instrument(
        tracer_provider=provider, set_logging_format=False
    )
    HTTPXClientInstrumentor().instrument()


def configure_metrics(settings: Settings) -> None:
    if is_executable():
        return

    if not settings.otel.metrics.enabled:
        logger.debug("Open Telemetry metrics not enabled")
        return

    if settings.otel.metrics.expose_as_prometheus:
        reader = PrometheusMetricReader(settings.otel.metrics.service_name)
    else:
        collector_endpoint = settings.otel.metrics.endpoint
        headers = {}
        if settings.otel.metrics.headers:
            for s in settings.otel.metrics.headers.split(","):
                k, v = s.split("=", 1)
                headers[k] = v

        exporter = OTLPMetricExporter(
            endpoint=collector_endpoint, headers=headers
        )

        reader = PeriodicExportingMetricReader(
            exporter, export_interval_millis=30000
        )

    resource = Resource(
        attributes={"service.name": settings.otel.metrics.service_name}
    )
    SystemMetricsInstrumentor().instrument()

    provider = MeterProvider(resource=resource, metric_readers=[reader])
    metrics.set_meter_provider(provider)

    meter = provider.get_meter(__name__)

    METRICS["scheduled-plans"] = meter.create_counter(
        name="scheduled_plans", description="number of scheduled plans"
    )


@contextmanager
def oltp_span(
    name: str, settings: Settings, attrs: dict[str, str] = None
) -> Span | NonRecordingSpan:
    if is_executable() or not settings.otel.traces.enabled:
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


def inc_metric_value(
    metric_name: str, attrs: dict[str, str] | None = None
) -> None:
    if not METRICS:
        return

    METRICS[metric_name].add(1, attributes=attrs)

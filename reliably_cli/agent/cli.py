import signal
from typing import Awaitable

import anyio
import trio
import typer
from anyio.abc import CancelScope
from hypercorn.config import Config
from hypercorn.trio import serve
from prometheus_client.exposition import _bake_output
from prometheus_client.registry import REGISTRY
from starlette.applications import Starlette
from starlette.requests import Request
from starlette.responses import Response
from starlette.routing import Route

from ..config import get_settings
from ..log import logger
from . import agent_runner, validate_agent_configuration

cli = typer.Typer()


@cli.command()
def run() -> None:
    try:
        validate_agent_configuration()
    except ValueError:
        raise typer.Exit(1)

    trio.run(_main, agent_runner)


##############################################################################
# Private
##############################################################################
async def signal_handler(scope: CancelScope):
    with anyio.open_signal_receiver(
        signal.SIGINT, signal.SIGTERM, signal.SIGHUP
    ) as signals:
        logger.debug("Listening for system signals...")
        async for signum in signals:
            logger.debug(
                f"Signal '{signal.strsignal(signum) or signum}' received"
            )
            if signum == signal.SIGHUP:
                logger.info("Reloading configuration")
                get_settings.cache_clear()
                get_settings()
                logger.info("Configuration reloaded")
            else:
                scope.cancel()
                return


async def _main(runner: Awaitable) -> None:
    logger.debug("Agent ready to run")

    settings = get_settings()
    otel_metrics = settings.otel.metrics

    async with anyio.create_task_group() as tg:
        tg.start_soon(signal_handler, tg.cancel_scope)
        tg.start_soon(runner, tg.cancel_scope)

        if otel_metrics.enabled and otel_metrics.expose_as_prometheus:
            tg.start_soon(_with_prometheus_server)

    logger.debug("Agent terminated. Exiting...")


async def _with_prometheus_server() -> None:
    config = Config.from_mapping(
        {"bind": "0.0.0.0:45380", "loglevel": "NOTSET"}
    )

    class Prometheus(Starlette):
        def __init__(self):
            super().__init__(routes=[Route("/metrics", self._index)])

        async def _index(self, request: Request) -> Response:
            status, headers, output = _bake_output(
                REGISTRY,
                request.headers.get("Accept"),
                request.headers.get("Accept-Encoding"),
                request.query_params,
                False,
            )

            return Response(
                content=output,
                status_code=int(status.split(" ")[0]),
                headers=dict(headers),
            )

    logger.debug("Prometheus endpoint at http://0.0.0.0:45380/metrics")
    await serve(Prometheus(), config)

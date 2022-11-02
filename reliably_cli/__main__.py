import signal
from pathlib import Path
from typing import Awaitable

import anyio
import trio
import typer
from anyio.abc import CancelScope

from .__version__ import __version__
from .agent import agent_runner, validate_agent_configuration
from .config import Settings, get_settings
from .log import configure_logger, logger
from .oltp import configure_instrumentation

cli = typer.Typer()


@cli.callback()
def main(
    config: Path = typer.Option(
        Path(typer.get_app_dir("reliably")) / "config.toml",
        envvar="RELIABLY_CLI_CONFIG",
    ),
):
    Settings.Config.toml_file = config
    configure_app()


@cli.command()
def agent() -> None:
    try:
        validate_agent_configuration()
    except ValueError as ve:
        print(ve)
        raise typer.Exit(1)

    trio.run(_main, agent_runner)


@cli.command()
def version(
    short: bool = typer.Option(
        False, "--short", "-s", help="Only shows the version string"
    )
) -> None:
    if short:
        print(__version__)
    else:
        print(f"Reliably CLI: {__version__}")


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
    logger.debug("Application ready to run")

    async with anyio.create_task_group() as tg:
        tg.start_soon(signal_handler, tg.cancel_scope)
        tg.start_soon(runner, tg.cancel_scope)

    logger.debug("Application terminated. Exiting...")


def configure_app() -> None:
    settings = get_settings()

    configure_logger(settings)

    logger.debug("Configuring application...")

    if settings.otel and settings.otel.enabled:
        configure_instrumentation(settings)

    logger.debug("Application configured")


if __name__ == "__main__":
    cli()

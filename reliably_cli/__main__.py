from pathlib import Path

import typer

from .__version__ import __version__
from .agent.cli import cli as agent_cli
from .config import get_settings
from .config.cli import cli as config_cli
from .config.types import Settings
from .log import configure_logger, logger
from .oltp import configure_metrics, configure_traces

cli = typer.Typer()
cli.add_typer(config_cli, name="config")
cli.add_typer(agent_cli, name="agent")


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
def configure_app() -> None:
    settings = get_settings()

    configure_logger(settings)

    logger.debug("Configuring application...")

    if settings.otel:
        configure_traces(settings)
        configure_metrics(settings)

    logger.debug("Application configured")


if __name__ == "__main__":
    cli()

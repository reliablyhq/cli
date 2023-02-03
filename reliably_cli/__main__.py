from pathlib import Path

import typer

from .__version__ import __version__
from .agent.cli import cli as agent_cli
from .config.cli import cli as config_cli
from .config.types import Settings
from .log import console
from .services.cli import cli as services_cli

cli = typer.Typer()
cli.add_typer(config_cli, name="config")
cli.add_typer(agent_cli, name="agent")
cli.add_typer(services_cli, name="service")


@cli.callback()
def main(
    config: Path = typer.Option(
        Path(typer.get_app_dir("reliably")) / "config.toml",
        envvar="RELIABLY_CLI_CONFIG",
    ),
):
    Settings.Config.toml_file = config


@cli.command()
def version(
    short: bool = typer.Option(
        False, "--short", "-s", help="Only shows the version string"
    )
) -> None:
    if short:
        console.print(__version__)
    else:
        console.print(f"Reliably CLI: {__version__}")


if __name__ == "__main__":
    cli()

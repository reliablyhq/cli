import os
from enum import Enum
from pathlib import Path

import typer
from rich import print as print_

from . import get_settings

cli = typer.Typer(help="View and read Reliably configuration entries")


class EntryType(str, Enum):
    token = "token"
    org = "org"
    agent = "agent"
    host = "host"


@cli.command(help="Show the current configuration content")
def view() -> None:
    encoding = "utf-8"
    toml_file = os.getenv("RELIABLY_CLI_CONFIG")
    if not toml_file:
        settings = get_settings()
        toml_file = settings.__config__.toml_file
        encoding = settings.__config__.env_file_encoding

    data = ""
    if Path(toml_file).exists():
        data = Path(toml_file).read_text(encoding=encoding)

    print_(data, end="")


@cli.command(help="Read one entry of the configuration")
def get(entry: EntryType) -> None:
    settings = get_settings()

    value = ""
    match entry.value:  # noqa
        case "token":
            value = settings.agent.token.get_secret_value()
        case "org":
            value = str(settings.organization.id)
        case "host":
            value = settings.service.host

    print_(value)


@cli.command(help="Create an empty configuration file")
def create() -> None:
    toml_file = os.getenv("RELIABLY_CLI_CONFIG")
    if not toml_file:
        settings = get_settings()
        toml_file = settings.__config__.toml_file

    Path(toml_file).touch(exist_ok=True)


@cli.command(help="Checks the configuration's presence")
def exists() -> None:
    toml_file = os.getenv("RELIABLY_CLI_CONFIG")
    if not toml_file:
        settings = get_settings()
        toml_file = settings.__config__.toml_file

    print("yes" if Path(toml_file).exists() else "no")


@cli.command(help="Get the path of the configuration")
def path() -> str:
    toml_file = os.getenv("RELIABLY_CLI_CONFIG")
    if not toml_file:
        settings = get_settings()
        toml_file = settings.__config__.toml_file

    print(Path(toml_file).absolute())

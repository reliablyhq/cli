import os
from enum import Enum
from pathlib import Path

import typer
from rich import print as print_

from . import get_settings

cli = typer.Typer()


class EntryType(str, Enum):
    token = "token"
    org = "org"
    agent = "agent"
    host = "host"


@cli.command()
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


@cli.command()
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

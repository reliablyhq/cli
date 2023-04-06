import contextlib
import json
import logging
import logging.handlers
import os
import traceback
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Generator
from uuid import UUID

import httpx
import typer
from chaoslib import convert_vars, merge_vars
from chaoslib.control import load_global_controls
from chaoslib.experiment import ensure_experiment_is_valid, run_experiment
from chaoslib.settings import CHAOSTOOLKIT_CONFIG_PATH, load_settings
from chaoslib.types import Journal, Schedule, Strategy
from ruamel.yaml import YAML

from ..client import api_client
from ..config import ensure_config_is_set, get_settings
from ..format import format_as
from ..log import console
from ..types import FormatOption, Plan

cli = typer.Typer(help="Manage and execute Reliably plans from your terminal")


@cli.command()
def get(
    plan_id: UUID,
    format: FormatOption = typer.Option("json", show_choices=True),
) -> None:
    """
    Get a plan and display it
    """
    ensure_config_is_set()
    p = load_plan(plan_id)
    console.print(format_as(p, format))


@cli.command(name="store-context")
def store_context(plan_id: UUID) -> None:
    """
    Store a plan context so the execution has what it needs to operate
    """
    ensure_config_is_set()
    p = load_plan(plan_id)
    store_plan_context(p)


def validate_vars(ctx: typer.Context, value: list[str]) -> dict[str, Any]:
    try:
        return convert_vars(value)
    except ValueError as x:
        raise typer.BadParameter(str(x))


@cli.command()
def execute(
    plan_id: UUID,
    result_file: Path = typer.Option("./result.json", writable=True),
    log_stdout: bool = typer.Option(False, is_flag=True),
    log_file: Path = typer.Option("./run.log", writable=True),
    set_status: bool = typer.Option(False, is_flag=True),
    skip_context: bool = typer.Option(False, is_flag=True),
    control_file: list[Path] = typer.Option(
        lambda: [], dir_okay=False, readable=True
    ),
    var_file: list[Path] = typer.Option(
        lambda: [], dir_okay=False, readable=True
    ),
) -> None:
    """
    Execute a plan
    """
    ensure_config_is_set()

    p = load_plan(plan_id)

    context = {}
    if not skip_context:
        context = store_plan_context(p)

    settings = get_settings()

    if os.getenv("RELIABLY_HOST") is None:
        os.environ["RELIABLY_HOST"] = settings.service.host.replace(
            "https://", ""
        )

    token = settings.service.token.get_secret_value()
    if os.getenv("RELIABLY_TOKEN") is None:
        os.environ["RELIABLY_TOKEN"] = token

    if os.getenv("RELIABLY_PLAN_ID") is None:
        os.environ["RELIABLY_PLAN_ID"] = str(p.id)

    experiment_id = p.definition.experiments[0]
    base_url = f"{settings.service.host}/api/v1/organization"
    base_url = f"{base_url}/{settings.organization.id}"
    experiment_url = f"{base_url}/experiments/{experiment_id}/raw"

    with console.status("Executing..."):
        if set_status:
            send_status(p.id, "running")

        with reconfigure_chaostoolkit_logger(log_file, log_stdout):
            try:
                journal = run_chaostoolkit(
                    experiment_url, context, var_file, control_file
                )
            except Exception as x:
                if set_status:
                    send_status(p.id, "error", f"Error: {x}")

                tb = "".join(traceback.format_exception(x))
                console.print(f"running experiment failed: {tb}")

                raise typer.Exit(code=1)

            result_file.absolute().write_text(json.dumps(journal, indent=2))
            show_result_url(journal)


###############################################################################
# Private functions
###############################################################################
def load_plan(plan_id: UUID) -> Plan:
    with console.status("Fetching plan..."):
        with api_client() as client:
            r = client.get(f"/plans/{str(plan_id)}")
            if r.status_code == 404:
                console.print("plan not found")
                raise typer.Exit(code=1)
            elif r.status_code == 401:
                console.print("not authorized. please verify your token or org")
                raise typer.Exit(code=1)
            elif r.status_code > 399:
                console.print(f"unexpected error: {r.status_code}: {r.json()}")
                raise typer.Exit(code=1)

            return Plan.parse_obj(r.json())


def store_plan_context(plan: Plan) -> dict[str, Any]:
    global_controls = {}

    with console.status("Storing context..."):
        with api_client() as client:
            for int_id in plan.definition.integrations:
                r = client.get(f"/integrations/{int_id}/control")
                control = r.json()
                if control:
                    ctrl_name = control.pop("name")
                    provider = control.get("provider", {})
                    if "secrets" in provider and provider["secrets"] is None:
                        del provider["secrets"]
                    if (
                        "arguments" in provider
                        and provider["arguments"] is None
                    ):
                        del provider["arguments"]
                    ctrl = {ctrl_name: control}
                    global_controls.update(ctrl)

    return global_controls


def send_status(plan_id: UUID, status: str, error: str | None = None) -> None:
    with api_client() as client:
        payload = {"status": status, "error": error}
        r = client.put(f"/plans/{str(plan_id)}/status", json=payload)
        if r.status_code == 404:
            console.print("plan not found")
            raise typer.Exit(code=1)
        elif r.status_code == 401:
            console.print("not authorized. please verify your token or org")
            raise typer.Exit(code=1)
        elif r.status_code > 399:
            console.print(f"unexpected error: {r.status_code}: {r.json()}")
            raise typer.Exit(code=1)


def run_chaostoolkit(
    experiment_url: str,
    context: dict[str, Any],
    var_files: list[Path],
    control_files: list[Path],
) -> Journal:
    logger = logging.getLogger("logzero_default")

    logger.info("#" * 80)
    logger.info(f"Starting Reliably experiment: {experiment_url}")

    settings = {
        "runtime": {
            "hypothesis": {"strategy": "default"},
            "rollbacks": {"strategy": "always"},
        },
        "controls": {},
    }
    settings_path = os.getenv(
        "CHAOSTOOLKIT_CONFIG_PATH", CHAOSTOOLKIT_CONFIG_PATH
    )
    if os.path.isfile(settings_path):
        settings = load_settings(settings_path)

    if "controls" not in settings:
        settings["controls"] = {}
    settings["controls"].update(context)

    experiment_vars = merge_vars({}, [str(f.absolute()) for f in var_files])

    load_global_controls(settings, [str(f.absolute()) for f in control_files])
    experiment = load_experiment(experiment_url)
    ensure_experiment_is_valid(experiment)

    rt = settings["runtime"]
    x_runtime = experiment.get("runtime")
    if x_runtime:
        rt.setdefault("rollbacks", {})["strategy"] = (
            x_runtime.get("rollbacks", {}).get("strategy") or "always"
        )
        rt.setdefault("hypothesis", {})["strategy"] = (
            x_runtime.get("hypothesis", {}).get("strategy") or "default"
        )

    schedule = Schedule(continuous_hypothesis_frequency=1.0, fail_fast=True)
    ssh_strategy = Strategy.from_string(rt["hypothesis"]["strategy"])

    journal = run_experiment(
        experiment,
        settings=settings,
        strategy=ssh_strategy,
        schedule=schedule,
        experiment_vars=experiment_vars,
    )

    return journal


@contextlib.contextmanager
def reconfigure_chaostoolkit_logger(
    log_file: Path,
    log_stdout: bool,
) -> Generator[logging.Logger, None, None]:
    ctk_logger = logging.getLogger("logzero_default")

    for handler in list(ctk_logger.handlers):
        ctk_logger.removeHandler(handler)

    fmt = UTCFormatter(
        fmt="[%(asctime)s %(levelname)s] [%(module)s:%(lineno)d] %(message)s",
    )

    handler = logging.FileHandler(log_file.absolute())
    handler.setLevel(logging.DEBUG)
    handler.setFormatter(fmt)
    ctk_logger.addHandler(handler)

    if log_stdout:
        handler = logging.StreamHandler()
        handler.setLevel(logging.DEBUG)
        handler.setFormatter(fmt)
        ctk_logger.addHandler(handler)

    yield ctk_logger


class UTCFormatter(logging.Formatter):
    def formatTime(
        self, record: logging.LogRecord, datefmt: str | None = None
    ) -> str:
        return datetime.fromtimestamp(
            record.created, tz=timezone.utc
        ).isoformat()


def show_result_url(journal: Journal) -> None:
    for x in journal.get("experiment", {}).get("extensions", []):
        if x["name"] == "reliably":
            url = x.get("execution_url")
            if url:
                console.print(f"Check results at {url}")


def load_experiment(url: str) -> dict[str, Any]:
    token = os.getenv("RELIABLY_TOKEN")
    with httpx.Client(http2=True, verify=True) as c:
        r = c.get(
            url,
            headers={
                "Authorization": f"Bearer {token}",
                "Accept": "application/json, application/x-yaml",
            },
        )

        if r.status_code != 200:
            console.print(f"failed to fetch experiment at {url}")
            raise typer.Exit(code=1)

        if r.headers["content-type"] == "application/json":
            return r.json()
        elif r.headers["content-type"] in ("text/yaml", "application/x-yaml"):
            yaml = YAML(typ="safe")
            return yaml.load(r.content.decode("utf-8"))

        console.print("unrecognized experiment format")
        raise typer.Exit(code=1)

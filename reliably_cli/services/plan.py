import contextlib
import json
import logging
import logging.handlers
import os
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Generator
from uuid import UUID

import typer
from chaoslib.control import load_global_controls
from chaoslib.experiment import ensure_experiment_is_valid, run_experiment
from chaoslib.loader import load_experiment
from chaoslib.types import Journal, Schedule, Strategy

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


@cli.command()
def execute(
    plan_id: UUID,
    result_file: Path = typer.Option("./result.json", writable=True),
    log_file: Path = typer.Option("./run.log", writable=True),
    skip_context: bool = typer.Option(False, is_flag=True),
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
    if os.getenv("CHAOSTOOLKIT_LOADER_AUTH_BEARER_TOKEN") is None:
        os.environ["CHAOSTOOLKIT_LOADER_AUTH_BEARER_TOKEN"] = token

    if os.getenv("RELIABLY_TOKEN") is None:
        os.environ["RELIABLY_TOKEN"] = token

    if os.getenv("RELIABLY_PLAN_ID") is None:
        os.environ["RELIABLY_PLAN_ID"] = str(p.id)

    experiment_id = p.definition.experiments[0]
    base_url = f"{settings.service.host}/api/v1/organization"
    base_url = f"{base_url}/{settings.organization.id}"
    experiment_url = f"{base_url}/experiments/{experiment_id}/raw"

    with console.status("Executing..."):
        with reconfigure_chaostoolkit_logger(log_file):
            journal = run_chaostoolkit(experiment_url, context)
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


def run_chaostoolkit(experiment_url: str, context: dict[str, Any]) -> Journal:
    logger = logging.getLogger("logzero_default")

    logger.info("#" * 80)
    logger.info(f"Starting Reliably experiment: {experiment_url}")

    settings = {
        "runtime": {
            "hypothesis": {"strategy": "default"},
            "rollbacks": {"strategy": "always"},
        },
        "controls": context,
    }

    load_global_controls(settings)
    experiment = load_experiment(experiment_url, settings, verify_tls=True)
    ensure_experiment_is_valid(experiment)

    x_runtime = experiment.get("runtime")
    if x_runtime:
        settings["runtime"]["rollbacks"]["strategy"] = x_runtime.get(
            "rollbacks", {}
        ).get("strategy", "always")
        settings["runtime"]["hypothesis"]["strategy"] = x_runtime.get(
            "hypothesis", {}
        ).get("strategy", "default")

    schedule = Schedule(continuous_hypothesis_frequency=1.0, fail_fast=True)
    experiment_vars = ({}, {})
    ssh_strategy = Strategy.DEFAULT

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

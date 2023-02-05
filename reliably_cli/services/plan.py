import json
import shutil
import subprocess
from pathlib import Path
from uuid import UUID

import typer

from ..client import api_client
from ..config import get_settings
from ..config.types import get_settings_directory_path
from ..format import format_as
from ..log import console
from ..types import FormatOption, Plan

cli = typer.Typer()


@cli.command()
def get(
    plan_id: UUID,
    format: FormatOption = typer.Option("json", show_choices=True),
) -> None:
    """
    Get a plan and display it
    """
    p = load_plan(plan_id)
    console.print(format_as(p, format))


@cli.command(name="store-context")
def store_context(plan_id: UUID) -> None:
    """
    Store a plan context so the execution has what it needs to operate
    """
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
    settings_dir = get_settings_directory_path()
    p = load_plan(plan_id)

    if not skip_context:
        store_plan_context(p)

    plan_dir = settings_dir / "plans" / str(p.id)
    ctk_settings_file = str(plan_dir / "ctk.yaml")

    args = [
        shutil.which("chaos"),
        "--no-version-check",
        "--verbose",
        "--log-file",
        log_file,
        "--log-file-level",
        "debug",
        "--settings",
        ctk_settings_file,
        "run",
        "--journal-path",
        str(result_file),
    ]

    for int_id in p.definition.integrations:
        args.extend(
            (
                "--control-file",
                str(plan_dir / "controls" / f"{str(int_id)}.json"),
            )
        )

    settings = get_settings()

    experiment_id = p.definition.experiments[0]
    base_url = f"{settings.service.host}/api/v1/organization"
    base_url = f"{base_url}/{settings.organization.id}"

    args.append(f"{base_url}/experiments/{experiment_id}/raw")

    with console.status("Executing..."):
        try:
            subprocess.run(args, capture_output=True, check=True)
        except subprocess.CalledProcessError as x:
            console.print(f"failed to execute plan. see {log_file.absolute()}")
            raise typer.Exit(code=x.returncode)


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
            elif r.status_code > 399:
                console.print("unexpected YYYYY error: ", r.json())
                raise typer.Exit(code=1)

            return Plan.parse_obj(r.json())


def store_plan_context(plan: Plan) -> None:
    settings_dir = get_settings_directory_path()

    plan_dir = settings_dir / "plans" / str(plan.id)
    plan_dir.mkdir(parents=True, exist_ok=True)

    (plan_dir / "ctk.yaml").write_text(
        "runtime:\n  rollbacks:\n    strategy: always\n"
    )

    with console.status("Storing context..."):
        with api_client() as client:
            for int_id in plan.definition.integrations:
                r = client.get(f"/integrations/{int_id}/control")
                control = r.json()
                if control:
                    ctrl_dir = plan_dir / "controls"
                    ctrl_dir.mkdir(parents=True, exist_ok=True)

                    ctrl_name = control.pop("name")
                    provider = control.get("provider", {})
                    if provider.get("secrets") is None:
                        del provider["secrets"]
                    if provider.get("arguments") is None:
                        del provider["arguments"]
                    ctrl = {ctrl_name: control}
                    (ctrl_dir / f"{str(int_id)}.json").write_text(
                        json.dumps(ctrl)
                    )

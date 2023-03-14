import typer

from .plan import cli as plan_cli

cli = typer.Typer(help="Interact with Reliably services")
cli.add_typer(plan_cli, name="plan")

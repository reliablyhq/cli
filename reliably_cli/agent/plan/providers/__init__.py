from typing import Any

from reliably_cli.agent.types import Plan
from reliably_cli.client import reliably_client
from reliably_cli.config import Settings
from reliably_cli.log import logger

__all__ = ["fetch_deployment", "fetch_experiment"]


async def fetch_deployment(
    plan: Plan, settings: Settings
) -> dict[str, Any] | None:
    dep_id = plan.definition.deployment.deployment_id

    async with reliably_client() as client:
        r = await client.get(f"/deployments/{dep_id}")
        if r.status_code > 399:
            logger.error(
                f"error fetching deployment {dep_id} for plan {plan.id}"
            )
            return

        return r.json()


async def fetch_experiment(
    plan: Plan, settings: Settings
) -> dict[str, Any] | None:
    exp_id = plan.definition.experiments[0]

    async with reliably_client() as client:
        r = await client.get(f"/experiments/{exp_id}")
        if r.status_code > 399:
            logger.error(
                f"error fetching experiment {exp_id} for plan {plan.id}"
            )
            return

        return r.json()

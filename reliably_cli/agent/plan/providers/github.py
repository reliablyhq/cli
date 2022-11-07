import os

import httpx
from anyio.abc import CancelScope

from reliably_cli.agent.types import Plan
from reliably_cli.config import get_settings
from reliably_cli.log import logger
from reliably_cli.oltp import oltp_span

from . import fetch_deployment

__all__ = ["schedule_plan"]


async def schedule_plan(plan: Plan) -> None:
    with CancelScope(shield=True) as scope:
        try:
            settings = get_settings()
            with oltp_span(
                "schedule-plan",
                settings=settings,
                attrs={"plan_id": str(plan.id), "deployment_type": "github"},
            ):
                logger.info(f"Schedule plan {plan.id} on GitHub")

                deployment = await fetch_deployment(plan, settings)
                gh_token = deployment["definition"]["token"]
                if not gh_token:
                    gh_token = os.getenv("GITHUB_TOKEN")

                exp_id = plan.definition.experiments[0]
                experiment_url = (
                    f"{settings.service.host}/api/v1"
                    f"/organization/{settings.organization.id}"
                    f"/experiments/{exp_id}/raw"
                )

                gh_api_url = settings.plan.providers.github.api_url
                gh_repo = settings.plan.providers.github.repo
                gh_workflow_id = settings.plan.providers.github.workflow_id
                gh_ref = settings.plan.providers.github.ref

                url = (
                    f"{gh_api_url}/repos/{gh_repo}/actions"
                    f"/workflows/{gh_workflow_id}/dispatches"
                )

                async with httpx.AsyncClient() as h:
                    r = await h.post(
                        url,
                        headers={
                            "Accept": "application/vnd.github+json",
                            "Authorization": f"Bearer {gh_token}",
                        },
                        json={
                            "ref": gh_ref,
                            "inputs": {"experiment-url": experiment_url},
                        },
                    )
                    logger.info(r.status_code)
                    logger.info(r.json())
        finally:
            scope.cancel()

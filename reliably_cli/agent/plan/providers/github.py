import os

import httpx
from anyio.abc import CancelScope

from reliably_cli.agent.types import Plan
from reliably_cli.config import get_settings
from reliably_cli.log import logger
from reliably_cli.oltp import oltp_span

__all__ = ["schedule_plan"]


async def schedule_plan(plan: Plan) -> None:
    plan_id = str(plan.id)

    gh_attrs = {
        "reliably.plan_id": plan_id,
        "reliably.deployment_type": "github",
        "reliably.gh_actor": os.getenv("GITHUB_ACTOR"),
        "reliably.gh_event_name": os.getenv("GITHUB_EVENT_NAME"),
        "reliably.gh_job_id": os.getenv("GITHUB_JOB"),
        "reliably.gh_ref": os.getenv("GITHUB_REF"),
        "reliably.gh_ref_type": os.getenv("GITHUB_REF_TYPE"),
        "reliably.gh_repo": os.getenv("GITHUB_REPOSITORY"),
        "reliably.gh_run_id": os.getenv("GITHUB_RUN_ID"),
        "reliably.gh_sha": os.getenv("GITHUB_SHA"),
    }

    with CancelScope(shield=True) as scope:
        try:
            settings = get_settings()
            with oltp_span(
                "schedule-plan",
                settings=settings,
                attrs=gh_attrs,
            ):
                logger.info(f"Schedule plan {plan.id} on GitHub")

                gh_token = os.getenv("GITHUB_TOKEN")

                if not gh_token:
                    raise RuntimeError(
                        "you must specify a suitable GitHub token, either in "
                        "the Reliably configuration file or via the "
                        "GITHUB_TOKEN environment variable"
                    )

                exp_id = plan.definition.experiments[0]
                experiment_url = (
                    f"{settings.service.host}/api/v1"
                    f"/organization/{settings.organization.id}"
                    f"/experiments/{exp_id}/raw"
                )
                logger.debug(f"Scheduling experiment {experiment_url}")

                gh_workflow_id = settings.plan.providers.github.workflow_id
                gh_api_url = os.getenv(
                    "GITHUB_API_URL", settings.plan.providers.github.api_url
                )
                gh_repo = os.getenv(
                    "GITHUB_REPOSITORY", settings.plan.providers.github.repo
                )
                gh_ref = os.getenv(
                    "GITHUB_REF_NAME", settings.plan.providers.github.ref
                )

                url = (
                    f"{gh_api_url}/repos/{gh_repo}/actions"
                    f"/workflows/{gh_workflow_id}/dispatches"
                )

                logger.debug(f"Calling plan {url}")
                async with httpx.AsyncClient() as h:
                    r = await h.post(
                        url,
                        headers={
                            "Accept": "application/vnd.github+json",
                            "Authorization": f"Bearer {gh_token}",
                        },
                        json={
                            "ref": gh_ref,
                            "inputs": {
                                "experiment-url": experiment_url,
                                "environment-name": "myenv",
                            },
                        },
                    )
                    if r.status_code == 204:
                        logger.info(f"Plan {plan_id} scheduled")
                    else:
                        logger.error(
                            f"Failed to schedule plan {plan_id}: "
                            f"{r.status_code} - {r.json()}"
                        )
        except Exception:
            logger.error(f"Failed to schedule plan {plan_id}", exc_info=True)
            raise
        finally:
            scope.cancel()

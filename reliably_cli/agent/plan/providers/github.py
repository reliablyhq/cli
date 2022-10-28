from anyio.abc import CancelScope
from reliably.agent.types import Plan
from reliably.config import get_settings
from reliably.log import logger
from reliably.oltp import oltp_span

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
                _ = await fetch_deployment(plan, settings)
        finally:
            scope.cancel()

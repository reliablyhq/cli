import anyio
import trio
from anyio.abc import TaskStatus
from anyio.streams.memory import (
    MemoryObjectReceiveStream,
    MemoryObjectSendStream,
)

from reliably_cli.client import reliably_client
from reliably_cli.config import Settings, get_settings
from reliably_cli.log import logger
from reliably_cli.oltp import oltp_span

from ..types import Plan
from .providers.github import schedule_plan as gh_schedule_plan

__all__ = ["fetch_plans", "process_plans"]


async def process_plans(
    stream: MemoryObjectReceiveStream,
    task_status: TaskStatus = anyio.TASK_STATUS_IGNORED,
) -> None:
    logger.debug("Processing plans starting")
    async with anyio.create_task_group() as tg:
        task_status.started()
        async with stream:
            async for plan in stream:
                if plan.definition.environment.provider == "github":
                    tg.start_soon(gh_schedule_plan, plan)

    logger.debug("Processing plans terminated")


async def fetch_plans(stream: MemoryObjectSendStream) -> None:
    logger.debug("Fetching plans starting")
    async with anyio.create_task_group() as tg:
        try:
            while True:
                event = anyio.Event()
                tg.start_soon(fetch_next_schedulable_plan, stream, event)
                await event.wait()

                settings = get_settings()
                if settings.plan.fetch_frequency <= 0:
                    break
                await trio.sleep(settings.plan.fetch_frequency)
        finally:
            # by closing the sending stream, this close the receiving stream
            # and terminate the processing of events automatically. neat.
            await stream.aclose()
    logger.debug("Fetching plans terminated")


###############################################################################
# Private
###############################################################################
def get_enabled_providers(settings: Settings) -> list[str]:
    enabled = []

    if settings.plan.providers.github.enabled:
        enabled.append("github")

    return enabled


async def fetch_next_schedulable_plan(
    stream: MemoryObjectSendStream, event: anyio.Event
) -> None:
    try:
        settings = get_settings()
        providers = get_enabled_providers(settings)

        for provider in providers:
            with oltp_span(
                "fetch-next-plan",
                settings=settings,
                attrs={"deployment_type": provider},
            ):
                async with reliably_client() as client:
                    logger.debug(
                        f"Looking for next plan to schedule on '{provider}'"
                    )
                    r = await client.get(
                        "/plans/schedulables/next",
                        params={"deployment_type": provider},
                    )
                    plan = r.json()
                    if plan is None:
                        logger.debug(f"No plans to schedule on '{provider}'")
                        return

                    plan = Plan.parse_obj(plan)
                    await stream.send(plan)
    finally:
        event.set()

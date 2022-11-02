import anyio
from anyio.abc import CancelScope

from ..config import get_settings
from ..log import logger
from .plan import fetch_plans, process_plans

__all__ = ["agent_runner", "validate_agent_configuration"]


async def agent_runner(scope: CancelScope) -> None:
    logger.debug("Agent now running...")

    if not get_settings().plan:
        print("missing [plan] section in configuration. noop.")
        scope.cancel()
        return

    fetch_stream, process_stream = anyio.create_memory_object_stream()

    async with anyio.create_task_group() as tg:
        await tg.start(process_plans, process_stream)
        tg.start_soon(fetch_plans, fetch_stream)

    logger.debug("Agent terminated")
    scope.cancel()


def validate_agent_configuration() -> None:
    config = get_settings()

    if not config.organization or not config.organization.id:
        raise ValueError("please set the organization identifier")

    if not config.agent or not config.agent.id:
        raise ValueError("please set the agent identifier")

    if not config.agent or not config.agent.token:
        raise ValueError("please set the agent token")

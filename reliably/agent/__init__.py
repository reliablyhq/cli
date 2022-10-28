import anyio
from anyio.abc import CancelScope

from ..log import logger
from .plan import fetch_plans, process_plans

__all__ = ["agent_runner"]


async def agent_runner(scope: CancelScope) -> None:
    logger.debug("Agent now running...")
    fetch_stream, process_stream = anyio.create_memory_object_stream()

    async with anyio.create_task_group() as tg:
        await tg.start(process_plans, process_stream)
        tg.start_soon(fetch_plans, fetch_stream)

    logger.debug("Agent terminated")
    scope.cancel()

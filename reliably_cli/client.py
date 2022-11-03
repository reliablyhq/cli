from contextlib import asynccontextmanager

import httpx

from .config import get_settings

__all__ = ["reliably_client"]


@asynccontextmanager
async def reliably_client() -> httpx.AsyncClient:
    settings = get_settings()
    host = settings.service.host
    org_id = settings.organization.id
    token = settings.agent.token

    async with httpx.AsyncClient(
        http2=True,
        base_url=f"{host}/api/v1/organization/{org_id}",
        headers={"Authorization": f"Bearer {token.get_secret_value()}"},
        timeout=30.0,
    ) as c:
        yield c

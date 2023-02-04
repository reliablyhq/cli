from contextlib import asynccontextmanager, contextmanager

import httpx

from .config import get_settings

__all__ = ["agent_client", "api_client"]


@asynccontextmanager
async def agent_client() -> httpx.AsyncClient:
    settings = get_settings()
    host = settings.service.host
    org_id = settings.organization.id
    token = settings.agent.token

    headers = {"Authorization": f"Bearer {token.get_secret_value()}"}

    async with httpx.AsyncClient(
        http2=True,
        base_url=f"{host}/api/v1/organization/{org_id}",
        headers=headers,
        timeout=30.0,
    ) as c:
        yield c


@contextmanager
def api_client() -> httpx.Client:
    settings = get_settings()
    host = settings.service.host
    token = settings.service.token
    org_id = settings.organization.id

    headers = {"Authorization": f"Bearer {token.get_secret_value()}"}

    with httpx.Client(
        http2=True,
        base_url=f"{host}/api/v1/organization/{org_id}",
        headers=headers,
        timeout=30.0,
    ) as c:
        yield c

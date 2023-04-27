from datetime import datetime
from enum import Enum
from typing import Any, Dict, List, Literal

import orjson
from pydantic import UUID4, BaseModel

__all__ = ["Plan", "FormatOption", "Organization"]


def _json_dumps(*args, **kwargs) -> str:  # type: ignore[no-untyped-def]
    return orjson.dumps(*args, **kwargs).decode("utf-8")


def _json_loads(obj: str, *args, **kwargs) -> Any:  # type: ignore[no-untyped-def]  # noqa
    return orjson.loads(obj.encode("utf-8"), *args, **kwargs)


class FormatOption(Enum):
    json = "json"
    yaml = "yaml"


class BaseSchema(BaseModel):
    class Config:
        json_loads = _json_loads
        json_dumps = _json_dumps


class PlanReliablyEnvironment(BaseSchema):
    provider: Literal["reliably_cloud"] = "reliably_cloud"
    id: UUID4


class PlanGitHubEnvironment(BaseSchema):
    provider: Literal["github"] = "github"
    name: str


class PlanGCPEnvironment(BaseSchema):
    provider: Literal["gcp"] = "gcp"


class PlanDeployment(BaseSchema):
    deployment_id: UUID4


class PlanScheduleNow(BaseSchema):
    type: Literal["now"] = "now"


class PlanScheduleCron(BaseSchema):
    type: Literal["cron"] = "cron"
    pattern: str


class PlanBase(BaseSchema):  # pragma: no cover
    environment: PlanGitHubEnvironment | PlanReliablyEnvironment | PlanGCPEnvironment | None  # noqa: E501
    deployment: PlanDeployment
    schedule: PlanScheduleNow | PlanScheduleCron
    experiments: List[UUID4]
    integrations: List[UUID4]


class Plan(BaseSchema):
    id: UUID4
    created_date: datetime
    definition: PlanBase
    ref: str
    status: str
    error: str | None


class IntegrationControlPythonProvider(BaseSchema):
    type: Literal["python"] = "python"
    module: str
    secrets: List[str] | None
    arguments: Dict[str, Any] | None


class IntegrationControl(BaseSchema):
    name: str
    provider: IntegrationControlPythonProvider


class Organization(BaseSchema):
    id: UUID4
    name: str
    created_date: datetime


class Organizations(BaseSchema):
    count: int
    items: list[Organization]

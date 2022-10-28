from datetime import datetime
from typing import List, Literal

from pydantic import UUID4, BaseModel

__all__ = ["Plan"]


class PlanGitHubEnvironment(BaseModel):
    provider: Literal["github"] = "github"
    name: str


class PlanGCPEnvironment(BaseModel):
    provider: Literal["gcp"] = "gcp"


class PlanDeployment(BaseModel):
    deployment_id: UUID4


class PlanScheduleNow(BaseModel):
    type: Literal["now"] = "now"


class PlanScheduleCron(BaseModel):
    type: Literal["cron"] = "cron"
    pattern: str


class PlanBase(BaseModel):  # pragma: no cover
    environment: PlanGitHubEnvironment | PlanGCPEnvironment
    deployment: PlanDeployment
    schedule: PlanScheduleNow | PlanScheduleCron
    experiments: List[UUID4]


class Plan(BaseModel):
    id: UUID4
    created_date: datetime
    definition: PlanBase
    ref: str
    status: str
    error: str | None

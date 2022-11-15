import os
from pathlib import Path
from typing import Any, Literal

import typer
from logzero import logger
from pydantic import UUID4, AnyUrl, BaseModel, BaseSettings, HttpUrl, SecretStr

try:
    import tomllib  # noqa
except ImportError:
    import tomli as tomllib

__all__ = ["Settings"]


class PlanProviderGitHubSection(BaseModel):
    enabled: bool = False
    token: SecretStr | None
    api_url: HttpUrl = "https://api.github.com"
    repo: str | None
    workflow_id: str = "plan.yaml"
    ref: str = "main"


class PlanProvidersSection(BaseModel):
    github: PlanProviderGitHubSection | None


class PlanSection(BaseModel):
    fetch_frequency: float = 0
    providers: PlanProvidersSection | None


class OrgSection(BaseModel):
    id: UUID4


class AgentSection(BaseModel):
    id: UUID4
    token: SecretStr


class ServiceSection(BaseModel):
    host: str = "https://app.reliably.com"


class LogSection(BaseModel):
    level: Literal["debug", "info", "error", "warn", "notset"] = "info"
    as_json: bool = False


class OTELInfoSection(BaseModel):
    enabled: bool = False
    service_name: str = "reliably"
    endpoint: AnyUrl | None
    headers: str | None


class OTELMetricsInfoSection(OTELInfoSection):
    expose_as_prometheus: bool = False


class OTELSection(BaseModel):
    traces: OTELInfoSection = OTELInfoSection()
    metrics: OTELMetricsInfoSection = OTELMetricsInfoSection()


class Settings(BaseSettings):
    service: ServiceSection | None
    agent: AgentSection | None
    organization: OrgSection | None
    plan: PlanSection | None
    log: LogSection = LogSection(level="info", as_json=False)
    otel: OTELSection = OTELSection()

    class Config:
        env_prefix = "reliably_"
        env_nested_delimiter = "_"
        env_file_encoding = "utf-8"
        toml_file = Path(typer.get_app_dir("reliably")) / "config.toml"

        @classmethod
        def customise_sources(
            cls,
            init_settings,
            env_settings,
            file_secret_settings,
        ):
            return (
                env_settings,
                toml_config_settings,
            )


def toml_config_settings(settings: BaseSettings) -> dict[str, Any]:
    toml_file = os.getenv("RELIABLY_CLI_CONFIG")
    if not toml_file:
        toml_file = settings.__config__.toml_file

    if not Path(toml_file).exists():
        return {}

    logger.debug(f"Reading configuration from '{toml_file}'")
    encoding = settings.__config__.env_file_encoding
    return tomllib.loads(Path(toml_file).read_text(encoding=encoding))

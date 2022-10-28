import functools
from pathlib import Path
from typing import Any, Literal

import typer
from pydantic import UUID4, AnyUrl, BaseModel, BaseSettings, SecretStr

try:
    import tomlib
except ImportError:
    import tomli as tomlib

__all__ = ["get_settings", "Settings"]


class PlanProviderGitHubSection(BaseModel):
    enabled: bool = False


class PlanProvidersSection(BaseModel):
    github: PlanProviderGitHubSection | None


class PlanSection(BaseModel):
    fetch_frequency: float = 3
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


class OTELSection(BaseModel):
    enabled: bool = False
    service_name: str = "reliably"
    endpoint: AnyUrl | None
    headers: str | None


class Settings(BaseSettings):
    service: ServiceSection
    agent: AgentSection
    organization: OrgSection
    plan: PlanSection
    log: LogSection | None
    otel: OTELSection | None

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


@functools.lru_cache
def get_settings() -> Settings:
    return Settings()


def toml_config_settings(settings: BaseSettings) -> dict[str, Any]:
    toml_file = settings.__config__.toml_file
    if not Path(toml_file).exists():
        return {}

    encoding = settings.__config__.env_file_encoding
    return tomlib.loads(Path(toml_file).read_text(encoding=encoding))

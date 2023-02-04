import os
from pathlib import Path
from typing import Any, Literal

import typer
from pydantic import UUID4, BaseModel, BaseSettings, HttpUrl, SecretStr

try:
    import tomllib  # noqa
except ImportError:
    import tomli as tomllib

__all__ = ["Settings", "get_settings_directory_path"]


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
    token: SecretStr | None


class ServiceSection(BaseModel):
    host: str = "https://app.reliably.com"
    token: SecretStr | None


class LogSection(BaseModel):
    level: Literal["debug", "info", "error", "warn", "notset"] = "info"
    as_json: bool = False


class Settings(BaseSettings):
    service: ServiceSection = ServiceSection()
    agent: AgentSection | None
    organization: OrgSection | None
    plan: PlanSection | None
    log: LogSection = LogSection(level="info", as_json=False)

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

    encoding = settings.__config__.env_file_encoding
    return tomllib.loads(Path(toml_file).read_text(encoding=encoding))


def get_settings_directory_path() -> Path:
    cfg_path = os.getenv("RELIABLY_CLI_CONFIG")
    if not cfg_path:
        cfg_path = Settings.__config__.toml_file

    p = Path(cfg_path).parent
    if not p.exists() and not p.is_dir():
        p.mkdir()

    return p.absolute()

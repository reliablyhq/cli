import functools

from .types import Settings

__all__ = ["get_settings"]


@functools.lru_cache
def get_settings() -> Settings:
    return Settings()

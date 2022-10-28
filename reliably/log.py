import logging

import logzero
from logzero import ForegroundColors, LogFormatter, logger
from opentelemetry.instrumentation.logging.constants import (
    DEFAULT_LOGGING_FORMAT,
)

from . import is_executable
from .config import Settings

__all__ = ["logger", "configure_logger"]
COLORS = {
    logging.DEBUG: ForegroundColors.CYAN,
    logging.INFO: ForegroundColors.GREEN,
    logging.WARNING: ForegroundColors.YELLOW,
    logging.ERROR: ForegroundColors.RED,
    logging.CRITICAL: ForegroundColors.RED,
}


def configure_logger(config: Settings) -> None:
    logzero.loglevel(logging.getLevelName(config.log.level.upper()))
    logzero.formatter(
        LogFormatter(
            fmt="%(color)s[%(asctime)s %(levelname)s]%(end_color)s %(message)s",
            datefmt="%Y-%m-%d %H:%M:%S",
            colors=COLORS,
        )
    )

    if config.log.as_json:
        logzero.json()

        if not is_executable() and config.otel and config.otel.enabled:
            logzero.formatter(
                LogFormatter(color=False, fmt=DEFAULT_LOGGING_FORMAT)
            )

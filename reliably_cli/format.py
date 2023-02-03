import io
import json

from ruamel.yaml import YAML

from .types import BaseSchema, FormatOption

__all__ = ["format_as"]


def format_as(
    entity: BaseSchema, fmt: FormatOption = FormatOption.json
) -> str | None:
    match fmt.value:  # noqa
        case "json":
            return entity.json(indent=2)
        case "yaml":
            with io.StringIO() as s:
                yaml = YAML()
                yaml.default_flow_style = False
                yaml.dump(json.loads(entity.json()), s)
                return s.getvalue()

    return None

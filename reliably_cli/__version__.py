from importlib.metadata import version, PackageNotFoundError

try:
    __version__ = version("reliably-cli")
except PackageNotFoundError:
    __version__ = 'unknown'

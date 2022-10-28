from typer.testing import CliRunner

from reliably_cli.__main__ import cli

runner = CliRunner()


def test_app_help():
    result = runner.invoke(cli, ["--help"])
    assert result.exit_code == 0

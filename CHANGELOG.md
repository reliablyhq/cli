# Changelog

## [Unreleased]
### Added
- Kubernetes live scan now provides suggestions for nodes
- Scan has new `extended` format for having more verbose output
- Suggestion examples are added to `sarif`, `codeclimate` and `extended` output formats

### Changed
- policy retrieval logic has been changed to incorporate the API Version into the path to the remote policy. Locally cached policy also includes the API Version.

## [0.5.0] - 2021-03-03

### Changed
- `discover` command has been replaced by `scan`; [69](https://github.com/reliablyhq/cli/issues/69)

## [0.4.0] - 2021-03-01
### Added
- Logout command has a `--yes` flag for logging out without being prompted
- Check for newest release when running the CLI; [38](https://github.com/reliablyhq/cli/issues/38)
- Setup Reliably workflow for CI/CD platform; [44](https://github.com/reliablyhq/cli/issues/44)
- Show suggestions level (info, warning, error) on discovery result; [52](https://github.com/reliablyhq/cli/issues/52)
- Discover `--level` flag allows to display only suggestions at specified level or higher
- Discover `--live` flag looks for weaknesses in a live Kubernetes cluster

### Changed
- `discover` command now requires to be authenticated to Reliably.

### Fixed
- `discover` command was not supporting folders starting with ../; [#51](https://github.com/reliablyhq/cli/issues/51)
- increased timeout for HTTP call to Reliably API for loggin in; [#43](https://github.com/reliablyhq/cli/issues/43)

## [0.3.0] - 2021-01-22
### Added

- Authentication management commands; see `reliably auth --help`
- Help topic for environment variables descriptions; see `reliably environment --help`
- Completion command for generating shell completion scripts; see `reliably completion --help`
- Coloring support for output & help; can be disabled with `--no-color` flag

## [0.2.1] - 2021-01-18
### Fixed
- Fixes locations path generation in sarif format; [#18](https://github.com/reliablyhq/cli/issues/18)

## [0.2.0] - 2020-12-18
### Changed

- the `discover` command now supports manifest file or folder path as main argument.
  It can also read a manifest from stdin by providing a '-' as argument.
  The `--dir` flag has been deprecated and will be removed.

## [0.1.0] - 2020-11-20
### Added

- Initial version

[Unreleased]: https://github.com/reliablyhq/cli/compare/v0.5.0...HEAD
[0.5.0]: https://github.com/reliablyhq/cli/releases/tag/v0.5.0
[0.4.0]: https://github.com/reliablyhq/cli/releases/tag/v0.4.0
[0.3.0]: https://github.com/reliablyhq/cli/releases/tag/v0.3.0
[0.2.1]: https://github.com/reliablyhq/cli/releases/tag/v0.2.1
[0.2.0]: https://github.com/reliablyhq/cli/releases/tag/v0.2.0
[0.1.0]: https://github.com/reliablyhq/cli/releases/tag/v0.1.0
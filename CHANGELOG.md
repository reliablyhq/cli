# Changelog

## [Unreleased]
### Added

- Authentication management commands; see `reliably auth --help`
- Help topic for environment variables descriptions; see `reliably environment --help`
- Completion command for generating shell completion scripts; see `reliably completion --help`

## [0.2.0] - 2020-12-18
### Changed

- the `discover` command now supports manifest file or folder path as main argument.
  It can also read a manifest from stdin by providing a '-' as argument.
  The `--dir` flag has been deprecated and will be removed.

## [0.1.0] - 2020-11-20
### Added

- Initial version

[Unreleased]: https://github.com/reliablyhq/cli/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/reliablyhq/cli/releases/tag/v0.2.0
[0.1.0]: https://github.com/reliablyhq/cli/releases/tag/v0.1.0
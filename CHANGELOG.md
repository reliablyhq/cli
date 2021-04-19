# Changelog

## [Unreleased]

## [0.10.0] - 2021-04-19
### Added

- added retrieval of service level indicators for AWS `Application Load Balancer` resources
- added `watch` flag for `slo report` command to continuously watch SLO report output

## [0.9.0] - 2021-04-12

### Changed

- `slo init` command now allows the initialization of multiple SLOs;

### Fixed
- Fixes sorting metrics providers in `slo init`; [#161](https://github.com/reliablyhq/cli/issues/161)
- Fixes the output not wrapping correctly in `slo report`.

## [0.8.0] - 2021-04-06

### Changed

- `scan` command now works only with sub-command for scanning a specific type of resource;
  To scan for kubernetes, you should now use `scan kubernetes` command instead.

## [0.7.0] - 2021-04-02

### Added
- added `reliably slo init` command that allows a user to describe a service level objective
- added `reliably slo report` command that allows a user to generate a report about their current slo
- added an AWS Cloudwatch provider that retrieves service level indicators for `Api Gateway` resources
- added an GCP Monitoring provider that retrieved service level indicators for `Load Balancer` resources

### Changed
- changed suggestion printing to exclude line numbers when they are ':1:1' - this indicates that we couldn't extract a line number and so are using defaults.


## [0.6.0] - 2021-03-24

### Added
- Kubernetes live scan now provides suggestions for nodes
- Scan has new `tabbed` format that provides tabbed formatted output
- Scan has new `extended` format for having more verbose output
- Suggestion examples are added to `sarif` and `extended` output formats

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

[Unreleased]: https://github.com/reliablyhq/cli/compare/v0.10.0...HEAD
[0.10.0]: https://github.com/reliablyhq/cli/releases/tag/v0.10.0
[0.9.0]: https://github.com/reliablyhq/cli/releases/tag/v0.9.0
[0.8.0]: https://github.com/reliablyhq/cli/releases/tag/v0.8.0
[0.7.0]: https://github.com/reliablyhq/cli/releases/tag/v0.7.0
[0.6.0]: https://github.com/reliablyhq/cli/releases/tag/v0.6.0
[0.5.0]: https://github.com/reliablyhq/cli/releases/tag/v0.5.0
[0.4.0]: https://github.com/reliablyhq/cli/releases/tag/v0.4.0
[0.3.0]: https://github.com/reliablyhq/cli/releases/tag/v0.3.0
[0.2.1]: https://github.com/reliablyhq/cli/releases/tag/v0.2.1
[0.2.0]: https://github.com/reliablyhq/cli/releases/tag/v0.2.0
[0.1.0]: https://github.com/reliablyhq/cli/releases/tag/v0.1.0
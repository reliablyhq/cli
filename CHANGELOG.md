# Changelog

## [Unreleased]

## [v0.21.0] - 2021-07-19

### Added

-   Add support for Datadog for SLO commands; init an objective with datadog queries & push indicators using the SLO agent
-   Added `--report-view/-R` flags to `reliably slo agent` command to allow a split view showing agent output and an SLO report

## [v0.20.4] - 2021-07-16

### Added

-   A very verbose `-V, --very-verbose` flag allows for logging request/response bodies

### Fixed

-   Report command accepts to/from time in RFC 3339 and time.String() output
-   Creates config folder before writing config file; [#350](https://github.com/reliablyhq/cli/issues/350)

### Changed

-   Reading manifest message in report command is displayed only in verbose mode

## [v0.20.3] - 2021-07-12

### Fixed

-   Datetime string parsing fixed to support microseconds, as provided by the `slo agent` command.

## [v0.20.2] - 2021-07-09

### Fixed

-   `reliably slo report` receives the correct number of results for each objective
-   Entity server hostname is not changed when overridding with reliably host environment variable
-   Do not lowercase the organization name in the entity server URLs

### Changed

-   `reliably slo report` now accepts a `-l, --selector` flag to filter objectives by labels
-   `reliably slo report` does not add additional filtering from manifest by default. To filter from manifest, explicitly include the path.
-   The service label is now optional in `reliably slo report`.

## [v0.20.1] - 2021-07-02

### Fixed

-   Invalid `/relatedto` path for graph visualisation: [#382](https://github.com/reliablyhq/cli/issues/382)

### Changed

-   Added RELIABLY_ORG usage to workflows templates
-   Added more descriptive error messaging when using `slo init` when not logged in to GCP

## [v0.20.0] - 2021-07-02

### Added

-   New `reliably slo related` command for SLO relationship visualisation
-   use user's default organization when not set in config nor set with RELIABLY_ORG env var; [#362](https://github.com/reliablyhq/cli/issues/362)

### Changed

-   Uses new GitHub action for scan command in GitHub workflow template
-   updated `slo sync` to support weakly define types

## [v0.19.4] - 2021-07-01

### Fixed

-   Release file not found, path is not correctly expanded from user home
-   Overridden organization with RELIABLY_ORG taken into account; [#361](https://github.com/reliablyhq/cli/issues/361)

## [v0.19.3] - 2021-06-30

### Fixed

-   Fixed overidden organization info merged with info from config; [#364](https://github.com/reliablyhq/cli/issues/364)

## [v0.19.2] - 2021-06-29

### Fixed

-   Remove a debug statement for `org set` command

## [v0.19.1] - 2021-06-28

### Added

-   Save user's default organization to config upon successful authentication; [#242](https://github.com/reliablyhq/cli/issues/242)

### Fixed

-   Use current organization set from config, instead of users default one

## [v0.19.0] - 2021-06-23

### Added

-   `org` command provides sub-commands to create, manage and delete organizations.

### Fixed

-   `auth login` crashes when config file is not present; [#350](https://github.com/reliablyhq/cli/issues/350)

## [v0.18.1] - 2021-06-18

### Fixed

-   CLI failed when config file is missing; [#341](https://github.com/reliablyhq/cli/issues/341)
-   Regression in `auth status` with overriden auth token from environment variable; [#346](https://github.com/reliablyhq/cli/issues/346)

## [v0.18.0] - 2021-06-17

### Added

-   `slo sync` command allows to synchronize the local objectives with Reliably servers
-   `slo agent` command runs as a background process to periodically fetch your SLIs and send them back to Reliably.

### Changed

-   Added the formats 'table' & 'text' to the kubernetes scan command. Put 'simple' amd 'tabbed' formats in the deprecated list  ; [#332](https://github.com/reliablyhq/cli/issues/332)
-   Fixed the trend ouput for terminal and included html span elements for markdown type reports  ; [#330](https://github.com/reliablyhq/cli/issues/330)
-   `slo init` and `slo report` have been updated with breaking changes to handle new manifest & objective structure.
-   The SLO manifest has been entirely refactored changed to match the Kubernete-like definition.

## [v0.17.0] - 2021-06-10

### Added

-   changes SLO report to support reports from templates; [#316](https://github.com/reliablyhq/cli/issues/316)

## [v0.16.1] - 2021-06-01

### Changed

-   changes format and content of the SLO Markdown Report; [#302](https://github.com/reliablyhq/cli/issues/302)
-   added the type of the sli to the text format report; [#297](https://github.com/reliablyhq/cli/issues/297)

## [v0.16.0] - 2021-05-19

### Added

-   `update` command for installing the latest version of the CLI, or a specific one; [#289](https://github.com/reliablyhq/cli/issues/289)

### Changed

-   `slo report` now supports multiple outputs '--output' & formats '--format', with comma-seperated values; [#287](https://github.com/reliablyhq/cli/issues/287)
-   `slo report` table output has been improved with some UX changes

### Fixed

-   no color output was not working for some subcommands; [#291](https://github.com/reliablyhq/cli/issues/291)

## [v0.15.0] - 2021-05-14

### Added

-   `slo report` sends the generated report to Reliably
-   `slo report` shows the SLO trend for the last 5 reports; [#255](https://github.com/reliablyhq/cli/issues/255)
-   the time period defined for an SLO is exported into the report; [#262](https://github.com/reliablyhq/cli/issues/262)

### Changed

-   `simple` & `tabbed` output formats for `slo report` command have been deprecated, use `text` & `table` in replacement. The deprecated ones will soon be removed without notice.; [#257](https://github.com/reliablyhq/cli/issues/257)
-   Removed SLO manifest management commands, `reliably slo pull/apply/edit`
-   `reliably slo report` now explicitly requires the user to provide a manifest file
-   `reliably slo init` now only creates a new manifest file or overrides an existing one

## [v0.14.0] - 2021-05-04

### Added

-   New commands for SLO manifest management: `slo edit`, `slo apply`, `slo pull`. The manifest is now centralized and backed up on Reliably's servers.

### Changed

-   `slo init` now sends the newly generated manifest to Reliably's API
-   `slo report` now uses the centralized manifest to generate the report. If a local manifest is found, `slo report` uses that local file in precedence over the remote one.

### Fixed

-   Users with an invalid token in config can now re-authenticate with `auth login`; [#246](https://github.com/reliablyhq/cli/issues/246)
-   Validate user input when user authenticate with token in interactive mode; [#248](https://github.com/reliablyhq/cli/issues/248)

## [v0.13.3] - 2021-05-03

### Added

-   `help` is now a command; help can be displayed for any command using either
    the `help` command or the `--help` flag

## [0.13.2] - 2021-04-30

### Fixed

-   Fixes SLO report `--format` & `--output` combined flags; [#241](https://github.com/reliablyhq/cli/issues/241)

## [0.13.1] - 2021-04-30

### Fixed

-   Fixes GCP client not closed; [#236](https://github.com/reliablyhq/cli/issues/236)

## [0.13.0] - 2021-04-29

### Added

-   `DEBUG` environment variable for turning on debug/verbose mode
-   `slo init` now suggests a SLO title by default; [#225](https://github.com/reliablyhq/cli/issues/225)

### Changed

-   Go minimal version has been upgraded to 1.16

### Fixed

-   No SLI metrics found for latest hour on AWS; [#226](https://github.com/reliablyhq/cli/issues/226)
-   Expose boolean if SLO is met as part of SLO result in the report; [#237](https://github.com/reliablyhq/cli/issues/237)

## [0.12.1] - 2021-04-26

### Added

-   `NO_COLOR` environment variable for disabling colored output

### Fixed

-   Fix bad prompt validation for some user inputs; [#219](https://github.com/reliablyhq/cli/issues/219)
-   Fix prompting for user does not respect the `--no-color` flag; [#221](https://github.com/reliablyhq/cli/issues/221)
-   Fix missing validation for AWS ARN value on `slo init`; [#223](https://github.com/reliablyhq/cli/issues/223)

## [0.12.0] - 2021-04-23

### Added

-   Added time observation window for SLO in manifest & prompt user in `slo init` generation; [#1681](https://github.com/reliablyhq/cli/issues/181)
-   `slo report` can now output the report into yaml format

### Fixed

-   Fixes missing SLOs without result in markdown report; [#208](https://github.com/reliablyhq/cli/issues/208)

## [0.11.0] - 2021-04-21

### Changed

-   `slo report` command now includes and output format of markdown
-   changed computation for latency SLO: percentage of 99 percentiles under threshold for 1-minute samples

## [0.10.0] - 2021-04-19

### Added

-   added retrieval of service level indicators for AWS `Application Load Balancer` resources
-   added `watch` flag for `slo report` command to continuously watch SLO report output

## [0.9.0] - 2021-04-12

### Changed

-   `slo init` command now allows the initialization of multiple SLOs;

### Fixed

-   Fixes sorting metrics providers in `slo init`; [#161](https://github.com/reliablyhq/cli/issues/161)
-   Fixes the output not wrapping correctly in `slo report`.

## [0.8.0] - 2021-04-06

### Changed

-   `scan` command now works only with sub-command for scanning a specific type of resource;
    To scan for kubernetes, you should now use `scan kubernetes` command instead.

## [0.7.0] - 2021-04-02

### Added

-   added `reliably slo init` command that allows a user to describe a service level objective
-   added `reliably slo report` command that allows a user to generate a report about their current slo
-   added an AWS Cloudwatch provider that retrieves service level indicators for `Api Gateway` resources
-   added an GCP Monitoring provider that retrieved service level indicators for `Load Balancer` resources

### Changed

-   changed suggestion printing to exclude line numbers when they are ':1:1' - this indicates that we couldn't extract a line number and so are using defaults.

## [0.6.0] - 2021-03-24

### Added

-   Kubernetes live scan now provides suggestions for nodes
-   Scan has new `tabbed` format that provides tabbed formatted output
-   Scan has new `extended` format for having more verbose output
-   Suggestion examples are added to `sarif` and `extended` output formats

### Changed

-   policy retrieval logic has been changed to incorporate the API Version into the path to the remote policy. Locally cached policy also includes the API Version.

## [0.5.0] - 2021-03-03

### Changed

-   `discover` command has been replaced by `scan`; [69](https://github.com/reliablyhq/cli/issues/69)

## [0.4.0] - 2021-03-01

### Added

-   Logout command has a `--yes` flag for logging out without being prompted
-   Check for newest release when running the CLI; [38](https://github.com/reliablyhq/cli/issues/38)
-   Setup Reliably workflow for CI/CD platform; [44](https://github.com/reliablyhq/cli/issues/44)
-   Show suggestions level (info, warning, error) on discovery result; [52](https://github.com/reliablyhq/cli/issues/52)
-   Discover `--level` flag allows to display only suggestions at specified level or higher
-   Discover `--live` flag looks for weaknesses in a live Kubernetes cluster

### Changed

-   `discover` command now requires to be authenticated to Reliably.

### Fixed

-   `discover` command was not supporting folders starting with ../; [#51](https://github.com/reliablyhq/cli/issues/51)
-   increased timeout for HTTP call to Reliably API for loggin in; [#43](https://github.com/reliablyhq/cli/issues/43)

## [0.3.0] - 2021-01-22

### Added

-   Authentication management commands; see `reliably auth --help`
-   Help topic for environment variables descriptions; see `reliably environment --help`
-   Completion command for generating shell completion scripts; see `reliably completion --help`
-   Coloring support for output & help; can be disabled with `--no-color` flag

## [0.2.1] - 2021-01-18

### Fixed

-   Fixes locations path generation in sarif format; [#18](https://github.com/reliablyhq/cli/issues/18)

## [0.2.0] - 2020-12-18

### Changed

-   the `discover` command now supports manifest file or folder path as main argument.
    It can also read a manifest from stdin by providing a '-' as argument.
    The `--dir` flag has been deprecated and will be removed.

## [0.1.0] - 2020-11-20

### Added

-   Initial version

[Unreleased]: https://github.com/reliablyhq/cli/compare/v0.21.0...HEAD

[v0.21.0]: https://github.com/reliablyhq/cli/compare/v0.20.4...v0.21.0

[v0.20.4]: https://github.com/reliablyhq/cli/compare/v0.20.3...v0.20.4

[v0.20.3]: https://github.com/reliablyhq/cli/compare/v0.20.2...v0.20.3

[v0.20.2]: https://github.com/reliablyhq/cli/compare/v0.20.1...v0.20.2

[v0.20.1]: https://github.com/reliablyhq/cli/compare/v0.20.0...v0.20.1

[v0.20.0]: https://github.com/reliablyhq/cli/compare/v0.19.4...v0.20.0

[v0.19.4]: https://github.com/reliablyhq/cli/compare/v0.19.3...v0.19.4

[v0.19.3]: https://github.com/reliablyhq/cli/compare/v0.19.2...v0.19.3

[v0.19.2]: https://github.com/reliablyhq/cli/compare/v0.19.1...v0.19.2

[v0.19.1]: https://github.com/reliablyhq/cli/compare/v0.19.0...v0.19.1

[v0.19.0]: https://github.com/reliablyhq/cli/compare/v0.18.1...v0.19.0

[v0.18.1]: https://github.com/reliablyhq/cli/compare/v0.18.0...v0.18.1

[v0.18.0]: https://github.com/reliablyhq/cli/compare/v0.17.0...v0.18.0

[v0.17.0]: https://github.com/reliablyhq/cli/compare/v0.16.1...v0.17.0

[v0.16.1]: https://github.com/reliablyhq/cli/compare/v0.16.0...v0.16.1

[v0.16.0]: https://github.com/reliablyhq/cli/compare/v0.15.0...v0.16.0

[v0.15.0]: https://github.com/reliablyhq/cli/compare/v0.14.0...v0.15.0

[v0.14.0]: https://github.com/reliablyhq/cli/compare/v0.13.3...v0.14.0

[v0.13.3]: https://github.com/reliablyhq/cli/compare/v0.13.2...v0.13.3

[0.13.2]: https://github.com/reliablyhq/cli/releases/tag/v0.13.2

[0.13.1]: https://github.com/reliablyhq/cli/releases/tag/v0.13.1

[0.13.0]: https://github.com/reliablyhq/cli/releases/tag/v0.13.0

[0.12.1]: https://github.com/reliablyhq/cli/releases/tag/v0.12.1

[0.12.0]: https://github.com/reliablyhq/cli/releases/tag/v0.12.0

[0.11.0]: https://github.com/reliablyhq/cli/releases/tag/v0.11.0

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

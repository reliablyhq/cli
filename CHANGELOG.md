# Changelog

## [Unreleased][]

[Unreleased]: https://github.com/reliablyhq/cli/compare/0.8.2...HEAD

## [0.8.2][]

[0.8.2]: https://github.com/reliablyhq/cli/compare/0.8.1...0.8.2

### Changed

* Bump dependencies
* Enable PyPI trusted publisher
* Automate release with changelog

## [0.8.1][]

[0.8.1]: https://github.com/reliablyhq/cli/compare/0.8.0...0.8.1

### Changed

* Bump dependencies

## [0.8.0][]

[0.8.0]: https://github.com/reliablyhq/cli/compare/0.7.1...0.8.0

### Changed

* Bump dependencies
* Load integration environment variables and secrets into memory when requested
- Reworked the Dockerfile to use pdm.lock

## [0.7.1][]

[0.7.1]: https://github.com/reliablyhq/cli/compare/0.7.0...0.7.1

### Changed

* Bump dependencies

## [0.7.0][]

[0.7.0]: https://github.com/reliablyhq/cli/compare/0.6.2...0.7.0

### Added

* The `--load-environment` flag to the `plan execute` command so that the CLI
  automatically fetches the environment, if any provided, for the given plan.
  This will fetch environment variables and secrets and load them into memory
  for Chaos Toolkit to use

## [0.6.2][]

[0.6.2]: https://github.com/reliablyhq/cli/compare/0.6.1...0.6.2

### Fixed

* Give more time for Pypi to propagate the new release

## [0.6.1][]

[0.6.1]: https://github.com/reliablyhq/cli/compare/0.6.0...0.6.1

### Fixed

* Dockerfile so it builds from the tag just released
* Fixed GitHub action ity uses buildx and sets the right tags

## [0.5.0][]

[0.5.0]: https://github.com/reliablyhq/cli/compare/0.4.0...0.5.0

### Added

* Extended the `config` command

## [0.4.0][]

[0.4.0]: https://github.com/reliablyhq/cli/compare/0.3.0...0.4.0

### Changed

- Various fixes to handling chaostoolkit execution properly
- Better messages when some config keys aren't set
- Added help messages to commands

## [0.3.0][]

[0.3.0]: https://github.com/reliablyhq/cli/compare/0.2.6...0.3.0

### Changed

- Ensure we can load Python libraries using pyoxdizer even when they rely
  on the `__file__` value

## [0.2.6][]

[0.2.6]: https://github.com/reliablyhq/cli/compare/0.2.6...HEAD

### Changed

- Add changelog file
- Bump dependencies
- Fix readme badges
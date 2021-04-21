<h2 align="center">
  <br>
  <p align="center"><img src="https://raw.githubusercontent.com/reliablyhq/cli/main/logo.png"></p>
</h2>

<h4 align="center">Reliably CLI</h4>

<p align="center">
   <a href="https://github.com/reliablyhq/cli/releases">
   <img alt="Release" src="https://img.shields.io/github/v/release/reliablyhq/cli">
   <a href="https://goreportcard.com/report/github.com/reliablyhq/cli">
   <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/reliablyhq/cli">
   <a href="#">
   <img alt="Build" src="https://github.com/reliablyhq/cli/actions/workflows/build.yaml/badge.svg">
   <a href="https://github.com/reliablyhq/cli/issues">
   
   <img alt="GitHub issues" src="https://img.shields.io/github/issues/reliablyhq/cli?style=flat-square&logo=github&logoColor=white">
   <a href="https://github.com/reliablyhq/cli/blob/master/LICENSE.md">
   <img alt="License" src="https://img.shields.io/github/license/reliablyhq/cli">
   <a href="#">
   <img alt="Go Version" src="https://img.shields.io/github/go-mod/go-version/reliablyhq/cli">
   <a href="https://pkg.go.dev/github.com/reliablyhq/cli">
</p>

<p align="center">
  <a href="#installation">Installation</a> •
  <a href="https://reliably.com/docs/">Documentation</a> •
  <a href="https://github.com/reliablyhq/cli/blob/main/CHANGELOG.md">ChangeLog</a> •
  <a href="#credits">Credits</a> •
</p>

---

# What does Reliably do for you?

Here is a really quick recap of the main features that will help you bringing
reliability to your daily activities:

* Declare and Report on your [SLO][slo]
* Scan for reliability risks in your Kubernetes manifests
* Run from anywhere, including your CI/CD so that you can run reliability checks

# Installation

Reliably CLI is available for macOS, Linux and Windows as
downloadable binaries from the [releases page][releases].

[releases]: https://github.com/reliablyhq/cli/releases

Reliably can also be installed as a [Krew plugin][krew].

[krew]: https://reliably.dev/docs/guides/scan-infrastructure/kubectl-plugin/

Please see the [documentation][docs] for further details.

[docs]: https://reliably.com/docs/

# Usage

Reliably is a CLI and runs from anywhere you can execute its binary. From
your terminal or embedded into a CI/CD pipeline for instance.

```console
$ reliably 
The Reliably Command Line Interface (CLI).

Usage:
  reliably [command]

Available Commands:
  auth        Login, logout, and verify your authentication
  completion  Generate shell completion scripts
  history     Show your scan history
  scan        Check for Reliably Suggestions
  slo         service level objective commands
  workflow    Setup your Reliably workflow

Flags:
  -h, --help       help for reliably
      --no-color   Disable color output
  -v, --verbose    verbose output
      --version    version for reliably

Use "reliably [command] --help" for more information about a command.

Environment variables:
  See 'reliably environment --help' for the list of supported environment variables.

Feedback:
  You can provide with feedback or report an issue at https://github.com/reliablyhq/cli/issues/new
```

Please follow the [guidelines][] to walk you through Reliably's main features.

[guidelines]: https://reliably.dev/docs/guides/

# Credits

This repository contains code from the Reliably CLI project as well as
some open-source works:
* [GitHub CLI](https://github.com/cli/cli)
* Secure Go [gosec](https://github.com/securego/gosec)
* Christopher Thorpe [nestedmaplookup.go](https://gist.github.com/ChristopherThorpe/fd3720efe2ba83c929bf4105719ee967)

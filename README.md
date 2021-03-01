# Reliably CLI

![Reliably](logo.png "Reliably CLI")

Reliability toolbox for developers from the command line.

### Installation

Reliably CLI is available for macOS, Linux and Windows as
downloadable binaries from the [releases page][releases].

[releases]: https://github.com/reliablyhq/cli/releases

### Installation as a Krew plugin

Reliably CLI can be used as a kubectl [Krew plugin][krew-home]
for macOS and Linux.

Once you have [Krew installed][krew-installation], install Reliably with the
following command.

```bash
$ krew install reliably
```

You can then use the CLI as a kubectl plugin:

```bash
$ kubectl reliably
```
[krew-home]: (https://krew.sigs.k8s.io/)
[krew-installation]: (https://krew.sigs.k8s.io/docs/user-guide/setup/install/)

### Authentication

Run `reliably auth login` to authenticate with your Reliably account.
This will run the interactive authentication flow by default.

You can also choose to login with an access token in a non-interactive mode:
`reliably auth login --with-token < my-access-token.txt`

Finally, `reliably` will respect tokens set as environment variable
using `RELIABLY_TOKEN`.

## Usage

To check your Kubernetes manifests for Reliably Advice and Suggestions, simply run:

```
$ reliably discover
```

It will scan for manifests recursively in your current working directory.

To indicate a specific file or folder, give it as a command argument:

```
$ reliably discover manifest.yaml
$ reliably discover ./manifests
```

You can also pipe into `discover` command, as it can read from stdin using
'-' as argument:

```
$ cat manifest.yaml | reliably discover -
```

The CLI supports multiple output formats, such as `simple` *(default)*,
`json`, `yaml`, `sarif`, `codeclimate`. To report in a specific format,
you can use the `--format` or `-f` flag, as follow:

```
$ reliably discover --format sarif
```

The CLI prints out the report to the standard output, by default, but it can
write the report to a local file. You can indicate the path of the report
with the `--output` or `-o` flag, as follow:

```
$ reliably discover --output ./report.txt
```

Please read the [documentation][docs] for more information.

[docs]: https://docs.reliably.com/

## Credits

This repository contains code from the Reliably CLI project as well as
some open-source works:
* [GitHub CLI](https://github.com/cli/cli)
* Secure Go [gosec](https://github.com/securego/gosec)
* Christopher Thorpe [nestedmaplookup.go](https://gist.github.com/ChristopherThorpe/fd3720efe2ba83c929bf4105719ee967)

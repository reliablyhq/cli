# Reliably CLI

![Reliably](logo.png "Reliably CLI")

Reliability toolbox for developers from the command line.

### Usage

To check your Kubernetes manifests for Reliably Advice and Suggestions, simply run:

```
$ reliably discover
```

It will scan for manifests recursively in your current working directory.

To indicate a specific folder, you can use the `--dir` flag, as follow:

```
$ reliably discover --dir ./manifests
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

#### Use as a Github Action

You can use Reliably as part of your Github workflows, by using our [Github Action](https://github.com/reliablyhq/gh-action)

```yaml
- name: Run Reliably
  uses: reliablyhq/gh-action@v1
```

#### Use as a docker container

You can run the CLI with our [docker image](https://github.com/orgs/reliablyhq/packages/container/package/cli%2Fcli)

```
$ docker run --rm \
  --volume=</path/to/manifests/folder>:/manifests \
  ghcr.io/reliablyhq/cli/cli \
  discover --dir /manifests
```


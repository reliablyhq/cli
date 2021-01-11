# Reliably CLI

![Reliably](logo.png "Reliably CLI")

Reliability toolbox for developers from the command line.

### Installation

Reliably CLI is available for macOS, Linux and Windows as
downloadable binaries from the [releases page][].

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


## Contribute

### How to checkout the code

```
$ go get github.com/reliablyhq/cli
```

The source code will be downloaded in the `$GOPATH/src` folder

### How to build

Run:

```
$ make
```

### How to run unit tests

Run:

```
$ make test
```

Or with verbose mode enabled:

```
$ make test/debug
```

#### Code coverage

You can also run tests and compute the code coverage

First, ensure your have installed the Go Tools:

```
$ go install golang.org/x/tools/cmd/cover
```

Then, run your tests with coverage enabled

```
$ make test/coverage
```

Finally, you can visualize the code covarage in a browser:

```
$ make show/coverage
```

### Go coding style

#### Format your source code

Gofmt is a tool that automatically formats Go source code.

You can re-format your code, in place, using the command:

```
$ make format
```

#### Organize your imports

First, install the `goimports` tool

```
$ go install golang.org/x/tools/cmd/goimports
```

Then, run:

```
make imports
```

#### Check for code style mistakes

First, install the `golint` tool

```
$ go install golang.org/x/lint/golint
```

Then, run it:

```
$ make lint
```

#### Check coding style on git prehook

It's best practice to enable the safeguard on git commit, that ensure
your code is properly following the coding guidelines.

To enable the hook, you need to make sure you have all these tools:
- `goimports`
- `golint`

First, create the pre-commit hook file

```console
$ touch .git/hooks/pre-commit
$ chmod +x .git/hooks/pre-commit
```

Then, paste the following script as the hook content:

```bash
#!/bin/sh

STAGED_GO_FILES=$(git diff --cached --name-only | grep ".go$")

if [[ "$STAGED_GO_FILES" = "" ]]; then
  exit 0
fi

PASS=true

for FILE in $STAGED_GO_FILES
do
  ${GOPATH}/bin/goimports -w $FILE

  ${GOPATH}/bin/golint "-set_exit_status" $FILE
  if [[ $? == 1 ]]; then
    PASS=false
  fi

#  go vet $FILE
#  if [[ $? != 0 ]]; then
#    PASS=false
#  fi
done

if ! $PASS; then
  printf "COMMIT FAILED\n"
  exit 1
else
  printf "COMMIT SUCCEEDED\n"
fi

exit 0
```

or Copy/Paste the `pre-commit.sh` file content into your hook.
It contains a colored-version of the above script.


#### SARIF report validation

To check the generated SARIF report, please use the online validator
available at: [https://sarifweb.azurewebsites.net/Validation](https://sarifweb.azurewebsites.net/Validation)

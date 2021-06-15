# Contribute

Contributors to this project are welcome as this is an open-source effort that seeks discussions and continuous improvement.

[join]: https://join.chaostoolkit.org/

From a code perspective, if you wish to contribute, you will need to use Golang 1.15 version. Then, fork this repository and submit a PR. The project cares for code readability and checks the code style to match best practices. Please also make sure you provide tests
whenever you submit a PR so we keep the code reliable.


The Reliably CLI project require all contributors must sign a
[Developer Certificate of Origin][dco] on each commit they would like to merge into the master branch of the repository. Please, make sure you can abide by the rules of the DCO before submitting a PR.

[dco]: https://github.com/probot/dco#how-it-works


## How to checkout the code

You must have a working [Go environment].

```
$ mkdir -p $GOPATH/src/github.com/reliablyhq
$ cd $GOPATH/src/github.com/reliablyhq
$ git clone github.com/reliablyhq/cli
$ cd cli
```

[Go environment]: https://golang.org/doc/install
## How to build

Run:

```
$ make
```

## How to run unit tests

Run:

```
$ make test
```

Or with verbose mode enabled:

```
$ make test/debug
```

### Code coverage

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

## Go coding style

All code must be formatted with `gofmt` and `goimports`.

### Format your source code

Gofmt is a tool that automatically formats Go source code.

You can re-format your code, in place, using the command:

```
$ make format
```

### Organize your imports

Imports should be added and sorted by
[`goimports`](https://godoc.org/golang.org/x/tools/cmd/goimports).

First, install the `goimports` tool

```
$ go install golang.org/x/tools/cmd/goimports
```

Then, run:

```
make imports
```

Imports shall be organized in three groups, separated by a newline:
* Standard library
* Third parties / External packages
* Internal CLI packages

```go
import (
  "fmt"
  "strings"

  "github.com/spf13/cobra"
  "github.com/spf13/viper"

  "github.com/reliablyhq/cli/core"
  "github.com/reliablyhq/cli/version"
)
```

### Check for code style mistakes

First, install the `golint` tool

```
$ go install golang.org/x/lint/golint
```

Then, run it:

```
$ make lint
```

### Check coding style on git prehook

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


### SARIF report validation

To check the generated SARIF report, please use the online validator
available at: [https://sarifweb.azurewebsites.net/Validation](https://sarifweb.azurewebsites.net/Validation)

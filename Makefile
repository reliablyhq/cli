VERSION=$(shell git describe --tags)
BUILD=$(shell date +%Y-%m-%d)

LDFLAGS := -X github.com/reliablyhq/cli/version.Version=${VERSION}
LDFLAGS := $(LDFLAGS) -X github.com/reliablyhq/cli/version.Date=${BUILD}
LDFLAGS := -ldflags "$(LDFLAGS)"

build:
	go build -o bin/reliably main.go
	#go build -gcflags="-m=2" -o bin/reliably main.go

build/docker:
	#docker build --no-cache -t reliably/cli:latest -f Dockerfile --progress plain .
	docker build -t reliably/cli:latest -f Dockerfile .
	# docker build  --build-arg VERSION=$(git describe --tags) --build-arg BUILD_DATE=$(date +%Y-%m-%d) -t reliably/cli:test -f Dockerfile .

compile:
	GOOS=darwin GOARCH=amd64 go build -o bin/main-darwin-amd64 main.go
	GOOS=linux GOARCH=amd64 go build -o bin/main-linux-amd64 main.go
	#GOOS=windows GOARCH=amd64 go build -o bin/main-windows-amd64 main.go

release:
	go build ${LDFLAGS} -o bin/reliably main.go

.PHONY: test
test:
	go test ./...

.PHONY: test/debug
test/debug:
	go test -v ./...

.PHONY: test/coverage
test/coverage:
	go test --coverprofile cover.out ./...

show/coverage:
	go tool cover -html=cover.out

requirements:
	go mod tidy -v

lint:
	${GOPATH}/bin/golint ./...

# format source code in place
format:
	go fmt ./...

imports:
	${GOPATH}/bin/goimports -w -l .

## Docs tasks
.PHONY: docs
docs: clean-docs markdown manpages

clean-docs:
	rm -rf ./docs

markdown:
	rm -rf ./docs/markdown
	mkdir -p docs/markdown
	go run ./cmd/doc markdown --output-dir ./docs/markdown

manpages:
	rm -rf ./docs/man
	mkdir -p docs/man/man1
	go run ${LDFLAGS} ./cmd/doc man --output-dir ./docs/man/man1

## Install/uninstall tasks are here for use on *nix platform. On Windows, there is no equivalent.

DESTDIR :=
prefix  := /usr/local
bindir  := ${prefix}/bin
mandir  := ${prefix}/share/man

.PHONY: install
install: release manpages
	install -d ${DESTDIR}${bindir}
	install -m755 bin/reliably ${DESTDIR}${bindir}/
	install -d ${DESTDIR}${mandir}/man1
	install -m644 ./docs/man/man1/* ${DESTDIR}${mandir}/man1/

.PHONY: uninstall
uninstall:
	rm -f ${DESTDIR}${bindir}/reliably ${DESTDIR}${mandir}/man1/reliably.1 ${DESTDIR}${mandir}/man1/reliably-*.1


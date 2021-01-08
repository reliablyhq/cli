build:
	go build -o bin/reliably main.go
	#go build -gcflags="-m" -o bin/reliably main.go

build/docker:
	#docker build --no-cache -t reliably/cli:latest -f Dockerfile --progress plain .
	docker build -t reliably/cli:latest -f Dockerfile .

compile:
	GOOS=darwin GOARCH=amd64 go build -o bin/main-darwin-amd64 main.go
	GOOS=linux GOARCH=amd64 go build -o bin/main-linux-amd64 main.go
	#GOOS=windows GOARCH=amd64 go build -o bin/main-windows-amd64 main.go

.PHONY: test
test:
	go test ./...

test/debug:
	go test -v ./...

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

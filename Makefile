BINARY_NAME=cctop
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS=-ldflags "-X github.com/fanghanjun/cctop/pkg/version.Version=$(VERSION) -X github.com/fanghanjun/cctop/pkg/version.BuildTime=$(BUILD_TIME)"

.PHONY: build test run lint coverage clean

build:
	CGO_ENABLED=0 go build $(LDFLAGS) -o bin/$(BINARY_NAME) .

test:
	go test ./... -v -race

coverage:
	go test ./... -coverprofile=coverage.out -race
	go tool cover -func=coverage.out

lint:
	golangci-lint run

run: build
	./bin/$(BINARY_NAME)

clean:
	rm -rf bin/ dist/ coverage.out

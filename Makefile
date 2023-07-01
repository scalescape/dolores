VERSION=$(shell git tag --sort=-version:refname | head -1)
SHA=$(shell git rev-parse --short HEAD)

LDFLAGS=-X 'main.Version=$(VERSION)' -X 'main.Sha=$(SHA)'

.PHONY: setup build build_linux test run clean all

.DEFAULT_GOAL: default

default: build test

setup:
	mkdir -p bin
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.46.2

install:
	go install --ldflags="${LDFLAGS}" ./cmd/dolores/

lint: setup
	./bin/golangci-lint run

test: lint
	go test ./...

gomod:
	go mod tidy

build: gomod
	go build --ldflags="${LDFLAGS}" -o ./bin ./cmd/dolores/

gorelease_snapshot: build
	goreleaser release --snapshot --rm-dist

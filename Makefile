VERSION=$(shell git tag --sort=-version:refname | head -1)
SHA=$(shell git rev-parse --short HEAD)

LDFLAGS=-X 'main.version=$(VERSION)' -X 'main.sha=$(SHA)'

.PHONY: setup build build_linux test run clean all

.DEFAULT_GOAL: default

default: build test

setup:
	mkdir -p bin
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.46.2

install:
	go install --ldflags="${LDFLAGS}" ./cmd/rom/

lint: setup
	./bin/golangci-lint run

test: lint
	go test ./...

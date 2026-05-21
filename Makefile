BINARY     := devup
MODULE     := github.com/RajarshiParmar/devup
CMD        := ./cmd/devup
BIN_DIR    := $(HOME)/bin

VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT     := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE       := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS    := -ldflags "\
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(DATE)"

.PHONY: all build install test lint clean doctor release help

all: build

## build: compile the binary to ./bin/devup
build:
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/$(BINARY) $(CMD)

## install: build and copy binary to ~/bin/devup
install: build
	@mkdir -p $(BIN_DIR)
	cp bin/$(BINARY) $(BIN_DIR)/$(BINARY)
	@echo "installed to $(BIN_DIR)/$(BINARY)"

## test: run all tests with race detector and coverage
test:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

## lint: run golangci-lint
lint:
	golangci-lint run ./...

## doctor: run devup doctor using the local build
doctor: build
	./bin/$(BINARY) doctor

## tidy: tidy and verify go modules
tidy:
	go mod tidy
	go mod verify

## clean: remove build artifacts
clean:
	rm -rf bin/ coverage.out

## release: create a release with goreleaser (requires a git tag)
release:
	goreleaser release --clean

## release-snapshot: dry-run release without publishing
release-snapshot:
	goreleaser release --snapshot --clean

## help: print this help message
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'

VERSION ?= 0.2.0
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo unknown)

LDFLAGS = -ldflags "-s -w \
	-X github.com/Dendro-X0/Orbit/internal/cli.Version=$(VERSION) \
	-X github.com/Dendro-X0/Orbit/internal/cli.Commit=$(COMMIT) \
	-X github.com/Dendro-X0/Orbit/internal/cli.BuildDate=$(DATE)"

.PHONY: build test ci

build:
	go build $(LDFLAGS) -o orbit ./cmd/orbit

test:
	go test ./...

ci: test build
	./orbit version

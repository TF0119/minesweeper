BINARY := bin/minesweeper
PKG := ./cmd/minesweeper

.DEFAULT_GOAL := help

## help: list the available targets
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## /  /'

## build: compile the binary to bin/minesweeper
build:
	go build -o $(BINARY) $(PKG)

## run: build and play
run: build
	./$(BINARY)

## test: run the test suite with the race detector
test:
	go test -race ./...

## cover: report test coverage per package
cover:
	go test -cover ./...

## fmt: rewrite sources with gofmt
fmt:
	gofmt -w .

## lint: gofmt check, go vet, and golangci-lint when installed
lint:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "These files need gofmt:"; echo "$$unformatted"; exit 1; \
	fi
	go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping"; \
	fi

## check: everything CI runs
check: lint test
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser check; \
	else \
		echo "goreleaser not installed, skipping config check"; \
	fi

## demo: re-record docs/demo.gif with vhs
demo: build
	vhs docs/demo.tape

## clean: remove build output
clean:
	rm -rf bin dist

.PHONY: help build run test cover fmt lint check demo clean

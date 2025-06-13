run *args:
	go run cmd/sandworm/main.go {{args}}

fmt:
	go fmt ./...
	go run golang.org/x/tools/cmd/goimports@latest -w .

test *args:
	go test -v -race -cover ./... {{args}}

lint:
	golangci-lint run ./...

build:
	go build -o bin/sandworm ./cmd/sandworm

update-deps:
	# Update deps to their latest compatible versions
	go get -u ./...
	go mod tidy

install:
	#!/usr/bin/env sh
	set -euo pipefail

	# Determine OS and architecture
	OS=$(uname -s | tr '[:upper:]' '[:lower:]')
	ARCH=$(uname -m)
	if [ "$ARCH" = "x86_64" ]; then
		ARCH="amd64"
	elif [ "$ARCH" = "arm64" ]; then
		ARCH="arm64"
	fi

	# Clean dist and build binary snapshot for local use
	goreleaser build --clean --snapshot

	# Assume v8.0 (x86-64-v3) since it has better performance in modern CPUs (at
	# the cost of compatibility with older CPUs)
	BINARY="dist/sandworm_${OS}_${ARCH}_v8.0/sandworm"

	# Overzealous check; if build above succeeded, binary should exist.
	if [ ! -f "$BINARY" ]; then
		echo "Error: Binary not found at $BINARY" >&2
		exit 1
	fi

	# Install as sandworm-dev to $GOBIN
	# (allows coexisting with the real sandworm binary elsewhere in PATH)
	BIN_PATH="$(go env GOBIN)"
	if [ -z "$BIN_PATH" ]; then
		echo "Error: GOBIN is not set" >&2
		exit 1
	fi

	echo "Installing to $BIN_PATH/sandworm-dev..." >&2
	cp "$BINARY" "$BIN_PATH/sandworm-dev"
	chmod +x "$BIN_PATH/sandworm-dev"

uninstall:
	rm -f "$(go env GOBIN)/sandworm-dev"

clean:
	rm -rf bin/ dist/

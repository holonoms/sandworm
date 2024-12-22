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

install version="dev":
	#!/usr/bin/env sh
	echo "Installing to $(go env GOPATH)/bin/sandworm..." >&2; \
	go install -ldflags="-X main.version={{version}}" ./cmd/sandworm
	echo "" >&2; \
	echo "If \$GOPATH is in your \$PATH, you can now run 'sandworm'." >&2; \
	echo "Otherwise, add the following to your shell configuration:" >&2; \
	echo "" >&2; \
	echo "    export GOPATH=\"\$(go env GOPATH)\"" >&2; \
	echo "    export PATH=\"\$GOPATH/bin:\$PATH\"" >&2; \

clean:
	rm -rf bin/

.PHONY: build test lint clean

# Build the binary
build:
	go build -o bin/sandworm ./cmd/sandworm

run:
	go run cmd/sandworm/main.go

# Format code
fmt:
	go fmt ./...
	go run golang.org/x/tools/cmd/goimports@latest -w .

# Run tests
test:
	go test -v -race -cover ./...

# Run linter
lint:
	golangci-lint run ./...

# Clean build artifacts
clean:
	rm -rf bin/

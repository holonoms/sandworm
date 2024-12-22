.PHONY: build test lint clean

# Run the source code
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

# Build the binary
build:
	go build -o bin/sandworm ./cmd/sandworm

# Clean build artifacts
clean:
	rm -rf bin/

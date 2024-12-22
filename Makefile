.PHONY: build test lint clean

# Build the binary
build:
	go build -o bin/sandworm ./cmd/sandworm

run:
	go run cmd/sandworm/main.go

# Run tests
test:
	go test -v -race -cover ./...

# Run linter
lint:
	golangci-lint run ./...

# Clean build artifacts
clean:
	rm -rf bin/

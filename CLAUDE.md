# Sandworm Development Guide

## Commands
- Build: `just build` - creates binary at bin/sandworm
- Run: `just run [args]` - run from source with arguments
- Test: `just test` - runs all tests with race detection and coverage
- Test single: `go test -v ./internal/package -run TestName`
- Lint: `just lint` - runs golangci-lint
- Format: `just fmt` - formats with gofmt and goimports
- Install: `just install` - builds and installs to $GOPATH/bin using goreleaser

## Code Style
- Follow standard Go conventions (gofmt compliant)
- Imports: stdlib first, then third-party, alphabetically sorted
- Error handling: wrap errors with context using `fmt.Errorf("context: %w", err)`
- Naming: PascalCase for exported, camelCase for unexported
- Documentation: all functions and packages have doc comments
- Tests: table-driven tests with t.Run subtests
- Structure: keep packages small and focused on single responsibility
- Permissions: use explicit octal literals (`0o644`, `0o755`)
- Organization: helper functions grouped with "MARK:" comments

## Project Structure
- `cmd/` - application entrypoints
- `internal/` - private implementation packages
- `bin/` - build artifacts (not committed)

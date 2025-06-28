[private]
default: help

help:
    just --list --justfile {{justfile()}}

# Build the application
build:
    go build -o bin/server ./cmd

# Run the application
run:
    go run ./cmd

# Run tests
test:
    go test -v ./...

# Run tests with coverage
test-coverage:
    go test -v -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out

# Format code
fmt:
    go fmt ./...

# Run linter (requires golangci-lint)
lint:
    golangci-lint run

# Clean build artifacts
clean:
    rm -rf bin/
    rm -f coverage.out

# Tidy dependencies
tidy:
    go mod tidy

# Makefile for stats application

.PHONY: help test test-verbose test-coverage build clean lint fmt vet run

# Default target
help:
	@echo "Available targets:"
	@echo "  test          - Run tests"
	@echo "  test-verbose  - Run tests with verbose output"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  build         - Build the application"
	@echo "  build-all     - Build for all platforms"
	@echo "  clean         - Clean build artifacts"
	@echo "  lint          - Run linter"
	@echo "  fmt           - Format code"
	@echo "  vet           - Run go vet"
	@echo "  run           - Run the application"

# Test targets
test:
	go test ./...

test-verbose:
	go test -v ./...

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Build targets
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o stats ./cmd/stats

build-all:
	@echo "Building for all platforms..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o stats-linux-amd64 ./cmd/stats
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -trimpath -o stats-linux-arm64 ./cmd/stats
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o stats-windows-amd64.exe ./cmd/stats
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o stats-darwin-amd64 ./cmd/stats
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -trimpath -o stats-darwin-arm64 ./cmd/stats
	@echo "Build complete. Artifacts:"
	@ls -la stats-*

# Clean target
clean:
	rm -f stats stats-* coverage.out coverage.html

# Code quality targets
lint:
	golangci-lint run

fmt:
	go fmt ./...

vet:
	go vet ./...

# Run target
run: build
	./stats

# Development targets
dev-deps:
	go mod download
	go mod verify

# Install development tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Security scan
security:
	gosec ./...

# All quality checks
check: fmt vet lint test
	@echo "All quality checks passed!"

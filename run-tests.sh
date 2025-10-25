#!/bin/bash

# Test runner script for stats application

set -e

echo "Running tests for stats application..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed or not in PATH"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "Go version: $GO_VERSION"

# Run tests
echo "Running unit tests..."
go test -v ./...

echo "Running tests with coverage..."
go test -v -coverprofile=coverage.out ./...

# Generate coverage report
echo "Generating coverage report..."
go tool cover -html=coverage.out -o coverage.html

echo "Coverage report generated: coverage.html"

# Run additional checks
echo "Running go vet..."
go vet ./...

echo "Running go fmt check..."
if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
    echo "Code is not formatted. Run 'go fmt ./...' to fix."
    exit 1
fi

echo "All tests passed!"

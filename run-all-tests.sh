#!/bin/bash

# Test runner script for stats application - runs individual test files

set -e

echo "Running all tests for stats application..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed or not in PATH"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "Go version: $GO_VERSION"

# Test files and their corresponding source files
declare -A test_files=(
    ["env_test.go"]="env.go"
    ["calcsize_test.go"]="calcsize.go"
    ["log_test.go"]="log.go main.go env.go logic.go sab.go models.go size.go calcsize.go html.go maintenance.go"
    ["models_test.go"]="models.go"
    ["html_test.go"]="html.go main.go env.go logic.go sab.go models.go size.go calcsize.go maintenance.go log.go"
    ["maintenance_test.go"]="maintenance.go log.go main.go env.go logic.go sab.go models.go size.go calcsize.go html.go"
    ["sab_test.go"]="sab.go models.go main.go env.go logic.go size.go calcsize.go html.go maintenance.go log.go"
    ["size_test.go"]="size.go log.go main.go env.go logic.go sab.go models.go calcsize.go html.go maintenance.go"
)

# Run tests for each file
for test_file in "${!test_files[@]}"; do
    source_file="${test_files[$test_file]}"
    echo "Running tests for $test_file..."
    
    if go test -v "./$test_file" $source_file; then
        echo "✅ $test_file passed"
    else
        echo "❌ $test_file failed"
        exit 1
    fi
done

echo "All tests passed! 🎉"

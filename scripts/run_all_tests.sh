#!/usr/bin/env bash
# Script to run all tests including integration tests

set -euo pipefail

echo "Running all tests (including integration tests)..."

# Check if Docker is available
if ! docker info &> /dev/null; then
    echo "Warning: Docker is not available. Some integration tests will be skipped."
fi

# Run all tests with longer timeout for integration tests
go test \
  -timeout 30m \
  -race \
  ./... \
  -v
#!/usr/bin/env bash
# Script to run only integration tests

set -euo pipefail

echo "Running integration tests only..."

# Check if Docker is available
if ! docker info &> /dev/null; then
    echo "Error: Docker is required for integration tests but is not available."
    exit 1
fi

# Environment variables to ensure integration tests run
unset SKIP_DOCKER_TESTS
unset SKIP_INTEGRATION_TESTS

# Run only integration tests
# We use -run to select tests that typically are integration tests
go test \
  -timeout 30m \
  -race \
  -run="TestIntegration|TestContainer|TestWithDocker" \
  ./... \
  -v
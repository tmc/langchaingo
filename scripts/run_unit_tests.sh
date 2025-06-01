#!/usr/bin/env bash
# Script to run unit tests (excluding integration tests that require external services)

set -euo pipefail

echo "Running unit tests (excluding integration tests)..."

# Tags to skip integration tests
SKIP_TAGS="integration,docker,testcontainers"

# Environment variables to skip tests that check for external services
export SKIP_DOCKER_TESTS=1
export SKIP_INTEGRATION_TESTS=1

# Run tests with timeout, excluding integration tests
go test \
  -timeout 5m \
  -tags="!${SKIP_TAGS// /,!}" \
  -short \
  ./... \
  -v
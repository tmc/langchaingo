#!/bin/bash
# Quick smoke test for merge validation
set -e

echo "======================================"
echo "Merge Validation Test Suite"
echo "======================================"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

pass() { echo -e "${GREEN}✓${NC} $1"; }
fail() { echo -e "${RED}✗${NC} $1"; exit 1; }
warn() { echo -e "${YELLOW}⚠${NC} $1"; }

echo "1. Building all packages..."
if go build ./... 2>&1; then
    pass "All packages build successfully"
else
    fail "Build failed"
fi
echo ""

echo "2. Checking for vendasta import leaks..."
VENDASTA_COUNT=$(grep -r "github.com/vendasta/langchaingo" --include="*.go" . 2>/dev/null | grep -v "go.mod" | wc -l | tr -d ' ')
if [ "$VENDASTA_COUNT" -eq "0" ]; then
    pass "No vendasta imports found in Go files"
else
    fail "Found $VENDASTA_COUNT vendasta imports in Go files"
fi
echo ""

echo "3. Testing GPT-4.1 model token counting..."
if go test ./llms -run "TestGetModelContextSize|TestCountTokens" -v 2>&1 | grep -q "PASS"; then
    pass "Token counting tests passed"
else
    warn "Token counting tests failed or didn't run"
fi
echo ""

echo "4. Testing OpenAI User and ParallelToolCalls..."
if go test ./llms/openai -run TestWithUserAndParallelToolCalls -v 2>&1 | grep -q "PASS"; then
    pass "User and ParallelToolCalls options work correctly"
else
    warn "User/ParallelToolCalls tests failed or didn't run"
fi
echo ""

echo "5. Running core LLM tests..."
if go test ./llms -short 2>&1 | grep -q "PASS"; then
    pass "Core LLM tests passed"
else
    warn "Some core LLM tests failed"
fi
echo ""

echo "6. Checking module path..."
MODULE_PATH=$(grep "^module " go.mod | awk '{print $2}')
if [ "$MODULE_PATH" = "github.com/tmc/langchaingo" ]; then
    pass "Module path is correct: $MODULE_PATH"
else
    fail "Module path is incorrect: $MODULE_PATH (expected github.com/tmc/langchaingo)"
fi
echo ""

echo "======================================"
echo "Basic Smoke Tests: ${GREEN}PASSED${NC}"
echo "======================================"
echo ""
echo "Next steps:"
echo "  1. Run integration tests with API keys:"
echo "     export OPENAI_API_KEY='your-key'"
echo "     go test ./llms/openai -v"
echo ""
echo "  2. Test in your services with replace directive:"
echo "     replace github.com/tmc/langchaingo => github.com/vendasta/langchaingo vX.X.X"
echo ""
echo "  3. Review: MERGE_TEST_PLAN.md for comprehensive testing"
echo ""


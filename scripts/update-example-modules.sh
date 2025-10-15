#!/bin/bash
# update-example-modules.sh - Update all example go.mod files to use local version
#
# This script updates example modules to use replace directives pointing to the
# local version of langchaingo instead of downloading from GitHub.
#
# Usage: ./scripts/update-example-modules.sh

set -e

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EXAMPLES_DIR="$REPO_ROOT/examples"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo ""
echo "════════════════════════════════════════════════════════════════"
echo "  Update Example Modules"
echo "════════════════════════════════════════════════════════════════"
echo ""

count=0

# Find all example go.mod files
while IFS= read -r -d '' gomod; do
    dir=$(dirname "$gomod")
    example_name=$(basename "$dir")
    
    echo -e "${BLUE}ℹ${NC} Processing: $example_name"
    
    # Check if go.mod already has a replace directive
    if grep -q "^replace github.com/.*/langchaingo" "$gomod"; then
        echo "  Already has replace directive, updating..."
        # Update existing replace (cross-platform sed)
        sed -i.bak 's|^replace github.com/.*/langchaingo.*|replace github.com/vendasta/langchaingo => ../..|' "$gomod" && rm -f "${gomod}.bak"
    else
        # Add replace directive at the end
        echo "" >> "$gomod"
        echo "replace github.com/vendasta/langchaingo => ../.." >> "$gomod"
    fi
    
    # Also update the module path and require if needed (cross-platform sed)
    sed -i.bak 's|module github.com/tmc/langchaingo/examples/|module github.com/vendasta/langchaingo/examples/|' "$gomod" && rm -f "${gomod}.bak"
    sed -i.bak 's|require github.com/tmc/langchaingo|require github.com/vendasta/langchaingo|' "$gomod" && rm -f "${gomod}.bak"
    
    # Run go mod tidy in the example directory
    (cd "$dir" && go mod tidy 2>/dev/null || true)
    
    count=$((count + 1))
done < <(find "$EXAMPLES_DIR" -name "go.mod" -type f -print0)

echo ""
echo "════════════════════════════════════════════════════════════════"
echo -e "${GREEN}✓${NC} Updated $count example modules"
echo "════════════════════════════════════════════════════════════════"
echo ""
echo "All examples now use: replace github.com/vendasta/langchaingo => ../.."
echo ""


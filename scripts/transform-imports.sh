#!/bin/bash
# transform-imports.sh - Bidirectional import transformer for langchaingo fork
#
# Usage:
#   ./scripts/transform-imports.sh to-tmc      # vendasta → tmc (before upstream merge)
#   ./scripts/transform-imports.sh to-vendasta # tmc → vendasta (after upstream merge)
#   ./scripts/transform-imports.sh --dry-run to-tmc  # Preview changes
#
# This script helps manage the fork by transforming imports between:
#   github.com/vendasta/langchaingo ←→ github.com/tmc/langchaingo

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FROM_ORG=""
TO_ORG=""
DRY_RUN=false

# Statistics
FILES_CHANGED=0
TOTAL_REPLACEMENTS=0

# Function to print colored output
print_info() { echo -e "${BLUE}ℹ${NC} $1"; }
print_success() { echo -e "${GREEN}✓${NC} $1"; }
print_warning() { echo -e "${YELLOW}⚠${NC} $1"; }
print_error() { echo -e "${RED}✗${NC} $1"; }

# Function to show usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS] DIRECTION

Transform imports between tmc and vendasta organizations.

DIRECTION:
    to-tmc          Transform vendasta → tmc (before upstream merge)
    to-vendasta     Transform tmc → vendasta (after upstream merge)

OPTIONS:
    --dry-run       Preview changes without modifying files
    --help          Show this help message

EXAMPLES:
    # Before merging from upstream
    $0 to-tmc

    # Preview changes
    $0 --dry-run to-tmc

    # After merging from upstream
    $0 to-vendasta

WORKFLOW:
    1. Ensure git working directory is clean
    2. Run transformation to-tmc before upstream merge
    3. Merge from upstream (conflicts will be minimal)
    4. Run transformation to-vendasta to restore fork's imports
    5. Commit the changes

EOF
    exit 1
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --help|-h)
            usage
            ;;
        to-tmc)
            FROM_ORG="vendasta"
            TO_ORG="tmc"
            shift
            ;;
        to-vendasta)
            FROM_ORG="tmc"
            TO_ORG="vendasta"
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            usage
            ;;
    esac
done

# Validate direction was specified
if [[ -z "$FROM_ORG" ]] || [[ -z "$TO_ORG" ]]; then
    print_error "Direction not specified"
    usage
fi

# Banner
echo ""
echo "════════════════════════════════════════════════════════════════"
echo "  LangChain Go Import Transformer"
echo "════════════════════════════════════════════════════════════════"
echo ""
print_info "Direction: ${FROM_ORG} → ${TO_ORG}"
if $DRY_RUN; then
    print_warning "DRY RUN MODE - No files will be modified"
fi
echo ""

# Check if we're in the right directory
if [[ ! -f "$REPO_ROOT/go.mod" ]] || ! grep -q "langchaingo" "$REPO_ROOT/go.mod"; then
    print_error "Not in langchaingo repository root"
    exit 1
fi

# Check git status
if ! git diff-index --quiet HEAD -- 2>/dev/null; then
    print_warning "Git working directory is not clean"
    print_warning "Uncommitted changes detected. Recommended to commit or stash first."
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Aborted by user"
        exit 1
    fi
fi

# Function to transform a file
transform_file() {
    local file="$1"
    local from_import="github.com/${FROM_ORG}/langchaingo"
    local to_import="github.com/${TO_ORG}/langchaingo"
    
    # Check if file contains the pattern
    if ! grep -q "$from_import" "$file" 2>/dev/null; then
        return 0
    fi
    
    # Count occurrences
    local count=$(grep -o "$from_import" "$file" | wc -l | tr -d ' ')
    
    if $DRY_RUN; then
        echo "  Would transform: $file ($count occurrences)"
        TOTAL_REPLACEMENTS=$((TOTAL_REPLACEMENTS + count))
        FILES_CHANGED=$((FILES_CHANGED + 1))
    else
        # Perform the replacement (macOS compatible)
        if [[ "$OSTYPE" == "darwin"* ]]; then
            sed -i '' "s|$from_import|$to_import|g" "$file"
        else
            sed -i "s|$from_import|$to_import|g" "$file"
        fi
        TOTAL_REPLACEMENTS=$((TOTAL_REPLACEMENTS + count))
        FILES_CHANGED=$((FILES_CHANGED + 1))
    fi
}

# Main transformation logic
echo "Scanning repository..."
echo ""

# 1. Transform main go.mod
print_info "Transforming main go.mod..."
if [[ -f "$REPO_ROOT/go.mod" ]]; then
    transform_file "$REPO_ROOT/go.mod"
fi

# 2. Transform all .go files
print_info "Transforming Go source files..."
while IFS= read -r -d '' file; do
    transform_file "$file"
done < <(find "$REPO_ROOT" -name "*.go" -type f -print0 | grep -v "/vendor/" | grep -v "/.git/")

# 3. Transform example go.mod files
print_info "Transforming example go.mod files..."
while IFS= read -r -d '' file; do
    transform_file "$file"
done < <(find "$REPO_ROOT/examples" -name "go.mod" -type f -print0 2>/dev/null || true)

# 4. Transform documentation files
print_info "Transforming documentation files..."
while IFS= read -r -d '' file; do
    transform_file "$file"
done < <(find "$REPO_ROOT" -type f \( -name "*.md" -o -name "*.mdx" \) -print0 | grep -v "/vendor/" | grep -v "/.git/" | grep -v "/node_modules/")

# Summary
echo ""
echo "════════════════════════════════════════════════════════════════"
echo "  Transformation Complete"
echo "════════════════════════════════════════════════════════════════"
echo ""

if $DRY_RUN; then
    print_warning "DRY RUN - No files were modified"
    echo ""
    echo "Summary (would change):"
else
    echo "Summary:"
fi

echo "  Files transformed: $FILES_CHANGED"
echo "  Total replacements: $TOTAL_REPLACEMENTS"
echo "  Direction: github.com/${FROM_ORG}/langchaingo → github.com/${TO_ORG}/langchaingo"
echo ""

if ! $DRY_RUN && [[ $FILES_CHANGED -gt 0 ]]; then
    print_success "Transformation completed successfully"
    echo ""
    print_info "Next steps:"
    if [[ "$TO_ORG" == "tmc" ]]; then
        echo "  1. Review changes: git diff"
        echo "  2. Commit: git add -A && git commit -m 'Transform imports: vendasta → tmc for upstream merge'"
        echo "  3. Merge from upstream: git merge upstream/main"
        echo "  4. After merge, run: $0 to-vendasta"
    else
        echo "  1. Review changes: git diff"
        echo "  2. Verify build: go build ./..."
        echo "  3. Run tests: go test ./..."
        echo "  4. Commit: git add -A && git commit -m 'Transform imports: tmc → vendasta after upstream merge'"
    fi
    echo ""
elif $DRY_RUN && [[ $FILES_CHANGED -gt 0 ]]; then
    echo "To apply these changes, run without --dry-run:"
    echo "  $0 ${TO_ORG##*-}"
    echo ""
elif [[ $FILES_CHANGED -eq 0 ]]; then
    print_warning "No files needed transformation"
    echo ""
    print_info "Current state check:"
    CURRENT_ORG=$(grep "^module github.com/.*/langchaingo" "$REPO_ROOT/go.mod" | sed -E 's|module github.com/([^/]+)/.*|\1|')
    echo "  Module is currently: github.com/${CURRENT_ORG}/langchaingo"
    echo ""
    if [[ "$CURRENT_ORG" == "$TO_ORG" ]]; then
        print_success "Repository is already using ${TO_ORG} organization"
    else
        print_warning "Repository is using ${CURRENT_ORG}, expected ${FROM_ORG}"
    fi
fi

echo ""


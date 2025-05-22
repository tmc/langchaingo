#!/bin/bash
# Script to help update tests to use httprr

set -e  # Exit on error

# Make sure we're in the right directory (project root)
if [ ! -d "internal/httprr" ]; then
    echo "ERROR: This script must be run from the project root directory"
    echo "Current directory: $(pwd)"
    echo "Expected directory structure should contain internal/httprr"
    exit 1
fi

# Function to find test files that might benefit from httprr
find_candidate_tests() {
    echo "Finding test files that might benefit from httprr..."
    
    # Find all test files that mention http.Client or DefaultClient
    grep -r --include="*_test.go" "http\.Client\|DefaultClient" \
        --exclude-dir=internal/httprr \
        --exclude-dir=vendor \
        . | sort
        
    echo ""
    echo "Finding test files that make HTTP requests..."
    
    # Find test files that use Get, Post, etc.
    grep -r --include="*_test.go" "\.Get(\|\.Post(\|\.Do(" \
        --exclude-dir=internal/httprr \
        --exclude-dir=vendor \
        . | sort
}

# Check if we need to install httprr
check_httprr_installation() {
    echo "Checking if httprr is properly installed..."
    
    if [ ! -d "internal/httprr" ]; then
        echo "ERROR: httprr package not found at internal/httprr"
        exit 1
    fi
    
    echo "httprr package is installed and ready to use"
}

# Create an example patched file to show how to update tests
create_example_patch() {
    local example_file="internal/httprr/scripts/example_test_update.diff"
    
    echo "Creating example patch to show how to update tests..."
    
    cat > "$example_file" << 'EOF'
--- old_test.go	2023-01-01 00:00:00.000000000 -0500
+++ new_test.go	2023-01-01 00:00:00.000000000 -0500
@@ -3,6 +3,8 @@
 import (
 	"net/http"
 	"testing"
+	"path/filepath"
+	"github.com/tmc/langchaingo/internal/httprr"
 )
 
 func TestExample(t *testing.T) {
@@ -10,7 +12,10 @@
 	// Test setup
 	
 	// Before: Using default client
-	client := http.DefaultClient
+	// After: Using httprr
+	recordingsDir := filepath.Join("testdata", "example_recordings")
+	httpHelper := httprr.NewAutoHelper(t, recordingsDir)
+	client := httpHelper.Client
 	
 	service := NewService(client)
 	
@@ -18,4 +23,7 @@
 	result := service.DoSomething()
 	
 	// Assert result
+	
+	// Optional: Add assertions about HTTP calls
+	httpHelper.AssertURLCalled("api.example.com")
 }
EOF
    
    echo "Example patch created at $example_file"
}

# Create directories for test recordings
create_testdata_dirs() {
    echo "Creating test data directories for recordings..."
    
    # Create base testdata directory if it doesn't exist
    mkdir -p testdata
    
    echo "Base testdata directory created"
    echo "For each test, you should create specific recording directories like:"
    echo "  mkdir -p testdata/my_test_recordings"
}

# Main script
main() {
    echo "======= httprr Test Update Helper ======="
    echo "This script helps identify and update tests to use httprr for HTTP recording and replay"
    echo ""
    
    check_httprr_installation
    create_example_patch
    create_testdata_dirs
    
    echo ""
    echo "Candidate test files to update:"
    find_candidate_tests
    
    echo ""
    echo "======= Next Steps ======="
    echo "1. Review the candidate test files above"
    echo "2. For each test file, update it to use httprr.NewAutoHelper"
    echo "3. Create test data directories for recordings"
    echo "4. Run tests in record mode first (default)"
    echo "5. Then run tests in replay mode: HTTPRR_MODE=replay go test ./..."
    echo ""
    echo "See internal/httprr/UPDATING_TESTS.md for more detailed instructions"
    echo "See internal/httprr/scripts/example_test_update.diff for example changes"
}

main "$@" 
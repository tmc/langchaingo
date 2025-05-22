#!/bin/bash
# find_defaultclient_tests.sh
# Finds test files that use http.DefaultClient and need to be updated to use httprr

set -e

echo "Looking for test files that use http.DefaultClient..."
echo

# Find test files that import net/http
TEST_FILES=$(find . -name "*_test.go" -type f -not -path "./internal/httprr/*" -exec grep -l "\"net/http\"" {} \;)

# Check each file for http.DefaultClient usage
for file in $TEST_FILES; do
    if grep -q "http\.DefaultClient" "$file"; then
        echo "Found usage in $file:"
        grep -n "http\.DefaultClient" "$file" | while read -r line; do
            echo "  $line"
        done
        echo
    fi
done

echo "Looking for files that create their own http.Client..."
echo

# Find test files that create new http.Client instances
for file in $TEST_FILES; do
    if grep -q "http\.Client{" "$file" || grep -q "&http\.Client" "$file"; then
        echo "Found client creation in $file:"
        grep -n -e "http\.Client{" -e "&http\.Client" "$file" | while read -r line; do
            echo "  $line"
        done
        echo
    fi
done

echo "Done! These files should be updated to use the httprr package."
echo "See internal/httprr/UPDATING_TESTS.md for instructions." 
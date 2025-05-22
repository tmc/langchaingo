# Implementation Plan for Migrating Tests to httprr

This document outlines a step-by-step approach for migrating existing tests in the codebase to use the httprr package for HTTP recording and replay.

## Phase 1: Preparation

1. **Identify Candidates**
   - Run the `update_tests.sh` script to identify test files that make HTTP requests
   - Prioritize tests based on:
     - Test importance/criticality
     - Test flakiness (prioritize tests that sometimes fail due to network issues)
     - Test execution time (prioritize slow tests that make multiple HTTP calls)

2. **Setup Test Infrastructure**
   - Create base directories for test recordings
   - Decide on naming conventions for recording directories
   - Add recording directories to .gitignore if not committing them

3. **Create Migration Documentation**
   - Document the migration process for developers
   - Set up CI checks for new tests to ensure they use httprr where appropriate

## Phase 2: Initial Migration

1. **Start with Simple Tests**
   - Choose simple tests that use http.DefaultClient directly
   - Apply the Pattern 1 approach from UPDATING_TESTS.md
   - Verify tests pass in both record and replay mode

2. **Move to Complex Tests**
   - Update tests that use custom HTTP clients
   - Apply the Pattern 2 approach from UPDATING_TESTS.md
   - Handle tests with special requirements (custom timeouts, etc.)

3. **Handle Special Cases**
   - Update tests that use http.DefaultClient indirectly
   - Apply the Pattern 3 approach with ReplaceDefaultClient
   - Create fixtures for tests that need predetermined responses

## Phase 3: Validation and Cleanup

1. **Verify Tests in CI**
   - Run tests in both record and replay mode in CI
   - Ensure tests are deterministic in replay mode
   - Check for improved test speed in replay mode

2. **Add Assertions**
   - Enhance tests with HTTP assertions (AssertURLCalled, etc.)
   - Add detailed debugging with DumpRecordings for verbose mode

3. **Optimize Recordings**
   - Clean up/minimize recordings where necessary
   - Consider committing recordings for critical tests
   - Set up periodic refresh of recordings for tests against evolving APIs

## Phase 4: Standardization

1. **Update Test Helpers**
   - Create standardized test helpers for common HTTP test patterns
   - Update test documentation to reflect new patterns

2. **Enforce in Code Reviews**
   - Add linter rules or code review guidelines
   - Document best practices for new tests

3. **Monitor Performance**
   - Track test execution time improvements
   - Monitor flakiness reduction
   - Adjust approach based on metrics

## Migration Checklist for Each Test

- [ ] Identify HTTP client usage pattern
- [ ] Choose appropriate httprr pattern
- [ ] Create recordings directory
- [ ] Update imports
- [ ] Replace client with httprr client
- [ ] Add assertions about HTTP interactions
- [ ] Verify test passes in record mode
- [ ] Verify test passes in replay mode
- [ ] Add debugging output for verbose mode
- [ ] Update documentation if needed

## Commands for Testing

```bash
# Run tests in record mode (default)
go test ./path/to/package

# Run tests in replay mode
HTTPRR_MODE=replay go test ./path/to/package

# Run tests with verbose output to see HTTP interactions
HTTPRR_MODE=replay go test -v ./path/to/package
```

## Resources

- [UPDATING_TESTS.md](../UPDATING_TESTS.md) - Detailed guide on updating tests
- [README.md](../README.md) - httprr package documentation
- [update_tests.sh](./update_tests.sh) - Script to identify candidate tests 
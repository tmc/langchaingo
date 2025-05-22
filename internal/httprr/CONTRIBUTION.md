# Contributing to httprr

The `httprr` package provides HTTP request and response recording functionality for tests, inspired by how Russ Cox would design such a package - simple, focused, and useful.

## Updating Tests to Use httprr

We're gradually updating all tests in the codebase to use `httprr` for HTTP traffic recording. This will help:

1. Make tests more robust by enabling assertions about HTTP requests
2. Provide better debugging information when tests fail
3. Ensure consistent handling of HTTP interactions across the codebase
4. Document the HTTP traffic for each test case
5. Allow tests to run without network connectivity
6. Make tests more deterministic by eliminating network flakiness

## How to Help

### Find Tests that Need Updating

Use the provided script to find tests that still use `http.DefaultClient` or create their own `http.Client` instances:

```bash
# From the root of the repository
./internal/httprr/scripts/find_defaultclient_tests.sh
```

### Update the Tests

Follow the guidelines in [UPDATING_TESTS.md](./UPDATING_TESTS.md) to update each test. The basic steps are:

1. Import the `httprr` package
2. Create a permanent recordings directory in testdata for the test
3. Use `NewAutoHelper` to create a test helper that can work in both record and replay modes
4. Replace `http.DefaultClient` or custom client with the `TestHelper.Client`
5. Add assertions about HTTP traffic where relevant
6. Add debugging output for verbose mode

### Testing Your Changes

After making changes:

1. Run the test you modified in record mode to capture HTTP traffic:
   ```bash
   # Record mode
   go test -v ./path/to/package
   ```

2. Run the test in replay mode to ensure it works offline:
   ```bash
   # Replay mode
   HTTPRR_MODE=replay go test -v ./path/to/package
   ```

3. Verify that any assertions you added about HTTP traffic work correctly

### Submit a Pull Request

When submitting a PR:

1. Include only one test file per PR to keep changes focused
2. Explain which tests you updated and why
3. Include the recorded HTTP traffic in testdata directories if possible (for fixed APIs)
4. Verify the tests pass in CI in replay mode

## Offline Testing in CI

One of the main benefits of `httprr` is enabling tests to run without network access. This is particularly useful in CI environments where:

- Network connections might be restricted
- Tests should be deterministic
- Network latency can slow down tests
- External APIs might be rate-limited

To run tests in replay mode in CI, add the following to your CI configuration:

```yaml
# GitHub Actions example
env:
  HTTPRR_MODE: replay

# ... rest of your CI config
```

## Design Principles for httprr

When extending the `httprr` package, adhere to these principles:

1. **Simplicity**: Keep the API simple and focused
2. **Standard library alignment**: Follow patterns from the standard library
3. **Test-focused**: Optimize for testing use cases
4. **Low overhead**: Minimize performance impact
5. **Compatibility**: Don't break existing code
6. **Offline-first**: Prioritize making tests work without network access

## Future Work

Some ideas for future improvements:

1. Enhance HTTP traffic replay functionality with better matching algorithms
2. Add more assertion helpers for common HTTP testing patterns
3. Create a way to match requests by method, URL pattern, etc.
4. Support for WebSocket and HTTP/2 specific features
5. Integration with the standard `httptest` package
6. Provide a way to manipulate recorded responses for testing edge cases

## Questions?

If you have questions about how to use `httprr` or need help updating tests, please reach out to the maintainers. 
# Guide to Updating Tests to Use httprr

This document explains how to update existing tests in the codebase to use the new `httprr` package for recording HTTP traffic.

## Why Use httprr?

The `httprr` package provides several benefits for tests:

1. **Visibility**: Record all HTTP traffic for inspection and debugging
2. **Assertions**: Make assertions about HTTP requests in tests
3. **Debugging**: Detailed output of HTTP interactions when tests fail
4. **Consistency**: Consistent approach for dealing with HTTP in tests
5. **Test Isolation**: Each test has its own recorder, preventing cross-test contamination
6. **Offline Testing**: Tests can run without network access once responses are recorded

## Update Patterns

### Pattern 1: Replace direct use of http.DefaultClient with AutoHelper

**Before:**
```go
func TestExample(t *testing.T) {
    // Test setup
    
    client := http.DefaultClient
    service := myservice.New(client)
    
    // Run test
}
```

**After:**
```go
import (
    "path/filepath"
    "github.com/tmc/langchaingo/internal/httprr"
)

func TestExample(t *testing.T) {
    // Test setup
    
    // Create a helper with automatic record/replay mode
    recordingsDir := filepath.Join("testdata", "example_recordings")
    httpHelper := httprr.NewAutoHelper(t, recordingsDir)
    
    service := myservice.New(httpHelper.Client)
    
    // Run test
    
    // Optional: Add assertions about HTTP calls
    httpHelper.AssertURLCalled("api.example.com")
}
```

### Pattern 2: Replace custom HTTP clients

**Before:**
```go
func TestExample(t *testing.T) {
    // Test setup
    
    client := &http.Client{
        Timeout: 10 * time.Second,
    }
    service := myservice.New(client)
    
    // Run test
}
```

**After:**
```go
import (
    "path/filepath"
    "github.com/tmc/langchaingo/internal/httprr"
)

func TestExample(t *testing.T) {
    // Test setup
    
    // Create a custom transport with your desired settings
    transport := &http.Transport{
        // Set custom transport settings...
    }
    
    // Create a recorder with the custom transport
    recorder := httprr.NewRecorder(transport)
    recorder.Dir = filepath.Join("testdata", "example_recordings") 
    recorder.Mode = httprr.ModeRecord // Will auto-switch to replay when HTTPRR_MODE=replay
    
    client := &http.Client{
        Transport: recorder,
        Timeout: 10 * time.Second,
    }
    
    service := myservice.New(client)
    
    // Run test
    
    // Optional: Use recorder directly for assertions
    records := recorder.Records()
    if len(records) != 1 {
        t.Errorf("Expected 1 HTTP request, got %d", len(records))
    }
}
```

### Pattern 3: Global Default Client Replacement

If you have tests that indirectly use `http.DefaultClient` and you can't inject a client, you can temporarily replace the default client:

```go
import (
    "path/filepath"
    "github.com/tmc/langchaingo/internal/httprr"
)

func TestExample(t *testing.T) {
    // Replace the default client
    recordingsDir := filepath.Join("testdata", "example_recordings")
    
    // Ensure the directory exists for recording mode
    if os.Getenv("HTTPRR_MODE") != "replay" {
        os.MkdirAll(recordingsDir, 0755)
    }
    
    // Create a recorder that will respond to environment settings
    recorder := httprr.NewRecorder(http.DefaultTransport)
    recorder.Dir = recordingsDir
    if os.Getenv("HTTPRR_MODE") == "replay" {
        recorder.Mode = httprr.ModeReplay
        recorder.loadRecordings()
    }
    
    originalClient := http.DefaultClient
    http.DefaultClient = &http.Client{Transport: recorder}
    defer func() { http.DefaultClient = originalClient }()
    
    // Run your test which will now use the recording client
    
    // Make assertions about HTTP requests
    if len(recorder.Records()) == 0 {
        t.Error("Expected HTTP requests, but none were recorded")
    }
}
```

## Example: Complete Test Update for Offline Support

**Before:**
```go
func TestAPIChain(t *testing.T) {
    // Set up LLM
    llm, err := openai.New()
    require.NoError(t, err)
    
    // Create chain with default client
    chain := NewAPIChain(llm, http.DefaultClient)
    
    // Run the chain
    result, err := Call(context.Background(), chain, queryInput)
    require.NoError(t, err)
    
    // Check result
    require.Contains(t, result["answer"], "expected text")
}
```

**After:**
```go
func TestAPIChain(t *testing.T) {
    // Set up LLM
    llm, err := openai.New()
    require.NoError(t, err)
    
    // Create HTTP recorder with auto record/replay mode
    recordingsDir := filepath.Join("testdata", "api_chain_recordings")
    httpHelper := httprr.NewAutoHelper(t, recordingsDir)
    
    // Create chain with recording client
    chain := NewAPIChain(llm, httpHelper.Client)
    
    // Run the chain
    result, err := Call(context.Background(), chain, queryInput)
    require.NoError(t, err)
    
    // Check result
    require.Contains(t, result["answer"], "expected text")
    
    // Add assertions about HTTP traffic
    httpHelper.AssertURLCalled("api.example.com")
    
    // Dump HTTP interactions for debugging when verbose
    if testing.Verbose() {
        httpHelper.DumpRecordings()
    }
}
```

## Running Tests Offline

After updating tests to use httprr, you can run them without network access:

1. First run tests in record mode to capture responses:
```bash
# Record mode (default)
go test ./...
```

2. Then run tests in replay mode, which won't make real HTTP requests:
```bash
# Replay mode
HTTPRR_MODE=replay go test ./...
```

This is particularly useful for:
- CI/CD pipelines where tests should be deterministic
- Development environments with limited internet access
- Tests that would otherwise be flaky due to network issues
- Speeding up test execution by eliminating network latency

## Further Steps for Large-Scale Adoption

1. Start by updating critical tests using one of the patterns above
2. Create a helper function in test files that wrap `NewAutoHelper` if you have many tests needing the same setup
3. Consider adding a test suite setup that globally replaces the client in an `init()` function or test suite setup if needed

## Tips for Effective Usage

1. Always use `filepath.Join()` for recording directories to ensure cross-platform compatibility
2. Create a convention for recording directories, such as `testdata/{test_name}_recordings`
3. Add `.gitignore` entries for recording directories if you don't want to commit them
4. For fixed APIs, commit the recordings to git to ensure tests can run offline
5. For evolving APIs, use CI/CD to periodically refresh recordings
6. Add URL assertions to verify expected endpoints are called
7. For API-heavy tests, save recordings to a permanent location for documentation
8. When debugging HTTP issues, use `DumpRecordings()` to see the full request/response details
9. Use verbose mode to see recordings only when needed: `go test -v` 
# httprr: HTTP Record and Replay for Testing

The `httprr` package provides deterministic HTTP record and replay functionality for testing. It allows tests to record real HTTP interactions during development and replay them during CI/testing, ensuring consistent and fast test execution.

## Quick Start

```go
func TestMyAPI(t *testing.T) {
    // Skip test gracefully if no credentials and no recording exists
    httprr.SkipIfNoCredentialsOrRecording(t, "API_KEY")
    
    // Create recorder/replayer
    rr, err := httprr.OpenForTest(t, http.DefaultTransport)
    if err != nil {
        t.Fatal(err)
    }
    defer rr.Close()
    
    // Use rr.Client() for all HTTP calls
    client := rr.Client()
    resp, err := client.Get("https://api.example.com/data")
    // ... test continues
}
```

## Core Concepts

### Recording vs Replay Modes

- **Recording Mode** (`-httprecord=.`): Makes real HTTP requests and saves them to `.httprr` files
- **Replay Mode** (default): Reads saved `.httprr` files and replays the responses

### Command-Line Flags

- `-httprecord=<regexp>`: Re-record traces for files matching the regexp pattern (use "." to match all)
- `-httprecord-delay=<ms>`: Add delay in milliseconds between HTTP requests during recording (helps avoid rate limits)

### File Management

- **Recording**: Always creates uncompressed `.httprr` files for easier debugging
- **Replay**: Automatically handles both `.httprr` and `.httprr.gz` files
- **Conflict Resolution**: Chooses the newer file if both compressed and uncompressed exist
- **Auto-cleanup**: Recording mode removes conflicting files automatically

## API Reference

### Core Functions

#### `OpenForTest(t *testing.T, rt http.RoundTripper) (*RecordReplay, error)`

The primary API for most test cases. Creates a recorder/replayer for the given test.

- **Recording mode**: Creates `testdata/TestName.httprr` 
- **Replay mode**: Loads existing recording
- **File naming**: Derived automatically from `t.Name()`
- **Directory**: Always uses `testdata/` subdirectory

#### `SkipIfNoCredentialsOrRecording(t *testing.T, envVars ...string)`

Gracefully skips tests when they cannot run (no API keys) and have no recorded data.

```go
// Skip if OPENAI_API_KEY not set AND no recording exists
httprr.SkipIfNoCredentialsOrRecording(t, "OPENAI_API_KEY")

// Skip if neither API_KEY nor BACKUP_KEY is set AND no recording exists  
httprr.SkipIfNoCredentialsOrRecording(t, "API_KEY", "BACKUP_KEY")
```

#### `Open(file string, rt http.RoundTripper) (*RecordReplay, error)`

Low-level API for custom file management. Most tests should use `OpenForTest` instead.

### RecordReplay Methods

#### `Client() *http.Client`
Returns an HTTP client that routes through the recorder/replayer.

#### `ScrubReq(scrubs ...func(*http.Request) error)`
Adds request scrubbing functions to sanitize sensitive data before recording.

```go
rr.ScrubReq(func(req *http.Request) error {
    req.Header.Set("Authorization", "Bearer test-api-key")
    return nil
})
```

#### `ScrubResp(scrubs ...func(*bytes.Buffer) error)`
Adds response scrubbing functions to sanitize sensitive data in responses.

#### `Recording() bool`
Reports whether the recorder is in recording mode.

#### `Close() error`
Closes the recorder/replayer. Use with `defer` for automatic cleanup.

## Usage Patterns

### Basic API Testing

```go
func TestOpenAIChat(t *testing.T) {
    httprr.SkipIfNoCredentialsOrRecording(t, "OPENAI_API_KEY")
    
    rr, err := httprr.OpenForTest(t, http.DefaultTransport)
    if err != nil {
        t.Fatal(err)
    }
    defer rr.Close()
    
    // Scrub sensitive data
    rr.ScrubReq(func(req *http.Request) error {
        req.Header.Set("Authorization", "Bearer test-api-key")
        return nil
    })
    
    // Create client with recording support
    llm, err := openai.New(openai.WithHTTPClient(rr.Client()))
    require.NoError(t, err)
    
    // Test continues with recorded/replayed HTTP calls
    response, err := llm.GenerateContent(ctx, messages)
    require.NoError(t, err)
}
```

### Helper Functions for Multiple Tests

```go
func createTestClient(t *testing.T) *MyAPIClient {
    t.Helper()
    httprr.SkipIfNoCredentialsOrRecording(t, "MY_API_KEY")
    
    rr, err := httprr.OpenForTest(t, http.DefaultTransport)
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { rr.Close() })
    
    return NewMyAPIClient(WithHTTPClient(rr.Client()))
}

func TestFeatureA(t *testing.T) {
    client := createTestClient(t)
    // ... test continues
}

func TestFeatureB(t *testing.T) {
    client := createTestClient(t)
    // ... test continues
}
```

### Multiple API Endpoints

```go
func TestMultiAPIIntegration(t *testing.T) {
    httprr.SkipIfNoCredentialsOrRecording(t, "OPENAI_API_KEY", "SERPAPI_KEY")
    
    rr, err := httprr.OpenForTest(t, http.DefaultTransport)
    if err != nil {
        t.Fatal(err)
    }
    defer rr.Close()
    
    // Both clients will use the same recording
    openaiClient := openai.New(openai.WithHTTPClient(rr.Client()))
    searchClient := serpapi.New(serpapi.WithHTTPClient(rr.Client()))
    
    // All HTTP calls are recorded/replayed together
}
```

## Command Line Usage

### Recording New Interactions

```bash
# Record all tests
go test ./... -httprecord=.

# Record specific test
go test ./pkg -httprecord=. -run TestSpecificFunction

# Record with pattern matching
go test ./... -httprecord="TestOpenAI.*"
```

### Running with Recorded Data

```bash
# Normal test run (uses recorded data)
go test ./...

# Skip tests that need credentials
OPENAI_API_KEY="" go test ./...  # Tests will skip gracefully
```

## File Management

### File Structure

```
testdata/
├── TestBasicFunction.httprr           # Uncompressed recording
├── TestWithSubtest-subcase.httprr     # Subtest recording  
├── TestOldFunction.httprr.gz          # Compressed recording
└── TestComplexAPI-setup.httprr        # Multi-part test
```

### File Naming Rules

- Test name: `TestMyFunction` → File: `TestMyFunction.httprr`
- With subtests: `TestMyFunction/subcase` → File: `TestMyFunction-subcase.httprr`
- Special chars: Replaced with hyphens for filesystem compatibility

### Compression Management

```bash
# Compress all recordings (for repository storage)
./internal/devtools/httprr-pack pack

# Check compression status
./internal/devtools/httprr-pack check

# Decompress for debugging
./internal/devtools/httprr-pack unpack
```

### Recording with Rate Limit Protection

When recording tests that make many API calls, use the delay flag to avoid hitting rate limits:

```bash
# Record with 1 second delay between requests
go test -httprecord=. -httprecord-delay=1000 ./...

# Record specific test with 500ms delay
go test -httprecord=. -httprecord-delay=500 -run TestMyAPI ./mypackage
```

## Best Practices

### 1. Always Use Graceful Skipping

```go
// ✅ Good: Test skips gracefully when it can't run
httprr.SkipIfNoCredentialsOrRecording(t, "API_KEY")

// ❌ Bad: Test fails when API key missing
rr, err := httprr.OpenForTest(t, http.DefaultTransport)
```

### 2. Scrub Sensitive Data

```go
// ✅ Good: Replace real API keys with test values
rr.ScrubReq(func(req *http.Request) error {
    req.Header.Set("Authorization", "Bearer test-api-key")
    return nil
})

// ❌ Bad: Real API keys recorded in files
// (No scrubbing - keys end up in repository)
```

### 3. Use Helper Functions

```go
// ✅ Good: Reusable test setup
func createTestLLM(t *testing.T) *openai.LLM {
    t.Helper()
    httprr.SkipIfNoCredentialsOrRecording(t, "OPENAI_API_KEY")
    // ... setup code
}

// ❌ Bad: Duplicate setup in every test
func TestA(t *testing.T) {
    httprr.SkipIfNoCredentialsOrRecording(t, "OPENAI_API_KEY")
    rr, err := httprr.OpenForTest(t, http.DefaultTransport)
    // ... repeated setup
}
```

### 4. Handle Cleanup Properly

```go
// ✅ Good: Automatic cleanup
defer rr.Close()

// or
t.Cleanup(func() { rr.Close() })

// ❌ Bad: Manual cleanup (can be forgotten)
// (No defer or cleanup)
```

## Troubleshooting

### Common Issues

#### "cached HTTP response not found"

**Problem**: Test is trying to make an HTTP request not in the recording.

**Solutions**:
```bash
# Re-record the test
go test ./pkg -httprecord=. -run TestName

# Check if you have required environment variables
export OPENAI_API_KEY="your-key-here"
go test ./pkg -httprecord=. -run TestName
```

#### "gzip: invalid header"

**Problem**: `.httprr.gz` file is corrupted or not actually compressed.

**Solutions**:
```bash
# Check and fix compression
./internal/devtools/httprr-pack check
./internal/devtools/httprr-pack pack

# Or remove the corrupted file and re-record
rm testdata/TestName.httprr.gz
go test ./pkg -httprecord=. -run TestName
```

#### Test skipped unexpectedly

**Problem**: Test is skipping when you expect it to run.

**Debug steps**:
```bash
# Check if environment variables are set
echo $OPENAI_API_KEY

# Check if recording exists
ls testdata/TestName.httprr*

# Run with verbose output
go test ./pkg -run TestName -v
```

### File Conflicts

The system automatically handles conflicts, but you can resolve manually:

```bash
# Check which file is newer
ls -la testdata/TestName.httprr*

# Remove older file (system will warn and use newer)
rm testdata/TestName.httprr.gz  # if .httprr is newer

# Or compress the newer one
gzip testdata/TestName.httprr
```

## Migration Guide

### From `OpenForTestWithSkip` (Old API)

```go
// ❌ Old API (removed)
rr := httprr.OpenForTestWithSkip(t, http.DefaultTransport, "API_KEY")
defer rr.Close()

// ✅ New API
httprr.SkipIfNoCredentialsOrRecording(t, "API_KEY")

rr, err := httprr.OpenForTest(t, http.DefaultTransport)
if err != nil {
    t.Fatal(err)
}
defer rr.Close()
```

### Benefits of New API

1. **Consistent Error Handling**: All `httprr` operations return errors
2. **Clear Separation**: Skip logic separate from file operations  
3. **Single Responsibility**: Each function has one clear purpose
4. **Better Documentation**: Self-documenting function names

## Advanced Usage

### Custom File Locations

```go
// For custom file management (rarely needed)
rr, err := httprr.Open("custom/path/recording.httprr", http.DefaultTransport)
if err != nil {
    t.Fatal(err)
}
defer rr.Close()
```

### Conditional Recording

```go
func TestWithConditionalRecording(t *testing.T) {
    // Only record if we have credentials
    if os.Getenv("API_KEY") != "" {
        // Will record new interactions
        rr, err := httprr.OpenForTest(t, http.DefaultTransport)
        // ...
    } else {
        // Will only replay existing recordings
        httprr.SkipIfNoCredentialsOrRecording(t, "API_KEY")
        rr, err := httprr.OpenForTest(t, http.DefaultTransport)
        // ...
    }
}
```

### Complex Scrubbing

```go
rr.ScrubReq(func(req *http.Request) error {
    // Remove API keys
    req.Header.Set("Authorization", "Bearer test-key")
    
    // Scrub request body
    if req.Body != nil {
        body := req.Body.(*httprr.Body)
        bodyStr := string(body.Data)
        bodyStr = strings.ReplaceAll(bodyStr, "real-secret", "test-secret")
        body.Data = []byte(bodyStr)
    }
    
    return nil
})

rr.ScrubResp(func(buf *bytes.Buffer) error {
    // Remove sensitive data from responses
    content := buf.String()
    content = strings.ReplaceAll(content, "sensitive-data", "redacted")
    buf.Reset()
    buf.WriteString(content)
    return nil
})
```

## Contributing

When adding new tests that use external APIs:

1. **Always use `SkipIfNoCredentialsOrRecording`** for graceful degradation
2. **Include appropriate scrubbing** to avoid committing secrets
3. **Record with real credentials** initially, then scrub the results
4. **Compress recordings** before committing to save repository space
5. **Document required environment variables** in test comments

For questions or issues with the httprr system, see the main project documentation or open an issue.
# HTTPRR - HTTP Record & Replay for LangChainGo

`httprr` is a testing utility that records and replays HTTP requests and responses, making it easy to test LLM interactions without making real API calls in your test suite.

## Overview

When testing applications that interact with LLM APIs, you often face these challenges:

1. **API Costs**: Real API calls cost money, especially during development and CI/CD
2. **Rate Limits**: APIs have rate limits that can slow down or break tests
3. **Reliability**: Network issues and API outages can cause flaky tests
4. **Reproducibility**: API responses may vary, making tests non-deterministic
5. **Offline Testing**: Need internet connectivity for tests to run

`httprr` solves these problems by:

- **Recording** real API interactions once
- **Replaying** them in subsequent test runs
- **Providing** test helpers specifically designed for LLM clients
- **Supporting** automatic mode detection (record vs replay)

## Quick Start

### Basic Usage

```go
package mypackage

import (
    "context"
    "testing"
    
    "github.com/tmc/langchaingo/httprr"
    "github.com/tmc/langchaingo/llms"
)

func TestMyLLMCode(t *testing.T) {
    // Create test helper with recordings directory
    helper := httprr.NewLLMTestHelper(t, "testdata/llm_recordings")
    defer helper.Cleanup()
    
    // Create OpenAI client with httprr
    client, err := helper.NewOpenAIClientWithToken("your-api-key")
    if err != nil {
        t.Fatal(err)
    }
    
    // Make API call - recorded on first run, replayed on subsequent runs
    response, err := client.GenerateContent(context.Background(), []llms.MessageContent{
        {
            Role: llms.ChatMessageTypeHuman,
            Parts: []llms.ContentPart{
                llms.TextContent{Text: "Hello, world!"},
            },
        },
    })
    
    // Test your response
    if err != nil {
        t.Fatal(err)
    }
    
    if len(response.Choices) == 0 {
        t.Fatal("Expected at least one choice")
    }
    
    // Verify HTTP interactions
    helper.AssertRequestCount(1)
    helper.AssertURLCalled("api.openai.com")
}
```

### Supported LLM Providers

#### OpenAI

```go
// Basic OpenAI client
client, err := helper.NewOpenAIClientWithToken("your-api-key")

// With additional options
client, err := helper.NewOpenAIClientWithToken("your-api-key",
    openai.WithModel("gpt-4"),
    openai.WithBaseURL("https://custom-endpoint.com"),
)

// Or use as an option with existing client creation
client, err := openai.New(
    openai.WithToken("your-api-key"),
    helper.WithOpenAIHTTPClient(), // Add httprr transport
)
```

#### Anthropic

```go
// Basic Anthropic client
client, err := helper.NewAnthropicClient("your-api-key")

// With options
client, err := helper.NewAnthropicClient("your-api-key",
    anthropic.WithModel("claude-3-sonnet-20240229"),
)
```

#### Ollama

```go
// Basic Ollama client (assumes local Ollama server)
client, err := helper.NewOllamaClient()

// With custom options
client, err := helper.NewOllamaClient(
    ollama.WithServerURL("http://localhost:11434"),
)
```

## Advanced Features

### Mode Control

Control recording/replay behavior with environment variables:

```bash
# Force recording mode (make real API calls)
export HTTPRR_MODE=record

# Force replay mode (use recorded responses)
export HTTPRR_MODE=replay

# Auto mode (default): replay if recordings exist, record otherwise
unset HTTPRR_MODE
```

### Directory Organization

Organize recordings by provider and test:

```
testdata/
├── llm_recordings/
│   ├── openai/
│   │   ├── basic_completion/
│   │   └── streaming_test/
│   ├── anthropic/
│   │   └── claude_test/
│   └── ollama/
│       └── local_test/
```

### Assertions and Debugging

```go
// Assert specific number of requests
helper.AssertRequestCount(2)

// Assert specific URLs were called
helper.AssertURLCalled("api.openai.com")
helper.AssertURLCalled("completions")

// Get all requested URLs
urls := helper.GetRequestURLs()
t.Logf("URLs called: %v", urls)

// Find specific responses
resp, body, err := helper.FindResponse("completions")
if err == nil {
    t.Logf("Found response: %d %s", resp.StatusCode, resp.Status)
}

// Dump all recordings for debugging
helper.DumpRecordings()

// Save recordings to custom location
err := helper.SaveRecordingsToDir("debug_recordings")
```

## Best Practices

### 1. Organize Test Data

Create a clear directory structure:

```
your_project/
├── testdata/
│   └── httprr_recordings/
│       ├── integration_tests/
│       ├── unit_tests/
│       └── examples/
└── your_test.go
```

### 2. Use Descriptive Directory Names

```go
// Good - descriptive
helper := httprr.NewLLMTestHelper(t, "testdata/openai_chat_completion")

// Bad - generic  
helper := httprr.NewLLMTestHelper(t, "testdata/test1")
```

### 3. Test Both Success and Error Cases

```go
func TestAPIError(t *testing.T) {
    helper := httprr.NewLLMTestHelper(t, "testdata/api_errors")
    defer helper.Cleanup()
    
    // Use invalid API key to test error handling
    client, _ := helper.NewOpenAIClientWithToken("invalid-key")
    
    _, err := client.GenerateContent(ctx, messages)
    if err == nil {
        t.Fatal("Expected error with invalid API key")
    }
    
    // Verify the error request was recorded
    helper.AssertURLCalled("api.openai.com")
}
```

### 4. Clean Up Sensitive Data

When committing recordings, ensure no sensitive data is included:

```go
// In CI or when sharing recordings, use placeholder API keys
apiKey := "test-api-key"
if os.Getenv("CI") != "" {
    apiKey = "placeholder-key-for-ci"
}
```

### 5. Use Environment-Specific Behavior

```go
func TestLLM(t *testing.T) {
    recordingsDir := "testdata/recordings"
    
    // In CI, always use replay mode
    if os.Getenv("CI") != "" {
        os.Setenv("HTTPRR_MODE", "replay")
        defer os.Unsetenv("HTTPRR_MODE")
    }
    
    helper := httprr.NewLLMTestHelper(t, recordingsDir)
    defer helper.Cleanup()
    
    // ... rest of test
}
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Test
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Run tests with httprr replay
      run: go test ./...
      env:
        HTTPRR_MODE: replay
```

This ensures that CI/CD runs use recorded responses instead of making real API calls.

## Troubleshooting

### Common Issues

1. **"No recordings found"**: Run tests locally first in record mode to create recordings
2. **"Request doesn't match"**: Check if request parameters changed; re-record if needed  
3. **"Network errors in replay mode"**: Ensure `HTTPRR_MODE=replay` is set correctly

### Debug Mode

Enable verbose logging:

```go
helper := httprr.NewLLMTestHelper(t, recordingsDir)
helper.DumpRecordings() // Shows all recorded interactions
```

## Examples

See `example_test.go` for comprehensive examples of:

- Basic recording/replay
- Multiple LLM providers
- Custom client configurations
- Advanced assertion methods
- CI/CD integration patterns
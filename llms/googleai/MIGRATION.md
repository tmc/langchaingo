# Google AI API Migration Guide

This document provides guidance for migrating from the legacy `github.com/google/generative-ai-go` package to the new official `google.golang.org/genai` package.

## Background

Google has deprecated the `github.com/google/generative-ai-go` package in favor of the new official `google.golang.org/genai` package. The new package provides:

- Official support and maintenance by Google
- Better performance and reliability
- Enhanced features and capabilities
- Consistent API design with other Google Cloud SDKs

## Migration Steps

### 1. Update Dependencies

Replace the old dependency:
```bash
go mod tidy
go get google.golang.org/genai
```

### 2. Update Imports

**Before (Legacy):**
```go
import "github.com/google/generative-ai-go/genai"
```

**After (New):**
```go
import "google.golang.org/genai"
```

### 3. Client Initialization

**Before (Legacy):**
```go
ctx := context.Background()
client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
```

**After (New):**
```go
ctx := context.Background()
client, err := genai.NewClient(ctx, &genai.ClientConfig{
    APIKey: apiKey,
})
```

### 4. Using the New Implementation

The langchaingo library now provides `GoogleApisAI` alongside the existing `GoogleAI` implementation:

**Legacy Implementation:**
```go
import "github.com/tmc/langchaingo/llms/googleai"

llm, err := googleai.New(
    googleai.WithAPIKey("your-api-key"),
    googleai.WithDefaultModel("gemini-1.5-flash"),
)
```

**New Implementation:**
```go
import "github.com/tmc/langchaingo/llms/googleai"

ctx := context.Background()
llm, err := googleai.NewGoogleApisAI(ctx,
    googleai.WithAPIKey("your-api-key"),
    googleai.WithDefaultModel("gemini-1.5-flash"),
)
defer llm.Close() // Important: Close the client when done
```

## Key Differences

### Client Management

The new implementation requires explicit client closure:
```go
defer llm.Close()
```

### Context Requirements

The new SDK requires a context for client initialization:
```go
ctx := context.Background()
client, err := googleai.NewGoogleApisAI(ctx, options...)
```

### Enhanced Error Handling

The new implementation provides more detailed error information and better error categorization.

### Improved Performance

The new SDK offers:
- Better connection pooling
- More efficient request batching
- Reduced memory footprint

## Feature Parity

| Feature | Legacy SDK | New SDK | Status |
|---------|------------|---------|--------|
| Text Generation | ✅ | ✅ | Complete |
| Image Input | ✅ | ✅ | Complete |
| Embeddings | ✅ | ✅ | Complete |
| Safety Settings | ✅ | ✅ | Complete |
| JSON Mode | ✅ | 🚧 | In Progress |
| Function Calls | ✅ | 🚧 | In Progress |
| Streaming | ✅ | 🚧 | Planned |

## Testing

Both implementations are thoroughly tested. To run tests for the new implementation:

```bash
export GEMINI_API_KEY="your-api-key"
go test ./llms/googleai -run TestGoogleApisAI
```

For record/replay testing:
```bash
export TEST_RECORD=true
export GEMINI_API_KEY="your-api-key"
go test ./llms/googleai -run TestGoogleApisAI
```

## Backwards Compatibility

The legacy `GoogleAI` implementation remains available for backwards compatibility. However, we recommend migrating to `GoogleApisAI` for new projects.

Both implementations share the same interface (`llms.Model`), making migration straightforward:

```go
// This works with both implementations
var llm llms.Model
llm = legacyGoogleAI  // or
llm = newGoogleApisAI

response, err := llm.GenerateContent(ctx, messages)
```

## Configuration Options

The new implementation supports all the same configuration options:

```go
client, err := googleai.NewGoogleApisAI(ctx,
    googleai.WithAPIKey("your-key"),
    googleai.WithDefaultModel("gemini-1.5-pro"),
    googleai.WithTemperature(0.7),
    googleai.WithTopP(0.9),
    googleai.WithTopK(40),
    googleai.WithMaxTokens(1000),
)
```

## Troubleshooting

### Common Issues

1. **"Client not closed" warnings**: Always call `defer client.Close()`
2. **Context canceled errors**: Ensure proper context lifecycle management
3. **API key issues**: Verify your API key is valid and has necessary permissions

### Getting Help

- Check the [Google AI Documentation](https://ai.google.dev/docs)
- Review the [googleapis/go-genai repository](https://github.com/googleapis/go-genai)
- Open an issue in the langchaingo repository for library-specific problems

## Roadmap

- **Phase 1** (Current): New implementation available alongside legacy
- **Phase 2** (Q2 2024): New implementation becomes default
- **Phase 3** (Q4 2024): Legacy implementation deprecated
- **Phase 4** (Q2 2025): Legacy implementation removed

We recommend starting migration planning now to ensure a smooth transition.
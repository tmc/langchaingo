# DeepSeek LLM Integration

This package provides dedicated support for DeepSeek models in langchaingo. While DeepSeek API is OpenAI-compatible, this package offers a more convenient interface with DeepSeek-specific features and best practices.

## Quick Start

### Using the Dedicated DeepSeek Package (Recommended)

```go
import (
    "context"
    "github.com/tmc/langchaingo/llms/deepseek"
)

// Initialize with API key from environment (DEEPSEEK_API_KEY or OPENAI_API_KEY)
llm, err := deepseek.New()

// Or specify options
llm, err := deepseek.New(
    deepseek.WithModel(deepseek.ModelReasoner),
    deepseek.WithToken("your-api-key"),
)

// Simple chat
response, err := llm.Chat(ctx, "What is the capital of France?")

// Chat with reasoning (for reasoning models)
reasoning, answer, err := llm.ChatWithReasoning(ctx, "Solve: 2+2*3")
```

### Using OpenAI Client Directly (Also Valid)

```go
import "github.com/tmc/langchaingo/llms/openai"

llm, err := openai.New(
    openai.WithModel("deepseek-reasoner"),
    openai.WithBaseURL("https://api.deepseek.com/v1"),
)
```

## Available Models

| Model | Constant | Description |
|-------|----------|-------------|
| `deepseek-reasoner` | `deepseek.ModelReasoner` | Advanced reasoning with step-by-step thinking |
| `deepseek-chat` | `deepseek.ModelChat` | General purpose chat model |
| `deepseek-coder` | `deepseek.ModelCoder` | Specialized for code generation |

## Features

- 🧠 **Dedicated DeepSeek Models**: Type-safe model constants
- 🔍 **Reasoning Access**: Easy access to step-by-step reasoning content  
- 🎯 **Convenience Methods**: Simplified APIs (`Chat`, `ChatWithReasoning`)
- ⚙️ **Full Compatibility**: All OpenAI options available through pass-through
- 🔧 **Better Developer Experience**: Purpose-built for DeepSeek

## Configuration Options

### DeepSeek-Specific Options

```go
llm, err := deepseek.New(
    deepseek.WithModel(deepseek.ModelCoder),          // Choose model
    deepseek.WithToken("sk-..."),                     // API key
    deepseek.WithBaseURL("https://custom-url.com"),   // Custom endpoint
)
```

### OpenAI Pass-Through Options

All OpenAI client options are supported:

```go
llm, err := deepseek.New(
    deepseek.WithOpenAIOption(openai.WithMaxTokens(1000)),
    deepseek.WithOpenAIOption(openai.WithTemperature(0.7)),
)
```

## Advanced Usage

### Reasoning Models

DeepSeek's reasoning models provide step-by-step thinking process:

```go
// Method 1: Using convenience method
reasoning, answer, err := llm.ChatWithReasoning(
    ctx, 
    "Explain why the sky is blue",
    llms.WithMaxTokens(1000),
)
fmt.Printf("Reasoning: %s\n", reasoning)
fmt.Printf("Answer: %s\n", answer)

// Method 2: Using GenerateContent directly
messages := []llms.MessageContent{
    llms.TextParts(llms.ChatMessageTypeHuman, "Explain quantum computing"),
}

resp, err := llm.GenerateContent(ctx, messages)
if len(resp.Choices) > 0 {
    choice := resp.Choices[0]
    fmt.Printf("Reasoning: %s\n", choice.ReasoningContent)
    fmt.Printf("Answer: %s\n", choice.Content)
}
```

### Streaming with Reasoning

```go
resp, err := llm.GenerateContent(ctx, messages,
    llms.WithStreamingReasoningFunc(func(ctx context.Context, reasoning, content []byte) error {
        if len(reasoning) > 0 {
            fmt.Printf("🧠 Thinking: %s", reasoning)
        }
        if len(content) > 0 {
            fmt.Printf("💬 Response: %s", content)
        }
        return nil
    }),
)
```

### Function Calling

```go
tools := []llms.Tool{
    {
        Type: "function",
        Function: &llms.FunctionDefinition{
            Name: "get_weather",
            Description: "Get current weather",
            Parameters: map[string]any{
                "type": "object",
                "properties": map[string]any{
                    "location": map[string]any{"type": "string"},
                },
            },
        },
    },
}

resp, err := llm.GenerateContent(ctx, messages, llms.WithTools(tools))
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DEEPSEEK_API_KEY` | DeepSeek API token | Required |
| `OPENAI_API_KEY` | Also works (for compatibility) | Alternative to DEEPSEEK_API_KEY |
| `DEEPSEEK_BASE_URL` | API base URL | `https://api.deepseek.com/v1` |

## Why Use the DeepSeek Package?

### Benefits Over Direct OpenAI Client

| Feature | DeepSeek Package | OpenAI Client Direct |
|---------|------------------|---------------------|
| Model Constants | ✅ Type-safe constants | ❌ String literals |
| Convenience Methods | ✅ `Chat`, `ChatWithReasoning` | ❌ Manual message building |
| Reasoning Access | ✅ Built-in support | ✅ Manual extraction |
| Documentation | ✅ DeepSeek-specific | ❌ Generic OpenAI docs |
| Future Features | ✅ DeepSeek-specific additions | ❌ May not fit OpenAI API |
| Configuration | ✅ Sensible defaults | ❌ Manual setup required |

### When to Use Each Approach

**Use DeepSeek Package When:**
- Building new applications with DeepSeek
- Want the best developer experience
- Using reasoning models extensively
- Need type safety for models

**Use OpenAI Client When:**
- Migrating existing OpenAI code
- Need to switch between OpenAI/DeepSeek dynamically
- Using advanced OpenAI-specific features
- Prefer minimal dependencies

## Migration Guide

### From OpenAI Client to DeepSeek Package

```go
// Before (OpenAI client)
llm, err := openai.New(
    openai.WithModel("deepseek-reasoner"),
    openai.WithBaseURL("https://api.deepseek.com/v1"),
    openai.WithToken(os.Getenv("DEEPSEEK_API_KEY")),
)

// After (DeepSeek package)
llm, err := deepseek.New(
    deepseek.WithModel(deepseek.ModelReasoner),
)
```

### Accessing Reasoning Content

```go
// Before (manual extraction)
resp, err := llm.GenerateContent(ctx, messages)
reasoning := resp.Choices[0].ReasoningContent
content := resp.Choices[0].Content

// After (convenience method)
reasoning, content, err := llm.ChatWithReasoning(ctx, "question")
```

## Examples

See the [complete example](../../examples/deepseek-completion-example/) for detailed usage patterns including both approaches.

## FAQ

**Q: Should I use the DeepSeek package or OpenAI client?**
A: For new projects, use the DeepSeek package for better developer experience. For existing OpenAI code, either approach works.

**Q: Do all OpenAI features work with DeepSeek?**  
A: Most do, but some advanced features may differ. Check DeepSeek's API documentation for specifics.

**Q: How do I access reasoning content?**
A: Use `GenerateWithReasoning()` or `ChatWithReasoning()` methods, or access `choice.ReasoningContent` from responses.

**Q: Can I switch between OpenAI and DeepSeek easily?**
A: Yes, but it's easier with the OpenAI client approach. The DeepSeek package is optimized for DeepSeek-specific usage.

## Related

- [DeepSeek API Documentation](https://platform.deepseek.com/api-docs/)
- [Complete Example](../../examples/deepseek-completion-example/)
- [OpenAI Client Documentation](../openai/README.md)
- GitHub Discussion: [#1212 "Use Deepseek with langchaingo"](https://github.com/tmc/langchaingo/discussions/1212)
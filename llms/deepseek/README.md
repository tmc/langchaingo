# DeepSeek Integration Guide

DeepSeek models are fully supported in LangChain Go through the OpenAI-compatible API. This guide explains how to use DeepSeek models effectively.

## Why Use OpenAI Client?

DeepSeek provides an OpenAI-compatible API, which means:
- **No separate client needed**: Use the existing `openai` package
- **Full feature compatibility**: All OpenAI client features work with DeepSeek
- **Easier maintenance**: Single codebase for multiple providers
- **Seamless switching**: Easy to switch between OpenAI and DeepSeek models

## Supported Models

- `deepseek-reasoner`: Advanced reasoning model with step-by-step thinking
- `deepseek-chat`: General chat model  
- `deepseek-coder`: Code-specialized model

## Basic Usage

```go
package main

import (
    "context"
    "log"
    
    "github.com/tmc/langchaingo/llms"
    "github.com/tmc/langchaingo/llms/openai"
)

func main() {
    // Initialize with DeepSeek model
    llm, err := openai.New(
        openai.WithModel("deepseek-reasoner"),
        openai.WithBaseURL("https://api.deepseek.com/v1"), // Optional: explicit base URL
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Use as normal
    response, err := llm.GenerateContent(context.Background(), []llms.MessageContent{
        llms.TextParts(llms.ChatMessageTypeHuman, "Hello!"),
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(response.Choices[0].Content)
}
```

## Reasoning Models

DeepSeek's reasoning models provide step-by-step thinking process:

```go
// Enable reasoning content streaming
completion, err := llm.GenerateContent(
    ctx,
    content,
    llms.WithStreamingReasoningFunc(func(_ context.Context, reasoningChunk []byte, chunk []byte) error {
        if len(reasoningChunk) > 0 {
            fmt.Printf("Reasoning: %s", string(reasoningChunk))
        }
        if len(chunk) > 0 {
            fmt.Printf("Answer: %s", string(chunk))
        }
        return nil
    }),
)

// Access reasoning content after completion
if len(completion.Choices) > 0 {
    choice := completion.Choices[0]
    fmt.Printf("Reasoning: %s\n", choice.ReasoningContent)
    fmt.Printf("Answer: %s\n", choice.Content)
}
```

## Configuration

### Environment Variables
```bash
export OPENAI_API_KEY="your-deepseek-api-key"
export OPENAI_BASE_URL="https://api.deepseek.com/v1"  # Optional
```

### Programmatic Configuration
```go
llm, err := openai.New(
    openai.WithToken("your-deepseek-api-key"),
    openai.WithBaseURL("https://api.deepseek.com/v1"),
    openai.WithModel("deepseek-reasoner"),
)
```

## Advanced Features

### Function Calling
```go
// DeepSeek supports function calling
tools := []llms.Tool{
    {
        Type: "function",
        Function: llms.FunctionDefinition{
            Name:        "get_weather",
            Description: "Get current weather",
            Parameters: map[string]any{
                "type": "object",
                "properties": map[string]any{
                    "location": map[string]any{
                        "type": "string",
                        "description": "City name",
                    },
                },
                "required": []string{"location"},
            },
        },
    },
}

response, err := llm.GenerateContent(ctx, messages, llms.WithTools(tools))
```

### Streaming with Reasoning
```go
response, err := llm.GenerateContent(
    ctx,
    messages,
    llms.WithStreamingReasoningFunc(func(ctx context.Context, reasoning, content []byte) error {
        // Handle reasoning chunks (DeepSeek's thinking process)
        if len(reasoning) > 0 {
            fmt.Printf("🧠 Thinking: %s", reasoning)
        }
        
        // Handle final answer chunks
        if len(content) > 0 {
            fmt.Printf("💬 Answer: %s", content)
        }
        
        return nil
    }),
)
```

## Best Practices

1. **Model Selection**: Choose the right model for your use case
   - `deepseek-reasoner`: Complex reasoning tasks
   - `deepseek-chat`: General conversation
   - `deepseek-coder`: Code generation and analysis

2. **Reasoning Content**: For reasoning models, always handle both reasoning and content
   ```go
   if choice.ReasoningContent != "" {
       log.Printf("Model reasoning: %s", choice.ReasoningContent)
   }
   ```

3. **Error Handling**: DeepSeek uses OpenAI-compatible error responses
   ```go
   if err != nil {
       // Handle same as OpenAI errors
       log.Printf("DeepSeek API error: %v", err)
   }
   ```

4. **Rate Limiting**: Follow DeepSeek's rate limiting guidelines
   ```go
   // Add retry logic for rate limits
   for retries := 0; retries < 3; retries++ {
       response, err := llm.GenerateContent(ctx, messages)
       if err == nil {
           break
       }
       time.Sleep(time.Second * time.Duration(retries+1))
   }
   ```

## Comparison with Dedicated Client

### Why not a separate DeepSeek package?

| Aspect | Current Approach (OpenAI client) | Separate Package |
|--------|----------------------------------|------------------|
| Maintenance | ✅ Single codebase | ❌ Duplicate code |
| Features | ✅ Full compatibility | ❌ May lag behind |
| Switching | ✅ Easy model swap | ❌ Code changes needed |
| Testing | ✅ Shared test suite | ❌ Separate tests |
| API Changes | ✅ Automatic support | ❌ Manual updates |

### When to consider a dedicated package?

A separate DeepSeek package would only be beneficial if:
- DeepSeek adds non-OpenAI-compatible features
- Significant performance optimizations are possible
- Custom authentication methods are required

## Examples

- [Basic DeepSeek Usage](../../examples/deepseek-completion-example/)
- [OpenAI Examples](../../examples/openai-completion-example/) (work with DeepSeek)

## Related

- [DeepSeek API Documentation](https://platform.deepseek.com/api-docs/)
- [OpenAI Client Documentation](../openai/README.md)
- GitHub Discussion: #1212 "Use Deepseek with langchaingo"
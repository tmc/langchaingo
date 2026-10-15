# Context Length Management in LangChain Go

This document addresses how LangChain Go handles prompts that exceed the defined context length, particularly in response to [Discussion #317](https://github.com/tmc/langchaingo/discussions/317).

## Overview

LangChain Go provides several mechanisms to handle context length limitations:

1. **Enhanced Token Buffer Memory** - Automatically manages conversation history within token limits
2. **Token Counting Utilities** - Accurately count tokens for different models
3. **Trimming Strategies** - Multiple approaches to reduce content when limits are exceeded
4. **Integration with Various Models** - Works with OpenAI, Google AI, and other providers

## Enhanced Token Buffer Memory

The `EnhancedTokenBuffer` in the `memory` package is the primary solution for managing context length:

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/tmc/langchaingo/llms/openai"
    "github.com/tmc/langchaingo/memory"
    "github.com/tmc/langchaingo/schema"
)

func main() {
    // Create an enhanced token buffer with automatic trimming
    memoryBuffer := memory.NewEnhancedTokenBuffer(
        memory.WithTokenLimit(2000),                    // Set token limit
        memory.WithEncodingModel("gpt-3.5-turbo"),     // Model for token counting
        memory.WithTrimStrategy(memory.TrimOldest),     // How to trim when limit exceeded
        memory.WithPreservePairs(true),                 // Keep human-AI pairs together
        memory.WithMinMessages(2),                      // Always preserve at least 2 messages
    )

    // Use with an LLM
    llm, err := openai.New()
    if err != nil {
        log.Fatal(err)
    }

    // The memory will automatically manage context length
    ctx := context.Background()
    
    // Add messages - they'll be automatically trimmed if limit exceeded
    memoryBuffer.SaveContext(ctx, map[string]any{
        "input":  "What is machine learning?",
        "output": "Machine learning is a subset of artificial intelligence...",
    })
    
    // Get memory variables - will be within token limit
    vars, err := memoryBuffer.LoadMemoryVariables(ctx, map[string]any{})
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Memory content: %v\n", vars)
}
```

### Advanced Configuration

```go
// Custom token counter
tokenCounter := &memory.TikTokenCounter{
    ModelName: "gpt-4",
}

// Create enhanced buffer with custom settings
memoryBuffer := memory.NewEnhancedTokenBuffer(
    memory.WithTokenLimit(4000),
    memory.WithTokenCounter(tokenCounter),
    memory.WithTrimStrategy(memory.TrimMiddle),  // Preserve first and last messages
    memory.WithPreservePairs(true),
    memory.WithMinMessages(4),
)
```

## Token Counting

### Built-in Token Counting

LangChain Go includes accurate token counting for various models:

```go
// Using tiktoken for OpenAI models
counter := &memory.TikTokenCounter{
    ModelName: "gpt-3.5-turbo",
}

tokenCount, err := counter.CountTokens("Your text here")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Token count: %d\n", tokenCount)

// Count tokens from chat messages
messages := []llms.ChatMessage{
    {Role: llms.ChatMessageTypeHuman, Content: "Hello"},
    {Role: llms.ChatMessageTypeAI, Content: "Hi there!"},
}

totalTokens, err := counter.CountTokensFromMessages(messages)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Total tokens: %d\n", totalTokens)
```

### Custom Token Counting

You can implement your own token counting logic:

```go
type CustomTokenCounter struct {
    // Your custom implementation
}

func (c *CustomTokenCounter) CountTokens(text string) (int, error) {
    // Implement your token counting logic
    return len(strings.Fields(text)) * 1.3, nil // Simple approximation
}

func (c *CustomTokenCounter) CountTokensFromMessages(messages []llms.ChatMessage) (int, error) {
    total := 0
    for _, msg := range messages {
        count, err := c.CountTokens(msg.GetContent())
        if err != nil {
            return 0, err
        }
        total += count + 3 // Add overhead for formatting
    }
    return total, nil
}
```

## Trimming Strategies

### Available Strategies

1. **TrimOldest** (default): Removes the oldest messages first
2. **TrimMiddle**: Preserves recent and initial messages, removes from the middle
3. **TrimByImportance**: Attempts to preserve more important messages (experimental)

```go
// Configure different trimming strategies
buffer1 := memory.NewEnhancedTokenBuffer(
    memory.WithTrimStrategy(memory.TrimOldest),
    memory.WithTokenLimit(1000),
)

buffer2 := memory.NewEnhancedTokenBuffer(
    memory.WithTrimStrategy(memory.TrimMiddle), 
    memory.WithTokenLimit(1000),
)
```

### Message Pair Preservation

When `PreservePairs` is enabled, the buffer ensures that human-AI message pairs are kept together:

```go
buffer := memory.NewEnhancedTokenBuffer(
    memory.WithPreservePairs(true),  // Keep human-AI pairs together
    memory.WithMinMessages(2),       // Always preserve at least 2 messages
    memory.WithTokenLimit(1000),
)
```

## Integration with Chains

The enhanced memory can be used with any chain that supports memory:

```go
// Use with LLM Chain
chain := chains.NewLLMChain(llm, prompt)
chain.Memory = memoryBuffer

// Use with Conversation Chain
conversationChain := chains.NewConversation(llm)
conversationChain.Memory = memoryBuffer

// The chain will automatically use the memory's context length management
result, err := chain.Call(ctx, map[string]any{
    "input": "Continue our conversation about machine learning",
})
```

## Model-Specific Context Limits

Different models have different context limits. The enhanced buffer can be configured accordingly:

```go
var tokenLimit int
var modelName string

switch modelName {
case "gpt-3.5-turbo":
    tokenLimit = 4096
case "gpt-3.5-turbo-16k":
    tokenLimit = 16384
case "gpt-4":
    tokenLimit = 8192
case "gpt-4-32k":
    tokenLimit = 32768
case "claude-2":
    tokenLimit = 100000
default:
    tokenLimit = 4096 // Conservative default
}

buffer := memory.NewEnhancedTokenBuffer(
    memory.WithTokenLimit(tokenLimit),
    memory.WithEncodingModel(modelName),
)
```

## Best Practices

1. **Set Appropriate Limits**: Leave room for your prompt template and expected response
   ```go
   // If model limit is 4096, use 3000 for memory to leave room for prompt/response
   memory.WithTokenLimit(3000)
   ```

2. **Use Model-Specific Counting**: Different models encode tokens differently
   ```go
   memory.WithEncodingModel("gpt-3.5-turbo") // Use exact model name
   ```

3. **Preserve Important Context**: Use `WithMinMessages` to ensure critical context is retained
   ```go
   memory.WithMinMessages(4) // Always keep at least 4 messages
   ```

4. **Monitor Token Usage**: Check token counts in your application
   ```go
   tokenCount, _ := buffer.GetTokenCount()
   fmt.Printf("Current memory token count: %d\n", tokenCount)
   ```

## Error Handling

The enhanced buffer gracefully handles various error conditions:

- **Model Not Supported**: Falls back to approximation methods
- **Empty Context**: Returns empty variables without error
- **Encoding Errors**: Uses fallback token counting

```go
buffer := memory.NewEnhancedTokenBuffer(
    memory.WithTokenLimit(1000),
    memory.WithEncodingModel("unsupported-model"), // Will use approximation
)

// Will work even with unsupported models
vars, err := buffer.LoadMemoryVariables(ctx, map[string]any{})
if err != nil {
    log.Printf("Memory error (usually safe to continue): %v", err)
}
```

## Migration from Basic Memory

If you're currently using basic memory types, migration is straightforward:

```go
// Old approach
oldMemory := memory.NewBuffer()

// New approach with automatic context management
newMemory := memory.NewEnhancedTokenBuffer(
    memory.WithTokenLimit(2000),
    memory.WithEncodingModel("gpt-3.5-turbo"),
)

// Same interface, enhanced functionality
vars, err := newMemory.LoadMemoryVariables(ctx, map[string]any{})
```

## Troubleshooting

### Common Issues

1. **Still Getting Context Length Errors**: 
   - Reduce the token limit further
   - Check if your prompt template is very long
   - Ensure you're using the correct model name

2. **Important Messages Being Trimmed**:
   - Increase `MinMessages`
   - Use `TrimMiddle` strategy to preserve recent and initial messages
   - Enable `PreservePairs` to keep conversation flow intact

3. **Inaccurate Token Counts**:
   - Verify you're using the exact model name
   - Consider implementing a custom `TokenCounter` for your specific use case

## Contributing

This context length management system is actively developed. If you encounter issues or have suggestions for improvements, please:

1. Check existing [GitHub discussions](https://github.com/tmc/langchaingo/discussions)
2. Open a new discussion for questions
3. Submit a PR for enhancements

For more examples, see the `/examples` directory in the repository.
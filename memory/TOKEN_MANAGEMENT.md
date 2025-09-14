# Token-Based Memory Management

This document describes the advanced token-based memory management features in langchaingo, addressing the enhancement request from [Discussion #124](https://github.com/tmc/langchaingo/discussions/124).

## Overview

The `EnhancedTokenBuffer` provides sophisticated token-aware memory management that automatically trims conversation history when token limits are exceeded. This prevents context length errors and optimizes LLM performance.

## Key Features

### 1. Automatic Token Management
- **Token Counting**: Pluggable token counting with support for model-specific encoders
- **Smart Trimming**: Multiple strategies for removing messages when limits are exceeded
- **Preservation Logic**: Options to preserve message pairs and maintain conversation flow

### 2. Multiple Trimming Strategies

#### TrimOldest (Default)
Removes the oldest messages first, preserving recent conversation context.

```go
buffer := memory.NewEnhancedTokenBuffer(
    memory.WithTokenLimit(2800),
    memory.WithTrimStrategy(memory.TrimOldest),
)
```

#### TrimMiddle
Preserves both the beginning and end of conversations, removing from the middle.

```go
buffer := memory.NewEnhancedTokenBuffer(
    memory.WithTokenLimit(2800),
    memory.WithTrimStrategy(memory.TrimMiddle),
)
```

#### TrimByImportance (Experimental)
Attempts to preserve more important messages based on heuristics.

```go
buffer := memory.NewEnhancedTokenBuffer(
    memory.WithTokenLimit(2800),
    memory.WithTrimStrategy(memory.TrimByImportance),
)
```

### 3. Token Counting Interfaces

#### TikTokenCounter
Provides accurate token counting using tiktoken-style encoding:

```go
counter := &memory.TikTokenCounter{ModelName: "gpt-3.5-turbo"}
buffer := memory.NewEnhancedTokenBuffer(
    memory.WithTokenCounter(counter),
    memory.WithTokenLimit(2800),
)
```

#### LLMTokenCounter
Uses an LLM's built-in token counting capabilities:

```go
llm, _ := openai.New()
counter := &memory.LLMTokenCounter{LLM: llm, Model: "gpt-3.5-turbo"}
buffer := memory.NewEnhancedTokenBuffer(
    memory.WithTokenCounter(counter),
)
```

#### Custom Token Counters
Implement the `TokenCounter` interface for custom counting logic:

```go
type CustomCounter struct{}

func (c *CustomCounter) CountTokens(text string) (int, error) {
    // Your custom token counting logic
    return count, nil
}

func (c *CustomCounter) CountTokensFromMessages(messages []llms.ChatMessage) (int, error) {
    // Your custom message token counting logic
    return count, nil
}
```

## Usage Examples

### Basic Usage

```go
import (
    "context"
    "github.com/tmc/langchaingo/memory"
    "github.com/tmc/langchaingo/llms/openai"
)

// Create enhanced token buffer
buffer := memory.NewEnhancedTokenBuffer(
    memory.WithTokenLimit(2800),                    // ~70% of GPT-3.5-turbo context
    memory.WithEncodingModel("gpt-3.5-turbo"),     // Model for token counting
    memory.WithPreservePairs(true),                 // Keep human-AI pairs intact
    memory.WithMinMessages(2),                      // Always preserve at least 2 messages
)

// Use in conversation
ctx := context.Background()
chain := chains.NewConversation(llm, buffer)

// The buffer automatically manages token limits
response, err := chain.Run(ctx, "What is machine learning?")
```

### Advanced Configuration

```go
// Create buffer with custom settings
buffer := memory.NewEnhancedTokenBuffer(
    memory.WithTokenLimit(3800),                    // Higher limit for GPT-4
    memory.WithEncodingModel("gpt-4"),             // GPT-4 token counting
    memory.WithTrimStrategy(memory.TrimMiddle),     // Preserve start and end
    memory.WithPreservePairs(true),                 // Maintain conversation pairs
    memory.WithMinMessages(4),                      // Preserve at least 4 messages
    memory.WithHumanPrefix("User"),                 // Custom prefixes
    memory.WithAIPrefix("Assistant"),
)

// Monitor token usage
tokenCount, _ := buffer.GetTokenCount(ctx)
fmt.Printf("Current tokens: %d/%d\n", tokenCount, buffer.GetTokenLimit())
```

### Integration with Existing Code

The enhanced buffer is a drop-in replacement for existing memory implementations:

```go
// Before
buffer := memory.NewConversationBuffer()

// After  
buffer := memory.NewEnhancedTokenBuffer(
    memory.WithTokenLimit(2800),
)

// Same interface
vars := buffer.MemoryVariables(ctx)
memoryData, err := buffer.LoadMemoryVariables(ctx, nil)
err = buffer.SaveContext(ctx, inputs, outputs)
```

## Configuration Options

### Token Management
- `WithTokenLimit(int)`: Set maximum tokens to keep in memory
- `WithEncodingModel(string)`: Specify model for token counting
- `WithTokenCounter(TokenCounter)`: Use custom token counter
- `WithLLM(llms.Model)`: Use LLM for token counting (fallback)

### Trimming Behavior  
- `WithTrimStrategy(TrimStrategy)`: Choose trimming strategy
- `WithPreservePairs(bool)`: Preserve human-AI message pairs
- `WithMinMessages(int)`: Minimum messages to always keep

### Memory Configuration
- `WithChatHistory(schema.ChatMessageHistory)`: Custom chat history storage
- `WithReturnMessages(bool)`: Return messages as slice vs string
- `WithInputKey(string)`: Key for input values
- `WithOutputKey(string)`: Key for output values
- `WithMemoryKey(string)`: Key for memory in LoadMemoryVariables
- `WithHumanPrefix(string)`: Prefix for human messages
- `WithAIPrefix(string)`: Prefix for AI messages

## Token Counting Accuracy

The token counting accuracy depends on the implementation used:

| Counter Type | Accuracy | Performance | Use Case |
|-------------|----------|-------------|----------|
| TikTokenCounter | High | Fast | Production use with OpenAI models |
| LLMTokenCounter | Medium | Slow | When LLM supports token counting |
| Custom | Varies | Varies | Specialized requirements |

## Performance Considerations

### Memory Trimming Overhead
- Trimming occurs after each `SaveContext` call
- O(n) complexity where n is the number of messages
- Consider batching conversations before trimming for high-throughput applications

### Token Counting Performance
- Token counting is called on every trim operation
- Cache token counts when possible
- Use efficient token counters for production workloads

### Memory Storage
- The buffer stores full message history until trimming
- Consider persistent storage for long conversations
- Use appropriate chat history implementations for your scale

## Migration from Basic Token Buffer

If you're currently using the basic `ConversationTokenBuffer`:

```go
// Before
buffer := memory.NewConversationTokenBuffer(llm, 2800)

// After
buffer := memory.NewEnhancedTokenBuffer(
    memory.WithTokenLimit(2800),
    memory.WithLLM(llm),
    memory.WithEncodingModel("gpt-3.5-turbo"), // Add model specification
)
```

The enhanced buffer provides the same functionality with additional configuration options.

## Best Practices

### 1. Choose Appropriate Token Limits
- GPT-3.5-turbo: 2800-3200 tokens (leave room for response)
- GPT-4: 3800-6000 tokens (depending on variant)
- Custom models: Check model documentation

### 2. Consider Your Trimming Strategy
- **TrimOldest**: Best for most conversations
- **TrimMiddle**: Good for preserving context and recent exchanges
- **TrimByImportance**: Experimental, use with testing

### 3. Preserve Message Pairs
Enable `PreservePairs` to maintain conversation flow:

```go
memory.WithPreservePairs(true)
```

### 4. Set Minimum Messages
Prevent over-trimming by setting minimum message count:

```go
memory.WithMinMessages(2) // Always keep at least one exchange
```

### 5. Monitor Token Usage
Regularly check token counts in production:

```go
if count, err := buffer.GetTokenCount(ctx); err == nil {
    log.Printf("Memory using %d/%d tokens", count, buffer.GetTokenLimit())
}
```

## Troubleshooting

### Common Issues

1. **Token counts seem inaccurate**
   - Verify you're using the correct encoding model
   - Consider using TikTokenCounter for OpenAI models
   - Check that token counter matches your LLM

2. **Messages being trimmed too aggressively**
   - Increase token limit
   - Set higher minimum message count
   - Check if preserve pairs is enabled

3. **Performance issues with large conversations**
   - Consider batching saves before trimming
   - Use more efficient token counter
   - Implement conversation archiving

### Debugging

Enable verbose logging to understand trimming behavior:

```go
// Check token count before and after operations
beforeCount, _ := buffer.GetTokenCount(ctx)
err := buffer.SaveContext(ctx, inputs, outputs)
afterCount, _ := buffer.GetTokenCount(ctx)

if beforeCount != afterCount {
    log.Printf("Trimmed from %d to %d tokens", beforeCount, afterCount)
}
```

## Future Enhancements

Planned improvements include:

- **Integration with tiktoken-go**: More accurate token counting
- **Importance scoring**: Better heuristics for message importance
- **Streaming trimming**: Trim during conversation for real-time applications
- **Compression**: Summarize old messages instead of removing them
- **Persistent storage**: Better integration with database storage

## Contributing

This implementation addresses [Discussion #124](https://github.com/tmc/langchaingo/discussions/124). 

To contribute improvements:
1. Add tests for new features
2. Ensure backward compatibility
3. Update documentation
4. Consider performance implications
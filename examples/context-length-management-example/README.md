# Context Length Management Example

This example demonstrates how LangChain Go automatically handles context length limitations using the Enhanced Token Buffer Memory. It addresses [Discussion #317](https://github.com/tmc/langchaingo/discussions/317) about context length management.

## What This Example Demonstrates

1. **Automatic Token Management**: How memory buffers automatically trim content when token limits are exceeded
2. **Token Counting**: Accurate token counting for different models using tiktoken
3. **Trimming Strategies**: Different approaches to removing content when limits are reached
4. **Conversation Preservation**: Keeping important conversation context intact
5. **Integration with Chains**: Using enhanced memory with conversation chains

## Features Showcased

- **Enhanced Token Buffer**: Automatically manages conversation history within token limits
- **Multiple Trimming Strategies**: 
  - `TrimOldest`: Removes oldest messages first (default)
  - `TrimMiddle`: Preserves recent and initial messages
- **Message Pair Preservation**: Keeps human-AI message pairs together
- **Custom Token Counting**: Ability to plug in custom token counting logic
- **Model-Specific Encoding**: Uses model-specific tokenizers for accuracy

## Running the Example

```bash
export OPENAI_API_KEY="your-openai-api-key"
go run context_length_management_example.go
```

## Key Configuration Options

```go
memoryBuffer := memory.NewEnhancedTokenBuffer(
    memory.WithTokenLimit(2000),                    // Maximum tokens to keep
    memory.WithEncodingModel("gpt-3.5-turbo"),     // Model for token counting
    memory.WithTrimStrategy(memory.TrimOldest),     // How to trim messages
    memory.WithPreservePairs(true),                 // Keep human-AI pairs together
    memory.WithMinMessages(2),                      // Minimum messages to preserve
)
```

## Real-World Applications

This pattern is essential for:

- **Long-running conversations**: Chatbots and virtual assistants
- **Document Q&A systems**: When processing large documents
- **Multi-turn reasoning**: Complex problem-solving that requires context
- **RAG applications**: Retrieval-augmented generation with conversation history

## Related Documentation

- [Context Length Management Guide](../../docs/CONTEXT_LENGTH_MANAGEMENT.md)
- [Memory Package Documentation](../../memory/)
- [Enhanced Token Buffer API](../../memory/enhanced_token_buffer.go)

## Troubleshooting

If you encounter context length errors:

1. **Reduce token limit**: Set a more conservative limit
2. **Check prompt size**: Ensure your prompt template isn't too large
3. **Adjust trimming strategy**: Use `TrimMiddle` to preserve important context
4. **Increase minimum messages**: Use `WithMinMessages()` to preserve critical context

## Contributing

This example addresses community questions about context length management. If you have suggestions for improvements or additional use cases, please contribute to the [GitHub discussion](https://github.com/tmc/langchaingo/discussions/317).
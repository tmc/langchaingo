# Ollama Context Example

This example demonstrates how to use Ollama's context feature for maintaining conversational memory.

## About Ollama Context

Ollama's context feature allows you to maintain state between requests, providing a form of short-term conversational memory. When you make a request to Ollama's `/api/generate` endpoint, it returns a `context` field containing the model's internal state. You can pass this context to subsequent requests to maintain continuity.

## Key Benefits

1. **Conversational Continuity**: The model remembers previous interactions
2. **Efficiency**: Avoids re-processing the same context repeatedly  
3. **Better Responses**: Follow-up questions can reference previous context

## Usage Pattern

```go
// 1. Make initial request (without context)
llm1, err := ollama.New(ollama.WithModel("llama2"))
response1, err := llm1.Call(ctx, "Tell me a joke about programming")

// 2. Extract context from response (when implemented)
// context := response1.Context

// 3. Use context in follow-up request
llm2, err := ollama.New(
    ollama.WithModel("llama2"),
    ollama.WithContext(context), // Use context from previous response
)
response2, err := llm2.Call(ctx, "Tell me another one")
// The model now knows "another one" refers to another programming joke
```

## Important Notes

- Context is primarily supported by Ollama's `/api/generate` endpoint
- The current langchaingo implementation uses `/api/chat` which has different context handling
- Context is model-specific and should only be reused with the same model
- Context has memory limits and will eventually be truncated

## Example Implementation

This example shows the API design and usage patterns. Full implementation requires:

1. ✅ `WithContext()` option function (implemented)
2. ⏳ Context extraction from responses (needs implementation)
3. ⏳ Integration with generate API when context is used (needs implementation)

## Running This Example

```bash
# Make sure Ollama is running locally
ollama serve

# Pull a model if you haven't already
ollama pull llama2

# Run the example
go run main.go
```

Note: This example will demonstrate the API usage pattern. Full context functionality requires additional implementation work.
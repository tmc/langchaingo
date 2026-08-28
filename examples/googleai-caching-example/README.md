# Google AI Caching Example

This example demonstrates how to use Google AI's context caching feature with langchaingo.

## What is Context Caching?

Context caching allows you to cache large amounts of content (system prompts, context documents, etc.) 
and reuse them across multiple requests. This can significantly reduce:

- **Latency**: Cached content is processed faster
- **Cost**: You only pay for the cached content once, not on every request
- **Token usage**: Cached tokens don't count against your rate limits

## Requirements

- Minimum cacheable content: **32,768 tokens** (~24,000 words)
- Supported models: gemini-2.5-pro and other Gemini 2.0+ models
- Google API key

## How It Works

1. **Create cached content** with a large system prompt or context
2. **Reference the cache** in subsequent requests using `WithCachedContent(cacheName)`
3. **Reuse the cache** across multiple different queries
4. **Update TTL** to extend cache lifetime if needed
5. **Delete cache** when no longer needed

## Running the Example

```bash
export GOOGLE_API_KEY="your-api-key-here"
go run main.go
```

## Expected Output

The example will:
1. Create a cached content with Go programming expertise (~32k+ tokens)
2. Make two different requests using the same cache
3. Show cached token usage in responses
4. List all cached contents
5. Clean up the cache when done

## Key Concepts

### Minimum Size Requirement

Google AI requires at least 32,768 tokens for caching. In practice, this is approximately:
- 24,000 words (assuming ~1.4 tokens per word)
- 120,000 characters (assuming ~4 characters per token)

### TTL (Time To Live)

Cached content automatically expires after the specified TTL. You can:
- Set initial TTL when creating: `1*time.Hour`, `24*time.Hour`, etc.
- Update TTL later: `helper.UpdateCachedContent(ctx, name, newTTL)`

### Cost Savings

Caching is particularly valuable when:
- You have a large system prompt used across many conversations
- You need to provide extensive context (documentation, knowledge base)
- You're building a chatbot with consistent personality/instructions
- You have reference materials that don't change often

## API Reference

### Creating a Cache

```go
cached, err := helper.CreateCachedContent(
    ctx,
    "gemini-2.5-pro",         // Model name
    messages,                 // Content to cache
    1*time.Hour,              // TTL
    "my-cache-name",          // Display name
)
```

### Using a Cache

```go
resp, err := client.GenerateContent(
    ctx,
    messages,
    googleai.WithCachedContent(cached.Name),  // Reference the cache
    llms.WithMaxTokens(500),
)
```

### Managing Caches

```go
// Get cache details
cache, err := helper.GetCachedContent(ctx, name)

// Update TTL
updated, err := helper.UpdateCachedContent(ctx, name, 2*time.Hour)

// List all caches
for cache, err := range helper.AllCachedContents(ctx) {
    // Process each cache
}

// Delete cache
err := helper.DeleteCachedContent(ctx, name)
```

## Best Practices

1. **Cache Reuse**: Create one cache and use it across many requests
2. **Appropriate TTL**: Set TTL based on how often content changes
3. **Clean Up**: Delete caches you no longer need to avoid accumulation
4. **Monitor Usage**: Check `CachedTokens` in response metadata
5. **Size Requirements**: Ensure content meets minimum 32,768 token threshold

## Limitations

- Minimum size: 32,768 tokens
- Maximum cached content per project: Check current quotas
- TTL range: Minimum 5 minutes, maximum 24 hours (may vary)
- Model support: Currently only Gemini 2.0+ models

## Learn More

- [Google AI Caching Documentation](https://ai.google.dev/gemini-api/docs/caching)
- [langchaingo Documentation](https://github.com/vxcontrol/langchaingo)

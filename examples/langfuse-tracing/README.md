# Langfuse Tracing Example

This example demonstrates how to use Langfuse tracing with LangChainGo to monitor and observe your LLM applications.

## What is Langfuse?

[Langfuse](https://langfuse.com) is an open-source LLM observability and analytics platform that helps you:

- 📊 **Monitor** LLM applications in production
- 🔍 **Trace** complex multi-step LLM workflows
- 📈 **Analyze** usage patterns and performance metrics  
- 🐛 **Debug** issues with detailed execution traces
- 💰 **Track costs** and token usage across models

## Features Demonstrated

This example shows how to trace:

1. **Basic LLM calls** - Simple question-answering with token usage tracking
2. **Chain execution** - Multi-step workflows with prompt templates
3. **Agent interactions** - Tool usage and decision-making processes
4. **Retrieval operations** - Document search and similarity matching
5. **Error handling** - Failed requests and exception tracking

## Setup

### 1. Get Langfuse Credentials

#### Option A: Langfuse Cloud (Recommended)
1. Sign up at [https://cloud.langfuse.com](https://cloud.langfuse.com)
2. Create a new project
3. Copy your public and secret keys from the project settings

#### Option B: Self-hosted Langfuse
1. Follow the [self-hosting guide](https://langfuse.com/docs/deployment/self-host)
2. Set your custom host URL in the environment variables

### 2. Set Environment Variables

```bash
export LANGFUSE_PUBLIC_KEY="pk_lf_..."
export LANGFUSE_SECRET_KEY="sk_lf_..."
export LANGFUSE_HOST="https://cloud.langfuse.com"  # Optional, defaults to cloud
export OPENAI_API_KEY="sk-..."  # Required for LLM calls
```

### 3. Install Dependencies

```bash
go mod tidy
```

### 4. Run the Example

```bash
go run main.go
```

## Code Overview

### Creating a Langfuse Handler

```go
handler, err := callbacks.NewLangfuseHandler(callbacks.LangfuseOptions{
    PublicKey: os.Getenv("LANGFUSE_PUBLIC_KEY"),
    SecretKey: os.Getenv("LANGFUSE_SECRET_KEY"),
    BaseURL:   os.Getenv("LANGFUSE_HOST"), // Optional
    UserID:    "user-123",
    SessionID: "session-456",
    Metadata: map[string]interface{}{
        "environment": "production",
        "version": "1.0.0",
    },
})
```

### Using with LLMs

```go
llm, err := openai.New(
    openai.WithCallback(handler),
)

// All LLM calls will now be traced
response, err := llms.GenerateFromSinglePrompt(ctx, llm, "Hello!")
```

### Using with Chains

```go
chain := chains.NewLLMChain(llm, template)

result, err := chains.Call(ctx, chain, inputs, 
    chains.WithCallback(handler),
)
```

### Using with Agents

```go
executor, err := agents.Initialize(
    llm,
    tools,
    agents.ZeroShotReactDescription,
    agents.WithCallback(handler),
)
```

### Manual Trace Management

```go
// Add metadata to current trace
handler.SetTraceMetadata(map[string]interface{}{
    "user_id": "user-123",
    "feature": "chat",
})

// Get the current trace ID
traceID := handler.GetTraceID()

// Flush pending traces
handler.Flush()
```

## What Gets Tracked

The Langfuse integration automatically captures:

### 🤖 LLM Calls
- Input prompts and messages
- Model responses and choices
- Token usage (input, output, total)
- Model parameters and settings
- Latency and timing information
- Error states and messages

### ⛓️ Chain Execution  
- Chain inputs and outputs
- Individual step execution
- Nested chain hierarchies
- Performance metrics

### 🤖 Agent Operations
- Tool selection and usage
- Agent reasoning and decisions
- Action sequences
- Final results

### 🔍 Retrieval Operations
- Search queries
- Retrieved documents
- Relevance scores
- Source metadata

## Viewing Traces

After running the example:

1. Open your Langfuse dashboard
2. Navigate to the "Traces" section
3. Find traces with session ID "langchaingo-demo" 
4. Explore the detailed execution tree

## Advanced Usage

### Custom Trace Hierarchies

```go
// Create nested spans for complex workflows
handler.HandleChainStart(ctx, map[string]any{"step": "preprocessing"})
handler.HandleLLMStart(ctx, []string{"process this data..."})
handler.HandleLLMGenerateContentEnd(ctx, response)
handler.HandleChainEnd(ctx, map[string]any{"result": "processed"})
```

### Error Tracking

```go
// Errors are automatically captured
handler.HandleLLMError(ctx, fmt.Errorf("API timeout"))
handler.HandleChainError(ctx, err)
handler.HandleToolError(ctx, err)
```

### Batch Operations

```go
// Process multiple items and track each
for i, item := range items {
    handler.SetTraceMetadata(map[string]interface{}{
        "batch_item": i,
        "item_id": item.ID,
    })
    
    // Process item...
}

// Flush all traces at the end
handler.Flush()
```

## Troubleshooting

### Common Issues

1. **Missing Environment Variables**
   ```
   Error: LANGFUSE_PUBLIC_KEY and LANGFUSE_SECRET_KEY environment variables are required
   ```
   → Set your Langfuse credentials in environment variables

2. **Network Connectivity**
   ```
   Error: Failed to send span to Langfuse: connection timeout
   ```
   → Check your internet connection and Langfuse host URL

3. **Invalid Credentials**
   ```
   Error: Langfuse API returned status: 401
   ```
   → Verify your public and secret keys are correct

### Debug Mode

Set `LANGFUSE_DEBUG=true` to enable verbose logging:

```bash
export LANGFUSE_DEBUG=true
go run main.go
```

## Performance Considerations

- Traces are sent asynchronously to avoid blocking your application
- Failed trace uploads are logged but don't affect your main application flow
- Use `handler.Flush()` at application shutdown to ensure all traces are sent
- Consider batching trace uploads for high-volume applications

## Next Steps

1. **Integrate with your application** - Add the Langfuse handler to your existing LangChainGo workflows
2. **Set up alerts** - Configure Langfuse to notify you of errors or performance issues  
3. **Analyze usage patterns** - Use Langfuse analytics to optimize your LLM applications
4. **Create dashboards** - Build custom views for your specific use cases

## Resources

- [Langfuse Documentation](https://langfuse.com/docs)
- [LangChainGo Documentation](https://tmc.github.io/langchaingo/docs/)
- [Langfuse Python SDK](https://langfuse.com/docs/sdk/python) (for comparison)
- [OpenAI API Documentation](https://platform.openai.com/docs/api-reference)
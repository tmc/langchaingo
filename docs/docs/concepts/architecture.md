# LangChain Go Architecture

This document explains the core architectural principles and design patterns that make LangChain Go powerful and idiomatic for Go developers.

## Design Philosophy

LangChain Go is built around several key principles:

### Interface-Driven Design

Every major component in LangChain Go is defined by interfaces, enabling:
- **Modularity**: Swap implementations without changing code
- **Testability**: Mock interfaces for comprehensive testing
- **Extensibility**: Add new providers and components easily

```go
type Model interface {
    GenerateContent(ctx context.Context, messages []MessageContent, options ...CallOption) (*ContentResponse, error)
}

type Chain interface {
    Call(ctx context.Context, inputs map[string]any, options ...ChainCallOption) (map[string]any, error)
    GetMemory() schema.Memory
    GetInputKeys() []string
    GetOutputKeys() []string
}
```

### Context-First Approach

All operations accept `context.Context` as the first parameter:
- **Cancellation**: Cancel long-running operations
- **Timeouts**: Set deadlines for API calls
- **Request Tracing**: Propagate request context through the call stack
- **Graceful Shutdown**: Handle application termination cleanly

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

response, err := llm.GenerateContent(ctx, messages)
```

### Go Idiomatic Patterns

#### Error Handling
LangChain Go uses Go's explicit error handling with typed errors:

```go
type Error struct {
    Code    ErrorCode
    Message string
    Cause   error
}

// Check for specific error types
if errors.Is(err, llms.ErrRateLimit) {
    // Handle rate limiting
}
```

#### Options Pattern
Functional options provide flexible configuration:

```go
llm, err := openai.New(
    openai.WithModel("gpt-4"),
    openai.WithTemperature(0.7),
    openai.WithMaxTokens(1000),
)
```

#### Channels and Goroutines
Leverage Go's concurrency primitives for streaming and parallel processing:

```go
// Streaming responses
response, err := llm.GenerateContent(ctx, messages, 
    llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
        select {
        case resultChan <- chunk:
        case <-ctx.Done():
            return ctx.Err()
        }
        return nil
    }),
)
```

## Core Components

### 1. Models Layer

The models layer provides abstractions for different types of language models:

```
┌─────────────────────────────────────────────────────────┐
│                    Models Layer                         │
├─────────────────┬─────────────────┬─────────────────────┤
│   Chat Models   │   LLM Models    │  Embedding Models   │
├─────────────────┼─────────────────┼─────────────────────┤
│  • OpenAI       │  • Completion   │  • OpenAI           │
│  • Anthropic    │  • Legacy APIs  │  • HuggingFace      │
│  • Google AI    │  • Local Models │  • Local Models     │
│  • Local (Ollama)│               │                     │
└─────────────────┴─────────────────┴─────────────────────┘
```

Each model type implements specific interfaces:
- `Model`: Unified interface for all language models
- `EmbeddingModel`: Specialized for generating embeddings
- `ChatModel`: Optimized for conversational interactions

### 2. Prompt Management

Prompts are first-class citizens with template support:

```go
template := prompts.NewPromptTemplate(
    "You are a {{.role}}. Answer this question: {{.question}}",
    []string{"role", "question"},
)

prompt, err := template.Format(map[string]any{
    "role":     "helpful assistant",
    "question": "What is Go?",
})
```

### 3. Memory Subsystem

Memory provides stateful conversation management:

```
┌─────────────────────────────────────────────────────────┐
│                  Memory Subsystem                       │
├─────────────────┬─────────────────┬─────────────────────┤
│  Buffer Memory  │ Window Memory   │  Summary Memory     │
├─────────────────┼─────────────────┼─────────────────────┤
│ • Simple buffer │ • Sliding window│ • Auto-summarization│
│ • Full history  │ • Fixed size    │ • Token management  │
│ • Fast access   │ • Recent focus  │ • Long conversations│
└─────────────────┴─────────────────┴─────────────────────┘
```

### 4. Chain Orchestration

Chains enable complex workflows:

```go
// Sequential chain example
chain1 := chains.NewLLMChain(llm, template1)
chain2 := chains.NewLLMChain(llm, template2)

sequential := chains.NewSequentialChain([]chains.Chain{chain1, chain2})
```

### 5. Agent Framework

Agents provide autonomous behavior:

```
┌─────────────────────────────────────────────────────────┐
│                  Agent Framework                        │
├─────────────────┬─────────────────┬─────────────────────┤
│     Agent       │     Tools       │    Executor         │
├─────────────────┼─────────────────┼─────────────────────┤
│ • Decision logic│ • Calculator    │ • Execution loop    │
│ • Tool selection│ • Web search    │ • Error handling    │
│ • ReAct pattern │ • File ops      │ • Result processing │
│ • Planning      │ • Custom tools  │ • Memory management │
└─────────────────┴─────────────────┴─────────────────────┘
```

## Data Flow

### Request Flow
```
User Input → Prompt Template → LLM → Output Parser → Response
     ↓             ↓              ↓         ↓           ↓
   Memory ←── Chain Logic ←── API Call ←── Processing ←── Memory
```

### Agent Flow
```
User Input → Agent Planning → Tool Selection → Tool Execution
     ↓              ↓              ↓              ↓
   Memory ←── Result Analysis ←── Tool Results ←── External APIs
     ↓              ↓
   Response ←── Final Answer
```

## Concurrency Model

LangChain Go embraces Go's concurrency model:

### Parallel Processing
```go
// Process multiple inputs concurrently
var wg sync.WaitGroup
results := make(chan string, len(inputs))

for _, input := range inputs {
    wg.Add(1)
    go func(inp string) {
        defer wg.Done()
        result, err := chain.Run(ctx, inp)
        if err == nil {
            results <- result
        }
    }(input)
}

wg.Wait()
close(results)
```

### Streaming
```go
// Stream processing with channels
type StreamProcessor struct {
    input  chan string
    output chan string
}

func (s *StreamProcessor) Process(ctx context.Context) {
    for {
        select {
        case input := <-s.input:
            // Process input
            result := processInput(input)
            s.output <- result
        case <-ctx.Done():
            return
        }
    }
}
```

## Extension Points

### Custom LLM Providers
Implement the `Model` interface:

```go
type CustomLLM struct {
    apiKey string
    client *http.Client
}

func (c *CustomLLM) GenerateContent(ctx context.Context, messages []MessageContent, options ...CallOption) (*ContentResponse, error) {
    // Custom implementation
}
```

### Custom Tools
Implement the `Tool` interface:

```go
type CustomTool struct {
    name        string
    description string
}

func (t *CustomTool) Name() string { return t.name }
func (t *CustomTool) Description() string { return t.description }
func (t *CustomTool) Call(ctx context.Context, input string) (string, error) {
    // Tool logic
}
```

### Custom Memory
Implement the `Memory` interface:

```go
type CustomMemory struct {
    storage map[string][]MessageContent
}

func (m *CustomMemory) ChatHistory() schema.ChatMessageHistory {
    // Return chat history implementation
}

func (m *CustomMemory) MemoryVariables() []string {
    return []string{"history"}
}
```

## Performance Considerations

### Connection Pooling
LLM providers use HTTP connection pooling for efficiency:

```go
client := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}
```

### Memory Management
- Use appropriate memory types for your use case
- Implement cleanup strategies for long-running applications
- Monitor memory usage in production

### Caching
Implement caching at multiple levels:
- LLM response caching
- Embedding caching
- Tool result caching

```go
type CachingLLM struct {
    llm   Model
    cache map[string]*ContentResponse
    mutex sync.RWMutex
}
```

## Error Handling Strategy

### Layered Error Handling
1. **Provider Level**: Handle API-specific errors
2. **Component Level**: Handle component-specific errors  
3. **Application Level**: Handle business logic errors

### Retry Logic
```go
func retryableCall(ctx context.Context, fn func() error) error {
    backoff := time.Second
    maxRetries := 3
    
    for i := 0; i < maxRetries; i++ {
        err := fn()
        if err == nil {
            return nil
        }
        
        if !isRetryable(err) {
            return err
        }
        
        select {
        case <-time.After(backoff):
            backoff *= 2
        case <-ctx.Done():
            return ctx.Err()
        }
    }
    
    return fmt.Errorf("max retries exceeded")
}
```

## Testing Architecture

### Interface Mocking
Use interfaces for comprehensive testing:

```go
type MockLLM struct {
    responses []string
    index     int
}

func (m *MockLLM) GenerateContent(ctx context.Context, messages []MessageContent, options ...CallOption) (*ContentResponse, error) {
    if m.index >= len(m.responses) {
        return nil, fmt.Errorf("no more responses")
    }
    
    response := &ContentResponse{
        Choices: []ContentChoice{{Content: m.responses[m.index]}},
    }
    m.index++
    return response, nil
}
```

### Integration Testing
Use testcontainers for external dependencies:

```go
func TestWithDatabase(t *testing.T) {
    ctx := context.Background()
    
    postgresContainer, err := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:13"),
        postgres.WithDatabase("test"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
    )
    require.NoError(t, err)
    defer postgresContainer.Terminate(ctx)
    
    // Test with real database
}
```

This architecture enables LangChain Go to be both powerful and maintainable, following Go's principles of simplicity, clarity, and performance.
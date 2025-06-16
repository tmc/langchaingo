# LangChainGo Architecture

This document explains LangChainGo's architecture and how it follows Go conventions.

## Modular adoption philosophy

**You don't need to adopt the entire LangChainGo framework.** The architecture is designed for selective adoption - use only the components that solve your specific problems:

- **Need an LLM client?** Use only the `llms` package
- **Want prompt templating?** Add the `prompts` package
- **Building conversational apps?** Include `memory` for state management
- **Creating autonomous agents?** Combine `agents`, `tools`, and `chains`

Each component is designed to work independently while providing seamless integration when combined. Start small and grow your usage as needed.

## Standard library alignment

LangChainGo follows Go's standard library patterns and philosophy. We model our interfaces after proven standard library designs:

- **`context.Context` first**: Like `database/sql`, `net/http`, and other standard library packages
- **Interface composition**: Small, focused interfaces that compose well (like `io.Reader`, `io.Writer`)
- **Constructor patterns**: `New()` functions with functional options (like `http.Client`)
- **Error handling**: Explicit errors with type assertions (like `net.OpError`, `os.PathError`)

When the standard library evolves, we evolve with it. Recent examples:
- Adopted `slog` patterns for structured logging
- Use `context.WithCancelCause` for richer cancellation
- Follow `testing/slogtest` patterns for handler validation

### Interface evolution

Our core interfaces will change as Go and the AI ecosystem evolve. We welcome discussion about better alignment with standard library patterns - open an issue if you see opportunities to make our APIs more Go-like.

Common areas for improvement:
- Method naming consistency with standard library conventions
- Error type definitions and handling patterns  
- Streaming patterns that match `io` package designs
- Configuration patterns that follow standard library examples

## Design philosophy

LangChainGo is built around several key principles:

### Interface-driven design

Every major component is defined by interfaces:
- **Modularity**: Swap implementations without changing code
- **Testability**: Mock interfaces for testing
- **Extensibility**: Add new providers and components

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

### Context-first approach

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

### Go idiomatic patterns

#### Error handling
Error handling uses Go's standard patterns with typed errors:

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

#### Options pattern
Functional options provide flexible configuration:

```go
llm, err := openai.New(
    openai.WithModel("gpt-4"),
    openai.WithTemperature(0.7),
    openai.WithMaxTokens(1000),
)
```

#### Channels and goroutines
Use Go's concurrency features for streaming and parallel processing:

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

## Core components

### 1. Models layer

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

### 2. Prompt management

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

### 3. Memory subsystem

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

### 4. Chain orchestration

Chains enable complex workflows:

```go
// Sequential chain example
chain1 := chains.NewLLMChain(llm, template1)
chain2 := chains.NewLLMChain(llm, template2)

sequential := chains.NewSequentialChain([]chains.Chain{chain1, chain2})
```

### 5. Agent framework

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

## Data flow

### Request flow
```
User Input → Prompt Template → LLM → Output Parser → Response
     ↓             ↓              ↓         ↓           ↓
   Memory ←── Chain Logic ←── API Call ←── Processing ←── Memory
```

### Agent flow
```
User Input → Agent Planning → Tool Selection → Tool Execution
     ↓              ↓              ↓              ↓
   Memory ←── Result Analysis ←── Tool Results ←── External APIs
     ↓              ↓
   Response ←── Final Answer
```

## Concurrency model

LangChainGo embraces Go's concurrency model:

### Parallel processing
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

## Extension points

### Custom LLM providers
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

### Custom tools
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

### Custom memory
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

## Performance considerations

### Connection pooling
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

### Memory management
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

## Error handling strategy

### Layered error handling
1. **Provider Level**: Handle API-specific errors
2. **Component Level**: Handle component-specific errors  
3. **Application Level**: Handle business logic errors

### Retry logic
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

## Testing architecture

### Interface mocking
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

### HTTP testing with httprr

For internal testing of HTTP-based LLM providers, LangChainGo uses [httprr](https://pkg.go.dev/github.com/tmc/langchaingo/internal/httprr) for recording and replaying HTTP interactions. This is an internal testing tool used by LangChainGo's own test suite to ensure reliable, fast tests without hitting real APIs.

#### Setting up httprr

```go
func TestOpenAIWithRecording(t *testing.T) {
    // Start httprr recorder
    recorder := httprr.New("testdata/openai_recording")
    defer recorder.Stop()
    
    // Configure HTTP client to use recorder
    client := &http.Client{
        Transport: recorder,
    }
    
    // Create LLM with custom client
    llm, err := openai.New(
        openai.WithHTTPClient(client),
        openai.WithToken("test-token"), // Will be redacted in recording
    )
    require.NoError(t, err)
    
    // Make actual API call (recorded on first run, replayed on subsequent runs)
    response, err := llm.GenerateContent(context.Background(), []llms.MessageContent{
        llms.TextParts(llms.ChatMessageTypeHuman, "Hello, world!"),
    })
    require.NoError(t, err)
    require.NotEmpty(t, response.Choices[0].Content)
}
```

#### Recording guidelines

1. **Initial Recording**: Run tests with real API credentials to create recordings
2. **Sensitive Data**: httprr automatically redacts common sensitive headers
3. **Deterministic Tests**: Recordings ensure consistent test results across environments
4. **Version Control**: Commit recording files for team consistency

#### Contributing with httprr

When contributing to LangChainGo's internal tests:

1. **Use httprr for new LLM providers**:
   ```go
   func TestNewProvider(t *testing.T) {
       recorder := httprr.New("testdata/newprovider_test")
       defer recorder.Stop()
       
       // Test implementation
   }
   ```

2. **Update recordings when APIs change**:
   ```bash
   # Delete old recordings
   rm testdata/provider_test.httprr
   
   # Re-run tests with real credentials
   PROVIDER_API_KEY=real_key go test
   ```

3. **Verify recordings are committed**:
   ```bash
   git add testdata/*.httprr
   git commit -m "test: update API recordings"
   ```

### Integration testing
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

This architecture follows Go's principles of simplicity, clarity, and performance.
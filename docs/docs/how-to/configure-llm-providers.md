# How to configure different LLM providers

This guide shows you how to configure and use different LLM providers with LangChainGo.

## OpenAI

### Basic configuration

```go
import "github.com/vxcontrol/langchaingo/llms/openai"

// Using environment variable OPENAI_API_KEY
llm, err := openai.New()

// Or with explicit API key
llm, err := openai.New(openai.WithToken("your-api-key"))
```

### Advanced configuration

```go
llm, err := openai.New(
    openai.WithToken("your-api-key"),
    openai.WithModel("gpt-4o"), // Specify model
    openai.WithBaseURL("https://custom-endpoint.com"), // Custom endpoint
    openai.WithOrganization("org-id"), // Organization ID
    openai.WithAPIVersion("2023-12-01"), // API version
)
```

### Azure OpenAI

```go
import "github.com/vxcontrol/langchaingo/llms/openai"

llm, err := openai.New(
    openai.WithToken("your-azure-api-key"),
    openai.WithBaseURL("https://your-resource.openai.azure.com"),
    openai.WithAPIVersion("2023-12-01-preview"),
    openai.WithAPIType(openai.APITypeAzure),
)
```

## Anthropic

### Basic configuration

```go
import "github.com/vxcontrol/langchaingo/llms/anthropic"

// Using environment variable ANTHROPIC_API_KEY
llm, err := anthropic.New()

// Or with explicit API key
llm, err := anthropic.New(anthropic.WithToken("your-api-key"))
```

### Model selection

```go
llm, err := anthropic.New(
    anthropic.WithModel("claude-sonnet-4-5-20250929"),
    anthropic.WithToken("your-api-key"),
)
```

## Google AI (Gemini)

### Basic configuration

```go
import (
    "context"
    "github.com/vxcontrol/langchaingo/llms/googleai"
)

// Using environment variable GOOGLE_API_KEY
llm, err := googleai.New(context.Background())

// Or with explicit API key
llm, err := googleai.New(
    context.Background(),
    googleai.WithAPIKey("your-api-key"),
)
```

### Model configuration

```go
llm, err := googleai.New(
    context.Background(),
    googleai.WithDefaultModel("gemini-2.5-flash"),
    googleai.WithAPIKey("your-api-key"),
)
```

## Vertex AI

### Basic configuration

```go
import (
    "context"
    "github.com/vxcontrol/langchaingo/llms/googleai"
    "github.com/vxcontrol/langchaingo/llms/googleai/vertex"
)

llm, err := vertex.New(
    context.Background(),
    googleai.WithCloudProject("your-project-id"),
    googleai.WithCloudLocation("us-central1"),
)
```

### With service account

```go
import (
    "context"
    "github.com/vxcontrol/langchaingo/llms/googleai"
    "github.com/vxcontrol/langchaingo/llms/googleai/vertex"
)

llm, err := vertex.New(
    context.Background(),
    googleai.WithCloudProject("your-project-id"),
    googleai.WithCloudLocation("us-central1"),
    googleai.WithCredentialsFile("path/to/service-account.json"),
)
```

## Local Models (Ollama)

### Basic configuration

```go
import "github.com/vxcontrol/langchaingo/llms/ollama"

// Default configuration (localhost:11434)
llm, err := ollama.New(ollama.WithModel("llama2"))

// Custom server
llm, err := ollama.New(
    ollama.WithServerURL("http://custom-server:11434"),
    ollama.WithModel("codellama"),
)
```

## Hugging Face

### Basic configuration

```go
import "github.com/vxcontrol/langchaingo/llms/huggingface"

// Using environment variable HF_TOKEN
llm, err := huggingface.New()

// Or with explicit token
llm, err := huggingface.New(huggingface.WithToken("your-hf-token"))
```

### Model selection

```go
llm, err := huggingface.New(
    huggingface.WithModel("microsoft/DialoGPT-medium"),
    huggingface.WithToken("your-hf-token"),
)
```

## Environment variables

Set up your environment with the appropriate API keys:

```bash
# OpenAI
export OPENAI_API_KEY="sk-..."

# Anthropic
export ANTHROPIC_API_KEY="sk-ant-..."

# Google AI
export GOOGLE_API_KEY="AI..."

# Hugging Face
export HF_TOKEN="hf_..."

# Vertex AI (using Application Default Credentials)
export GOOGLE_APPLICATION_CREDENTIALS="path/to/service-account.json"
```

## Provider-specific features

### OpenAI functions

```go
tools := []llms.Tool{
    {
        Type: "function",
        Function: &llms.FunctionDefinition{
            Name:        "get_weather",
            Description: "Get current weather",
            Parameters: map[string]any{
                "type": "object",
                "properties": map[string]any{
                    "location": map[string]any{
                        "type":        "string",
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

### Anthropic system messages

```go
messages := []llms.MessageContent{
    llms.TextParts(llms.ChatMessageTypeSystem, "You are a helpful assistant."),
    llms.TextParts(llms.ChatMessageTypeHuman, "Hello!"),
}
```

### Reasoning (thinking)

Reasoning-capable models (OpenAI o-series/GPT-5, Anthropic Claude extended thinking,
Gemini 2.5/3.x, and others) are controlled with provider-neutral options. The adapter
resolves the correct wire form for the target model, so a request is never sent in a
shape the model rejects:

```go
// Effort-based reasoning (token budget of 0 lets the model choose).
resp, err := llm.GenerateContent(ctx, messages,
    llms.WithReasoning(llms.ReasoningMedium, 0),
)

// Adaptive reasoning on the newest generations (e.g. Claude 4.6+).
resp, err = llm.GenerateContent(ctx, messages,
    llms.WithAdaptiveReasoning(llms.ReasoningHigh),
)

// Turn thinking off where the model allows disabling it.
resp, err = llm.GenerateContent(ctx, messages, llms.WithReasoningDisabled())
```

### Structured output

Request schema-constrained JSON with a provider-neutral option. The SDK sends the
schema through each provider's native structured-output mechanism and validates the
response against the original schema:

```go
schema := json.RawMessage(`{
    "type": "object",
    "properties": {"answer": {"type": "string"}},
    "required": ["answer"],
    "additionalProperties": false
}`)

resp, err := llm.GenerateContent(ctx, messages,
    llms.WithStructuredOutput(llms.StructuredOutputConfig{
        Name:   "answer_schema",
        Schema: schema,
    }),
)
```

### Streaming responses

```go
// Works with most providers. The callback receives a streaming.Chunk
// (import "github.com/vxcontrol/langchaingo/llms/streaming").
response, err := llm.GenerateContent(
    ctx,
    messages,
    llms.WithStreamingFunc(func(ctx context.Context, chunk streaming.Chunk) error {
        fmt.Print(chunk.String())
        return nil
    }),
)
```

## Error handling

```go
response, err := llm.GenerateContent(ctx, messages)
if err != nil {
    // Check for specific error types
    if errors.Is(err, llms.ErrRateLimit) {
        // Handle rate limiting
        time.Sleep(time.Second * 60)
        // Retry...
    } else if errors.Is(err, llms.ErrQuotaExceeded) {
        // Handle quota exceeded
        log.Fatal("API quota exceeded")
    } else {
        // Handle other errors
        log.Printf("LLM error: %v", err)
    }
}
```

## Best practices

1. **Use environment variables**: Store API keys securely in environment variables
2. **Handle rate limits**: Implement retry logic with exponential backoff
3. **Model selection**: Choose the right model for your use case and budget
4. **Error handling**: Implement robust error handling for different failure modes
5. **Resource management**: Use context for timeouts and cancellation
6. **Testing**: Use mock providers for testing (see testing guide)

## Provider comparison

| Provider | Strengths | Use cases |
|----------|-----------|-----------|
| OpenAI | High quality, function calling | General purpose, agents |
| Anthropic | Safety, long context | Research, content analysis |
| Google AI | Free tier, fast | Experimentation, mobile apps |
| Vertex AI | Enterprise features | Production, compliance |
| Ollama | Privacy, offline | Local development, sensitive data |
| Hugging Face | Open models, variety | Research, experimentation |
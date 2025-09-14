# How to configure different LLM providers

This guide shows you how to configure and use different LLM providers with LangChainGo.

## OpenAI

### Basic configuration

```go
import "github.com/tmc/langchaingo/llms/openai"

// Using environment variable OPENAI_API_KEY
llm, err := openai.New()

// Or with explicit API key
llm, err := openai.New(openai.WithToken("your-api-key"))
```

### Advanced configuration

```go
llm, err := openai.New(
    openai.WithToken("your-api-key"),
    openai.WithModel("gpt-4"), // Specify model
    openai.WithBaseURL("https://custom-endpoint.com"), // Custom endpoint
    openai.WithOrganization("org-id"), // Organization ID
    openai.WithAPIVersion("2023-12-01"), // API version
)
```

### Azure OpenAI

```go
import "github.com/tmc/langchaingo/llms/openai"

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
import "github.com/tmc/langchaingo/llms/anthropic"

// Using environment variable ANTHROPIC_API_KEY
llm, err := anthropic.New()

// Or with explicit API key
llm, err := anthropic.New(anthropic.WithToken("your-api-key"))
```

### Model selection

```go
llm, err := anthropic.New(
    anthropic.WithModel("claude-3-opus-20240229"),
    anthropic.WithToken("your-api-key"),
)
```

## Google AI (Gemini)

### Basic configuration

```go
import "github.com/tmc/langchaingo/llms/googleai"

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
    googleai.WithDefaultModel("gemini-pro"),
    googleai.WithAPIKey("your-api-key"),
)
```

## Vertex AI

### Basic configuration

```go
import (
    "github.com/tmc/langchaingo/llms/googleai"
    "github.com/tmc/langchaingo/llms/googleai/vertex"
)

llm, err := vertex.New(
    context.Background(),
    googleai.WithCloudProject("your-project-id"),
    googleai.WithCloudLocation("us-central1"),
)
```

### With service account

```go
llm, err := vertexai.New(
    context.Background(),
    vertexai.WithProjectID("your-project-id"),
    vertexai.WithLocation("us-central1"),
    vertexai.WithCredentialsFile("path/to/service-account.json"),
)
```

## Local Models (Ollama)

### Basic configuration

```go
import "github.com/tmc/langchaingo/llms/ollama"

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
import "github.com/tmc/langchaingo/llms/huggingface"

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
tools := []openai.Tool{
    {
        Type: "function",
        Function: openai.FunctionDefinition{
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

### Streaming responses

```go
// Works with most providers
response, err := llm.GenerateContent(
    ctx, 
    messages, 
    llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
        fmt.Print(string(chunk))
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
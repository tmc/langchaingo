# New Features Guide

This guide covers the latest features and capabilities added to LangChainGo. Learn how to use cutting-edge functionality to build better AI applications.

## Latest Features (2024-2025)

### 🆕 Ollama Reasoning Mode Support

Ollama now supports reasoning mode with the `think` parameter, enabling advanced reasoning capabilities for compatible models.

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/tmc/langchaingo/llms/ollama"
)

func main() {
    ctx := context.Background()
    
    // Initialize Ollama with reasoning capabilities
    llm, err := ollama.New(
        ollama.WithServerURL("http://localhost:11434"),
        ollama.WithModel("qwen2.5:7b"), // Use a reasoning-capable model
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Enable reasoning mode
    response, err := llm.GenerateContent(ctx, messages, 
        llms.WithThinkParameter(true),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Response:", response.Choices[0].Content)
}
```

### 🆕 Bedrock Tool Calling for Claude Models

Amazon Bedrock now supports tool calling with Anthropic Claude models, enabling function calling capabilities.

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"

    "github.com/tmc/langchaingo/llms"
    "github.com/tmc/langchaingo/llms/bedrock"
)

func main() {
    ctx := context.Background()
    
    // Initialize Bedrock with Claude
    llm, err := bedrock.New(
        bedrock.WithModelProvider("anthropic"),
        bedrock.WithModel("claude-3-haiku-20240307"),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Define a tool for weather information
    weatherTool := llms.Tool{
        Name:        "get_weather",
        Description: "Get current weather information for a location",
        Parameters: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "location": map[string]interface{}{
                    "type":        "string",
                    "description": "The city and state, e.g. San Francisco, CA",
                },
            },
            "required": []string{"location"},
        },
    }
    
    messages := []llms.MessageContent{
        {
            Role: llms.ChatMessageTypeHuman,
            Parts: []llms.ContentPart{
                llms.TextContent{Text: "What's the weather like in San Francisco?"},
            },
        },
    }
    
    response, err := llm.GenerateContent(ctx, messages,
        llms.WithTools([]llms.Tool{weatherTool}),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Handle tool calls
    for _, choice := range response.Choices {
        if choice.ToolCalls != nil {
            for _, toolCall := range choice.ToolCalls {
                fmt.Printf("Tool: %s\n", toolCall.FunctionCall.Name)
                fmt.Printf("Args: %s\n", toolCall.FunctionCall.Arguments)
            }
        }
    }
}
```

### 🆕 Amazon Nova Model Support

Bedrock now includes support for Amazon's new Nova foundation models with advanced multimodal capabilities.

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/tmc/langchaingo/llms/bedrock"
)

func main() {
    ctx := context.Background()
    
    // Use Amazon Nova Lite model
    llm, err := bedrock.New(
        bedrock.WithModelProvider("amazon"),
        bedrock.WithModel("amazon.nova-lite-v1:0"),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    response, err := llm.Call(ctx, "Explain quantum computing in simple terms")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Nova Response:", response)
}
```

### 🆕 OpenAI O1/O3 Reasoning Models Temperature Fix

Proper handling of temperature parameters for OpenAI's reasoning models (o1, o3) which don't support temperature settings.

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/tmc/langchaingo/llms"
    "github.com/tmc/langchaingo/llms/openai"
)

func main() {
    ctx := context.Background()
    
    // Use O1 model for complex reasoning
    llm, err := openai.New(openai.WithModel("o1-preview"))
    if err != nil {
        log.Fatal(err)
    }
    
    // Temperature is automatically ignored for reasoning models
    response, err := llm.GenerateContent(ctx, messages,
        llms.WithTemperature(0.7), // This will be ignored for o1 models
    )
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("O1 Response:", response.Choices[0].Content)
}
```

### 🆕 Enhanced HTTP Recording for Tests

Improved HTTP recording capabilities for better test reliability and debugging.

```go
package main

import (
    "testing"
    
    "github.com/tmc/langchaingo/internal/httprr"
)

func TestMyLLMCall(t *testing.T) {
    // Record HTTP interactions for reproducible tests
    transport := httprr.NewRecorder("testdata/my_test.httprr")
    defer transport.Close()
    
    // Use the recorder in your HTTP client
    // Your test code here...
}
```

### 🆕 Google AI Embedding Model Override Fix

Fixed issue where user-provided embedding models were being overridden by defaults.

```go
package main

import (
    "context"
    "log"

    "github.com/tmc/langchaingo/embeddings"
    "github.com/tmc/langchaingo/llms/googleai"
)

func main() {
    ctx := context.Background()
    
    // Your custom embedding model will now be respected
    googleai, err := googleai.New(
        googleai.WithEmbeddingModel("text-embedding-004"), // This won't be overridden
    )
    if err != nil {
        log.Fatal(err)
    }
    
    embeddings, err := googleai.EmbedDocuments(ctx, []string{"Hello world"})
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Generated %d embeddings", len(embeddings))
}
```

### 🆕 Proxy Support for GoogleAI and Llama.cpp

Added SOCKS5 and HTTP proxy support for GoogleAI and llama.cpp providers.

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/tmc/langchaingo/llms/googleai"
    "github.com/tmc/langchaingo/llms/llamacpp"
)

func main() {
    ctx := context.Background()
    
    // GoogleAI with proxy
    googleLLM, err := googleai.New(
        googleai.WithAPIKey("your-api-key"),
        googleai.WithHTTPProxy("http://proxy.example.com:8080"),
        // Or SOCKS5
        // googleai.WithSOCKS5Proxy("socks5://proxy.example.com:1080"),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Llama.cpp with proxy  
    llamaLLM, err := llamacpp.New(
        llamacpp.WithServerURL("http://localhost:8080"),
        llamacpp.WithHTTPProxy("http://proxy.example.com:8080"),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    response, err := googleLLM.Call(ctx, "Hello from behind proxy!")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Response:", response)
}
```

## Recent Enhancements

### 🔧 Improved Agent Robustness

MRKL and OpenAI Functions agents now have better error handling and recovery mechanisms.

```go
package main

import (
    "context"
    "log"

    "github.com/tmc/langchaingo/agents"
    "github.com/tmc/langchaingo/llms/openai"
    "github.com/tmc/langchaingo/tools"
)

func main() {
    ctx := context.Background()
    
    llm, err := openai.New()
    if err != nil {
        log.Fatal(err)
    }
    
    // Create agent with improved error handling
    agent := agents.NewOpenAIFunctionsAgent(llm, []tools.Tool{
        // Your tools here
    }, agents.WithMaxIterations(5)) // Better iteration control
    
    executor := agents.NewExecutor(agent)
    
    result, err := executor.Run(ctx, "Complex task requiring multiple steps")
    if err != nil {
        log.Fatal(err)
    }
    
    log.Println("Result:", result)
}
```

### 🔧 Enhanced Vector Store Capabilities

Improved performance and reliability across all vector store implementations.

### 🔧 Better Memory Management

More efficient memory usage and conversation history handling.

## Using New Features in Production

### Best Practices

1. **Feature Flags**: Use environment variables to toggle new features
```go
if os.Getenv("ENABLE_REASONING_MODE") == "true" {
    // Use reasoning mode
}
```

2. **Gradual Rollout**: Test new features with a subset of users first

3. **Monitoring**: Monitor performance and costs when using new models

4. **Fallback Strategies**: Always have fallback options for new capabilities

### Example Production Setup

```go
package main

import (
    "context"
    "os"
    "log"

    "github.com/tmc/langchaingo/llms/openai"
)

func createLLM() (*openai.LLM, error) {
    opts := []openai.Option{}
    
    // Use new default model unless overridden
    if model := os.Getenv("CUSTOM_MODEL"); model != "" {
        opts = append(opts, openai.WithModel(model))
    }
    
    // Add proxy support if configured
    if proxy := os.Getenv("HTTP_PROXY"); proxy != "" {
        opts = append(opts, openai.WithHTTPProxy(proxy))
    }
    
    return openai.New(opts...)
}

func main() {
    llm, err := createLLM()
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    response, err := llm.Call(ctx, "Production-ready LangChainGo!")
    if err != nil {
        log.Fatal(err)
    }
    
    log.Println("Response:", response)
}
```

## What's Coming Next

### Planned Features

- Enhanced multimodal support across more providers
- Improved streaming capabilities
- Better integration with vector databases
- Advanced agent frameworks
- Performance optimizations

### How to Stay Updated

1. **GitHub Releases**: Watch the [repository](https://github.com/tmc/langchaingo) for releases
2. **Community**: Join discussions in GitHub Issues
3. **Documentation**: Check back regularly for updated guides

---

*This guide covers features as of LangChainGo v0.1.14+. New features are added regularly.*
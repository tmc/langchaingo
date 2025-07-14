# AWS Bedrock LLM Provider

A comprehensive Go implementation for AWS Bedrock LLM models with support for multiple providers, streaming, tool calling, and multimodal capabilities.

## Features

### 🤖 **Extensive Model Support**
- **AI21**: Jamba 1.5 Large/Mini, Jamba Instruct
- **Amazon**: Nova Premier/Pro/Lite/Micro, Titan Text Premier  
- **Anthropic**: Claude 4 Opus/Sonnet, Claude 3.7 Sonnet, Claude 3.5 Haiku/Sonnet
- **Cohere**: Command-R, Command-R+, Command Text/Light
- **Meta**: Llama 3.3 70B, Llama 3.1 8B, Llama 3 70B/8B
- **DeepSeek**: R1

### 🚀 **Advanced Capabilities**
- **Streaming**: Real-time response streaming for most models
- **Tool Calling**: Full function calling support with streaming
- **Reasoning**: Extended thinking support for Claude 3.7+ and Nova models
- **Multimodal**: Text, image, and document processing
- **Two APIs**: Modern Converse API + Legacy model-specific APIs

## Quick Start

```go
import "github.com/vxcontrol/langchaingo/llms/bedrock"

// Basic usage
llm, err := bedrock.New()

// With Converse API (recommended)
llm, err := bedrock.New(
    bedrock.WithConverseAPI(),
    bedrock.WithModel(bedrock.ModelAnthropicClaudeSonnet4),
)

resp, err := llm.GenerateContent(ctx, messages)
```

## Streaming

```go
streamingFunc := func(ctx context.Context, chunk streaming.Chunk) error {
    fmt.Print(chunk.Content)
    return nil
}

resp, err := llm.GenerateContent(ctx, messages,
    llms.WithStreamingFunc(streamingFunc),
)
```

## Tool Calling

```go
tools := []llms.Tool{{
    Type: "function",
    Function: &llms.FunctionDefinition{
        Name:        "calculator",
        Description: "Perform arithmetic operations",
        Parameters: map[string]any{
            "type": "object",
            "properties": map[string]any{
                "operation": map[string]any{
                    "type": "string",
                    "enum": []string{"add", "subtract", "multiply", "divide"},
                },
                "a": map[string]any{"type": "number"},
                "b": map[string]any{"type": "number"},
            },
            "required": []string{"operation", "a", "b"},
        },
    },
}}

resp, err := llm.GenerateContent(ctx, messages,
    llms.WithTools(tools),
)
```

## Configuration Options

```go
// Custom AWS client
client := bedrockruntime.NewFromConfig(cfg)
llm, err := bedrock.New(bedrock.WithClient(client))

// Model selection
llm, err := bedrock.New(bedrock.WithModel(bedrock.ModelAmazonNovaProV1))

// Enable Converse API (recommended for new applications)
llm, err := bedrock.New(bedrock.WithConverseAPI())

// Callback handler
llm, err := bedrock.New(bedrock.WithCallback(handler))
```

## Key Differences

### Converse API vs Legacy API
- **Converse API**: Unified interface, better tool calling, reasoning support
- **Legacy API**: Direct model-specific implementations, broader model compatibility

### Model Capabilities
- **Tool Calling**: Nova, Anthropic Claude 4/3.7/3.5, Cohere Command-R+
- **Reasoning**: Anthropic Claude 4/3.7, Amazon Nova models
- **Streaming**: Most models (except AI21 Jamba series)
- **Multimodal**: Nova Pro/Lite, Claude models

## Environment Setup

```bash
export AWS_ACCESS_KEY_ID=your_access_key
export AWS_SECRET_ACCESS_KEY=your_secret_key  
export AWS_REGION=us-east-1
```

## Testing

```bash
# Basic functionality
go test -v -run TestAmazonOutput

# Streaming support  
go test -v -run TestAmazonStreamingOutput

# Tool calling
go test -v -run TestAmazonToolCalling
```

For comprehensive examples, see the [examples directory](../../examples/) in the langchaingo repository.

# GitHub Models LLM Integration for LangChain Go

This package provides integration with GitHub Models API for LangChain Go, allowing you to use various LLMs available through GitHub's Models API.

## Installation

```bash
go get github.com/tmc/langchaingo
```

## Usage

### Basic Usage

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/githubmodels"
)

func main() {
	// The GITHUB_TOKEN environment variable will be used automatically if not provided
	llm, err := githubmodels.New()
	if err != nil {
		log.Fatalf("Failed to create GitHub Models client: %v", err)
	}

	ctx := context.Background()
	result, err := llms.GenerateFromSinglePrompt(ctx, llm, "What is the capital of France?")
	if err != nil {
		log.Fatalf("Error generating text: %v", err)
	}

	fmt.Println(result)
}
```

### Configuration Options

You can configure the GitHub Models client with various options:

```go
llm, err := githubmodels.New(
    githubmodels.WithToken("your-github-token"), // Explicitly set token instead of using env var
    githubmodels.WithModel("anthropic/claude-3-sonnet"), // Choose a different model
    githubmodels.WithHTTPClient(customClient), // Use a custom HTTP client
    githubmodels.WithCallbacksHandler(myCallbacksHandler), // Add callbacks
)
```

### Available Models

GitHub Models provides access to various models including:

- `openai/gpt-4.1` (default)
- `anthropic/claude-3-sonnet`
- `anthropic/claude-3-haiku`
- `mistral/mistral-large`
- `mistral/mistral-small`

For the latest list of available models, refer to the GitHub documentation.

### Chat Completion

For a more complex conversation with multiple messages:

```go
messages := []llms.MessageContent{
    {
        Role: llms.ChatMessageTypeSystem,
        Parts: []llms.ContentPart{
            llms.TextContent{Text: "You are a helpful assistant."},
        },
    },
    {
        Role: llms.ChatMessageTypeHuman,
        Parts: []llms.ContentPart{
            llms.TextContent{Text: "What is the capital of France?"},
        },
    },
}

response, err := llm.GenerateContent(ctx, messages)
if err != nil {
    log.Fatalf("Error generating chat completion: %v", err)
}

fmt.Println(response.Choices[0].Content)
```

## Authentication

You'll need a GitHub token with the appropriate permissions to access the GitHub Models API. You can set this token as the `GITHUB_TOKEN` environment variable or provide it directly using the `WithToken` option.

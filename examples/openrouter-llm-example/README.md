# OpenRouter LLM Example

This example demonstrates how to use OpenRouter with langchaingo. OpenRouter provides a unified API to access various LLM models through a single endpoint.

## About OpenRouter

OpenRouter is an AI routing service that provides:
- Access to multiple LLM providers (OpenAI, Anthropic, Google, Meta, etc.) through a single API
- Automatic failover and load balancing
- Usage tracking and analytics
- Support for both free and premium models

## Setup

1. Get an OpenRouter API key from https://openrouter.ai/
2. Set the environment variable:
   ```bash
   export OPENROUTER_API_KEY="your-api-key-here"
   ```

## Usage

OpenRouter uses an OpenAI-compatible API, so you can use the OpenAI client with a custom base URL:

```go
llm, err := openai.New(
    openai.WithModel("meta-llama/llama-3.2-3b-instruct:free"),
    openai.WithBaseURL("https://openrouter.ai/api/v1"),
    openai.WithToken(apiKey),
)
```

## Available Models

OpenRouter provides access to many models including:
- **Free tier**: `meta-llama/llama-3.2-3b-instruct:free`, `google/gemma-2-9b-it:free`
- **Premium**: GPT-4, Claude, Gemini Pro, and many more
- Check https://openrouter.ai/models for the full list

## Running the Example

```bash
go run openrouter_llm_example.go
```

## Features

- ✅ Streaming responses supported
- ✅ Compatible with OpenAI client
- ✅ Access to multiple model providers
- ✅ Automatic handling of provider-specific quirks
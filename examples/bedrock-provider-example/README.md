# Bedrock Provider Example

This example demonstrates how to use the Bedrock LLM with different model providers, including support for Nova models and inference profiles.

## Features

- Automatic provider detection from model ID
- Explicit provider specification for edge cases
- Support for Nova models (e.g., `amazon.nova-lite-v1:0`)
- Support for inference profiles (e.g., `us.amazon.nova-lite-v1:0`)

## Prerequisites

1. AWS credentials configured (via environment variables or AWS credentials file)
2. Access to the Bedrock models you want to use

## Usage

```bash
# Using default Titan model
go run main.go

# Using Nova model
go run main.go -model "amazon.nova-lite-v1:0"

# Using inference profile
go run main.go -model "us.amazon.nova-lite-v1:0"

# Using Anthropic model with explicit provider (for edge cases)
go run main.go -model "us.anthropic.claude-3-7-sonnet-20250219-v1:0" -provider "anthropic"

# Custom prompt
go run main.go -prompt "What is the capital of France?"

# Verbose output
go run main.go -verbose
```

## Environment Variables

Set these environment variables before running:

```bash
export AWS_ACCESS_KEY_ID=your_key_id
export AWS_SECRET_ACCESS_KEY=your_secret_key
export AWS_REGION=us-east-1
```

## Supported Providers

The Bedrock integration automatically detects the provider from the model ID:

- **Nova**: Models containing `.nova-` (e.g., `amazon.nova-lite-v1:0`, `us.amazon.nova-pro-v1:0`)
- **Anthropic**: Models containing `anthropic` (e.g., `anthropic.claude-3-sonnet-20240229-v1:0`)
- **Amazon**: Models containing `amazon` (excluding Nova)
- **Meta**: Models containing `meta` (e.g., `meta.llama3-1-405b-instruct-v1:0`)
- **Cohere**: Models containing `cohere` (e.g., `cohere.command-r-plus-v1:0`)
- **AI21**: Models containing `ai21` (e.g., `ai21.jamba-1-5-large-v1:0`)

## Model Provider Option

For cases where automatic detection doesn't work correctly (e.g., custom inference endpoints), you can explicitly specify the provider:

```go
llm, err := bedrock.New(
    bedrock.WithModel("custom.endpoint.model-id"),
    bedrock.WithModelProvider("anthropic"), // Explicitly set provider
)
```
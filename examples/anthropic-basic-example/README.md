# Anthropic API Demo

This example demonstrates basic usage of the Anthropic Claude API through LangChainGo.

## Features

- Basic text completion
- Complex reasoning tasks
- System messages for context
- Token usage tracking

## Setup

1. Get an Anthropic API key from https://console.anthropic.com
2. Set the environment variable:
   ```bash
   export ANTHROPIC_API_KEY="your-api-key"
   ```
3. Run the example:
   ```bash
   go run main.go
   ```

## Examples Included

1. **Basic Completion**: Simple question answering
2. **Complex Reasoning**: Logic puzzle solving
3. **System Design**: Architecture planning with system context

## Supported Models

This example uses Claude 3.5 Sonnet, one of Anthropic's most capable models for general tasks.

## Token Usage

The demo tracks input and output tokens for each request to help monitor usage and costs.
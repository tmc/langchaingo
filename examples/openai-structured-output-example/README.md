# OpenAI Structured Output Example

This example demonstrates how to use the unified structured output API with OpenAI's GPT-4o models.

## Overview

The structured output API provides a consistent way to request JSON-formatted responses that conform to a specific schema across different LLM providers (OpenAI, Google, Anthropic).

## Key Features

- **Unified API**: Same code works across OpenAI, Google, and Anthropic
- **Schema Definition**: Define your expected JSON structure once
- **Strict Mode**: OpenAI's strict mode enforces exact schema adherence
- **Type Safety**: Parse responses into Go structs

## Running the Example

```bash
export OPENAI_API_KEY=your-api-key-here
go run main.go
```

## How It Works

1. **Define Schema**: Create a `llms.StructuredOutputDefinition` with your desired JSON structure
2. **Generate Content**: Use `llms.WithStructuredOutput(schema)` option when calling `GenerateContent`
3. **Parse Response**: The LLM returns valid JSON matching your schema

## Benefits Over Provider-Specific APIs

- **Portability**: Switch between providers without changing code
- **Consistency**: Same schema definition works everywhere
- **Simplicity**: No need to learn provider-specific formats

## Supported Models

- GPT-4o (2024-08-06 and later)
- GPT-4o-mini (2024-07-18 and later)

## Related Examples

- `openai-jsonformat-example`: Provider-specific JSON format API
- `openai-function-call-example`: Function calling with structured parameters

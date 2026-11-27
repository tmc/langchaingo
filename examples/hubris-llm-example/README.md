# Hubris LLM Example

This example demonstrates how to use [Hubris](https://hubris.pro) with langchaingo. Hubris is an OpenAI-compatible LLM gateway billed in Russian rubles: one API key and one balance for 500+ models from OpenAI, Anthropic, Google, DeepSeek, Qwen, Z.ai, Moonshot, xAI and MiniMax.

## About Hubris

- OpenAI-compatible API at `https://api.hubris.pro/v1` (chat completions, embeddings, images, audio)
- Model ids use the full `vendor/model` form, e.g. `anthropic/claude-sonnet-5`, `openai/gpt-5.6-luna`, `google/gemini-3.7-flash`
- Billing in rubles; usage and spending limits per API key

## Setup

1. Register at https://hubris.pro and create an API key at https://hubris.pro/keys
2. Set the environment variable:
   ```bash
   export HUBRIS_API_KEY="sk-gw-..."
   ```

## Usage

Hubris uses an OpenAI-compatible API, so you can use the OpenAI client with a custom base URL:

```go
llm, err := openai.New(
    openai.WithModel("anthropic/claude-haiku-4.5"),
    openai.WithBaseURL("https://api.hubris.pro/v1"),
    openai.WithToken(apiKey),
)
```

## Available Models

The full catalog with prices is at https://hubris.pro/models. A few popular ids:
- `anthropic/claude-sonnet-5`, `anthropic/claude-haiku-4.5`
- `openai/gpt-5.6-luna`, `openai/gpt-4o-mini`
- `google/gemini-3.7-flash`
- `deepseek/deepseek-v4-flash-0731`

## Running the Example

```bash
go run . -model anthropic/claude-haiku-4.5 -prompt "Write a haiku about Go programming language."
```

Flags: `-model`, `-prompt`, `-temp` (0.0-2.0), `-stream` (default true).

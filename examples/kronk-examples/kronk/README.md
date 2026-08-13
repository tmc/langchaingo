# Kronk Abstraction Layer

A thin wrapper over the [ardanlabs/kronk](https://github.com/ardanlabs/kronk) SDK that provides a simplified, consistent chat API for all kronk examples.

## What It Does

- Downloads and installs llama.cpp libraries on first run
- Downloads the configured GGUF model from HuggingFace
- Initializes the kronk backend and loads the model
- Implements LangChainGo's `llms.Model` and `embeddings.EmbedderClient`
- Supports native tool calls and multi-turn tool results
- Supports text, image, audio, and video message parts with compatible models
- Streams normal and reasoning content and returns token/performance usage
- Maps LangChainGo JSON mode to Kronk's grammar-constrained JSON output
- Integrates with LangChainGo callback handlers

Media messages require a compatible multimodal model and projection file.
`llms.BinaryPart` supports image, audio, and video MIME types. Kronk requires
base64 media data, so `llms.ImageURLPart` must contain a data URL rather than a
remote HTTP URL.

## Usage

```go
import (
	"context"
	"log"

	"github.com/tmc/langchaingo/examples/kronk-examples/kronk"
	"github.com/tmc/langchaingo/llms"
)

client, err := kronk.New(
	ctx,
	"unsloth/Qwen3-0.6B-Q8_0",
	kronk.WithAutoTune(true),
	kronk.WithContextWindow(8192),
	kronk.WithNSeqMax(2),
)
if err != nil {
	log.Fatal(err)
}
defer client.Unload(context.Background())

messages := []llms.MessageContent{
	llms.TextParts(llms.ChatMessageTypeSystem, "You are a helpful assistant."),
	llms.TextParts(llms.ChatMessageTypeHuman, "Hello!"),
}
response, err := client.GenerateContent(ctx, messages,
	llms.WithMaxTokens(256),
	llms.WithTemperature(0.6),
	llms.WithTopK(20),
	llms.WithTopP(0.95),
)
```

Native tools use the standard LangChainGo `llms.WithTools` and
`llms.WithToolChoice` options. Append returned `llms.ToolCall` values to an AI
message, execute them, and append each result as an `llms.ToolCallResponse` in
a tool message. See `../function-example` for a complete round trip.

## API

| Method | Description |
|--------|-------------|
| `kronk.New(ctx, modelSource, ...options)` | Install libs + model, init backend, load model |
| `client.GenerateContent(ctx, messages, ...options)` | Generate or stream a LangChainGo response |
| `client.Call(ctx, prompt, ...options)` | Generate a text response for one prompt |
| `client.CreateEmbedding(ctx, texts)` | Embed text using a Kronk embedding model |
| `client.Unload(ctx)` | Free model resources |

Kronk constructor options mirror the SDK's load-time model configuration.
LangChainGo call options map `MaxTokens`, `Temperature`, `TopK`,
`TopP`, `Seed`, `StopWords`, `RepetitionPenalty`, `FrequencyPenalty`, and
`PresencePenalty` to Kronk sampling parameters. Kronk-specific sampling
defaults can also be set at load time with `kronk.WithDefaultParams`.
Because LangChainGo call options do not distinguish an omitted numeric value
from an explicit zero, configure zero-valued sampling defaults (such as greedy
temperature) with `kronk.WithDefaultParams`.

Additional supported call options include `llms.WithTools`,
`llms.WithToolChoice`, `llms.WithJSONMode`, `llms.WithResponseMIMEType`,
`llms.WithStreamingFunc`, and `llms.WithStreamingReasoningFunc`.

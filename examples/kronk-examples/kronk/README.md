# Kronk Abstraction Layer

A thin wrapper over the [ardanlabs/kronk](https://github.com/ardanlabs/kronk) SDK that provides a simplified, consistent chat API for all kronk examples.

## What It Does

- Downloads and installs llama.cpp libraries on first run
- Downloads the configured GGUF model from HuggingFace
- Initializes the kronk backend and loads the model
- Exposes a simple `Chat` / `ChatStream` API with `Message` types

## Usage

```go
import "github.com/tmc/langchaingo/examples/kronk-examples/kronk"

client, err := kronk.New(ctx, kronk.Config{
    ModelSource:  "unsloth/Qwen3-0.6B-Q8_0",
    SystemPrompt: "You are a helpful assistant.",
    AutoTune:     true,
})
if err != nil {
    log.Fatal(err)
}
defer client.Unload(context.Background())

answer, err := client.ChatStream(ctx, func(chunk string) error {
    fmt.Print(chunk)
    return nil
}, kronk.UserMessage("Hello!"))
```

## API

| Method | Description |
|--------|-------------|
| `kronk.New(ctx, Config)` | Install libs + model, init backend, load model |
| `client.Chat(ctx, ...Message)` | One-shot chat, returns full response |
| `client.ChatStream(ctx, onChunk, ...Message)` | Streaming chat, calls `onChunk` per token |
| `client.Unload(ctx)` | Free model resources |
| `client.SDK()` | Access the underlying `*kronk.Kronk` for advanced use |

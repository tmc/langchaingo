# callbacks/otel

OpenTelemetry callback handler for LangChainGo.

This package implements the `callbacks.Handler` interface and creates
OpenTelemetry trace spans for LLM, chain, tool, and retriever lifecycle
events. It uses GenAI semantic convention attribute keys and is designed
to be safe by default — no prompt or completion text is captured unless
you explicitly opt in.

## Installation

This package is part of `github.com/tmc/langchaingo`. Import it directly:

```go
import otelcallback "github.com/tmc/langchaingo/callbacks/otel"
```

Because it introduces `go.opentelemetry.io/otel` as a dependency, it lives
in its own subpackage so users who do not need tracing pay no dependency cost.

## Quick Start

```go
package main

import (
    "context"

    "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"

    otelcallback "github.com/tmc/langchaingo/callbacks/otel"
    "github.com/tmc/langchaingo/llms/openai"
)

func main() {
    // 1. Set up an exporter (stdout here; swap for OTLP in production).
    exporter, _ := stdouttrace.New(stdouttrace.WithPrettyPrint())
    tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter))
    defer tp.Shutdown(context.Background())

    // 2. Create the handler.
    handler := otelcallback.NewHandler(
        otelcallback.WithTracerProvider(tp),
    )

    // 3. Pass it to your LLM or chain via the callbacks option.
    llm, _ := openai.New(openai.WithCallbacksHandler(handler))
    _ = llm
}
```

## Options

| Option | Default | Description |
|---|---|---|
| `WithTracerProvider(tp)` | Global OTel provider | Use a specific `trace.TracerProvider`. If omitted, the global OTel provider is used (no-op until the app configures a real one). |
| `WithCaptureInput(bool)` | `false` | Capture prompt / input text as span attributes. **See privacy warning below.** |
| `WithCaptureOutput(bool)` | `false` | Capture completion / output text as span attributes. **See privacy warning below.** |

## Instrumented Events

| Lifecycle event | Span name | Span kind | Key attributes |
|---|---|---|---|
| `HandleLLMStart` | `llm.call` | `CLIENT` | `gen_ai.operation.name`, `gen_ai.prompt.count` |
| `HandleLLMGenerateContentStart` / `End` | `llm.generate_content` | `CLIENT` | `gen_ai.operation.name`, `gen_ai.prompt.message_count`, `gen_ai.usage.input_tokens`, `gen_ai.usage.output_tokens` |
| `HandleChainStart` / `End` | `chain` | `INTERNAL` | `gen_ai.operation.name` |
| `HandleToolStart` / `End` | `tool.call` | `INTERNAL` | `gen_ai.operation.name` |
| `HandleRetrieverStart` / `End` | `retriever.query` | `INTERNAL` | `gen_ai.operation.name`, `retriever.document_count` |
| `HandleLLMError` / `HandleChainError` / `HandleToolError` | *(ends active span)* | — | `error.type`, OTel exception event, span status `ERROR` |
| `HandleStreamingFunc` | *(no span)* | — | Intentionally omitted — per-chunk spans would be extremely noisy |

Token usage (`gen_ai.usage.input_tokens`, `gen_ai.usage.output_tokens`) is
read from `ContentResponse.Choices[0].GenerationInfo["InputTokens"]` and
`["OutputTokens"]`. Not all LLM providers populate these fields.

## Privacy Warning

**Prompt and completion content is NOT recorded by default.**

If you enable content capture, the raw text of every prompt and every
model completion will appear as span attributes in your trace backend.
Only enable this if:

- You understand what data is being sent to users' prompts in your application.
- Your trace backend (Jaeger, Tempo, Honeycomb, etc.) is appropriately secured.
- You have reviewed your organisation's data retention and privacy policies.

To opt in explicitly:

```go
handler := otelcallback.NewHandler(
    otelcallback.WithTracerProvider(tp),
    otelcallback.WithCaptureInput(true),  // records prompt text
    otelcallback.WithCaptureOutput(true), // records completion text
)
```

## Running the Example

A working example using the stdout exporter is in:

```
examples/otel-callback-example/
```

Run it with:

```bash
cd examples/otel-callback-example
go run .
```

It prints JSON-formatted OTel spans to stdout so you can see exactly what
the handler emits without needing any external collector.

## Semantic Conventions Note

The `gen_ai.*` attribute keys used in this package follow the
[OpenTelemetry GenAI semantic conventions](https://opentelemetry.io/docs/specs/semconv/gen-ai/).
They are defined as local constants in `handler.go` because
`go.opentelemetry.io/otel/semconv` does not yet export stable GenAI
constants. When that changes, the constants will be updated to use the
official package.

## Running Tests

```bash
go test ./callbacks/otel/...
```

Tests use `go.opentelemetry.io/otel/sdk/trace/tracetest` with an in-memory
span recorder — no external services or mocks required.

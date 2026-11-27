// otel-callback-example demonstrates how to use the OpenTelemetry callback
// handler with LangChainGo. It uses an in-memory exporter so you can see
// the spans printed to stdout without any external collector.
package main

import (
	"context"
	"fmt"
	"log"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	otelcallback "github.com/tmc/langchaingo/callbacks/otel"
	"github.com/tmc/langchaingo/llms"
)

func main() {
	// 1. Set up a stdout OTel exporter (prints JSON spans to stdout).
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatal(err)
	}

	// 2. Build a TracerProvider with the stdout exporter.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("langchaingo-otel-example"),
		)),
	)
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()

	// 3. Create the OTel callback handler, passing in our TracerProvider.
	handler := otelcallback.NewHandler(
		otelcallback.WithTracerProvider(tp),
		// otelcallback.WithCaptureInput(true),  // uncomment to capture prompts
		// otelcallback.WithCaptureOutput(true), // uncomment to capture completions
	)

	ctx := context.Background()

	// 4. Simulate the lifecycle events that LangChainGo would fire.
	//    In real usage you pass handler to an LLM or chain via its callbacks option.
	fmt.Println("=== Simulating LLM GenerateContent lifecycle ===")

	handler.HandleLLMGenerateContentStart(ctx, []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "What is the capital of France?"),
	})

	handler.HandleLLMGenerateContentEnd(ctx, &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content:    "Paris",
				StopReason: "stop",
				GenerationInfo: map[string]any{
					"InputTokens":  10,
					"OutputTokens": 3,
				},
			},
		},
	})

	fmt.Println("\n=== Simulating Tool lifecycle ===")

	handler.HandleToolStart(ctx, "search: Paris weather")
	handler.HandleToolEnd(ctx, "Sunny, 22°C")

	fmt.Println("\n=== Simulating Retriever lifecycle ===")

	handler.HandleRetrieverStart(ctx, "capital of France")
	handler.HandleRetrieverEnd(ctx, "capital of France", nil)

	fmt.Println("\nDone. Spans printed above.")
}

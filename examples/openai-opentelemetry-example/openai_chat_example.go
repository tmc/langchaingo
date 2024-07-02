package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"go.opentelemetry.io/otel"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// 1. Setup OpenTelemetry SDK
	otelShutdown, err := setupOTelSDK(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	// 2. Create a LangChain-Go CallbacksHandler for OpenTelemetry tracing
	h, err := callbacks.NewOpenTelemetryCallbacksHandler(otel.Tracer("langchaingo"), callbacks.WithLogPrompts(true), callbacks.WithLogCompletions(true))
	if err != nil {
		log.Fatal(err)
	}

	// 3. Pass the CallbackHandler in options when setting up the LangChain-Go OpenAI LLM
	llm, err := openai.New(openai.WithCallback(h))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a company branding design wizard."),
		llms.TextParts(llms.ChatMessageTypeHuman, "What would be a good company jingle? Only provide the jingle."),
	}

	// 4. Use LLM instance as normal
	if _, err := llm.GenerateContent(ctx, content,
		llms.WithMaxTokens(1024),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			fmt.Print(string(chunk))
			return nil
		})); err != nil {
		log.Fatal(err)
	}
}

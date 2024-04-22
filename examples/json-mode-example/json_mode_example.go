package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
)

var flagBackend = flag.String("backend", "openai", "backend to use")

func main() {
	flag.Parse()
	ctx := context.Background()
	llm, err := initBackend(ctx)
	if err != nil {
		log.Fatal(err)
	}
	completion, err := llms.GenerateFromSinglePrompt(ctx,
		llm,
		"Who was first man to walk on the moon? Respond in json format, include `first_man` in response keys.",
		llms.WithTemperature(0.0),
		llms.WithJSONMode(),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(completion)
}

func initBackend(ctx context.Context) (llms.Model, error) {
	switch *flagBackend {
	case "openai":
		return openai.New()
	case "ollama":
		return ollama.New(ollama.WithModel("mistral"))
	case "anthropic":
		return anthropic.New(anthropic.WithModel("claude-2.1"))
	case "googleai":
		return googleai.New(ctx, googleai.WithAPIKey(os.Getenv("GOOGLE_AI_API_KEY")))
	default:
		return nil, fmt.Errorf("unknown backend: %s", *flagBackend)
	}
}

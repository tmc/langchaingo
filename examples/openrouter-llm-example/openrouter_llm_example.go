package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/vendasta/langchaingo/llms"
	"github.com/vendasta/langchaingo/llms/openai"
)

func main() {
	// Command-line flags
	model := flag.String("model", "meta-llama/llama-3.2-3b-instruct:free", "OpenRouter model to use (see https://openrouter.ai/models)")
	prompt := flag.String("prompt", "Write a haiku about Go programming language.", "Prompt to send to the model")
	temperature := flag.Float64("temp", 0.8, "Temperature for response generation (0.0-2.0)")
	streaming := flag.Bool("stream", true, "Use streaming mode")
	flag.Parse()

	// OpenRouter provides access to multiple LLM providers through a unified API
	// Get your API key from https://openrouter.ai/
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("Please set OPENROUTER_API_KEY environment variable\n" +
			"Get your key from https://openrouter.ai/")
	}

	// Create an OpenAI-compatible client configured for OpenRouter
	llm, err := openai.New(
		openai.WithModel(*model),
		openai.WithBaseURL("https://openrouter.ai/api/v1"),
		openai.WithToken(apiKey),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	fmt.Println("🚀 OpenRouter CLI - langchaingo")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Model: %s\n", *model)
	fmt.Printf("Temperature: %.1f\n", *temperature)
	fmt.Printf("Streaming: %v\n", *streaming)
	fmt.Printf("Prompt: %s\n", *prompt)
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println()

	// Generate response
	opts := []llms.CallOption{
		llms.WithTemperature(*temperature),
	}

	if *streaming {
		opts = append(opts, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			fmt.Print(string(chunk))
			return nil
		}))
	}

	response, err := llms.GenerateFromSinglePrompt(ctx, llm, *prompt, opts...)

	if !*streaming && err == nil {
		fmt.Println(response)
	}

	fmt.Println()

	if err != nil {
		if strings.Contains(err.Error(), "429") {
			fmt.Println("⚠️  Rate limit reached. Free tier models are limited to 1 request per minute.")
			fmt.Println("    Try using a different model with -model flag")
		} else {
			log.Printf("Error: %v\n", err)
		}
		os.Exit(1)
	}

	fmt.Println("✅ Success!")
}

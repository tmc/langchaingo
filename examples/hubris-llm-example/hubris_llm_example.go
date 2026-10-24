package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	// Command-line flags
	model := flag.String("model", "anthropic/claude-haiku-4.5", "Hubris model to use (see https://hubris.pw/models)")
	prompt := flag.String("prompt", "Write a haiku about Go programming language.", "Prompt to send to the model")
	temperature := flag.Float64("temp", 0.8, "Temperature for response generation (0.0-2.0)")
	streaming := flag.Bool("stream", true, "Use streaming mode")
	flag.Parse()

	// Hubris is an OpenAI-compatible LLM gateway billed in Russian rubles:
	// one API key for models from OpenAI, Anthropic, Google, DeepSeek, Qwen and others.
	// Get your API key from https://hubris.pw/keys
	apiKey := os.Getenv("HUBRIS_API_KEY")
	if apiKey == "" {
		log.Fatal("Please set HUBRIS_API_KEY environment variable\n" +
			"Get your key from https://hubris.pw/keys")
	}

	// Create an OpenAI-compatible client configured for Hubris
	llm, err := openai.New(
		openai.WithModel(*model),
		openai.WithBaseURL("https://api.hubris.pw/v1"),
		openai.WithToken(apiKey),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	fmt.Println("🚀 Hubris CLI - langchaingo")
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
			fmt.Println("⚠️  Rate limit reached (key spending limit or upstream 429).")
			fmt.Println("    Check the key limits at https://hubris.pw/keys or try another model with -model")
		} else {
			log.Printf("Error: %v\n", err)
		}
		os.Exit(1)
	}

	fmt.Println("✅ Success!")
}


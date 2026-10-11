// structured_output_example demonstrates grammar-constrained JSON generation
// with the kronk LangChainGo adapter.
//
// Run from this directory:
//
//	go run structured_output_example.go
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/tmc/langchaingo/examples/kronk-examples/kronk"
	"github.com/tmc/langchaingo/llms"
)

var flagModel = flag.String("model", "unsloth/Qwen3-0.6B-Q8_0", "GGUF model source")

func main() {
	flag.Parse()
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	client, err := kronk.New(context.Background(), *flagModel, kronk.WithAutoTune(true))
	if err != nil {
		return fmt.Errorf("create kronk client: %w", err)
	}
	defer func() {
		if err := client.Unload(context.Background()); err != nil {
			log.Printf("unload model: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	response, err := client.GenerateContent(ctx, []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "Return only a JSON object."),
		llms.TextParts(llms.ChatMessageTypeHuman, `Describe Go using keys "language", "compiled", and "strengths".`),
	}, llms.WithJSONMode())
	if err != nil {
		return fmt.Errorf("generate content: %w", err)
	}
	if len(response.Choices) == 0 {
		return errors.New("model returned no choices")
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(response.Choices[0].Content), &result); err != nil {
		return fmt.Errorf("decode model JSON %q: %w", response.Choices[0].Content, err)
	}

	formatted, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("format model JSON: %w", err)
	}
	fmt.Println(string(formatted))
	fmt.Printf("\nUsage: %#v\n", response.Choices[0].GenerationInfo)
	return nil
}

// stream_example demonstrates streaming a response from a local GGUF model
// running via llama.cpp through the kronk abstraction layer. It uses
// llms.GenerateContent with llms.WithStreamingFunc to print each token chunk
// to the console as it arrives.
//
// The first time you run this program the system will download and install
// the llama.cpp libraries and the model. Subsequent runs load from disk.
//
// Run from this directory:
//
//	go run stream_example.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tmc/langchaingo/examples/kronk-examples/kronk"
	"github.com/tmc/langchaingo/llms"
)

func main() {
	modelSource := "unsloth/Qwen3-0.6B-Q8_0"
	if v := os.Getenv("KRONK_TEST_MODEL"); v != "" {
		modelSource = v
	}

	fmt.Println("Initializing kronk (first run downloads libraries and model)...")
	client, err := kronk.New(
		context.Background(),
		modelSource,
		kronk.WithAutoTune(true),
	)
	if err != nil {
		log.Fatalf("create kronk client: %v", err)
	}

	defer func() {
		fmt.Println("\nUnloading Kronk")
		if err := client.Unload(context.Background()); err != nil {
			fmt.Printf("failed to unload model: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a company branding design wizard."),
		llms.TextParts(llms.ChatMessageTypeHuman, "What would be a good company name for a company that produces Go-backed LLM tools?"),
	}

	fmt.Println("\nMODEL>")
	completion, err := client.GenerateContent(ctx, content, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		fmt.Print(string(chunk))
		return nil
	}))
	if err != nil {
		log.Fatalf("generate content: %v", err)
	}
	_ = completion
}

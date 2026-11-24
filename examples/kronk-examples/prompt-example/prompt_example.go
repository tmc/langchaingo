// prompt_example demonstrates a minimal chat session built on the kronk
// abstraction layer. It uses llms.GenerateFromSinglePrompt to ask a single
// question to a local GGUF model running via llama.cpp.
//
// The first time you run this program the system will download and install
// the llama.cpp libraries and the model. Subsequent runs load from disk.
//
// Run from this directory:
//
//	go run prompt_example.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/tmc/langchaingo/examples/kronk-examples/kronk"
	"github.com/tmc/langchaingo/llms"
)

var (
	flagModel  = flag.String("model", "unsloth/Qwen3-0.6B-Q8_0", "GGUF model source (HuggingFace URL or provider/modelID)")
	flagPrompt = flag.String("prompt", "In one sentence, what makes Go a good language for concurrent programming?", "question to ask the model")
)

func main() {
	flag.Parse()

	fmt.Println("Initializing kronk (first run downloads libraries and model)...")
	client, err := kronk.New(
		context.Background(),
		*flagModel,
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

	fmt.Printf("\nQUESTION: %s\n", *flagPrompt)
	fmt.Println("\nMODEL>")

	// client implements llms.Model, so it can be passed directly to
	// llms.GenerateFromSinglePrompt. WithStreamingFunc streams each chunk
	// to the console as it arrives.
	answer, err := llms.GenerateFromSinglePrompt(ctx, client, *flagPrompt,
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			fmt.Print(string(chunk))
			return nil
		}),
	)
	if err != nil {
		log.Fatalf("chat: %v", err)
	}

	fmt.Printf("\n\n--- full response ---\n%s\n", answer)
}

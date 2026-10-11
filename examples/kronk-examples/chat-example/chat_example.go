// chat_example demonstrates an interactive chat session built on the kronk
// abstraction layer. Unlike the prompt-example which asks a single hard-coded
// question, this example maintains a conversation history and lets you chat
// back and forth with a local GGUF model running via llama.cpp.
//
// The first time you run this program the system will download and install
// the llama.cpp libraries and the model. Subsequent runs load from disk.
//
// Run from this directory:
//
//	go run chat_example.go
//
// Type "quit" or press Ctrl+D to end the conversation.
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/tmc/langchaingo/examples/kronk-examples/kronk"
	"github.com/tmc/langchaingo/llms"
)

var flagModel = flag.String("model", "unsloth/Qwen3-0.6B-Q8_0", "GGUF model source (HuggingFace URL or provider/modelID)")

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

	if err := chat(client); err != nil {
		log.Fatalf("chat: %v", err)
	}
}

// chat runs an interactive REPL loop. It maintains the full conversation
// history in messages and passes it to the model on each turn so the model
// has context from prior exchanges.
func chat(client *kronk.Client) error {
	reader := bufio.NewReader(os.Stdin)

	// messages holds the running conversation history, including the system
	// prompt supplied to the model with every turn.
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a concise, friendly assistant. Answer clearly and keep responses brief."),
	}

	fmt.Println("\nChat ready. Type your message and press Enter. Type 'quit' to exit.")

	for {
		fmt.Print("\nYOU> ")

		input, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println()
				return nil
			}
			return fmt.Errorf("read input: %w", err)
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		if input == "quit" || input == "exit" {
			return nil
		}

		// Append the user's message to the conversation history.
		messages = append(messages, llms.TextParts(llms.ChatMessageTypeHuman, input))

		fmt.Print("\nMODEL> ")

		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)

		// GenerateContent streams the response chunk-by-token via
		// WithStreamingFunc. The full response is also captured in
		// resp.Choices[0].Content so we can append it to history.
		resp, err := client.GenerateContent(ctx, messages,
			llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
				fmt.Print(string(chunk))
				return nil
			}),
		)
		cancel()
		if err != nil {
			fmt.Printf("\n[error: %v]\n", err)
			// Remove the user message we just added so the history stays clean
			// for the next attempt.
			messages = messages[:len(messages)-1]
			continue
		}

		var answer string
		if len(resp.Choices) > 0 {
			answer = resp.Choices[0].Content
		}
		if answer == "" {
			answer = "(no response)"
		}

		// Append the model's response to the conversation history.
		messages = append(messages, llms.TextParts(llms.ChatMessageTypeAI, answer))

		fmt.Println()
	}
}

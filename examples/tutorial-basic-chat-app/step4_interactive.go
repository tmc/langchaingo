package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/vendasta/langchaingo/llms"
	"github.com/vendasta/langchaingo/llms/openai"
)

// Step 4: Interactive Chat
func interactiveChat() {
	// Initialize LLM
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Chat Application Started (type 'quit' to exit)")
	fmt.Println("----------------------------------------")

	for {
		fmt.Print("You: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "quit" {
			break
		}

		response, err := llms.GenerateFromSinglePrompt(ctx, llm, input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("AI: %s\n\n", response)
	}
}
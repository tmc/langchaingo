package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/memory"
)

// Step 6: Advanced Chat with Chains
func advancedChat() {
	// Initialize LLM
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}

	// Create conversation memory
	chatMemory := memory.NewConversationBuffer()

	// Create conversation chain
	// The built-in conversation chain includes a default prompt template
	// and handles memory automatically
	conversationChain := chains.NewConversation(llm, chatMemory)

	ctx := context.Background()
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Advanced Chat Application (type 'quit' to exit)")
	fmt.Println("----------------------------------------")

	for {
		fmt.Print("You: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "quit" {
			break
		}

		// Run the chain with the input
		result, err := chains.Run(ctx, conversationChain, input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("AI: %s\n\n", result)
	}

	fmt.Println("Goodbye!")
}
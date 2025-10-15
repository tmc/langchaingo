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
	"github.com/vendasta/langchaingo/memory"
)

// Step 5: Chat with Memory
func chatWithMemory() {
	// Initialize LLM
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}

	// Create conversation memory
	chatMemory := memory.NewConversationBuffer()
	ctx := context.Background()
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Chat with Memory (type 'quit' to exit)")
	fmt.Println("----------------------------------------")

	for {
		fmt.Print("You: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "quit" {
			break
		}

		// Get conversation history
		messages, _ := chatMemory.ChatHistory.Messages(ctx)

		// Format the conversation
		var conversation string
		for _, msg := range messages {
			conversation += msg.GetContent() + "\n"
		}

		// Add current input to the conversation
		fullPrompt := conversation + "Human: " + input + "\nAssistant:"

		// Generate response
		response, err := llms.GenerateFromSinglePrompt(ctx, llm, fullPrompt)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		// Save to memory
		chatMemory.ChatHistory.AddUserMessage(ctx, input)
		chatMemory.ChatHistory.AddAIMessage(ctx, response)

		fmt.Printf("AI: %s\n\n", response)
	}
}

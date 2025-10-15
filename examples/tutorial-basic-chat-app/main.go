package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/vendasta/langchaingo/chains"
	"github.com/vendasta/langchaingo/llms"
	"github.com/vendasta/langchaingo/llms/openai"
	"github.com/vendasta/langchaingo/memory"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "step3", "basic":
			runBasicChat()
		case "step4", "interactive":
			runInteractiveChat()
		case "step5", "memory":
			runChatWithMemory()
		case "step6", "advanced":
			runAdvancedChat()
		default:
			fmt.Println("Usage: go run . [step3|step4|step5|step6|basic|interactive|memory|advanced]")
			fmt.Println("If no argument provided, runs the advanced chat (step6)")
		}
	} else {
		// Default to advanced chat
		runAdvancedChat()
	}
}

// Step 3: Basic Chat Application
func runBasicChat() {
	fmt.Println("=== Step 3: Basic Chat ===")

	// Initialize the OpenAI LLM
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}

	// Create a context
	ctx := context.Background()

	// Send a message to the LLM
	response, err := llms.GenerateFromSinglePrompt(
		ctx,
		llm,
		"Hello! How can you help me today?",
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("AI:", response)
}

// Step 4: Interactive Chat
func runInteractiveChat() {
	fmt.Println("=== Step 4: Interactive Chat ===")

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

// Step 5: Chat with Memory
func runChatWithMemory() {
	fmt.Println("=== Step 5: Chat with Memory ===")

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

// Step 6: Advanced Chat with Chains and Prompt Templates
func runAdvancedChat() {
	fmt.Println("=== Step 6: Advanced Chat with Chains ===")

	// Initialize LLM
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}

	// Create conversation memory
	chatMemory := memory.NewConversationBuffer()

	// Create conversation chain
	// This uses the default conversation template with built-in memory handling
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
		// The chain automatically manages conversation history
		result, err := chains.Run(ctx, conversationChain, input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("AI: %s\n\n", result)
	}

	fmt.Println("Goodbye!")
}

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/deepseek"
)

func main() {
	ctx := context.Background()

	// Example 1: DeepSeek Chat for general conversations
	fmt.Println("=== DeepSeek Chat Example ===")
	chatModel, err := deepseek.New(
		deepseek.WithModel(deepseek.ModelDeepSeekChat),
	)
	if err != nil {
		log.Fatal(err)
	}

	chatResponse, err := chatModel.Call(ctx, "Explain the concept of recursion in programming in simple terms.", 
		llms.WithMaxTokens(200),
		llms.WithTemperature(0.7),
	)
	if err != nil {
		log.Printf("Chat model error: %v", err)
	} else {
		fmt.Printf("Chat Response: %s\n\n", chatResponse)
	}

	// Example 2: DeepSeek Coder for programming tasks
	fmt.Println("=== DeepSeek Coder Example ===")
	coderModel, err := deepseek.New(
		deepseek.WithModel(deepseek.ModelDeepSeekCoder),
	)
	if err != nil {
		log.Fatal(err)
	}

	codePrompt := "Write a Python function to calculate the factorial of a number using recursion."
	codeResponse, err := coderModel.Call(ctx, codePrompt,
		llms.WithMaxTokens(300),
		llms.WithTemperature(0.1), // Lower temperature for more deterministic code
	)
	if err != nil {
		log.Printf("Coder model error: %v", err)
	} else {
		fmt.Printf("Code Response: %s\n\n", codeResponse)
	}

	// Example 3: DeepSeek Reasoner for complex reasoning tasks
	fmt.Println("=== DeepSeek Reasoner Example ===")
	reasonerModel, err := deepseek.New(
		deepseek.WithModel(deepseek.ModelDeepSeekReasoner),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Using GenerateContent to access reasoning process
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a mathematical problem solver. Show your reasoning step by step."),
		llms.TextParts(llms.ChatMessageTypeHuman, "If a train travels at 60 mph for 2.5 hours, then slows down to 40 mph for the next 1.5 hours, what is the total distance traveled?"),
	}

	reasoningResponse, err := reasonerModel.GenerateContent(ctx, messages,
		llms.WithMaxTokens(400),
		llms.WithTemperature(0.3),
	)
	if err != nil {
		log.Printf("Reasoner model error: %v", err)
	} else if len(reasoningResponse.Choices) > 0 {
		choice := reasoningResponse.Choices[0]
		if choice.ReasoningContent != "" {
			fmt.Printf("Reasoning Process: %s\n", choice.ReasoningContent)
		}
		fmt.Printf("Final Answer: %s\n\n", choice.Content)
	}

	// Example 4: Streaming with reasoning
	fmt.Println("=== DeepSeek Reasoner Streaming Example ===")
	streamingMessages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "Solve this logic puzzle: Three friends Alice, Bob, and Charlie have different favorite colors (red, blue, green). Alice doesn't like red. Bob doesn't like blue. Charlie likes green. What color does each person like?"),
	}

	_, err = reasonerModel.GenerateContent(ctx, streamingMessages,
		llms.WithMaxTokens(500),
		llms.WithTemperature(0.2),
		llms.WithStreamingReasoningFunc(func(_ context.Context, reasoningChunk []byte, chunk []byte) error {
			if len(reasoningChunk) > 0 {
				fmt.Printf("[Reasoning] %s", string(reasoningChunk))
			}
			if len(chunk) > 0 {
				fmt.Printf("[Answer] %s", string(chunk))
			}
			return nil
		}),
	)
	if err != nil {
		log.Printf("Streaming error: %v", err)
	}
}
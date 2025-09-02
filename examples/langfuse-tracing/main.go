package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	// Initialize Langfuse handler
	langfuseHandler, err := setupLangfuseHandler()
	if err != nil {
		log.Printf("Warning: Failed to setup Langfuse handler: %v", err)
		log.Println("Continuing without Langfuse tracing...")
		runExamplesWithoutTracing()
		return
	}

	fmt.Printf("🔍 Langfuse tracing enabled - Trace ID: %s\n", langfuseHandler.GetTraceID())
	fmt.Printf("🌐 View traces at: https://cloud.langfuse.com\n\n")

	// Run examples with tracing
	runBasicLLMExample(langfuseHandler)
	runStreamingExample(langfuseHandler)
	runErrorExample(langfuseHandler)

	// Flush any pending traces
	if err := langfuseHandler.Flush(); err != nil {
		log.Printf("Error flushing traces: %v", err)
	}

	fmt.Printf("\n✅ All examples completed. Check your Langfuse dashboard for traces!\n")
}

func setupLangfuseHandler() (*callbacks.LangfuseHandler, error) {
	// Get credentials from environment
	publicKey := os.Getenv("LANGFUSE_PUBLIC_KEY")
	secretKey := os.Getenv("LANGFUSE_SECRET_KEY")
	baseURL := os.Getenv("LANGFUSE_HOST") // Optional: defaults to https://cloud.langfuse.com

	if publicKey == "" || secretKey == "" {
		return nil, fmt.Errorf("LANGFUSE_PUBLIC_KEY and LANGFUSE_SECRET_KEY environment variables are required")
	}

	opts := callbacks.LangfuseOptions{
		BaseURL:   baseURL,
		PublicKey: publicKey,
		SecretKey: secretKey,
		UserID:    "demo-user",
		SessionID: "langchaingo-demo",
		Metadata: map[string]interface{}{
			"example":     "langchaingo-integration",
			"environment": "demo",
			"sdk_version": "1.0.0",
		},
	}

	return callbacks.NewLangfuseHandler(opts)
}

func runBasicLLMExample(handler *callbacks.LangfuseHandler) {
	fmt.Println("🤖 Running Basic LLM Example...")

	// Set trace metadata for this example
	handler.SetTraceMetadata(map[string]interface{}{
		"example_type": "basic_llm",
		"model":        "gpt-3.5-turbo",
	})

	// Initialize OpenAI LLM with callback
	llm, err := openai.New(
		openai.WithCallback(handler),
	)
	if err != nil {
		log.Printf("Error creating LLM: %v", err)
		return
	}

	ctx := context.Background()
	prompt := "What are the key benefits of using LangChain for building LLM applications? Please be concise."

	fmt.Printf("  Prompt: %s\n", prompt)

	response, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		handler.HandleLLMError(ctx, err)
		log.Printf("Error generating response: %v", err)
		return
	}

	fmt.Printf("  Response: %s\n\n", response)
}

func runStreamingExample(handler *callbacks.LangfuseHandler) {
	fmt.Println("📡 Running Streaming Example...")

	handler.SetTraceMetadata(map[string]interface{}{
		"example_type": "streaming",
		"stream":       true,
	})

	// Initialize LLM with streaming
	llm, err := openai.New(
		openai.WithCallback(handler),
	)
	if err != nil {
		log.Printf("Error creating LLM: %v", err)
		return
	}

	ctx := context.Background()

	// Create a streaming request
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "Tell me a very short joke about programming."},
			},
		},
	}

	fmt.Println("  Prompt: Tell me a very short joke about programming.")
	fmt.Print("  Response: ")

	// Note: For simplicity in this demo, we're using GenerateContent instead of streaming
	// In a real streaming scenario, you'd handle the stream chunks individually
	response, err := llm.GenerateContent(ctx, messages)
	if err != nil {
		handler.HandleLLMError(ctx, err)
		log.Printf("Error generating streaming response: %v", err)
		return
	}

	if len(response.Choices) > 0 {
		fmt.Printf("%s\n\n", response.Choices[0].Content)
	}
}

func runErrorExample(handler *callbacks.LangfuseHandler) {
	fmt.Println("❌ Running Error Handling Example...")

	handler.SetTraceMetadata(map[string]interface{}{
		"example_type": "error_handling",
		"expect_error": true,
	})

	ctx := context.Background()

	// Simulate an error scenario
	handler.HandleLLMStart(ctx, []string{"This will simulate an error"})
	
	// Simulate an error (e.g., API timeout, invalid request, etc.)
	simulatedError := fmt.Errorf("simulated API timeout error")
	handler.HandleLLMError(ctx, simulatedError)

	fmt.Printf("  Simulated error: %s\n", simulatedError)
	fmt.Println("  ✓ Error was properly traced to Langfuse\n")
}

func runExamplesWithoutTracing() {
	fmt.Println("🔧 Running examples without Langfuse tracing...")
	fmt.Println("To enable Langfuse tracing, set the following environment variables:")
	fmt.Println("  - LANGFUSE_PUBLIC_KEY: Your Langfuse public key")
	fmt.Println("  - LANGFUSE_SECRET_KEY: Your Langfuse secret key")
	fmt.Println("  - LANGFUSE_HOST: (Optional) Your Langfuse host URL")
	fmt.Println()

	// Run a simple example without tracing
	llm, err := openai.New()
	if err != nil {
		log.Printf("Error creating LLM: %v", err)
		return
	}

	ctx := context.Background()
	prompt := "Hello! This is a test without tracing."

	response, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		log.Printf("Error generating response: %v", err)
		return
	}

	fmt.Printf("Response: %s\n", response)
}
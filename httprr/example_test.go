package httprr_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/tmc/langchaingo/httprr"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// Example_LLMTesting demonstrates how to use httprr for testing LLM interactions.
func Example_LLMTesting() {
	// This example shows how to record and replay LLM API calls for testing

	t := &testing.T{} // In real tests, this would be provided by the test function
	recordingsDir := filepath.Join("testdata", "example_recordings")

	// Create an LLM test helper with httprr
	helper := httprr.NewLLMTestHelper(t, recordingsDir)
	defer helper.Cleanup()

	// Create an OpenAI client that will use httprr for recording/replaying
	// Note: In a real test, you'd use a real API key or set HTTPRR_MODE=replay
	client, err := helper.NewOpenAIClientWithToken("test-api-key")
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		return
	}

	// Make an API call - this will be recorded on first run, replayed on subsequent runs
	ctx := context.Background()
	response, err := client.GenerateContent(ctx, []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "What is the capital of France?"},
			},
		},
	})

	if err != nil {
		fmt.Printf("Error generating content: %v\n", err)
		return
	}

	// In a real test, you would make assertions about the response
	if len(response.Choices) > 0 {
		fmt.Printf("Response: %s\n", response.Choices[0].Content)
	}

	// Verify that exactly one HTTP request was made
	helper.AssertRequestCount(1)

	// Output: (This would show actual output in a real test with recordings)
}

// Example_ReplayMode demonstrates using httprr in replay mode.
func Example_ReplayMode() {
	// Set environment variable to force replay mode
	os.Setenv("HTTPRR_MODE", "replay")
	defer os.Unsetenv("HTTPRR_MODE")

	// Now all HTTP requests will be replayed from existing recordings
	// This is useful for CI/CD where you don't want to make real API calls
	fmt.Println("Running in replay mode - no real API calls will be made")
}

// Example_CustomClient shows how to use httprr with custom client configurations.
func Example_CustomClient() {
	t := &testing.T{}
	recordingsDir := filepath.Join("testdata", "custom_recordings")

	helper := httprr.NewLLMTestHelper(t, recordingsDir)
	defer helper.Cleanup()

	// You can still use all the normal OpenAI options
	client, err := helper.NewOpenAIClientWithToken("test-api-key",
		openai.WithModel("gpt-4"),
		openai.WithBaseURL("https://custom-api.example.com"),
	)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// The client will use httprr for HTTP transport while respecting all other options
	fmt.Printf("Client created with custom options and httprr transport\n")

	// Output: Client created with custom options and httprr transport
}

// Example_MultipleProviders shows testing multiple LLM providers.
func Example_MultipleProviders() {
	t := &testing.T{}

	// Test OpenAI
	openaiHelper := httprr.NewLLMTestHelper(t, "testdata/openai")
	defer openaiHelper.Cleanup()

	// Test Anthropic  
	anthropicHelper := httprr.NewLLMTestHelper(t, "testdata/anthropic")
	defer anthropicHelper.Cleanup()

	// Test Ollama
	ollamaHelper := httprr.NewLLMTestHelper(t, "testdata/ollama")
	defer ollamaHelper.Cleanup()

	fmt.Println("Created helpers for testing multiple LLM providers")

	// Output: Created helpers for testing multiple LLM providers
}

// Example_AdvancedUsage shows advanced httprr features.
func Example_AdvancedUsage() {
	t := &testing.T{}
	helper := httprr.NewLLMTestHelper(t, "testdata/advanced")
	defer helper.Cleanup()

	// Create client
	client, _ := helper.NewOpenAIClientWithToken("test-key")

	// In a real test, you would make requests here...
	
	// Then use advanced assertion methods
	urls := helper.GetRequestURLs()
	fmt.Printf("Requested URLs: %v\n", urls)

	// Find specific responses
	resp, body, err := helper.FindResponse("completions")
	if err == nil {
		fmt.Printf("Found completion response with status: %d\n", resp.StatusCode)
		fmt.Printf("Response body length: %d bytes\n", len(body))
	}

	// Dump all recordings for debugging
	helper.DumpRecordings()

	// Save recordings to a specific directory
	err = helper.SaveRecordingsToDir("custom_output")
	if err != nil {
		fmt.Printf("Error saving recordings: %v\n", err)
	}

	fmt.Println("Advanced httprr features demonstrated")

	// Output: Advanced httprr features demonstrated
}
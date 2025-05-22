package httprr_test

import (
	"context"
	"os"
	"testing"

	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// ExampleTestWithRecording demonstrates how to use httprr for testing LLM interactions.
func ExampleTestWithRecording(t *testing.T) {
	// Skip if no API key is available
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	helper := httprr.NewLLMTestHelper("openai_basic_test")
	defer helper.Reset(false) // Keep recordings for inspection

	// Create OpenAI client with recording
	client, err := helper.TestOpenAI(openai.WithModel("gpt-3.5-turbo"))
	if err != nil {
		t.Fatalf("Failed to create OpenAI client: %v", err)
	}

	ctx := context.Background()
	
	// Test basic completion - this will be recorded
	response, err := helper.CallWithRecording(ctx, client, "Hello, world!")
	if err != nil {
		t.Fatalf("Failed to call OpenAI: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response")
	}

	// Test chat completion - this will also be recorded
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "What is the capital of France?"},
			},
		},
	}

	chatResponse, err := helper.GenerateContentWithRecording(ctx, client, messages)
	if err != nil {
		t.Fatalf("Failed to generate content: %v", err)
	}

	if len(chatResponse.Choices) == 0 {
		t.Error("Expected at least one choice in response")
	}
}

// ExampleCustomRecordingMode shows how to use custom recording modes.
func ExampleCustomRecordingMode(t *testing.T) {
	helper := httprr.NewLLMTestHelper(
		"custom_mode_test",
		httprr.WithMode(httprr.ModeReplay), // Force replay mode
		httprr.WithRecordingsDir("custom_recordings"),
	)

	// This client will only replay existing recordings
	httpClient := helper.HTTPClient()
	
	// Use the client with any HTTP-based service
	_ = httpClient // Use as needed
}

// ExampleManualRecorder demonstrates manual use of the recorder.
func ExampleManualRecorder(t *testing.T) {
	recorder := httprr.New("testdata/manual_recordings", httprr.ModeRecordOnce)
	
	// Get an HTTP client with recording
	client := recorder.Client()
	
	// Use with OpenAI
	openaiClient, err := openai.New(
		openai.WithHTTPClient(client),
		openai.WithModel("gpt-3.5-turbo"),
	)
	if err != nil {
		t.Fatalf("Failed to create OpenAI client: %v", err)
	}

	ctx := context.Background()
	if os.Getenv("OPENAI_API_KEY") != "" {
		_, err = openaiClient.Call(ctx, "Test prompt")
		if err != nil {
			t.Logf("Call failed (expected in replay mode): %v", err)
		}
	}
	
	// Reset recordings if needed
	err = recorder.Reset(false)
	if err != nil {
		t.Logf("Reset failed: %v", err)
	}
}
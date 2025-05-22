package httprr

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

func TestLLMTestHelper_OpenAI(t *testing.T) {
	// Skip if no OpenAI API key is available
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping OpenAI test")
	}

	// Create a recordings directory
	recordingsDir := filepath.Join("testdata", "openai_recordings")
	
	// Create the LLM test helper
	helper := NewLLMTestHelper(t, recordingsDir)
	defer helper.Cleanup()

	// Create an OpenAI client with httprr
	client, err := helper.NewOpenAIClientWithToken(apiKey)
	if err != nil {
		t.Fatalf("Failed to create OpenAI client: %v", err)
	}

	// Make a simple completion request
	ctx := context.Background()
	response, err := client.GenerateContent(ctx, []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "Say hello"},
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed to generate content: %v", err)
	}

	if len(response.Choices) == 0 {
		t.Fatalf("No response choices returned")
	}

	// Verify we recorded the HTTP request
	helper.AssertRequestCount(1)
	helper.AssertURLCalled("api.openai.com")
	
	t.Logf("Response: %s", response.Choices[0].Content)
}

func TestLLMTestHelper_Anthropic(t *testing.T) {
	// Skip if no Anthropic API key is available
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping Anthropic test")
	}

	// Create a recordings directory
	recordingsDir := filepath.Join("testdata", "anthropic_recordings")
	
	// Create the LLM test helper
	helper := NewLLMTestHelper(t, recordingsDir)
	defer helper.Cleanup()

	// Create an Anthropic client with httprr
	client, err := helper.NewAnthropicClient(apiKey)
	if err != nil {
		t.Fatalf("Failed to create Anthropic client: %v", err)
	}

	// Make a simple completion request
	ctx := context.Background()
	response, err := client.GenerateContent(ctx, []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "Say hello"},
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed to generate content: %v", err)
	}

	if len(response.Choices) == 0 {
		t.Fatalf("No response choices returned")
	}

	// Verify we recorded the HTTP request
	helper.AssertRequestCount(1)
	helper.AssertURLCalled("api.anthropic.com")
	
	t.Logf("Response: %s", response.Choices[0].Content)
}

func TestLLMTestHelper_Ollama(t *testing.T) {
	// Skip if Ollama is not available
	// Note: This test assumes Ollama is running locally
	// In a real test, you might want to check if Ollama is available first
	
	// Create a recordings directory
	recordingsDir := filepath.Join("testdata", "ollama_recordings")
	
	// Create the LLM test helper
	helper := NewLLMTestHelper(t, recordingsDir)
	defer helper.Cleanup()

	// Create an Ollama client with httprr
	client, err := helper.NewOllamaClient()
	if err != nil {
		t.Fatalf("Failed to create Ollama client: %v", err)
	}

	// Make a simple completion request
	ctx := context.Background()
	response, err := client.GenerateContent(ctx, []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "Say hello"},
			},
		},
	}, llms.WithModel("llama2"))
	
	// Note: This might fail if Ollama is not running or model is not available
	// In that case, we just log and continue
	if err != nil {
		t.Logf("Ollama request failed (this is expected if Ollama is not running): %v", err)
		return
	}

	if len(response.Choices) == 0 {
		t.Fatalf("No response choices returned")
	}

	// Verify we recorded the HTTP request
	helper.AssertRequestCount(1)
	
	t.Logf("Response: %s", response.Choices[0].Content)
}

// Example showing how to create test data directory structure
func TestLLMTestHelper_DirectorySetup(t *testing.T) {
	// This test demonstrates how to set up a proper directory structure
	// for your LLM test recordings
	
	testDataDir := filepath.Join("testdata", "llm_recordings")
	
	// Create subdirectories for different LLM providers
	dirs := []string{
		filepath.Join(testDataDir, "openai"),
		filepath.Join(testDataDir, "anthropic"),
		filepath.Join(testDataDir, "ollama"),
		filepath.Join(testDataDir, "huggingface"),
	}
	
	for _, dir := range dirs {
		helper := NewLLMTestHelper(t, dir)
		defer helper.Cleanup()
		
		// Verify the helper is properly configured
		if helper.RecordingsDir != dir {
			t.Errorf("Expected recordings dir %s, got %s", dir, helper.RecordingsDir)
		}
	}
}
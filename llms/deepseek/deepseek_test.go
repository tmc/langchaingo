package deepseek

import (
	"context"
	"os"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

func TestNew(t *testing.T) {
	if os.Getenv("DEEPSEEK_API_KEY") == "" {
		t.Skip("DEEPSEEK_API_KEY not set")
	}

	client, err := New()
	if err != nil {
		t.Fatalf("Failed to create DeepSeek client: %v", err)
	}

	if client == nil {
		t.Fatal("Expected non-nil client")
	}
}

func TestNewWithOptions(t *testing.T) {
	client, err := New(
		WithAPIKey("test-key"),
		WithModel(ModelDeepSeekCoder),
		WithBaseURL("https://custom-api.example.com"),
	)
	if err != nil {
		t.Fatalf("Failed to create DeepSeek client with options: %v", err)
	}

	if client == nil {
		t.Fatal("Expected non-nil client")
	}
}

func TestCall(t *testing.T) {
	if os.Getenv("DEEPSEEK_API_KEY") == "" {
		t.Skip("DEEPSEEK_API_KEY not set")
	}

	client, err := New(WithModel(ModelDeepSeekChat))
	if err != nil {
		t.Fatalf("Failed to create DeepSeek client: %v", err)
	}

	ctx := context.Background()
	response, err := client.Call(ctx, "Hello, how are you?", llms.WithMaxTokens(10))
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	if response == "" {
		t.Fatal("Expected non-empty response")
	}

	t.Logf("Response: %s", response)
}

func TestGenerateContent(t *testing.T) {
	if os.Getenv("DEEPSEEK_API_KEY") == "" {
		t.Skip("DEEPSEEK_API_KEY not set")
	}

	client, err := New(WithModel(ModelDeepSeekReasoner))
	if err != nil {
		t.Fatalf("Failed to create DeepSeek client: %v", err)
	}

	ctx := context.Background()
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "What is 2+2?"),
	}

	response, err := client.GenerateContent(ctx, messages, llms.WithMaxTokens(50))
	if err != nil {
		t.Fatalf("GenerateContent failed: %v", err)
	}

	if len(response.Choices) == 0 {
		t.Fatal("Expected at least one choice in response")
	}

	if response.Choices[0].Content == "" {
		t.Fatal("Expected non-empty content in response")
	}

	t.Logf("Response: %s", response.Choices[0].Content)
	if response.Choices[0].ReasoningContent != "" {
		t.Logf("Reasoning: %s", response.Choices[0].ReasoningContent)
	}
}

func TestModels(t *testing.T) {
	expectedModels := []string{
		ModelDeepSeekChat,
		ModelDeepSeekCoder,
		ModelDeepSeekReasoner,
		ModelDeepSeekV3,
	}

	for _, model := range expectedModels {
		if model == "" {
			t.Errorf("Model constant should not be empty")
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	config := defaultConfig()

	if config.BaseURL != DefaultBaseURL {
		t.Errorf("Expected BaseURL %s, got %s", DefaultBaseURL, config.BaseURL)
	}

	if config.Model != ModelDeepSeekChat {
		t.Errorf("Expected Model %s, got %s", ModelDeepSeekChat, config.Model)
	}
}
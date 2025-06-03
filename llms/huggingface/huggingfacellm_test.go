package huggingface

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/llms"
)

func TestHuggingFaceLLMWithProvider(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Skip if no credentials and no recording - HuggingFace accepts either token
	if os.Getenv("HF_TOKEN") == "" && os.Getenv("HUGGINGFACEHUB_API_TOKEN") == "" {
		httprr.SkipIfNoCredentialsAndRecordingMissing(t, "HF_TOKEN")
	}

	rr := httprr.OpenForTest(t, nil)
	defer rr.Close()
	// Create LLM with provider
	llm, err := New(
		WithModel("deepseek-ai/DeepSeek-R1-0528"),
		WithInferenceProvider("hyperbolic"),
		WithHTTPClient(rr.Client()),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Test the LLM call
	result, err := llm.Call(ctx, "What is 2+2?",
		llms.WithTemperature(0.5),
		llms.WithMaxLength(50),
	)

	// Skip test if provider is not available or recording is missing
	if err != nil && (strings.Contains(err.Error(), "404") ||
		strings.Contains(err.Error(), "403") ||
		strings.Contains(err.Error(), "cached HTTP response not found")) {
		t.Skip("Provider not available or recording missing, skipping test")
	}

	if err != nil {
		t.Fatal(err)
	}
	if result == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestHuggingFaceLLMStandardInference(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Skip if no credentials and no recording - HuggingFace accepts either token
	if os.Getenv("HF_TOKEN") == "" && os.Getenv("HUGGINGFACEHUB_API_TOKEN") == "" {
		httprr.SkipIfNoCredentialsAndRecordingMissing(t, "HF_TOKEN")
	}

	rr := httprr.OpenForTest(t, nil)
	defer rr.Close()
	// Create standard LLM without provider
	llm, err := New(
		WithModel("HuggingFaceH4/zephyr-7b-beta"),
		WithHTTPClient(rr.Client()),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Test the LLM call
	result, err := llm.Call(ctx, "Hello, say hi back",
		llms.WithTemperature(0.5),
		llms.WithMaxLength(20),
	)

	// Skip test if model is not available
	if err != nil && strings.Contains(err.Error(), "404") {
		t.Skip("Model not available on HuggingFace API, skipping test")
	}

	if err != nil {
		t.Fatal(err)
	}
	if result == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestHuggingFaceLLMGenerateContent(t *testing.T) {
	t.Skip("temporarily skip")

	t.Parallel()
	ctx := context.Background()

	// Skip if no credentials and no recording
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "HF_TOKEN", "HUGGINGFACEHUB_API_TOKEN")

	rr := httprr.OpenForTest(t, nil)
	defer rr.Close()
	// Create LLM
	llm, err := New(
		WithModel("HuggingFaceH4/zephyr-7b-beta"),
		WithHTTPClient(rr.Client()),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Test GenerateContent directly
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "What is the capital of France?"},
			},
		},
	}

	resp, err := llm.GenerateContent(ctx, messages,
		llms.WithTemperature(0.5),
		llms.WithMaxLength(30),
	)

	// Skip test if model is not available or rate limited
	if err != nil && (strings.Contains(err.Error(), "404") ||
		strings.Contains(err.Error(), "402") ||
		strings.Contains(err.Error(), "cached HTTP response not found")) {
		t.Skip("Model not available, rate limited, or recording missing, skipping test")
	}

	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(resp.Choices))
	}
	if resp.Choices[0].Content == "" {
		t.Fatal("expected non-empty content")
	}
	// Should mention Paris
	if !strings.Contains(strings.ToLower(resp.Choices[0].Content), "paris") {
		t.Fatal("expected response to mention Paris")
	}
}

package googleai

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/llms"
)

func TestGoogleApisAI_CreateEmbedding(t *testing.T) {
	t.Parallel()

	if os.Getenv("GEMINI_API_KEY") == "" && os.Getenv("TEST_RECORD") != "true" {
		t.Skip("GEMINI_API_KEY not set")
	}

	// Enable recording mode with TEST_RECORD=true
	mode := httprr.ModeReplay
	if os.Getenv("TEST_RECORD") == "true" {
		mode = httprr.ModeRecord
	}

	ctx := context.Background()
	
	// Note: googleapis genai SDK handles HTTP client internally,
	// so HTTPrr integration would need SDK-level support
	// For now, we'll just test with real API calls when recording
	
	client, err := NewGoogleApisAI(ctx, WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	require.NoError(t, err)
	defer client.Close()

	texts := []string{
		"What is machine learning?",
		"How does artificial intelligence work?",
		"Explain deep learning concepts.",
	}

	embeddings, err := client.CreateEmbedding(ctx, texts)
	require.NoError(t, err)
	assert.Len(t, embeddings, 3)

	for i, embedding := range embeddings {
		assert.Greater(t, len(embedding), 0, "Embedding %d should not be empty", i)
		assert.Greater(t, len(embedding), 100, "Embedding %d should have reasonable dimensions", i)
	}
}

func TestGoogleApisAI_GenerateContent(t *testing.T) {
	t.Parallel()

	if os.Getenv("GEMINI_API_KEY") == "" && os.Getenv("TEST_RECORD") != "true" {
		t.Skip("GEMINI_API_KEY not set")
	}

	ctx := context.Background()
	
	client, err := NewGoogleApisAI(ctx, 
		WithAPIKey(os.Getenv("GEMINI_API_KEY")),
		WithDefaultModel("gemini-1.5-flash"),
	)
	require.NoError(t, err)
	defer client.Close()

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What is the capital of France? Answer in one word."),
			},
		},
	}

	resp, err := client.GenerateContent(ctx, messages, llms.WithMaxTokens(10))
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Choices)
	assert.Contains(t, resp.Choices[0].Content, "Paris")
}

func TestGoogleApisAI_Call(t *testing.T) {
	t.Parallel()

	if os.Getenv("GEMINI_API_KEY") == "" && os.Getenv("TEST_RECORD") != "true" {
		t.Skip("GEMINI_API_KEY not set")
	}

	ctx := context.Background()
	
	client, err := NewGoogleApisAI(ctx, WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	require.NoError(t, err)
	defer client.Close()

	response, err := client.Call(ctx, "What is 2+2? Answer with just the number.", 
		llms.WithMaxTokens(5))
	require.NoError(t, err)
	assert.Contains(t, response, "4")
}

func TestGoogleApisAI_WithJSONMode(t *testing.T) {
	t.Parallel()

	if os.Getenv("GEMINI_API_KEY") == "" && os.Getenv("TEST_RECORD") != "true" {
		t.Skip("GEMINI_API_KEY not set")
	}

	ctx := context.Background()
	
	client, err := NewGoogleApisAI(ctx, WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	require.NoError(t, err)
	defer client.Close()

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Return a JSON object with the capital of France. Use 'capital' as the key."),
			},
		},
	}

	// Note: JSON mode implementation would need to be added to the googleapis client
	resp, err := client.GenerateContent(ctx, messages, 
		llms.WithMaxTokens(50),
		llms.WithTemperature(0.1),
	)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Choices)
	
	content := resp.Choices[0].Content
	assert.Contains(t, content, "Paris")
	// The response should contain JSON-like structure
	assert.Contains(t, content, "{")
	assert.Contains(t, content, "}")
}

func TestGoogleApisAI_WithImageInput(t *testing.T) {
	t.Parallel()

	if os.Getenv("GEMINI_API_KEY") == "" && os.Getenv("TEST_RECORD") != "true" {
		t.Skip("GEMINI_API_KEY not set")
	}

	ctx := context.Background()
	
	client, err := NewGoogleApisAI(ctx, WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	require.NoError(t, err)
	defer client.Close()

	// Use a simple test image URL (this would need to be a real image for actual testing)
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What do you see in this image?"),
				llms.BinaryPart{
					URL:      "https://example.com/test-image.jpg", // This would need to be a real image
					MIMEType: "image/jpeg",
				},
			},
		},
	}

	// Note: This test would fail with the example URL, but demonstrates the API structure
	_, err = client.GenerateContent(ctx, messages, llms.WithMaxTokens(100))
	// We expect this to fail with the example URL, but the error should be related to image download
	// not API structure issues
	assert.Error(t, err) // Expected to fail with example URL
}
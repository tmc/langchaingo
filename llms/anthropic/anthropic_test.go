package anthropic

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/llms"
)

func newTestClient(t *testing.T, opts ...Option) llms.Model {
	t.Helper()
	
	// Check if we need an API key (only for recording mode)
	if httprr.GetTestMode() == httprr.TestModeRecord {
		if anthropicKey := os.Getenv("ANTHROPIC_API_KEY"); anthropicKey == "" {
			t.Skip("ANTHROPIC_API_KEY not set")
			return nil
		}
	} else {
		// For replay mode, provide a fake API key if none is set
		if os.Getenv("ANTHROPIC_API_KEY") == "" {
			opts = append([]Option{WithToken("fake-api-key-for-testing")}, opts...)
		}
	}
	
	// Create HTTP client with httprr support
	httpClient := httprr.TestClient(t, "anthropic_"+t.Name())
	
	// Prepend HTTP client option
	opts = append([]Option{WithHTTPClient(httpClient)}, opts...)

	llm, err := New(opts...)
	require.NoError(t, err)
	return llm
}

func TestAnthropicCall(t *testing.T) {
	t.Parallel()
	llm := newTestClient(t)

	resp, err := llm.Call(context.Background(), "What is the capital of France?")
	require.NoError(t, err)
	assert.NotEmpty(t, resp)
	assert.Contains(t, strings.ToLower(resp), "paris")
}

func TestAnthropicGenerateContent(t *testing.T) {
	t.Parallel()
	llm := newTestClient(t)

	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Tell me a short fact about artificial intelligence."}},
		},
	}

	resp, err := llm.GenerateContent(context.Background(), content)
	require.NoError(t, err)
	
	assert.NotEmpty(t, resp.Choices)
	c1 := resp.Choices[0]
	assert.NotEmpty(t, c1.Content)
	assert.Contains(t, strings.ToLower(c1.Content), "artificial")
}

func TestAnthropicWithStreaming(t *testing.T) {
	t.Parallel()
	llm := newTestClient(t)

	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Write a short greeting."}},
		},
	}

	var sb strings.Builder
	resp, err := llm.GenerateContent(context.Background(), content,
		llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			sb.Write(chunk)
			return nil
		}))

	require.NoError(t, err)
	
	assert.NotEmpty(t, resp.Choices)
	c1 := resp.Choices[0]
	assert.NotEmpty(t, c1.Content)
	assert.NotEmpty(t, sb.String())
	// Both the complete response and streamed content should contain greeting-like text
	assert.True(t, strings.Contains(strings.ToLower(c1.Content), "hello") || 
		strings.Contains(strings.ToLower(c1.Content), "greet") ||
		strings.Contains(strings.ToLower(c1.Content), "hi"))
}
package maritaca

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

func newTestClient(t *testing.T, opts ...Option) *LLM {
	t.Helper()
	
	// Check if we need an API key (only for recording mode)
	if httprr.GetTestMode() == httprr.TestModeRecord {
		if token := os.Getenv("MARITACA_KEY"); token == "" {
			t.Skip("MARITACA_KEY not set")
			return nil
		} else {
			opts = append([]Option{WithToken(token), WithModel("sabia-2-medium")}, opts...)
		}
	} else {
		// For replay mode, provide fake credentials
		opts = append([]Option{WithToken("fake-token"), WithModel("sabia-2-medium")}, opts...)
	}
	
	// Create HTTP client with httprr support
	httpClient := httprr.TestClient(t, "maritaca_"+t.Name())
	opts = append([]Option{WithHTTPClient(httpClient)}, opts...)

	c, err := New(opts...)
	require.NoError(t, err)
	return c
}

func TestGenerateContent(t *testing.T) {
	t.Parallel()
	llm := newTestClient(t)

	parts := []llms.ContentPart{
		llms.TextContent{Text: "How many feet are in a nautical mile?"},
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	rsp, err := llm.GenerateContent(context.Background(), content)

	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "feet", strings.ToLower(c1.Content))
}

func TestWithStreaming(t *testing.T) {
	t.Parallel()
	llm := newTestClient(t)

	parts := []llms.ContentPart{
		llms.TextContent{Text: "How many feet are in a nautical mile?"},
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	var sb strings.Builder
	rsp, err := llm.GenerateContent(context.Background(), content,
		llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			sb.Write(chunk)
			return nil
		}))
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "feet", strings.ToLower(c1.Content))
	assert.Regexp(t, "feet", strings.ToLower(sb.String()))
}

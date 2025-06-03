package llamafile

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

// isLlamafileAvailable checks if the llamafile server is available
func isLlamafileAvailable() bool {
	// Check if CI environment variable is set - skip if in CI
	if os.Getenv("CI") != "" {
		return false
	}

	// Check if LLAMAFILE_HOST is set
	host := os.Getenv("LLAMAFILE_HOST")
	if host == "" {
		host = "http://127.0.0.1:8080"
	}

	// Try to connect to llamafile server
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(host + "/health")
	if err != nil {
		// Try without /health endpoint
		resp, err = client.Get(host)
		if err != nil {
			return false
		}
	}
	defer resp.Body.Close()

	return resp.StatusCode < 500
}

func newTestClient(t *testing.T) *LLM {
	t.Helper()
	options := []Option{
		WithEmbeddingSize(2048),
		WithTemperature(0.8),
	}
	c, err := New(options...)
	require.NoError(t, err)
	return c
}

func TestGenerateContent(t *testing.T) {
	if !isLlamafileAvailable() {
		t.Skip("llamafile is not available")
	}
	t.Parallel()
	ctx := context.Background()
	llm := newTestClient(t)

	parts := []llms.ContentPart{
		llms.TextContent{Text: "Brazil is a country? the answer should just be yes or no"},
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	rsp, err := llm.GenerateContent(ctx, content)
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "yes", strings.ToLower(c1.Content))
}

func TestWithStreaming(t *testing.T) {
	if !isLlamafileAvailable() {
		t.Skip("llamafile is not available")
	}
	t.Parallel()
	ctx := context.Background()
	llm := newTestClient(t)

	parts := []llms.ContentPart{
		llms.TextContent{Text: "Brazil is a country? answer yes or no"},
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	var sb strings.Builder
	rsp, err := llm.GenerateContent(ctx, content,
		llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			sb.Write(chunk)
			return nil
		}))
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "yes", strings.ToLower(c1.Content))
	assert.Regexp(t, "yes", strings.ToLower(sb.String()))
}

func TestCreateEmbedding(t *testing.T) {
	t.Parallel()
	if !isLlamafileAvailable() {
		t.Skip("llamafile is not available")
	}
	ctx := context.Background()
	llm := newTestClient(t)

	embeddings, err := llm.CreateEmbedding(ctx, []string{"hello", "world"})
	require.NoError(t, err)
	assert.Len(t, embeddings, 2)
}

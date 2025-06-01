package ollama

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

func newTestClient(t *testing.T, opts ...Option) *LLM {
	t.Helper()
	var ollamaModel string
	if ollamaModel = os.Getenv("OLLAMA_TEST_MODEL"); ollamaModel == "" {
		t.Skip("OLLAMA_TEST_MODEL not set")
		return nil
	}

	opts = append([]Option{WithModel(ollamaModel)}, opts...)

	c, err := New(opts...)
	require.NoError(t, err)
	return c
}

func TestGenerateContent(t *testing.T) {
	ctx := context.Background()
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

	rsp, err := llm.GenerateContent(ctx, content)
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "feet", strings.ToLower(c1.Content))
}

func TestWithFormat(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
	llm := newTestClient(t, WithFormat("json"))

	parts := []llms.ContentPart{
		llms.TextContent{Text: "How many feet are in a nautical mile?"},
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
	assert.Regexp(t, "feet", strings.ToLower(c1.Content))

	// check whether we got *any* kind of JSON object.
	var result map[string]any
	err = json.Unmarshal([]byte(c1.Content), &result)
	require.NoError(t, err)
}

func TestWithStreaming(t *testing.T) {
	ctx := context.Background()
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
	rsp, err := llm.GenerateContent(ctx, content,
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

func TestWithKeepAlive(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
	llm := newTestClient(t, WithKeepAlive("1m"))

	parts := []llms.ContentPart{
		llms.TextContent{Text: "How many feet are in a nautical mile?"},
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	resp, err := llm.GenerateContent(ctx, content)
	require.NoError(t, err)

	assert.NotEmpty(t, resp.Choices)
	c1 := resp.Choices[0]
	assert.Regexp(t, "feet", strings.ToLower(c1.Content))

	vector, err := llm.CreateEmbedding(ctx, []string{"test embedding with keep_alive"})
	require.NoError(t, err)
	assert.NotEmpty(t, vector)
}

func TestWithPullModel(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
	// This test uses a small model to minimize download time
	// Skip if not explicitly enabled via environment variable
	if os.Getenv("OLLAMA_TEST_PULL") == "" {
		t.Skip("OLLAMA_TEST_PULL not set, skipping pull test")
	}

	// Use a small model for testing
	llm, err := New(WithModel("gemma:2b"), WithPullModel())
	require.NoError(t, err)

	parts := []llms.ContentPart{
		llms.TextContent{Text: "Say hello"},
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	// The model should be pulled automatically before generating content
	resp, err := llm.GenerateContent(ctx, content)
	require.NoError(t, err)

	assert.NotEmpty(t, resp.Choices)
	c1 := resp.Choices[0]
	assert.NotEmpty(t, c1.Content)
}

func TestWithPullTimeout(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
	// This test verifies timeout functionality
	// Skip if not explicitly enabled via environment variable
	if os.Getenv("OLLAMA_TEST_PULL") == "" {
		t.Skip("OLLAMA_TEST_PULL not set, skipping pull timeout test")
	}

	// Use a very short timeout that should fail for any real model pull
	llm, err := New(
		WithModel("llama2:70b"), // Large model that would take time to download
		WithPullModel(),
		WithPullTimeout(1*time.Millisecond), // Extremely short timeout
	)
	require.NoError(t, err)

	parts := []llms.ContentPart{
		llms.TextContent{Text: "Say hello"},
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	// This should fail with a timeout error
	_, err = llm.GenerateContent(ctx, content)
	require.Error(t, err)
	// The error should contain "context deadline exceeded" or similar
	assert.Contains(t, err.Error(), "context")
}

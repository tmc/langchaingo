package ollama

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

func newTestClient(t *testing.T) *LLM {
	t.Helper()
	var ollamaModel string
	if ollamaModel = os.Getenv("OLLAMA_TEST_MODEL"); ollamaModel == "" {
		t.Skip("OLLAMA_TEST_MODEL not set")
		return nil
	}

	c, err := New(WithModel(ollamaModel))
	require.NoError(t, err)
	return c
}

func TestGenerateContent(t *testing.T) {
	t.Parallel()
	llm := newTestClient(t)

	parts := []schema.ContentPart{
		schema.TextContent{Text: "How many feet are in a nautical mile?"},
	}
	content := []schema.MessageContent{
		{
			Role:  schema.ChatMessageTypeHuman,
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

	parts := []schema.ContentPart{
		schema.TextContent{Text: "How many feet are in a nautical mile?"},
	}
	content := []schema.MessageContent{
		{
			Role:  schema.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	var sb strings.Builder
	rsp, err := llm.GenerateContent(context.Background(), content,
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			sb.Write(chunk)
			return nil
		}))
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "feet", strings.ToLower(c1.Content))
	assert.Regexp(t, "feet", strings.ToLower(sb.String()))
}

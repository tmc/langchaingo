package llamafile

import (
	"context"

	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

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
	t.Skip("llamafile is not available")
	t.Parallel()
	llm := newTestClient(t)

	parts := []llms.ContentPart{
		llms.TextContent{Text: "Brazil is a country? the answer should just be yes or no"},
	}
	content := []llms.MessageContent{
		{
			Role:  schema.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	rsp, err := llm.GenerateContent(context.Background(), content)
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "yes", strings.ToLower(c1.Content))
}

func TestWithStreaming(t *testing.T) {
	t.Skip("llamafile is not available")
	t.Parallel()
	llm := newTestClient(t)

	parts := []llms.ContentPart{
		llms.TextContent{Text: "Brazil is a country? answer yes or no"},
	}
	content := []llms.MessageContent{
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
	assert.Regexp(t, "yes", strings.ToLower(c1.Content))
	assert.Regexp(t, "yes", strings.ToLower(sb.String()))
}

func TestCreateEmbedding(t *testing.T) {
	t.Parallel()
	t.Skip("llamafile is not available")
	llm := newTestClient(t)

	embeddings, err := llm.CreateEmbedding(context.Background(), []string{"hello", "world"})
	require.NoError(t, err)
	assert.Len(t, embeddings, 2)
}

// teste example complet
func TestLLM(t *testing.T) {
	t.Parallel()

	llm, err := New()

	if err != nil {
		t.Fatal(err)
	}

	parts := []llms.ContentPart{
		llms.TextContent{Text: "Brazil is a country? answer yes or no"},
	}
	content := []llms.MessageContent{
		{
			Role:  schema.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	_, err = embeddings.NewEmbedder(llm)
	if err != nil {
		t.Fatal(err)
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
	assert.Regexp(t, "yes", strings.ToLower(c1.Content))
	assert.Regexp(t, "yes", strings.ToLower(sb.String()))
}

package openai

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tmc/langchaingo/llms/openai"
)

func TestOpenaiEmbeddings(t *testing.T) {
	t.Parallel()

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	e, err := NewOpenAI()
	require.NoError(t, err)

	_, err = e.EmbedQuery(context.Background(), "Hello world!")
	require.NoError(t, err)

	embeddings, err := e.EmbedDocuments(context.Background(), []string{"Hello world", "The world is ending", "good bye"})
	require.NoError(t, err)
	assert.Len(t, embeddings, 3)
}

func TestOpenaiEmbeddingsWithOptions(t *testing.T) {
	t.Parallel()

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	client, err := openai.New()
	require.NoError(t, err)

	e, err := NewOpenAI(WithClient(*client), WithBatchSize(1), WithStripNewLines(false))
	require.NoError(t, err)

	_, err = e.EmbedQuery(context.Background(), "Hello world!")
	require.NoError(t, err)

	embeddings, err := e.EmbedDocuments(context.Background(), []string{"Hello world"})
	require.NoError(t, err)
	assert.Len(t, embeddings, 1)
}

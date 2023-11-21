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

func TestOpenaiEmbeddingsQueryVsDocuments(t *testing.T) {
	// Verifies that we get the same embedding for the same string, regardless
	// of which method we call.
	t.Parallel()

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	e, err := NewOpenAI()
	require.NoError(t, err)

	text := "hi there"

	eq, err := e.EmbedQuery(context.Background(), text)
	require.NoError(t, err)

	eb, err := e.EmbedDocuments(context.Background(), []string{text})
	require.NoError(t, err)

	// Using strict equality should be OK here because we expect the same values
	// for the same string, deterministically.
	assert.Equal(t, eq, eb[0])
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

func TestOpenaiEmbeddingsWithAzureAPI(t *testing.T) {
	t.Parallel()

	// Azure OpenAI Key is used as OPENAI_API_KEY
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	// Azure OpenAI URL is used as OPENAI_BASE_URL
	if openaiBase := os.Getenv("OPENAI_BASE_URL"); openaiBase == "" {
		t.Skip("OPENAI_BASE_URL not set")
	}

	client, err := openai.New(
		openai.WithAPIType(openai.APITypeAzure),
		// Azure deployment that uses desired model the name depends on what we define in the Azure deployment section
		openai.WithModel("model"),
		// Azure deployment that uses embeddings model, the name depends on what we define in the Azure deployment section
		openai.WithEmbeddingModel("model-embedding"),
	)
	require.NoError(t, err)

	e, err := NewOpenAI(WithClient(*client), WithBatchSize(1), WithStripNewLines(false))
	require.NoError(t, err)

	_, err = e.EmbedQuery(context.Background(), "Hello world!")
	require.NoError(t, err)

	embeddings, err := e.EmbedDocuments(context.Background(), []string{"Hello world"})
	require.NoError(t, err)
	assert.Len(t, embeddings, 1)
}

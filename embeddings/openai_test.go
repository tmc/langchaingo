package embeddings

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/llms/openai"
)

func newOpenAIEmbedder(t *testing.T, opts ...Option) *EmbedderImpl {
	t.Helper()
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	t.Cleanup(func() { rr.Close() })
	llm, err := openai.New(openai.WithHTTPClient(rr.Client()))
	require.NoError(t, err)

	embedder, err := NewEmbedder(llm, opts...)
	require.NoError(t, err)

	return embedder
}

func TestOpenaiEmbeddings(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	e := newOpenAIEmbedder(t)
	_, err := e.EmbedQuery(ctx, "Hello world!")
	require.NoError(t, err)

	embeddings, err := e.EmbedDocuments(ctx, []string{"Hello world", "The world is ending", "good bye"})
	require.NoError(t, err)
	assert.Len(t, embeddings, 3)
}

func TestOpenaiEmbeddingsQueryVsDocuments(t *testing.T) {
	ctx := context.Background()
	// Verifies that we get the same embedding for the same string, regardless
	// of which method we call.
	t.Parallel()

	e := newOpenAIEmbedder(t)
	text := "hi there"
	eq, err := e.EmbedQuery(ctx, text)
	require.NoError(t, err)

	eb, err := e.EmbedDocuments(ctx, []string{text})
	require.NoError(t, err)

	// Using strict equality should be OK here because we expect the same values
	// for the same string, deterministically.
	assert.Equal(t, eq, eb[0])
}

func TestOpenaiEmbeddingsWithOptions(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	e := newOpenAIEmbedder(t, WithBatchSize(1), WithStripNewLines(false))

	_, err := e.EmbedQuery(ctx, "Hello world!")
	require.NoError(t, err)

	embeddings, err := e.EmbedDocuments(ctx, []string{"Hello world"})
	require.NoError(t, err)
	assert.Len(t, embeddings, 1)
}

func TestOpenaiEmbeddingsWithAzureAPI(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	t.Cleanup(func() { rr.Close() })
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
		openai.WithHTTPClient(rr.Client()),
	)
	require.NoError(t, err)

	e, err := NewEmbedder(client, WithBatchSize(1), WithStripNewLines(false))
	require.NoError(t, err)

	_, err = e.EmbedQuery(ctx, "Hello world!")
	require.NoError(t, err)

	embeddings, err := e.EmbedDocuments(ctx, []string{"Hello world"})
	require.NoError(t, err)
	assert.Len(t, embeddings, 1)
}

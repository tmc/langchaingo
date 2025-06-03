package huggingface

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/llms/huggingface"
)

func TestHuggingfaceEmbeddings(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "HF_TOKEN")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	t.Cleanup(func() { rr.Close() })

	// Create HuggingFace client with httprr HTTP client
	hfClient, err := huggingface.New(huggingface.WithHTTPClient(rr.Client()))
	require.NoError(t, err)

	e, err := NewHuggingface(WithClient(*hfClient))
	require.NoError(t, err)

	_, err = e.EmbedQuery(ctx, "Hello world!")
	require.NoError(t, err)

	embeddings, err := e.EmbedDocuments(ctx, []string{"Hello world", "The world is ending", "good bye"})
	require.NoError(t, err)
	assert.Len(t, embeddings, 3)
}

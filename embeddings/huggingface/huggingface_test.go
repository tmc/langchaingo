package huggingface

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/0xDezzy/langchaingo/internal/httprr"
	"github.com/0xDezzy/langchaingo/llms/huggingface"
)

func TestHuggingfaceEmbeddings(t *testing.T) {

	t.Skip("temporary skip")
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "HF_TOKEN")

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	// Only run tests in parallel when not recording (to avoid rate limits)
	if rr.Replaying() {
		t.Parallel()
	}

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

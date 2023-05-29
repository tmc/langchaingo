package embeddings

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVertexAIPaLMEmbeddings(t *testing.T) {
	t.Parallel()

	if gcpProjectID := os.Getenv("GOOGLE_CLOUD_PROJECT"); gcpProjectID == "" {
		t.Skip("GOOGLE_CLOUD_PROJECT not set")
	}
	e, err := NewVertexAIPaLM()
	require.NoError(t, err)

	_, err = e.EmbedQuery(context.Background(), "Hello world!")
	require.NoError(t, err)

	embeddings, err := e.EmbedDocuments(context.Background(), []string{"Hello world", "The world is ending", "good bye"})
	require.NoError(t, err)
	assert.Len(t, embeddings, 3)
}

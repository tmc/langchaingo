package huggingface

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHuggingfaceEmbeddings(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	if huggingfaceKey := os.Getenv("HUGGINGFACEHUB_API_TOKEN"); huggingfaceKey == "" {
		t.Skip("HUGGINGFACEHUB_API_TOKEN not set")
	}
	e, err := NewHuggingface()
	require.NoError(t, err)

	_, err = e.EmbedQuery(ctx, "Hello world!")
	require.NoError(t, err)

	embeddings, err := e.EmbedDocuments(ctx, []string{"Hello world", "The world is ending", "good bye"})
	require.NoError(t, err)
	assert.Len(t, embeddings, 3)
}

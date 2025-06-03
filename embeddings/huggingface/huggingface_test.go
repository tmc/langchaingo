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

	// Check both HF_TOKEN and HUGGINGFACEHUB_API_TOKEN
	hfToken := os.Getenv("HF_TOKEN")
	huggingfaceToken := os.Getenv("HUGGINGFACEHUB_API_TOKEN")
	if hfToken == "" && huggingfaceToken == "" {
		t.Skip("HF_TOKEN or HUGGINGFACEHUB_API_TOKEN not set")
	}

	e, err := NewHuggingface()
	require.NoError(t, err)

	_, err = e.EmbedQuery(ctx, "Hello world!")
	require.NoError(t, err)

	embeddings, err := e.EmbedDocuments(ctx, []string{"Hello world", "The world is ending", "good bye"})
	require.NoError(t, err)
	assert.Len(t, embeddings, 3)
}

package jina

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJinaEmbeddings(t *testing.T) {
	t.Parallel()

	if jinakey := os.Getenv("JINA_API_KEY"); jinakey == "" {
		t.Skip("JINA_API_KEY not set")
	}
	j, err := NewJina()
	require.NoError(t, err)

	_, err = j.EmbedQuery(context.Background(), "Hello world!")
	require.NoError(t, err)

	embeddings, err := j.EmbedDocuments(context.Background(), []string{"Hello world", "The world is ending", "good bye"})
	require.NoError(t, err)
	assert.Len(t, embeddings, 3)
}

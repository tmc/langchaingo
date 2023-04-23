package embeddings

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenaiEmbeddings(t *testing.T) {
	t.Parallel()

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	e, err := NewOpenAI()
	require.NoError(t, err)

	_, err = e.EmbedQuery("Hello world!")
	require.NoError(t, err)

	embeddings, err := e.EmbedDocuments([]string{"Hello world", "The world is ending", "bye bye"})
	require.NoError(t, err)
	assert.Len(t, embeddings, 3)
}

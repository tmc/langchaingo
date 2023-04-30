package embeddings

import (
	"context"
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

	_, err = e.EmbedQuery(context.Background(), "Hello world!")
	require.NoError(t, err)

	embeddings, err := e.EmbedDocuments(context.Background(), []string{"Hello world", "The world is ending", "good bye"})
	require.NoError(t, err)
	assert.Len(t, embeddings, 3)
}

func TestBatchTexts(t *testing.T) {
	t.Parallel()

	cases := []struct {
		texts     []string
		batchSize int
		expected  [][]string
	}{
		{
			texts:     []string{"foo bar zoo"},
			batchSize: 4,
			expected:  [][]string{{"foo ", "bar ", "zoo"}},
		},
		{
			texts:     []string{"foo bar zoo", "foo"},
			batchSize: 7,
			expected:  [][]string{{"foo bar", " zoo"}, {"foo"}},
		},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.expected, batchTexts(tc.texts, tc.batchSize))
	}
}

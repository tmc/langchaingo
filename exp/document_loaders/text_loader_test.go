package document_loaders

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextLoader(t *testing.T) {
	t.Parallel()

	loader := NewTextLoaderFromFile("./testdata/test.txt")

	docs, err := loader.Load()
	require.NoError(t, err)
	require.Len(t, docs, 1)

	expectedPageContent := "Foo Bar Baz"
	assert.Equal(t, docs[0].PageContent, expectedPageContent)

	expectedMetadata := map[string]any{
		"source": "./testdata/test.txt",
	}

	assert.Equal(t, docs[0].Metadata, expectedMetadata)
}

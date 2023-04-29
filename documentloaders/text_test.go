package documentloaders

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextLoader(t *testing.T) {
	t.Parallel()
	file, err := os.Open("./testdata/test.txt")
	assert.NoError(t, err)

	loader := NewText(file)

	docs, err := loader.Load(context.Background())
	require.NoError(t, err)
	require.Len(t, docs, 1)

	expectedPageContent := "Foo Bar Baz"
	assert.Equal(t, docs[0].PageContent, expectedPageContent)

	expectedMetadata := map[string]any{}
	assert.Equal(t, docs[0].Metadata, expectedMetadata)
}

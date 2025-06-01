package documentloaders

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextLoader(t *testing.T) {
	t.Parallel()
	file, err := os.Open("./testdata/test.txt")
	require.NoError(t, err)

	loader := NewText(file)

	docs, err := loader.Load(t.Context())
	require.NoError(t, err)
	require.Len(t, docs, 1)

	expectedPageContent := "Foo Bar Baz"
	assert.Equal(t, expectedPageContent, docs[0].PageContent)

	expectedMetadata := map[string]any{}
	assert.Equal(t, expectedMetadata, docs[0].Metadata)
}

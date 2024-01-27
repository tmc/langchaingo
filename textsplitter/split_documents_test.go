package textsplitter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

func TestItSplitsDocumentsRetainingTheCustomIDs(t *testing.T) {
	docs := []schema.Document{
		{
			PageContent: "Item 1",
			Metadata:    map[string]any{},
			CustomID:    stringPtr("1"),
		},
		{
			PageContent: "Item 2",
			Metadata:    map[string]any{},
			CustomID:    stringPtr("2"),
		},
		{
			PageContent: "Item 2",
			Metadata:    map[string]any{},
		},
	}

	splitter := textsplitter.NewTokenSplitter(textsplitter.WithChunkSize(512))

	result, err := textsplitter.SplitDocuments(splitter, docs)
	require.NoError(t, err)
	assert.Equal(t, result, docs)
}

func stringPtr(s string) *string {
	return &s
}

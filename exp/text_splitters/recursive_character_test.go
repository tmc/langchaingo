package text_splitters

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/schema"
)

func TestRecursiveCharacterSplitter(t *testing.T) {
	t.Parallel()

	type test struct {
		text         string
		chunkOverlap int
		chunkSize    int
		expectedDocs []schema.Document
	}

	tests := []test{
		{
			text:         "Hi.\nI'm Harrison.\n\nHow?\na\nb",
			chunkOverlap: 1,
			chunkSize:    20,
			expectedDocs: []schema.Document{
				{
					PageContent: "Hi.\nI'm Harrison.",
					Metadata:    map[string]any{}, /* "loc": LOCMetadata{
					Lines: LineData{
						From: 1,
						To:   2,
					}, */
				},
				{
					PageContent: "How?\na\nb",
					Metadata:    map[string]any{}, /* "loc": LOCMetadata{
						Lines: LineData{
							From: 4,
							To:   6,
						},
					} */
				},
			},
		},
		{
			text:         "Hi.\nI'm Harrison.\n\nHow?\na\nbHi.\nI'm Harrison.\n\nHow?\na\nb",
			chunkOverlap: 1,
			chunkSize:    40,
			expectedDocs: []schema.Document{
				{
					PageContent: "Hi.\nI'm Harrison.",
					Metadata:    map[string]any{}, /* "loc": LOCMetadata{
						Lines: LineData{
							From: 1,
							To:   2,
						},
					}}, */
				},
				{
					PageContent: "How?\na\nbHi.\nI'm Harrison.\n\nHow?\na\nb",
					Metadata:    map[string]any{}, /* "loc": LOCMetadata{
						Lines: LineData{
							From: 4,
							To:   11,
						},
					}}, */
				},
			},
		},
	}

	splitter := NewRecursiveCharactersSplitter()
	for _, test := range tests {
		splitter.ChunkOverlap = test.chunkOverlap
		splitter.ChunkSize = test.chunkSize

		docs, err := CreateDocuments(splitter, []string{test.text}, []map[string]any{})
		assert.NoError(t, err)
		assert.Equal(t, test.expectedDocs, docs)
	}
}

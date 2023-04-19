package textSplitters

import (
	"reflect"
	"testing"

	"github.com/tmc/langchaingo/exp/schema"
)

type test struct {
	text         string
	chunkOverlap int
	chunkSize    int
	expectedDocs []schema.Document
}

var tests = []test{
	{

		text:         "Hi.\nI'm Harrison.\n\nHow?\na\nb",
		chunkOverlap: 1,
		chunkSize:    20,
		expectedDocs: []schema.Document{
			{
				PageContent: "Hi.\nI'm Harrison.",
				Metadata: map[string]any{"loc": LOCMetadata{
					Lines: LineData{
						From: 1,
						To:   2,
					},
				}},
			},
			{
				PageContent: "How?\na\nb",
				Metadata: map[string]any{"loc": LOCMetadata{
					Lines: LineData{
						From: 4,
						To:   6,
					},
				}},
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
				Metadata: map[string]any{"loc": LOCMetadata{
					Lines: LineData{
						From: 1,
						To:   2,
					},
				}},
			},
			{
				PageContent: "How?\na\nbHi.\nI'm Harrison.\n\nHow?\na\nb",
				Metadata: map[string]any{"loc": LOCMetadata{
					Lines: LineData{
						From: 4,
						To:   11,
					},
				}},
			},
		},
	},
}

func TestRecursiveCharacterSplitter(t *testing.T) {
	splitter := NewRecursiveCharactersSplitter()

	for _, test := range tests {
		splitter.ChunkOverlap = test.chunkOverlap
		splitter.ChunkSize = test.chunkSize

		docs, err := CreateDocuments(splitter, []string{test.text}, []map[string]any{})
		if err != nil {
			t.Errorf("Unexpected error creating documents with recursive character splitter: %s ", err.Error())
		}

		if !reflect.DeepEqual(docs, test.expectedDocs) {
			t.Logf("Result creating documents with recursive character splitter not equal expected. \n Got:\n %v \nWant:\n %v \n", docs, test.expectedDocs)
		}
	}

}

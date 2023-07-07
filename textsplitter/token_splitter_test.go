package textsplitter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/schema"
)

func TestTokenSplitter(t *testing.T) {
	t.Parallel()
	type testCase struct {
		text         string
		chunkOverlap int
		chunkSize    int
		expectedDocs []schema.Document
	}
	//nolint:dupword
	testCases := []testCase{
		{
			text:         "Hi.\nI'm Harrison.\n\nHow?\na\nb",
			chunkOverlap: 1,
			chunkSize:    20,
			expectedDocs: []schema.Document{
				{PageContent: "Hi.\nI'm Harrison.\n\nHow?\na\nb", Metadata: map[string]any{}},
			},
		},
		{
			text:         "Hi.\nI'm Harrison.\n\nHow?\na\nbHi.\nI'm Harrison.\n\nHow?\na\nb",
			chunkOverlap: 1,
			chunkSize:    40,
			expectedDocs: []schema.Document{
				{PageContent: "Hi.\nI'm Harrison.\n\nHow?\na\nbHi.\nI'm Harrison.\n\nHow?\na\nb", Metadata: map[string]any{}},
			},
		},
		{
			text:         "name: Harrison\nage: 30",
			chunkOverlap: 1,
			chunkSize:    40,
			expectedDocs: []schema.Document{
				{PageContent: "name: Harrison\nage: 30", Metadata: map[string]any{}},
			},
		},
		{
			text: `name: Harrison
age: 30

name: Joe
age: 32`,
			chunkOverlap: 1,
			chunkSize:    40,
			expectedDocs: []schema.Document{
				{PageContent: "name: Harrison\nage: 30\n\nname: Joe\nage: 32", Metadata: map[string]any{}},
			},
		},
		{
			text: `Hi.
I'm Harrison.

How? Are? You?
Okay then f f f f.
This is a weird text to write, but gotta test the splittingggg some how.

Bye!

-H.`,
			chunkOverlap: 1,
			chunkSize:    10,
			expectedDocs: []schema.Document{
				{PageContent: "Hi.\nI'm Harrison.\n\nHow? Are?", Metadata: map[string]any{}},
				{PageContent: "? You?\nOkay then f f f f.\n", Metadata: map[string]any{}},
				{PageContent: ".\nThis is a weird text to write, but", Metadata: map[string]any{}},
				{PageContent: " but gotta test the splittingggg some how.\n\n", Metadata: map[string]any{}},
				{PageContent: ".\n\nBye!\n\n-H.", Metadata: map[string]any{}},
			},
		},
	}
	splitter := NewTokenSplitter()
	for _, tc := range testCases {
		splitter.ChunkOverlap = tc.chunkOverlap
		splitter.ChunkSize = tc.chunkSize

		docs, err := CreateDocuments(splitter, []string{tc.text}, nil)
		assert.NoError(t, err)
		assert.Equal(t, tc.expectedDocs, docs)
	}
}

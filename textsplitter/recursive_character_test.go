package textsplitter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/schema"
)

//nolint:dupword
func TestRecursiveCharacterSplitter(t *testing.T) {
	t.Parallel()
	type testCase struct {
		text         string
		chunkOverlap int
		chunkSize    int
		expectedDocs []schema.Document
	}
	testCases := []testCase{
		{
			text:         "Hi.\nI'm Harrison.\n\nHow?\na\nb",
			chunkOverlap: 1,
			chunkSize:    20,
			expectedDocs: []schema.Document{
				{PageContent: "Hi.\nI'm Harrison.", Metadata: map[string]any{}},
				{PageContent: "How?\na\nb", Metadata: map[string]any{}},
			},
		},
		{
			text:         "Hi.\nI'm Harrison.\n\nHow?\na\nbHi.\nI'm Harrison.\n\nHow?\na\nb",
			chunkOverlap: 1,
			chunkSize:    40,
			expectedDocs: []schema.Document{
				{PageContent: "Hi.\nI'm Harrison.", Metadata: map[string]any{}},
				{PageContent: "How?\na\nbHi.\nI'm Harrison.\n\nHow?\na\nb", Metadata: map[string]any{}},
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
				{PageContent: "name: Harrison\nage: 30", Metadata: map[string]any{}},
				{PageContent: "name: Joe\nage: 32", Metadata: map[string]any{}},
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
				{PageContent: "Hi.", Metadata: map[string]any{}},
				{PageContent: "I'm", Metadata: map[string]any{}},
				{PageContent: "Harrison.", Metadata: map[string]any{}},
				{PageContent: "How? Are?", Metadata: map[string]any{}},
				{PageContent: "You?", Metadata: map[string]any{}},
				{PageContent: "Okay then", Metadata: map[string]any{}},
				{PageContent: "f f f f.", Metadata: map[string]any{}},
				{PageContent: "This is a", Metadata: map[string]any{}},
				{PageContent: "a weird", Metadata: map[string]any{}},
				{PageContent: "text to", Metadata: map[string]any{}},
				{PageContent: "write, but", Metadata: map[string]any{}},
				{PageContent: "gotta test", Metadata: map[string]any{}},
				{PageContent: "the", Metadata: map[string]any{}},
				{PageContent: "splittingg", Metadata: map[string]any{}},
				{PageContent: "ggg", Metadata: map[string]any{}},
				{PageContent: "some how.", Metadata: map[string]any{}},
				{PageContent: "Bye!\n\n-H.", Metadata: map[string]any{}},
			},
		},
	}
	splitter := NewRecursiveCharacter()
	for _, tc := range testCases {
		splitter.ChunkOverlap = tc.chunkOverlap
		splitter.ChunkSize = tc.chunkSize

		docs, err := CreateDocuments(splitter, []string{tc.text}, nil)
		assert.NoError(t, err)
		assert.Equal(t, tc.expectedDocs, docs)
	}
}

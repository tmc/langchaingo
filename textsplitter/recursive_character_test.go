package textsplitter

import (
	"strings"
	"testing"

	"github.com/averikitsch/langchaingo/schema"
	"github.com/pkoukk/tiktoken-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:dupword,funlen
func TestRecursiveCharacterSplitter(t *testing.T) {
	tokenEncoder, _ := tiktoken.GetEncoding("cl100k_base")

	t.Parallel()
	type testCase struct {
		text          string
		chunkOverlap  int
		chunkSize     int
		separators    []string
		expectedDocs  []schema.Document
		keepSeparator bool
		LenFunc       func(string) int
	}
	testCases := []testCase{
		{
			text:         "哈里森\n很高兴遇见你\n欢迎来中国",
			chunkOverlap: 0,
			chunkSize:    10,
			separators:   []string{"\n\n", "\n", " "},
			expectedDocs: []schema.Document{
				{PageContent: "哈里森\n很高兴遇见你", Metadata: map[string]any{}},
				{PageContent: "欢迎来中国", Metadata: map[string]any{}},
			},
		},
		{
			text:         "Hi, Harrison. \nI am glad to meet you",
			chunkOverlap: 1,
			chunkSize:    20,
			separators:   []string{"\n", "$"},
			expectedDocs: []schema.Document{
				{PageContent: "Hi, Harrison.", Metadata: map[string]any{}},
				{PageContent: "I am glad to meet you", Metadata: map[string]any{}},
			},
		},
		{
			text:         "Hi.\nI'm Harrison.\n\nHow?\na\nbHi.\nI'm Harrison.\n\nHow?\na\nb",
			chunkOverlap: 1,
			chunkSize:    40,
			separators:   []string{"\n\n", "\n", " ", ""},
			expectedDocs: []schema.Document{
				{PageContent: "Hi.\nI'm Harrison.", Metadata: map[string]any{}},
				{PageContent: "How?\na\nbHi.\nI'm Harrison.\n\nHow?\na\nb", Metadata: map[string]any{}},
			},
		},
		{
			text:         "name: Harrison\nage: 30",
			chunkOverlap: 1,
			chunkSize:    40,
			separators:   []string{"\n\n", "\n", " ", ""},
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
			separators:   []string{"\n\n", "\n", " ", ""},
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
			separators:   []string{"\n\n", "\n", " ", ""},
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
		{
			text:          "Hi, Harrison. \nI am glad to meet you",
			chunkOverlap:  0,
			chunkSize:     10,
			separators:    []string{"\n", "$"},
			keepSeparator: true,
			expectedDocs: []schema.Document{
				{PageContent: "Hi, Harrison. ", Metadata: map[string]any{}},
				{PageContent: "\nI am glad to meet you", Metadata: map[string]any{}},
			},
		},
		{
			text:          strings.Repeat("The quick brown fox jumped over the lazy dog. ", 2),
			chunkOverlap:  0,
			chunkSize:     10,
			separators:    []string{" "},
			keepSeparator: true,
			LenFunc:       func(s string) int { return len(tokenEncoder.Encode(s, nil, nil)) },
			expectedDocs: []schema.Document{
				{PageContent: "The quick brown fox jumped over the lazy dog.", Metadata: map[string]any{}},
				{PageContent: "The quick brown fox jumped over the lazy dog.", Metadata: map[string]any{}},
			},
		},
	}
	splitter := NewRecursiveCharacter()
	for _, tc := range testCases {
		splitter.ChunkOverlap = tc.chunkOverlap
		splitter.ChunkSize = tc.chunkSize
		splitter.Separators = tc.separators
		splitter.KeepSeparator = tc.keepSeparator
		if tc.LenFunc != nil {
			splitter.LenFunc = tc.LenFunc
		}

		docs, err := CreateDocuments(splitter, []string{tc.text}, nil)
		require.NoError(t, err)
		assert.Equal(t, tc.expectedDocs, docs)
	}
}

package textsplitter

import (
	"testing"

	"github.com/pkoukk/tiktoken-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/schema"
)

func TestTokenSplitter(t *testing.T) {
	t.Parallel()
	if _, err := tiktoken.GetEncoding(_defaultTokenEncoding); err != nil {
		t.Skipf("tokenizer unavailable: %v", err)
	}
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
		require.NoError(t, err)
		assert.Equal(t, tc.expectedDocs, docs)
	}
}

func TestTokenSplitterStopsAtEndOfText(t *testing.T) {
	t.Parallel()
	if _, err := tiktoken.GetEncoding(_defaultTokenEncoding); err != nil {
		t.Skipf("tokenizer unavailable: %v", err)
	}

	text := "one two three four five six seven eight nine"
	testCases := []struct {
		chunkSize    int
		chunkOverlap int
		expected     []string
	}{
		{
			chunkSize:    5,
			chunkOverlap: 1,
			expected:     []string{"one two three four five", " five six seven eight nine"},
		},
		{
			chunkSize:    3,
			chunkOverlap: 2,
			expected: []string{
				"one two three", " two three four", " three four five", " four five six",
				" five six seven", " six seven eight", " seven eight nine",
			},
		},
		{
			chunkSize:    20,
			chunkOverlap: 5,
			expected:     []string{text},
		},
	}
	for _, tc := range testCases {
		splitter := NewTokenSplitter(WithChunkSize(tc.chunkSize), WithChunkOverlap(tc.chunkOverlap))
		chunks, err := splitter.SplitText(text)
		require.NoError(t, err)
		assert.Equal(t, tc.expected, chunks, "size=%d overlap=%d", tc.chunkSize, tc.chunkOverlap)
	}
}

func TestTokenSplitterRejectsInvalidChunkConfig(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		chunkSize    int
		chunkOverlap int
	}{
		{chunkSize: 3, chunkOverlap: 3},
		{chunkSize: 3, chunkOverlap: 4},
		{chunkSize: 3, chunkOverlap: -1},
		{chunkSize: 0, chunkOverlap: 0},
	}
	for _, tc := range testCases {
		splitter := NewTokenSplitter(WithChunkSize(tc.chunkSize), WithChunkOverlap(tc.chunkOverlap))
		chunks, err := splitter.SplitText("one two three four five")
		require.ErrorIs(t, err, ErrInvalidChunkConfig, "size=%d overlap=%d", tc.chunkSize, tc.chunkOverlap)
		assert.Nil(t, chunks)
	}
}

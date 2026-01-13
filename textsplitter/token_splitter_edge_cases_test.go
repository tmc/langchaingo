package textsplitter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/schema"
)

func TestTokenSplitterEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("empty text", func(t *testing.T) {
		t.Parallel()
		splitter := NewTokenSplitter()
		docs, err := CreateDocuments(splitter, []string{""}, nil)
		require.NoError(t, err)
		assert.Equal(t, []schema.Document{
			{PageContent: "", Metadata: map[string]any{}},
		}, docs)
	})

	t.Run("single character", func(t *testing.T) {
		t.Parallel()
		splitter := NewTokenSplitter(WithChunkSize(1))
		docs, err := CreateDocuments(splitter, []string{"a"}, nil)
		require.NoError(t, err)
		assert.Equal(t, []schema.Document{
			{PageContent: "a", Metadata: map[string]any{}},
		}, docs)
	})

	t.Run("whitespace only", func(t *testing.T) {
		t.Parallel()
		splitter := NewTokenSplitter(WithChunkSize(10))
		docs, err := CreateDocuments(splitter, []string{"   \n\t  "}, nil)
		require.NoError(t, err)
		assert.Len(t, docs, 1)
		assert.Equal(t, "   \n\t  ", docs[0].PageContent)
	})

	t.Run("very large chunk size", func(t *testing.T) {
		t.Parallel()
		text := "This is a test text that should not be split because the chunk size is very large."
		splitter := NewTokenSplitter(WithChunkSize(10000))
		docs, err := CreateDocuments(splitter, []string{text}, nil)
		require.NoError(t, err)
		assert.Len(t, docs, 1)
		assert.Equal(t, text, docs[0].PageContent)
	})

	t.Run("zero chunk overlap", func(t *testing.T) {
		t.Parallel()
		text := "This is a longer text that should be split into multiple chunks without any overlap between them."
		splitter := NewTokenSplitter(
			WithChunkSize(10),
			WithChunkOverlap(0),
		)
		docs, err := CreateDocuments(splitter, []string{text}, nil)
		require.NoError(t, err)
		assert.Greater(t, len(docs), 1)

		// Verify no overlap by checking that no two consecutive chunks share content
		for i := 1; i < len(docs); i++ {
			prev := strings.TrimSpace(docs[i-1].PageContent)
			curr := strings.TrimSpace(docs[i].PageContent)
			if prev != "" && curr != "" {
				// Should not have overlapping words at boundaries
				prevWords := strings.Fields(prev)
				currWords := strings.Fields(curr)
				if len(prevWords) > 0 && len(currWords) > 0 {
					assert.NotEqual(t, prevWords[len(prevWords)-1], currWords[0],
						"Chunks should not overlap when overlap is 0")
				}
			}
		}
	})

	t.Run("chunk overlap equals chunk size", func(t *testing.T) {
		t.Parallel()
		text := "Word1 Word2 Word3 Word4 Word5 Word6 Word7 Word8"
		splitter := NewTokenSplitter(
			WithChunkSize(5),
			WithChunkOverlap(5),
		)
		docs, err := CreateDocuments(splitter, []string{text}, nil)
		require.NoError(t, err)
		assert.Greater(t, len(docs), 1)

		// With overlap equal to chunk size, chunks should have significant overlap
		for i := 1; i < len(docs); i++ {
			assert.NotEmpty(t, docs[i].PageContent)
		}
	})

	t.Run("unicode and special characters", func(t *testing.T) {
		t.Parallel()
		text := "Hello 世界! 🌍 This contains émojis and spëcial characters: àáâãäå"
		splitter := NewTokenSplitter(WithChunkSize(20))
		docs, err := CreateDocuments(splitter, []string{text}, nil)
		require.NoError(t, err)
		assert.NotEmpty(t, docs)

		// Verify all content is preserved
		combined := ""
		for _, doc := range docs {
			combined += doc.PageContent
		}
		// Remove potential whitespace differences for comparison
		assert.Contains(t, combined, "世界")
		assert.Contains(t, combined, "🌍")
		assert.Contains(t, combined, "émojis")
		assert.Contains(t, combined, "àáâãäå")
	})

	t.Run("very long single word", func(t *testing.T) {
		t.Parallel()
		longWord := strings.Repeat("a", 1000)
		splitter := NewTokenSplitter(WithChunkSize(10))
		docs, err := CreateDocuments(splitter, []string{longWord}, nil)
		require.NoError(t, err)
		assert.Greater(t, len(docs), 1)

		// Verify the long word is split appropriately
		combined := ""
		for _, doc := range docs {
			combined += doc.PageContent
		}
		assert.Equal(t, longWord, combined)
	})

	t.Run("newlines and formatting preservation", func(t *testing.T) {
		t.Parallel()
		text := "Line 1\n\nLine 2\n\n\nLine 3\n\t\tIndented line\n"
		splitter := NewTokenSplitter(WithChunkSize(15))
		docs, err := CreateDocuments(splitter, []string{text}, nil)
		require.NoError(t, err)
		assert.NotEmpty(t, docs)

		// Verify formatting is preserved in the split
		combined := ""
		for _, doc := range docs {
			combined += doc.PageContent
		}
		assert.Contains(t, combined, "\n\n")
		assert.Contains(t, combined, "\t\t")
	})
}

func TestTokenSplitterDifferentModels(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		modelName string
		text      string
	}{
		{
			name:      "gpt-4 model",
			modelName: "gpt-4",
			text:      "This is a test for GPT-4 tokenization.",
		},
		{
			name:      "gpt-3.5-turbo model",
			modelName: "gpt-3.5-turbo",
			text:      "This is a test for GPT-3.5-turbo tokenization.",
		},
		{
			name:      "text-davinci-003 model",
			modelName: "text-davinci-003",
			text:      "This is a test for text-davinci-003 tokenization.",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			splitter := NewTokenSplitter(
				WithModelName(tc.modelName),
				WithChunkSize(10),
			)
			docs, err := CreateDocuments(splitter, []string{tc.text}, nil)
			require.NoError(t, err)
			assert.NotEmpty(t, docs)

			// Verify content is preserved
			combined := ""
			for _, doc := range docs {
				combined += doc.PageContent
			}
			assert.Contains(t, combined, "test")
			assert.Contains(t, combined, "tokenization")
		})
	}
}

func TestTokenSplitterDifferentEncodings(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		encodingName string
		text         string
	}{
		{
			name:         "cl100k_base encoding",
			encodingName: "cl100k_base",
			text:         "Testing cl100k_base encoding with various tokens.",
		},
		{
			name:         "p50k_base encoding",
			encodingName: "p50k_base",
			text:         "Testing p50k_base encoding with various tokens.",
		},
		{
			name:         "r50k_base encoding",
			encodingName: "r50k_base",
			text:         "Testing r50k_base encoding with various tokens.",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			splitter := NewTokenSplitter(
				WithEncodingName(tc.encodingName),
				WithChunkSize(15),
			)
			docs, err := CreateDocuments(splitter, []string{tc.text}, nil)
			require.NoError(t, err)
			assert.NotEmpty(t, docs)

			// Verify content is preserved
			combined := ""
			for _, doc := range docs {
				combined += doc.PageContent
			}
			assert.Contains(t, combined, "Testing")
			assert.Contains(t, combined, "encoding")
		})
	}
}

func TestTokenSplitterSpecialTokens(t *testing.T) {
	t.Parallel()

	t.Run("with allowed special tokens", func(t *testing.T) {
		t.Parallel()
		text := "This text contains <|endoftext|> special token."
		splitter := NewTokenSplitter(
			WithAllowedSpecial([]string{"<|endoftext|>"}),
			WithChunkSize(20),
		)
		docs, err := CreateDocuments(splitter, []string{text}, nil)
		require.NoError(t, err)
		assert.NotEmpty(t, docs)

		// Verify special token is preserved
		combined := ""
		for _, doc := range docs {
			combined += doc.PageContent
		}
		assert.Contains(t, combined, "<|endoftext|>")
	})

	t.Run("with disallowed special tokens", func(t *testing.T) {
		t.Parallel()
		text := "This is normal text without special tokens."
		splitter := NewTokenSplitter(
			WithDisallowedSpecial([]string{"all"}),
			WithChunkSize(20),
		)
		docs, err := CreateDocuments(splitter, []string{text}, nil)
		require.NoError(t, err)
		assert.NotEmpty(t, docs)

		// Verify content is preserved
		combined := ""
		for _, doc := range docs {
			combined += doc.PageContent
		}
		assert.Contains(t, combined, "normal text")
	})
}

func TestTokenSplitterErrorHandling(t *testing.T) {
	t.Parallel()

	t.Run("invalid model name", func(t *testing.T) {
		t.Parallel()
		splitter := NewTokenSplitter(WithModelName("invalid-model-name"))
		_, err := CreateDocuments(splitter, []string{"test"}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tiktoken")
	})

	t.Run("invalid encoding name", func(t *testing.T) {
		t.Parallel()
		splitter := NewTokenSplitter(WithEncodingName("invalid-encoding"))
		_, err := CreateDocuments(splitter, []string{"test"}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tiktoken")
	})
}

func TestTokenSplitterConsistency(t *testing.T) {
	t.Parallel()

	t.Run("consistent splitting", func(t *testing.T) {
		t.Parallel()
		text := "This is a consistent test text that should be split the same way every time."
		splitter := NewTokenSplitter(
			WithChunkSize(10),
			WithChunkOverlap(2),
		)

		// Split the same text multiple times
		docs1, err1 := CreateDocuments(splitter, []string{text}, nil)
		require.NoError(t, err1)

		docs2, err2 := CreateDocuments(splitter, []string{text}, nil)
		require.NoError(t, err2)

		// Results should be identical
		assert.Equal(t, docs1, docs2)
	})

	t.Run("order preservation", func(t *testing.T) {
		t.Parallel()
		text := "First sentence. Second sentence. Third sentence. Fourth sentence."
		splitter := NewTokenSplitter(WithChunkSize(8))
		docs, err := CreateDocuments(splitter, []string{text}, nil)
		require.NoError(t, err)

		// Verify order is preserved
		combined := ""
		for _, doc := range docs {
			combined += doc.PageContent
		}

		firstPos := strings.Index(combined, "First")
		secondPos := strings.Index(combined, "Second")
		thirdPos := strings.Index(combined, "Third")
		fourthPos := strings.Index(combined, "Fourth")

		assert.True(t, firstPos < secondPos)
		assert.True(t, secondPos < thirdPos)
		assert.True(t, thirdPos < fourthPos)
	})
}

package embeddings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBatchTexts(t *testing.T) {
	t.Parallel()

	cases := []struct {
		texts     []string
		batchSize int
		expected  [][]string
	}{
		{
			texts:     []string{"foo bar zoo"},
			batchSize: 4,
			expected:  [][]string{{"foo ", "bar ", "zoo"}},
		},
		{
			texts:     []string{"foo bar zoo", "foo"},
			batchSize: 7,
			expected:  [][]string{{"foo bar", " zoo"}, {"foo"}},
		},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.expected, BatchTexts(tc.texts, tc.batchSize))
	}
}

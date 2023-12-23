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
			texts:     []string{},
			batchSize: 1,
			expected:  [][]string{},
		},
		{
			texts:     []string{"foo bar zoo"},
			batchSize: 4,
			expected:  [][]string{{"foo bar zoo"}},
		},
		{
			texts:     []string{"foo bar zoo", "foo"},
			batchSize: 7,
			expected:  [][]string{{"foo bar zoo", "foo"}},
		},
		{
			texts:     []string{"foo", "bar", "zoo"},
			batchSize: 2,
			expected:  [][]string{{"foo", "bar"}, {"zoo"}},
		},
		{
			texts:     []string{"foo", "bar", "zoo", "baz", "qux"},
			batchSize: 2,
			expected:  [][]string{{"foo", "bar"}, {"zoo", "baz"}, {"qux"}},
		},
		{
			texts:     []string{"foo", "bar", "zoo", "baz"},
			batchSize: 2,
			expected:  [][]string{{"foo", "bar"}, {"zoo", "baz"}},
		},
		{
			texts:     []string{"foo", "bar", "zoo", "baz", "qux"},
			batchSize: 3,
			expected:  [][]string{{"foo", "bar", "zoo"}, {"baz", "qux"}},
		},
		{
			texts:     []string{"foo", "bar", "zoo", "baz", "qux"},
			batchSize: 6,
			expected:  [][]string{{"foo", "bar", "zoo", "baz", "qux"}},
		},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.expected, BatchTexts(tc.texts, tc.batchSize))
	}
}

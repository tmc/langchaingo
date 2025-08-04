package outputparser_test

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yincongcyincong/langchaingo/outputparser"
)

func TestCommaSeparatedList(t *testing.T) {
	t.Parallel()
	
	testCases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "foo, bar, baz",
			expected: []string{"foo", "bar", "baz"},
		},
		{
			input:    "	foo, bar  , b az ",
			expected: []string{"foo", "bar", "b az"},
		},
		{
			input:    " foo bar  , baz",
			expected: []string{"foo bar", "baz"},
		},
	}
	
	parser := outputparser.NewCommaSeparatedList()
	
	for _, tc := range testCases {
		output, err := parser.Parse(tc.input)
		require.NoError(t, err)
		assert.Equal(t, tc.expected, output)
	}
}

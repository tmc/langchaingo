package output_parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommaSeparatedListOutputParser(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input          string
		expectedOutput []string
	}

	tests := []testCase{
		{
			input:          "foo, bar, baz",
			expectedOutput: []string{"foo", "bar", "baz"},
		},
		{
			input:          "	foo, bar  , b az ",
			expectedOutput: []string{"foo", "bar", "b az"},
		},
		{
			input:          " foo bar  , baz",
			expectedOutput: []string{"foo bar", "baz"},
		},
	}

	parser := NewCommaSeparatedList()
	for _, test := range tests {
		output, err := parser.Parse(test.input)
		assert.NoError(t, err)
		assert.Equal(t, test.expectedOutput, output)
	}
}

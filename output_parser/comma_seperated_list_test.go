package output_parser_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tmc/langchaingo/output_parser"
)

func TestCommaSeparatedList(t *testing.T) {
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

	parser := output_parser.NewCommaSeparatedList()

	for _, tc := range testCases {
		output, err := parser.Parse(tc.input)
		if err != nil {
			t.Fatal(err)
		}

		if !cmp.Equal(output, tc.expected) {
			t.Logf("expected %v, got: %v", tc.expected, output)
		}
	}
}

package outputparser_test

import (
	"reflect"
	"testing"

	"github.com/vendasta/langchaingo/outputparser"
)

func TestCommaSeparatedList(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "basic comma separated",
			input:    "foo, bar, baz",
			expected: []string{"foo", "bar", "baz"},
		},
		{
			name:     "extra whitespace",
			input:    "	foo, bar  , b az ",
			expected: []string{"foo", "bar", "b az"},
		},
		{
			name:     "spaces in values",
			input:    " foo bar  , baz",
			expected: []string{"foo bar", "baz"},
		},
	}

	parser := outputparser.NewCommaSeparatedList()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := parser.Parse(tc.input)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", tc.input, err)
			}
			if !reflect.DeepEqual(tc.expected, output) {
				t.Errorf("Parse(%q) = %v, want %v", tc.input, output, tc.expected)
			}
		})
	}
}

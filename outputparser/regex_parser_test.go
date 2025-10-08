package outputparser_test

import (
	"reflect"
	"testing"

	"github.com/tmc/langchaingo/outputparser"
)

func TestRegexParser(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input      string
		expression string
		expected   map[string]string
	}{
		{
			input:      "testing_foo, testing_bar, testing_baz",
			expression: `(?P<foo>\w+), (?P<bar>\w+), (?P<baz>\w+)`,
			expected: map[string]string{
				"foo": "testing_foo",
				"bar": "testing_bar",
				"baz": "testing_baz",
			},
		},
		{
			input:      "Score: 100",
			expression: `Score: (?P<score>\d+)`,
			expected: map[string]string{
				"score": "100",
			},
		},
		{
			input:      "Score: 100",
			expression: `Score: (?P<score>\d+)(?:\s(?P<test>\d+))*`,
			expected: map[string]string{
				"score": "100",
				"test":  "",
			},
		},
	}

	for _, tc := range testCases {
		parser := outputparser.NewRegexParser(tc.expression)

		actual, err := parser.Parse(tc.input)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !reflect.DeepEqual(actual, tc.expected) {
			t.Errorf("Expected %v, got %v", tc.expected, actual)
		}
	}
}

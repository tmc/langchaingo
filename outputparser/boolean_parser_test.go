package outputparser_test

import (
	"testing"

	"github.com/tmc/langchaingo/outputparser"
)

func TestBooleanParser(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input    string
		err      error
		expected bool
	}{
		{
			input:    "Yes",
			expected: true,
		},
		{
			input:    "NO",
			expected: false,
		},
		{
			input:    "ok",
			err:      outputparser.ParseError{},
			expected: false,
		},
	}

	for _, tc := range testCases {
		parser := outputparser.NewBooleanParser()

		actual, err := parser.Parse(tc.input)
		if tc.err != nil && err == nil {
			t.Errorf("Expected error %v, got nil", tc.err)
		}

		if actual != tc.expected {
			t.Errorf("Expected %v, got %v", tc.expected, actual)
		}
	}
}

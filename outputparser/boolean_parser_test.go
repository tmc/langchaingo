package outputparser_test

import (
	"testing"

	"github.com/averikitsch/langchaingo/outputparser"
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
			input:    "YESNO",
			err:      outputparser.ParseError{},
			expected: false,
		},
		{
			input:    "ok",
			err:      outputparser.ParseError{},
			expected: false,
		},
		{
			input:    "true",
			expected: true,
		},
		{
			input:    "false",
			expected: false,
		},
		{
			input:    "True",
			expected: true,
		},
		{
			input:    "False",
			expected: false,
		},
		{
			input:    "TRUE",
			expected: true,
		},
		{
			input:    "FALSE",
			expected: false,
		},
		{
			input:    "'TRUE'",
			expected: true,
		},
		{
			input:    "`TRUE`",
			expected: true,
		},
		{
			input:    "'TRUE`",
			expected: true,
		},
	}

	for _, tc := range testCases {
		parser := outputparser.NewBooleanParser()

		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			result, err := parser.Parse(tc.input)
			if err != nil && tc.err == nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if err == nil && tc.err != nil {
				t.Errorf("Expected error %v, got nil", tc.err)
			}

			if result != tc.expected {
				t.Errorf("Expected %v, but got %v", tc.expected, result)
			}
		})
	}
}

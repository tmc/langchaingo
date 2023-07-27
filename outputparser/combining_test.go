package outputparser

import (
	"errors"
	"reflect"
	"testing"

	"github.com/tmc/langchaingo/schema"
)

func TestCombine(t *testing.T) {
	t.Parallel()

	testText := "```" + `json
{
    "answer": "Paris",
    "source": "https://en.wikipedia.org/wiki/France"
}
` + "```" + `

//Confidence: A, Explanation: Paris is the capital of France according to Wikipedia.`

	invalidText := "\n\n\n\n"

	structuredParser := NewStructured(
		[]ResponseSchema{
			{Name: "answer", Description: "The answer to the question"},
			{Name: "source", Description: "A link to the source"},
		},
	)

	regexParser := NewRegexParser("Confidence: (?P<confidence>A|B|C), Explanation: (?P<explanation>.*)")

	validParsers := []schema.OutputParser[any]{
		structuredParser,
		regexParser,
	}

	validOutput := map[string]any{
		"answer":      "Paris",
		"source":      "https://en.wikipedia.org/wiki/France",
		"confidence":  "A",
		"explanation": "Paris is the capital of France according to Wikipedia.",
	}

	invalidParsers := []schema.OutputParser[any]{
		structuredParser,
	}

	testCases := []struct {
		text     string
		parsers  Combining
		expected map[string]any
		err      error
	}{
		{
			text:     testText,
			parsers:  NewCombining(validParsers),
			expected: validOutput,
			err:      nil,
		},
		{
			text:     testText,
			parsers:  NewCombining(invalidParsers),
			expected: nil,
			err: ParseError{
				Text:   testText,
				Reason: "Combining parser requires at least 2 parsers, got 1",
			},
		},
		{
			text:     invalidText,
			parsers:  NewCombining(validParsers),
			expected: nil,
			err: ParseError{
				Text:   invalidText,
				Reason: "Texts count (3) does not match parsers count (2)",
			},
		},
	}

	for _, tc := range testCases {
		actual, err := tc.parsers.Parse(tc.text)
		if tc.err != nil && !errors.Is(tc.err, err) {
			t.Errorf("Expected error %v, got %v", err, tc.err)
		}

		if !reflect.DeepEqual(actual, tc.expected) {
			t.Errorf("Expected %v, got %v", tc.expected, actual)
		}
	}
}

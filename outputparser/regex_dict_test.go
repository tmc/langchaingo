package outputparser_test

import (
	"reflect"
	"testing"

	"github.com/tmc/langchaingo/outputparser"
)

func TestRegexDict(t *testing.T) {
	t.Parallel()

	testText := `We have just received a new result from the LLM, and our next step is
to filter and read its format using regular expressions to identify specific fields,
such as:

- Action: Search
- Action Input: How to use this class?
- Additional Fields: "N/A"

To assist us in this task, we use the regex_dict class. This class allows us to send a
dictionary containing an output key and the expected format, which in turn enables us to
retrieve the result of the matching formats and extract specific information from it.

To exclude irrelevant information from our return dictionary, we can instruct the LLM to
use a specific command that notifies us when it doesn't know the answer. We call this
variable the "no_update_value", and for our current case, we set it to "N/A". Therefore,
we expect the result to only contain the following fields:
{
 {key = action, value = search}
 {key = action_input, value = "How to use this class?"}.
}`

	testCases := []struct {
		noUpdateValue string
		outputKeys    map[string]string
		expected      map[string]string
	}{
		{
			noUpdateValue: "",
			outputKeys: map[string]string{
				"action":       "Action",
				"action_input": "Action Input",
			},
			expected: map[string]string{
				"action":       "Search",
				"action_input": "How to use this class?",
			},
		},
		{
			noUpdateValue: "Search",
			outputKeys: map[string]string{
				"action":       "Action",
				"action_input": "Action Input",
			},
			expected: map[string]string{
				"action_input": "How to use this class?",
			},
		},
	}

	for _, tc := range testCases {
		parser := outputparser.NewRegexDict(tc.outputKeys, tc.noUpdateValue)

		actual, err := parser.Parse(testText)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !reflect.DeepEqual(actual, tc.expected) {
			t.Errorf("Expected %v, got %v", tc.expected, actual)
		}
	}
}

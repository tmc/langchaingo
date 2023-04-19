package outputParsers

import (
	"reflect"
	"testing"
)

type testCase struct {
	input          string
	expectedOutput []string
}

var tests = []testCase{
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

func TestCommaSeparatedListOutputParser(t *testing.T) {
	parser := NewCommaSeparatedList()

	for _, test := range tests {
		output, err := parser.Parse(test.input)
		if err != nil {
			t.Logf("Unexpected error %s", err.Error())
		}

		if !reflect.DeepEqual(output, test.expectedOutput) {
			t.Logf("Parsing with comma separated list did not get expected. Got: %v. Expected %v", output, test.expectedOutput)
		}
	}
}

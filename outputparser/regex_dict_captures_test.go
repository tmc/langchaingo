package outputparser_test

import (
	"reflect"
	"testing"

	"github.com/tmc/langchaingo/outputparser"
)

func TestRegexDictCapturingFormat(t *testing.T) {
	t.Parallel()
	for _, format := range []string{"(Action|Task)", "(?P<label>Action|Task)", "((Action)|Task)"} {
		t.Run(format, func(t *testing.T) {
			t.Parallel()
			parser := outputparser.NewRegexDict(map[string]string{"action": format}, "N/A")
			got, err := parser.Parse("Action: Search")
			if err != nil {
				t.Fatal(err)
			}
			if want := map[string]string{"action": "Search"}; !reflect.DeepEqual(got, want) {
				t.Fatalf("got %#v, want %#v", got, want)
			}
			got, err = parser.Parse("Task: N/A")
			if err != nil {
				t.Fatal(err)
			}
			if want := map[string]string{}; !reflect.DeepEqual(got, want) {
				t.Fatalf("got %#v, want %#v", got, want)
			}
		})
	}
}

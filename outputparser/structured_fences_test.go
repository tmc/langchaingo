package outputparser

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestStructuredFencesInValues(t *testing.T) {
	t.Parallel()
	parser := NewStructured([]ResponseSchema{{Name: "answer", Description: "The answer"}})
	for _, value := range []string{"Use ``` for a code fence", "```go\nfmt.Println(1)\n```", "A ```json example"} {
		t.Run(value, func(t *testing.T) {
			t.Parallel()
			want := map[string]string{"answer": value}
			encoded, err := json.Marshal(want)
			if err != nil {
				t.Fatal(err)
			}
			got, err := parser.Parse("```json\n" + string(encoded) + "\n```")
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("got %#v, want %#v", got, want)
			}
			got, err = parser.ParseWithPrompt("```json\n"+string(encoded)+"\n```", nil)
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("ParseWithPrompt got %#v, %v; want %#v", got, err, want)
			}
		})
	}
}

func TestStructuredFenceBoundaries(t *testing.T) {
	t.Parallel()
	parser := NewStructured([]ResponseSchema{{Name: "answer"}})
	for _, tc := range []struct {
		name    string
		text    string
		wantErr bool
	}{
		{"missing closing fence", "```json\n{\"answer\": \"ok\"}", true},
		{"garbage before closing fence", "```json\n{\"answer\": \"ok\"} invalid\n```", true},
		{"invalid JSON", "```json\n{\"answer\": }\n```", true},
		{"surrounding prose", "Result:\n```json\n{\"answer\": \"ok\"}\n```\nExplanation.", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := parser.Parse(tc.text)
			if (err != nil) != tc.wantErr {
				t.Fatalf("got %#v, %v; want error: %v", got, err, tc.wantErr)
			}
			if !tc.wantErr && !reflect.DeepEqual(got, map[string]string{"answer": "ok"}) {
				t.Fatalf("got %#v, want answer ok", got)
			}
		})
	}
}

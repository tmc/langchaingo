package outputparser

import (
	"fmt"
	"testing"
)

func TestDefined(t *testing.T) {
	t.Parallel()
	type Shape struct {
		Name     string `json:"shapeName" describe:"shape name"`
		NumSides int    `json:"numSides" describe:"number of sides"`
	}

	tests := map[string]struct {
		input    any
		expected string
		wantErr  bool
	}{
		"non-struct": {
			input:   []struct{}{},
			wantErr: true,
		},
		"empty struct": {
			input:   struct{}{},
			wantErr: true,
		},
		"no fields with tag": {
			input:   struct{ Field string }{},
			wantErr: true,
		},
		"string field": {
			input: struct {
				Color string `json:"color" describe:"shape color"`
			}{},
			expected: "interface _Root {\n\tcolor: string; // shape color\n}",
		},
		"anonymous struct field": {
			input: struct {
				Shape struct {
					Color string `describe:"color"` // json tag omitted
				} `json:"shape" describe:"most common 4 sided shape"`
			}{},
			expected: `interface _Root {
	shape: Shape; // most common 4 sided shape
}
interface Shape {
	Color: string; // color
}`,
		},
		"named struct field": {
			input: struct {
				Shape Shape `json:"shape" describe:"most common 4 sided shape"`
			}{},
			expected: `interface _Root {
	shape: Shape; // most common 4 sided shape
}
interface Shape {
	shapeName: string; // shape name
	numSides: int; // number of sides
}`,
		},
		"string array field": {
			input: struct {
				Foods []string `json:"foods" describe:"top 5 foods in the world"`
			}{},
			expected: "interface _Root {\n\tfoods: string[]; // top 5 foods in the world\n}",
		},
		"array-of-structs field": {
			input: struct {
				Foods []struct {
					Name string `json:"name"`
					Temp int    `json:"temp" describe:"temperature usually served at"`
				} `json:"foods" describe:"top 5 foods in the world"`
			}{},
			expected: `interface _Root {
	foods: Foods[]; // top 5 foods in the world
}
interface Foods {
	name: string;
	temp: int; // temperature usually served at
}`,
		},
	}

	for name, test := range tests {
		if output, err := NewDefined(test.input); test.wantErr && err == nil {
			t.Errorf("%s: missing expected error", name)
		} else if !test.wantErr && err != nil {
			t.Errorf("%s: %v", name, err)
		} else if output.schema != test.expected {
			t.Errorf("got '%s'; want '%s'", output, test.expected)
		}
	}
}

func TestDefinedParse(t *testing.T) {
	t.Parallel()
	var book struct {
		Chapters []struct {
			Title string `json:"title" describe:"chapter title"`
		} `json:"chapters" describe:"chapters"`
	}
	parser, newErr := NewDefined(book)
	if newErr != nil {
		t.Error(newErr)
	}

	titles := []string{
		"A Hello There",
		"The Meaty Middle",
		"The Grand Finale",
	}

	output, parseErr := parser.Parse(fmt.Sprintf("```json\n%s\n```", fmt.Sprintf(
		`{"chapters": [{"title": "%s"}, {"title": "%s"}, {"title": "%s"}]}`, titles[0], titles[1], titles[2],
	)))
	if parseErr != nil {
		t.Error(parseErr)
	}
	if count := len(output.Chapters); count != 3 {
		t.Errorf("got %d chapters; want 3", count)
	}
	for i, chapter := range output.Chapters {
		title := titles[i]
		if chapter.Title != titles[i] {
			t.Errorf("got '%s'; want '%s'", chapter.Title, title)
		}
	}
}

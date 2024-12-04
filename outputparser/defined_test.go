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
		"not tagged with describe": {
			input: struct {
				Color string
				Size  int `json:"size"`
			}{},
			expected: "interface _Root {\n\tColor: string;\n\tsize: int;\n}",
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
			t.Errorf("got '%s'; want '%s'", output.schema, test.expected)
		}
	}
}

type book struct {
	Chapters []struct {
		Title string `json:"title" describe:"chapter title"`
	} `json:"chapters" describe:"chapters"`
}

func getParseTests() map[string]struct {
	input    string
	expected *book
	wantErr  bool
} {
	titles := []string{
		"A Hello There",
		"The Meaty Middle",
		"The Grand Finale",
	}

	return map[string]struct {
		input    string
		expected *book
		wantErr  bool
	}{
		"empty": {
			input:    "",
			wantErr:  true,
			expected: nil,
		},
		"invalid": {
			input:    "invalid",
			wantErr:  true,
			expected: nil,
		},
		"valid": {
			input: fmt.Sprintf("```json\n%s\n```", fmt.Sprintf(
				`{"chapters": [{"title": "%s"}, {"title": "%s"}, {"title": "%s"}]}`, titles[0], titles[1], titles[2],
			)),
			wantErr: false,
			expected: &book{
				Chapters: []struct {
					Title string `json:"title" describe:"chapter title"`
				}{
					{Title: titles[0]},
					{Title: titles[1]},
					{Title: titles[2]},
				},
			},
		},
		"valid-without-json-tag": {
			input: fmt.Sprintf("```\n%s\n```", fmt.Sprintf(
				`{"chapters": [{"title": "%s"}, {"title": "%s"}, {"title": "%s"}]}`, titles[0], titles[1], titles[2],
			)),
			wantErr: false,
			expected: &book{
				Chapters: []struct {
					Title string `json:"title" describe:"chapter title"`
				}{
					{Title: titles[0]},
					{Title: titles[1]},
					{Title: titles[2]},
				},
			},
		},
		"valid-without-tags": {
			input: fmt.Sprintf("\n%s\n", fmt.Sprintf(
				`{"chapters": [{"title": "%s"}, {"title": "%s"}, {"title": "%s"}]}`, titles[0], titles[1], titles[2],
			)),
			wantErr: false,
			expected: &book{
				Chapters: []struct {
					Title string `json:"title" describe:"chapter title"`
				}{
					{Title: titles[0]},
					{Title: titles[1]},
					{Title: titles[2]},
				},
			},
		},
		"llm-explanation-and-tags": {
			input: fmt.Sprintf("Sure! Here's the JSON:\n\n```json\n%s\n```\n\nLet me know if you need anything else.", fmt.Sprintf(
				`{"chapters": [{"title": "%s"}, {"title": "%s"}, {"title": "%s"}]}`, titles[0], titles[1], titles[2],
			)),
			wantErr: false,
			expected: &book{
				Chapters: []struct {
					Title string `json:"title" describe:"chapter title"`
				}{
					{Title: titles[0]},
					{Title: titles[1]},
					{Title: titles[2]},
				},
			},
		},
		"llm-explanation-and-valid": {
			input: fmt.Sprintf("Sure! Here's the JSON:\n\n%s\n\nLet me know if you need anything else.", fmt.Sprintf(
				`{"chapters": [{"title": "%s"}, {"title": "%s"}, {"title": "%s"}]}`, titles[0], titles[1], titles[2],
			)),
			wantErr: false,
			expected: &book{
				Chapters: []struct {
					Title string `json:"title" describe:"chapter title"`
				}{
					{Title: titles[0]},
					{Title: titles[1]},
					{Title: titles[2]},
				},
			},
		},
	}
}

func TestDefinedParse(t *testing.T) {
	t.Parallel()
	for name, test := range getParseTests() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			parser, newErr := NewDefined(book{})
			if newErr != nil {
				t.Error(newErr)
			}
			output, parseErr := parser.Parse(test.input)
			switch {
			case parseErr != nil && !test.wantErr:
				t.Errorf("%s: unexpected error: %v", name, parseErr)
			case parseErr == nil && test.wantErr:
				t.Errorf("%s: expected error", name)
			case parseErr == nil && test.expected != nil:
				if count := len(output.Chapters); count != len(test.expected.Chapters) {
					t.Errorf("%s: got %d chapters; want %d", name, count, len(test.expected.Chapters))
				}
				for i, chapter := range output.Chapters {
					title := test.expected.Chapters[i].Title
					if chapter.Title != title {
						t.Errorf("%s: got '%s'; want '%s'", name, chapter.Title, title)
					}
				}
			}
		})
	}
}

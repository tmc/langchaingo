package prompts

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// ExampleSelector is not included because ExampleSelector is not implemented.
type testCaseStruct struct {
	name          string
	examplePrompt PromptTemplate
	examples      []map[string]string
	prefix        string
	suffix        string
	input         map[string]interface{}
	partialInput  map[string]interface{}
	options       FewShotCallOptions
	wantErr       bool
	expected      string
}

func TestFewShotPrompt_Format(t *testing.T) {
	examplePrompt := NewPromptTemplate("{{.question}}: {{.answer}}", []string{"question", "answer"})
	t.Parallel()
	testCases := []testCaseStruct{
		{
			"prefix only", examplePrompt,
			[]map[string]string{},
			"This is a {{.foo}} test.", "",
			map[string]interface{}{"foo": "bar"},
			nil,
			FewShotCallOptions{},
			false, "This is a bar test.",
		},
		{
			"suffix only", examplePrompt,
			[]map[string]string{},
			"", "This is a {{.foo}} test.",
			map[string]interface{}{"foo": "bar"},
			nil,
			FewShotCallOptions{},
			false, "This is a bar test.",
		},
		{
			"insufficient InputVariables w err",
			examplePrompt,
			[]map[string]string{},
			"",
			"This is a {{.foo}} test.",
			map[string]interface{}{"bar": "bar"},
			nil,
			FewShotCallOptions{},
			true,
			`template: template:1:12: executing "template" at <.foo>: map has no entry for key "foo"`,
		},
		{
			"inputVariables neither Examples nor ExampleSelector w err",
			examplePrompt,
			nil,
			"",
			"",
			map[string]interface{}{"bar": "bar"},
			nil,
			FewShotCallOptions{},
			true,
			ErrNoExample.Error(),
		},
		{
			"functionality test",
			examplePrompt,
			[]map[string]string{{"question": "foo", "answer": "bar"}, {"question": "baz", "answer": "foo"}},
			"This is a test about {{.content}}.",
			"Now you try to talk about {{.new_content}}.",
			map[string]interface{}{"content": "animals", "new_content": "party"},
			nil,
			FewShotCallOptions{ExampleSeparator: "\n"},
			false,
			"This is a test about animals.\nfoo: bar\nbaz: foo\nNow you try to talk about party.",
		},
		{
			"functionality test with partial input",
			examplePrompt,
			[]map[string]string{{"question": "foo", "answer": "bar"}, {"question": "baz", "answer": "foo"}},
			"This is a test about {{.content}}.",
			"Now you try to talk about {{.new_content}}.",
			map[string]interface{}{"content": "animals"},
			map[string]interface{}{"new_content": func() string { return "party" }},
			FewShotCallOptions{ExampleSeparator: "\n"},
			false,
			"This is a test about animals.\nfoo: bar\nbaz: foo\nNow you try to talk about party.",
		},
		{
			"invalid template w err",
			examplePrompt,
			[]map[string]string{{"question": "foo", "answer": "bar"}, {"question": "baz", "answer": "foo"}},
			"This is a test about {{.wrong_content}}.",
			"Now you try to talk about {{.new_content}}.",
			map[string]interface{}{"content": "animals"},
			map[string]interface{}{"new_content": func() string { return "party" }},
			FewShotCallOptions{ExampleSeparator: "\n"},
			true,
			"template: template:1:23: executing \"template\" at <.wrong_content>: map has no entry for key " +
				"\"wrong_content\"",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runTest(t, tc)
		})
	}
}

func runTest(t *testing.T, tc struct {
	name          string
	examplePrompt PromptTemplate
	examples      []map[string]string
	prefix        string
	suffix        string
	input         map[string]interface{}
	partialInput  map[string]interface{}
	options       FewShotCallOptions
	wantErr       bool
	expected      string
},
) {
	t.Helper()
	p, err := NewFewShotPrompt(tc.examplePrompt, tc.examples, nil, tc.prefix, tc.suffix, tc.input,
		tc.partialInput, tc.options)
	if tc.wantErr {
		checkError(t, err, tc.expected)
		return
	}
	fp, err := p.Format(tc.input)
	if checkError(t, err, "") {
		return
	}
	got := fmt.Sprint(fp)
	if diff := cmp.Diff(tc.expected, got); diff != "" {
		t.Errorf("unexpected prompt output (-want +got):\n%s", diff)
	}
}

func checkError(t *testing.T, err error, expected string) bool {
	t.Helper()
	if err != nil {
		if expected != "" && err.Error() != expected {
			t.Errorf("unexpected error: got %q, want %q", err.Error(), expected)
		}
		return true
	}
	return false
}

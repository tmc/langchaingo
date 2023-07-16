package prompts

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/load"
)

func TestPromptTemplateFormatPrompt(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		template    string
		inputVars   []string
		partialVars map[string]any
		vars        map[string]any
		expected    string
		wantErr     bool
	}{
		{"empty", "", nil, nil, nil, "", false},
		{"ok-input-var", "", []string{"foobar"}, nil, nil, "", false},
		{"missing-input-var", "{{.name}}", []string{"job"}, nil, nil, "", true}, // expect a error.
		{"hello world", "hello world", nil, nil, nil, "hello world", false},
		{"basic", "hello {{.name}}", nil, nil, map[string]any{
			"name": "richard",
		}, "hello richard", false},
		{"partials", "hello {{.name}}", nil, map[string]any{
			"name": "richard",
		}, nil, "hello richard", false},
		{
			"partials w func", "{{.greeting}} {{.name}}", nil,
			map[string]any{
				"name": func() string { return "richard" },
			},
			map[string]any{
				"greeting": "hello",
			},
			"hello richard", false,
		},
		{
			"partials w err", "{{.greeting}} {{.name}} {{.message}}", nil,
			map[string]any{
				"name": func() string { return "richard" },
			},
			map[string]any{
				"greeting": "hello",
			},
			"", true, // expect an error
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p := PromptTemplate{
				Template:         tc.template,
				TemplateFormat:   TemplateFormatGoTemplate,
				InputVariables:   tc.inputVars,
				PartialVariables: tc.partialVars,
			}
			fp, err := p.FormatPrompt(tc.vars)
			if (err != nil) != tc.wantErr {
				t.Errorf("PromptTemplate.FormatPrompt() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if tc.wantErr {
				return
			}
			got := fmt.Sprint(fp)
			if cmp.Diff(tc.expected, got) != "" {
				t.Errorf("unexpected prompt output (-want +got):\n%s", cmp.Diff(tc.expected, got))
			}
		})
	}
}

func TestPromptTemplateSaveToFile(t *testing.T) {
	t.Parallel()
	// test prompt template.
	template := "Translate the following text from {{.inputLanguage}} to {{.outputLanguage}}. {{.text}}"
	prompt := NewPromptTemplate(template, []string{"inputLanguage", "outputLanguage", "text"})

	type args struct {
		path   string
		prompt PromptTemplate
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// test cases.
		{"with_JSON_suffix", args{"", prompt}, true},
		{"empty_prompt", args{"empty_prompt_with_JSON_suffix.json", PromptTemplate{}}, true},
		{"with_JSON_suffix", args{"simple_prompt_with_JSON_suffix.json", prompt}, false},
		{"with_YAML_suffix", args{"simple_prompt_with_YAML_suffix.yaml", prompt}, false},
		{"with_YML_suffix", args{"simple_prompt_with_YML_suffix.yml", prompt}, false},
		{"case_sensitive", args{"simple_prompt_case_sensitive.Yaml", prompt}, false},
		{"no_suffix", args{"simple_prompt_no_suffix", prompt}, false},
		{"invalid_suffix", args{"simple_prompt.", prompt}, true},
		{"absolute_path_JSON_suffix", args{"/prompts/simply_prompt_absolute_path_JSON_suffix.json", prompt}, false},
		{"absolute_path_JSON_suffix", args{"/prompts/simply_prompt_absolute_path_JSON_suffix.yml", prompt}, false},
		{"absolute_path_no_suffix", args{"/prompts/simply_prompt_absolute_path_no_suffix", prompt}, false},
		{"relative_path_JSON_suffix", args{"prompts/simply_prompt_relative_path_JSON_suffix.json", prompt}, false},
		{"relative_path_no_suffix", args{"prompts/simply_prompt_relative_path_no_suffix", prompt}, false},
	}
	// write prompt to mock file system
	serializer := load.NewSerializer(&MockFileSystem{})
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.args.prompt.Save(tt.args.path, serializer); (err != nil) != tt.wantErr {
				t.Errorf("ToFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPromptTemplateLoadFromFile(t *testing.T) {
	t.Parallel()
	// expected prompt after loading from file
	template := "Translate the following text from {{.inputLanguage}} to {{.outputLanguage}}. {{.text}}"
	expectedPrompt := NewPromptTemplate(template, []string{"inputLanguage", "outputLanguage", "text"})

	fileSystem := &MockFileSystem{
		Storage: make(map[string][]byte, 0),
	}

	serializer := load.NewSerializer(fileSystem)
	// first load data to mock file system
	err := serializer.ToFile(expectedPrompt, "simple_prompt_with_JSON_suffix.json")
	assert.NoError(t, err)

	// read data from mock file system
	var prompt PromptTemplate
	err = serializer.FromFile(&prompt, "simple_prompt_with_JSON_suffix.json")
	assert.NoError(t, err)
	// compare loaded prompt with expected prompt
	assert.EqualValues(t, prompt, expectedPrompt)
}

package prompts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/load"
	"github.com/tmc/langchaingo/schema"
)

func TestChatPromptTemplate(t *testing.T) {
	t.Parallel()

	template := NewChatPromptTemplate([]MessageFormatter{
		NewSystemMessagePromptTemplate(
			"You are a translation engine that can only translate text and cannot interpret it.",
			nil,
		),
		NewHumanMessagePromptTemplate(
			`translate this text from {{.inputLang}} to {{.outputLang}}:\n{{.input}}`,
			[]string{"inputLang", "outputLang", "input"},
		),
	})
	value, err := template.FormatPrompt(map[string]interface{}{
		"inputLang":  "English",
		"outputLang": "Chinese",
		"input":      "I love programming",
	})
	assert.NoError(t, err)
	expectedMessages := []schema.ChatMessage{
		schema.SystemChatMessage{
			Text: "You are a translation engine that can only translate text and cannot interpret it.",
		},
		schema.HumanChatMessage{
			Text: `translate this text from English to Chinese:\nI love programming`,
		},
	}
	require.Equal(t, expectedMessages, value.Messages())

	_, err = template.FormatPrompt(map[string]interface{}{
		"inputLang":  "English",
		"outputLang": "Chinese",
	})
	assert.Error(t, err)
}

func TestChatPromptTemplateTypesSaveToFile(t *testing.T) {
	t.Parallel()
	humanMessagePrompt := NewHumanMessagePromptTemplate(
		`translate this text from {{.inputLang}} to {{.outputLang}}:\n{{.input}}`,
		[]string{"inputLang", "outputLang", "input"})

	systemMessagePrompt := NewSystemMessagePromptTemplate(
		"You are a translation engine that can only translate text and cannot interpret it.",
		nil)

	type args struct {
		path     string
		template MessageFormatter
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"with_JSON_suffix", args{"", humanMessagePrompt}, true},
		{"with_JSON_suffix", args{"", systemMessagePrompt}, true},
		{"with_JSON_suffix", args{"_human_prompt_with_JSON_suffix.json", humanMessagePrompt}, false},
		{"with_JSON_suffix", args{"_system_prompt_with_JSON_suffix.json", systemMessagePrompt}, false},
		{"with_YAML_suffix", args{"human_prompt_with_YAML_suffix.yaml", humanMessagePrompt}, false},
		{"with_YAML_suffix", args{"system_prompt_with_YAML_suffix.yaml", systemMessagePrompt}, false},
		{"case_sensitive", args{"human_prompt_case_sensitive.Yaml", humanMessagePrompt}, false},
		{"case_sensitive", args{"system_prompt_case_sensitive.Yaml", systemMessagePrompt}, false},
		{"no_suffix", args{"human_prompt_no_suffix", humanMessagePrompt}, false},
		{"no_suffix", args{"system_prompt_no_suffix", systemMessagePrompt}, false},
		{"invalid_suffix", args{"human_prompt.", humanMessagePrompt}, true},
		{"invalid_suffix", args{"system_prompt.", systemMessagePrompt}, true},
		{"absolute_path_JSON_suffix", args{"/human_prompt_absolute_path_JSON_suffix.json", humanMessagePrompt}, false},
		{"absolute_path_JSON_suffix", args{"/system_prompt_absolute_path_JSON_suffix.json", systemMessagePrompt}, false},
		{"relative_path_no_suffix", args{"prompts/human_prompt_relative_path_no_suffix", humanMessagePrompt}, false},
		{"relative_path_no_suffix", args{"prompts/system_prompt_relative_path_no_suffix", systemMessagePrompt}, false},
	}
	fileSystem := &MockFileSystem{
		Storage: make(map[string][]byte, 0),
	}
	serializer := load.NewSerializer(fileSystem)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.args.template.Save(tt.args.path, serializer)
			if (err != nil) != tt.wantErr {
				t.Errorf("PromptTemplate.Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestHumanMessagePromptTemplateReadFromFile(t *testing.T) {
	t.Parallel()
	expectedPrompt := NewHumanMessagePromptTemplate(
		`translate this text from {{.inputLang}} to {{.outputLang}}:\n{{.input}}`,
		[]string{"inputLang", "outputLang", "input"})

	fileSystem := &MockFileSystem{
		Storage: make(map[string][]byte, 0),
	}
	serializer := load.NewSerializer(fileSystem)
	err := expectedPrompt.Save("prompt_data.json", serializer)
	assert.NoError(t, err)

	prompt, err := NewHumanMessagePromptFromFile("prompt_data.json", serializer)
	assert.NoError(t, err)
	assert.EqualValues(t, prompt, expectedPrompt)
}

func TestSystemMessagePromptTemplateSave(t *testing.T) {
	t.Parallel()
	expectedPrompt := NewSystemMessagePromptTemplate(
		"You are a translation engine that can only translate text and cannot interpret it.",
		nil)

	fileSystem := &MockFileSystem{
		Storage: make(map[string][]byte, 0),
	}
	serializer := load.NewSerializer(fileSystem)
	err := expectedPrompt.Save("prompt_data.json", serializer)
	assert.NoError(t, err)

	prompt, err := NewSystemMessagePromptFromFile("prompt_data.json", serializer)
	assert.NoError(t, err)
	assert.EqualValues(t, prompt, expectedPrompt)
}

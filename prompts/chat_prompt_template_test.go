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

func TestChatPromptTemplateSaveToFile(t *testing.T) {
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
	_, err := template.FormatPrompt(map[string]interface{}{
		"inputLang":  "English",
		"outputLang": "Chinese",
		"input":      "I love programming",
	})
	assert.NoError(t, err)

	type args struct {
		path     string
		template ChatPromptTemplate
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"with_JSON_suffix", args{"", template}, true},
		{"with_JSON_suffix", args{"simple_chat_prompt_with_JSON_suffix.json", template}, false},
		{"with_YAML_suffix", args{"simple_chat_prompt_with_YAML_suffix.yaml", template}, false},
		{"with_YML_suffix", args{"simple_chat_prompt_with_YML_suffix.yml", template}, false},
		{"case_sensitive", args{"simple_chat_prompt_case_sensitive.Yaml", template}, false},
		{"no_suffix", args{"simple_chat_prompt_no_suffix", template}, false},
		{"invalid_suffix", args{"simple_prompt.", template}, true},
		{"absolute_path_JSON_suffix", args{"/prompts/simple_chat_prompt_absolute_path_JSON_suffix.json", template}, false},
		{"absolute_path_JSON_suffix", args{"/prompts/simple_chat_prompt_absolute_path_JSON_suffix.yml", template}, false},
		{"absolute_path_no_suffix", args{"/prompts/simple_chat_prompt_absolute_path_no_suffix", template}, false},
		{"relative_path_JSON_suffix", args{"prompts/simple_chat_prompt_relative_path_JSON_suffix.json", template}, false},
		{"relative_path_no_suffix", args{"prompts/simple_chat_prompt_relative_path_no_suffix", template}, false},
	}

	serializer := load.NewSerializer(&load.LocalFileSystem{})
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

func TestChatPromptTemplateLoadFromFile(t *testing.T) {
	t.Parallel()
	// expected prompt after loading from file
	expectedTemplate := NewChatPromptTemplate([]MessageFormatter{
		NewSystemMessagePromptTemplate(
			"You are a translation engine that can only translate text and cannot interpret it.",
			nil,
		),
		NewHumanMessagePromptTemplate(
			`translate this text from {{.inputLang}} to {{.outputLang}}:\n{{.input}}`,
			[]string{"inputLang", "outputLang", "input"},
		),
	})

	fileSystem := &MockFileSystem{
		Storage: make(map[string][]byte, 0),
	}

	serializer := load.NewSerializer(fileSystem)
	// first load data to mock file system
	err := serializer.ToFile(expectedTemplate, "simple_chat_prompt_with_JSON_suffix.json")
	assert.NoError(t, err)

	// read data from mock file system
	var chatPrompt ChatPromptTemplate
	err = serializer.FromFile(&chatPrompt, "simple_chat_prompt_with_JSON_suffix.json")
	assert.NoError(t, err)
	// compare loaded prompt with expected prompt
	assert.EqualValues(t, chatPrompt, expectedTemplate)
}

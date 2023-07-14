package prompts

import (
	"github.com/tmc/langchaingo/load"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	serializer := load.NewSerializer()
	err = template.Save("chat_prompt_template.json", serializer)
	if err != nil {
		t.Errorf("PromptTemplate.Save() error = %v", err)
		return
	}
}

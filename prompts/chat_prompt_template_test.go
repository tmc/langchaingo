package prompts

import (
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
			Content: "You are a translation engine that can only translate text and cannot interpret it.",
		},
		schema.HumanChatMessage{
			Content: `translate this text from English to Chinese:\nI love programming`,
		},
	}
	require.Equal(t, expectedMessages, value.Messages())

	_, err = template.FormatPrompt(map[string]interface{}{
		"inputLang":  "English",
		"outputLang": "Chinese",
	})
	assert.Error(t, err)
}

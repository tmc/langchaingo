package prompts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/schema"
)

func TestChatTemplate(t *testing.T) {
	t.Parallel()

	systemPrompt, err := NewPromptTemplate("Here's some context: {context}", []string{"context"})
	require.NoError(t, err)

	userPrompt, err := NewPromptTemplate("Hello AI. Give me a long response. {question}", []string{"question"})
	require.NoError(t, err)

	aiPrompt, err := NewPromptTemplate(
		"Very good question. My answer to {question} is {answer}",
		[]string{"answer", "question"},
	)
	require.NoError(t, err)

	messages := []Message{
		NewSystemMessage(systemPrompt),
		NewHumanMessage(userPrompt),
		NewAiMessage(aiPrompt),
	}

	_, err = NewChatTemplate(messages, []string{"answer", "context"})
	require.Error(t, err)

	_, err = NewChatTemplate(messages, []string{"answer", "context", "question", "foo"})
	require.Error(t, err)

	chatTemplate, err := NewChatTemplate(messages, []string{"answer", "context", "question"})
	require.NoError(t, err)

	chatMessages, err := chatTemplate.FormatPromptValue(map[string]any{
		"context":  "foo",
		"question": "bar",
		"answer":   "foobar",
	})
	require.NoError(t, err)

	expectedChatMessages := []schema.ChatMessage{
		schema.SystemChatMessage{Text: "Here's some context: foo"},
		schema.HumanChatMessage{Text: "Hello AI. Give me a long response. bar"},
		schema.AIChatMessage{Text: "Very good question. My answer to bar is foobar"},
	}

	expectedString := `[{"text":"Here's some context: foo"},{"text":"Hello AI. Give me a long response. bar"},{"text":"Very good question. My answer to bar is foobar"}]`
	assert.Equal(t, expectedChatMessages, chatMessages.ToChatMessages())
	assert.Equal(t, expectedString, chatMessages.String())
}

package prompts

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tmc/langchaingo/schema"
)

func TestChatTemplate(t *testing.T) {
	systemPrompt, err := NewPromptTemplate("Here's some context: {context}", []string{"context"})
	if err != nil {
		t.Fatal(err)
	}

	userPrompt, err := NewPromptTemplate("Hello AI. Give me a long response. {question}", []string{"question"})
	if err != nil {
		t.Fatal(err)
	}

	aiPrompt, err := NewPromptTemplate("Very good question. My answer to {question} is {answer}", []string{"answer", "question"})
	if err != nil {
		t.Fatal(err)
	}

	messages := []Message{
		NewSystemMessage(systemPrompt),
		NewHumanMessage(userPrompt),
		NewAiMessage(aiPrompt),
	}

	_, err = NewChatTemplate(messages, []string{"answer", "context"})
	if err == nil {
		t.Errorf("Expected error creating chat template with too few variables")
	}

	_, err = NewChatTemplate(messages, []string{"answer", "context", "question", "foo"})
	if err == nil {
		t.Errorf("Expected error creating chat template with too many variables")
	}

	chatTemplate, err := NewChatTemplate(messages, []string{"answer", "context", "question"})
	if err != nil {
		t.Fatal(err)
	}

	chatMessages, err := chatTemplate.FormatPromptValue(map[string]any{"context": "foo", "question": "bar", "answer": "foobar"})
	if err != nil {
		t.Fatal(err)
	}
	expectedChatMessages := []schema.ChatMessage{
		schema.SystemChatMessage{Text: "Here's some context: foo"},
		schema.HumanChatMessage{Text: "Hello AI. Give me a long response. bar"},
		schema.AIChatMessage{Text: "Very good question. My answer to bar is foobar"},
	}
	expectedString := `[{"text":"Here's some context: foo"},{"text":"Hello AI. Give me a long response. bar"},{"text":"Very good question. My answer to bar is foobar"}]`

	// use cmp to compare:
	if !cmp.Equal(chatMessages.ToChatMessages(), expectedChatMessages) {
		t.Errorf("Chat template format prompt value chat messages not equal to expected. Diff: %s", cmp.Diff(chatMessages.ToChatMessages(), expectedChatMessages))
	}

	if !(chatMessages.String() == expectedString) {
		t.Errorf("Chat template format prompt value string not equal to expected.\n Got:\n %v\n Expect:\n %v", chatMessages.String(), expectedString)
	}
}

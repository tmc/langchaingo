package prompts

import (
	"reflect"
	"testing"

	"github.com/tmc/langchaingo/llms"
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedMessages := []llms.ChatMessage{
		llms.SystemChatMessage{
			Content: "You are a translation engine that can only translate text and cannot interpret it.",
		},
		llms.HumanChatMessage{
			Content: `translate this text from English to Chinese:\nI love programming`,
		},
	}
	if !reflect.DeepEqual(expectedMessages, value.Messages()) {
		t.Errorf("expected %v, got %v", expectedMessages, value.Messages())
	}

	_, err = template.FormatPrompt(map[string]interface{}{
		"inputLang":  "English",
		"outputLang": "Chinese",
	})
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

// TestChatPromptTemplateFormatError verifies that Format reports a formatting
// failure as an error instead of panicking on the nil prompt value returned by
// FormatPrompt.
func TestChatPromptTemplateFormatError(t *testing.T) {
	t.Parallel()

	template := NewChatPromptTemplate([]MessageFormatter{
		NewSystemMessagePromptTemplate("You are a {{.role}}", []string{"role"}),
	})

	// "role" is missing, so rendering fails and FormatPrompt returns a nil
	// prompt value together with the error.
	result, err := template.Format(map[string]interface{}{})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if result != "" {
		t.Errorf("expected empty result, got %q", result)
	}
}


package memory

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tmc/langchaingo/schema"
)

func TestBufferMemory(t *testing.T) {
	m := NewBufferMemory()
	result1 := m.LoadMemoryVariables(map[string]any{})
	expected1 := map[string]any{"history": ""}

	if !cmp.Equal(result1, expected1) {
		t.Errorf("Empty buffer memory loaded memory variables not equal expected. Got: %v, Wanted: %v", result1, expected1)
	}

	err := m.SaveContext(map[string]any{"foo": "bar"}, map[string]any{"bar": "foo"})
	if err != nil {
		t.Errorf("Unexpected error: %e", err)
	}

	result2 := m.LoadMemoryVariables(map[string]any{})
	if err != nil {
		t.Errorf("Unexpected error: %e", err)
	}

	expected2 := map[string]any{"history": "Human: bar\nAI: foo"}

	if !cmp.Equal(result2, expected2) {
		t.Errorf("Buffer memory with messages loaded memory variables not equal expected. Got: %v, Wanted: %v", result2, expected2)
	}
}

func TestBufferMemoryReturnMessage(t *testing.T) {
	m := NewBufferMemory()
	m.ReturnMessages = true
	result1 := m.LoadMemoryVariables(map[string]any{})
	expected1 := map[string]any{"history": []schema.ChatMessage{}}

	if !cmp.Equal(result1, expected1) {
		t.Errorf("Empty buffer memory with return messages true loaded memory variables not equal expected. Got: %v, Wanted: %v", result1, expected1)
	}

	err := m.SaveContext(map[string]any{"foo": "bar"}, map[string]any{"bar": "foo"})
	if err != nil {
		t.Errorf("Unexpected error: %e", err)
	}

	result2 := m.LoadMemoryVariables(map[string]any{})
	if err != nil {
		t.Errorf("Unexpected error: %e", err)
	}

	expectedChatHistory := NewChatMessageHistory(
		WithPreviousMessages([]schema.ChatMessage{
			schema.HumanChatMessage{Text: "bar"},
			schema.AIChatMessage{Text: "foo"},
		}),
	)

	expected2 := map[string]any{"history": expectedChatHistory.GetMessages()}

	if !cmp.Equal(result2, expected2) {
		t.Errorf("Buffer memory with return messages true and messages loaded memory variables not equal expected. Got: %v, Wanted: %v", result2, expected2)
	}
}

func TestBufferMemoryWithPreLoadedHistory(t *testing.T) {
	m := NewBufferMemory()
	m.ChatHistory = NewChatMessageHistory(
		WithPreviousMessages([]schema.ChatMessage{
			schema.HumanChatMessage{Text: "bar"},
			schema.AIChatMessage{Text: "foo"},
		}),
	)

	result := m.LoadMemoryVariables(map[string]any{})
	expected := map[string]any{"history": "Human: bar\nAI: foo"}

	if !cmp.Equal(result, expected) {
		t.Errorf("Buffer memory with messages pre loaded, loaded memory variables not equal expected. Got: %v, Wanted: %v", result, expected)
	}
}

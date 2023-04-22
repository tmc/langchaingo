package memory

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tmc/langchaingo/schema"
)

func TestBufferMemory(t *testing.T) {
	m := NewBuffer()
	expected1 := map[string]any{"history": ""}
	result1, err := m.LoadMemoryVariables(map[string]any{})
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(result1, expected1) {
		t.Fatalf("Empty buffer memory loaded memory variables not equal expected. Got: %v, Wanted: %v", result1, expected1)
	}

	err = m.SaveContext(map[string]any{"foo": "bar"}, map[string]any{"bar": "foo"})
	if err != nil {
		t.Fatal(err)
	}

	result2, err := m.LoadMemoryVariables(map[string]any{})
	if err != nil {
		t.Fatal(err)
	}

	expected2 := map[string]any{"history": "Human: bar\nAI: foo"}

	if !cmp.Equal(result2, expected2) {
		t.Fatalf("Buffer memory with messages loaded memory variables not equal expected. Got: %v, Wanted: %v", result2, expected2)
	}
}

func TestBufferMemoryReturnMessage(t *testing.T) {
	m := NewBuffer()
	m.ReturnMessages = true
	expected1 := map[string]any{"history": []schema.ChatMessage{}}
	result1, err := m.LoadMemoryVariables(map[string]any{})
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(result1, expected1) {
		t.Fatalf("Empty buffer memory with return messages true loaded memory variables not equal expected. Got: %v, Wanted: %v", result1, expected1)
	}

	err = m.SaveContext(map[string]any{"foo": "bar"}, map[string]any{"bar": "foo"})
	if err != nil {
		t.Fatal(err)
	}

	result2, err := m.LoadMemoryVariables(map[string]any{})
	if err != nil {
		t.Fatal(err)
	}

	expectedChatHistory := NewChatMessageHistory(
		WithPreviousMessages([]schema.ChatMessage{
			schema.HumanChatMessage{Text: "bar"},
			schema.AIChatMessage{Text: "foo"},
		}),
	)

	expected2 := map[string]any{"history": expectedChatHistory.Messages()}

	if !cmp.Equal(result2, expected2) {
		t.Fatalf("Buffer memory with return messages true and messages loaded memory variables not equal expected. Got: %v, Wanted: %v", result2, expected2)
	}
}

func TestBufferMemoryWithPreLoadedHistory(t *testing.T) {
	m := NewBuffer()
	m.ChatHistory = NewChatMessageHistory(
		WithPreviousMessages([]schema.ChatMessage{
			schema.HumanChatMessage{Text: "bar"},
			schema.AIChatMessage{Text: "foo"},
		}),
	)

	expected := map[string]any{"history": "Human: bar\nAI: foo"}
	result, err := m.LoadMemoryVariables(map[string]any{})
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(result, expected) {
		t.Fatalf("Buffer memory with messages pre loaded, loaded memory variables not equal expected. Got: %v, Wanted: %v", result, expected)
	}
}

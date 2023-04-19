package memory

import (
	"reflect"
	"testing"

	"github.com/tmc/langchaingo/exp/schema"
)

func TestBufferMemory(t *testing.T) {
	m := NewBufferMemory()
	result1, err := m.LoadMemoryVariables(InputValues{})
	if err != nil {
		t.Errorf("Unexpected error: %e", err)
	}

	expected1 := InputValues{"history": ""}

	if !reflect.DeepEqual(result1, expected1) {
		t.Errorf("Empty buffer memory loaded memory variables not equal expected. Got: %v, Wanted: %v", result1, expected1)
	}

	err = m.SaveContext(InputValues{"foo": "bar"}, InputValues{"bar": "foo"})
	if err != nil {
		t.Errorf("Unexpected error: %e", err)
	}

	result2, err := m.LoadMemoryVariables(InputValues{})
	if err != nil {
		t.Errorf("Unexpected error: %e", err)
	}

	expected2 := InputValues{"history": "Human: bar\nAI: foo"}

	if !reflect.DeepEqual(result2, expected2) {
		t.Errorf("Buffer memory with messages loaded memory variables not equal expected. Got: %v, Wanted: %v", result2, expected2)
	}
}

func TestBufferMemoryReturnMessage(t *testing.T) {
	m := NewBufferMemory()
	m.ReturnMessages = true
	result1, err := m.LoadMemoryVariables(InputValues{})
	if err != nil {
		t.Errorf("Unexpected error: %e", err)
	}

	expected1 := InputValues{"history": []schema.ChatMessage{}}

	if !reflect.DeepEqual(result1, expected1) {
		t.Errorf("Empty buffer memory with return messages true loaded memory variables not equal expected. Got: %v, Wanted: %v", result1, expected1)
	}

	err = m.SaveContext(InputValues{"foo": "bar"}, InputValues{"bar": "foo"})
	if err != nil {
		t.Errorf("Unexpected error: %e", err)
	}

	result2, err := m.LoadMemoryVariables(InputValues{})
	if err != nil {
		t.Errorf("Unexpected error: %e", err)
	}

	expectedChatHistory := NewChatMessageHistory(
		WithPreviousMessages([]schema.ChatMessage{
			schema.HumanChatMessage{Text: "bar"},
			schema.AiChatMessage{Text: "foo"},
		}),
	)

	expected2 := InputValues{"history": expectedChatHistory.GetMessages()}

	if !reflect.DeepEqual(result2, expected2) {
		t.Errorf("Buffer memory with return messages true and messages loaded memory variables not equal expected. Got: %v, Wanted: %v", result2, expected2)
	}
}

func TestBufferMemoryWithPreLoadedHistory(t *testing.T) {
	m := NewBufferMemory()
	m.ChatHistory = NewChatMessageHistory(
		WithPreviousMessages([]schema.ChatMessage{
			schema.HumanChatMessage{Text: "bar"},
			schema.AiChatMessage{Text: "foo"},
		}),
	)
	result, err := m.LoadMemoryVariables(InputValues{})
	if err != nil {
		t.Errorf("Unexpected error: %e", err)
	}

	expected := InputValues{"history": "Human: bar\nAI: foo"}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Buffer memory with messages pre loaded, loaded memory variables not equal expected. Got: %v, Wanted: %v", result, expected)
	}
}

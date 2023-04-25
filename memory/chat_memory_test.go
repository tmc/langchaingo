package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/schema"
)

func TestChatMessageHistory(t *testing.T) {
	h := NewChatMessageHistory()
	h.AddAIMessage("foo")
	h.AddUserMessage("bar")
	assert.Equal(t, []schema.ChatMessage{
		schema.AIChatMessage{Text: "foo"},
		schema.HumanChatMessage{Text: "bar"},
	}, h.Messages())

	h = NewChatMessageHistory(
		WithPreviousMessages([]schema.ChatMessage{
			schema.AIChatMessage{Text: "foo"},
			schema.SystemChatMessage{Text: "bar"},
		}),
	)
	h.AddUserMessage("zoo")
	assert.Equal(t, []schema.ChatMessage{
		schema.AIChatMessage{Text: "foo"},
		schema.SystemChatMessage{Text: "bar"},
		schema.HumanChatMessage{Text: "zoo"},
	}, h.Messages())
}

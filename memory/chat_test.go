package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/schema"
)

func TestChatMessageHistory(t *testing.T) {
	t.Parallel()

	h := NewChatMessageHistory()
	h.AddAIMessage("foo")
	h.AddUserMessage("bar")
	assert.Equal(t, []schema.ChatMessage{
		schema.AIChatMessage{Content: "foo"},
		schema.HumanChatMessage{Content: "bar"},
	}, h.Messages())

	h = NewChatMessageHistory(
		WithPreviousMessages([]schema.ChatMessage{
			schema.AIChatMessage{Content: "foo"},
			schema.SystemChatMessage{Content: "bar"},
		}),
	)
	h.AddUserMessage("zoo")
	assert.Equal(t, []schema.ChatMessage{
		schema.AIChatMessage{Content: "foo"},
		schema.SystemChatMessage{Content: "bar"},
		schema.HumanChatMessage{Content: "zoo"},
	}, h.Messages())
}

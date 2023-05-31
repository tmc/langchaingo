package history

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/schema"
)

func TestSimpleChatMessageHistory(t *testing.T) {
	t.Parallel()

	h := NewSimpleChatMessageHistory()
	h.AddAIMessage("foo")
	h.AddUserMessage("bar")
	msgs, err := h.Messages()
	assert.NoError(t, err)

	assert.Equal(t, []schema.ChatMessage{
		schema.AIChatMessage{Text: "foo"},
		schema.HumanChatMessage{Text: "bar"},
	}, msgs)

	h = NewSimpleChatMessageHistory(
		WithPreviousMessages([]schema.ChatMessage{
			schema.AIChatMessage{Text: "foo"},
			schema.SystemChatMessage{Text: "bar"},
		}),
	)
	h.AddUserMessage("zoo")

	msgs, err = h.Messages()
	assert.NoError(t, err)

	assert.Equal(t, []schema.ChatMessage{
		schema.AIChatMessage{Text: "foo"},
		schema.SystemChatMessage{Text: "bar"},
		schema.HumanChatMessage{Text: "zoo"},
	}, msgs)
}

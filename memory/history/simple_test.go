package history

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/schema"
)

func TestSimpleChatMessageHistory(t *testing.T) {
	t.Parallel()

	h := NewSimpleChatMessageHistory()
	err := h.AddAIMessage("foo")
	assert.NoError(t, err)
	err = h.AddUserMessage("bar")
	assert.NoError(t, err)
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
	err = h.AddUserMessage("zoo")
	assert.NoError(t, err)

	msgs, err = h.Messages()
	assert.NoError(t, err)

	assert.Equal(t, []schema.ChatMessage{
		schema.AIChatMessage{Text: "foo"},
		schema.SystemChatMessage{Text: "bar"},
		schema.HumanChatMessage{Text: "zoo"},
	}, msgs)
}

package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/schema"
)

func TestChatMessageHistory(t *testing.T) {
	t.Parallel()

	h := NewChatMessageHistory()
	err := h.AddAIMessage("foo")
	assert.NoError(t, err)
	err = h.AddUserMessage("bar")
	assert.NoError(t, err)

	messages, err := h.Messages()
	assert.NoError(t, err)

	assert.Equal(t, []schema.ChatMessage{
		schema.AIChatMessage{Content: "foo"},
		schema.HumanChatMessage{Content: "bar"},
	}, messages)

	h = NewChatMessageHistory(
		WithPreviousMessages([]schema.ChatMessage{
			schema.AIChatMessage{Content: "foo"},
			schema.SystemChatMessage{Content: "bar"},
		}),
	)
	err = h.AddUserMessage("zoo")
	assert.NoError(t, err)

	messages, err = h.Messages()
	assert.NoError(t, err)

	assert.Equal(t, []schema.ChatMessage{
		schema.AIChatMessage{Content: "foo"},
		schema.SystemChatMessage{Content: "bar"},
		schema.HumanChatMessage{Content: "zoo"},
	}, messages)
}

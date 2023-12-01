package memory

import (
	"context" //nolint:gci
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/schema"
	"testing" //nolint:gci
)

func TestBufferWindowReturnMessage(t *testing.T) {
	t.Parallel()

	m := NewConversationWindowBuffer()
	m.ReturnMessages = true
	m.K = 1
	expected1 := map[string]any{"history": []schema.ChatMessage{}}
	result1, err := m.LoadMemoryVariables(context.Background(), map[string]any{})
	require.NoError(t, err)
	assert.Equal(t, expected1, result1)

	err = m.SaveContext(context.Background(), map[string]any{"foo": "bar"}, map[string]any{"bar": "foo"})
	require.NoError(t, err)

	result2, err := m.LoadMemoryVariables(context.Background(), map[string]any{})
	require.NoError(t, err)

	expectedChatHistory := NewChatMessageHistory(
		WithPreviousMessages([]schema.ChatMessage{
			schema.HumanChatMessage{Content: "bar"},
			schema.AIChatMessage{Content: "foo"},
		}),
	)

	messages, err := expectedChatHistory.Messages(context.Background())
	require.NoError(t, err)
	expected2 := map[string]any{"history": messages}
	assert.Equal(t, expected2, result2)

	err = m.SaveContext(context.Background(), map[string]any{"foo": "bar"}, map[string]any{"bar": "foo"})
	err = m.SaveContext(context.Background(), map[string]any{"foo": "bar1"}, map[string]any{"bar": "foo1"})
	require.NoError(t, err)

	result2, err = m.LoadMemoryVariables(context.Background(), map[string]any{})
	require.NoError(t, err)

	expectedChatHistory = NewChatMessageHistory(
		WithPreviousMessages([]schema.ChatMessage{
			schema.HumanChatMessage{Content: "bar1"},
			schema.AIChatMessage{Content: "foo1"},
		}),
	)

	messages, err = expectedChatHistory.Messages(context.Background())
	require.NoError(t, err)
	expected2 = map[string]any{"history": messages}
	assert.Equal(t, expected2, result2)
}

/*
=== RUN   TestBufferWindowReturnMessage
=== PAUSE TestBufferWindowReturnMessage
=== CONT  TestBufferWindowReturnMessage
--- PASS: TestBufferWindowReturnMessage (72.37s)
PASS
*/

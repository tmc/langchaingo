package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tmc/langchaingo/schema"
)

func TestBufferMemory(t *testing.T) {
	t.Parallel()

	m := NewConversationBuffer()
	result1, err := m.LoadMemoryVariables(context.Background(), map[string]any{})
	require.NoError(t, err)
	expected1 := map[string]any{"history": ""}
	assert.Equal(t, expected1, result1)

	err = m.SaveContext(context.Background(), map[string]any{"foo": "bar"}, map[string]any{"bar": "foo"})
	require.NoError(t, err)

	result2, err := m.LoadMemoryVariables(context.Background(), map[string]any{})
	require.NoError(t, err)

	expected2 := map[string]any{"history": "Human: bar\nAI: foo"}
	assert.Equal(t, expected2, result2)
}

func TestBufferMemoryReturnMessage(t *testing.T) {
	t.Parallel()

	m := NewConversationBuffer()
	m.ReturnMessages = true
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
	assert.NoError(t, err)
	expected2 := map[string]any{"history": messages}
	assert.Equal(t, expected2, result2)
}

func TestBufferMemoryWithPreLoadedHistory(t *testing.T) {
	t.Parallel()

	m := NewConversationBuffer(WithChatHistory(NewChatMessageHistory(
		WithPreviousMessages([]schema.ChatMessage{
			schema.HumanChatMessage{Content: "bar"},
			schema.AIChatMessage{Content: "foo"},
		}),
	)))

	result, err := m.LoadMemoryVariables(context.Background(), map[string]any{})
	require.NoError(t, err)
	expected := map[string]any{"history": "Human: bar\nAI: foo"}
	assert.Equal(t, expected, result)
}

type testChatMessageHistory struct{}

var _ schema.ChatMessageHistory = testChatMessageHistory{}

func (t testChatMessageHistory) AddUserMessage(context.Context, string) error {
	return nil
}

func (t testChatMessageHistory) AddAIMessage(context.Context, string) error {
	return nil
}

func (t testChatMessageHistory) AddMessage(context.Context, schema.ChatMessage) error {
	return nil
}

func (t testChatMessageHistory) Clear() error {
	return nil
}

func (t testChatMessageHistory) SetMessages(context.Context, []schema.ChatMessage) error {
	return nil
}

func (t testChatMessageHistory) Messages(context.Context) ([]schema.ChatMessage, error) {
	return []schema.ChatMessage{
		schema.HumanChatMessage{Content: "user message test"},
		schema.AIChatMessage{Content: "ai message test"},
	}, nil
}

func TestBufferMemoryWithChatHistoryOption(t *testing.T) {
	t.Parallel()

	chatMessageHistory := testChatMessageHistory{}
	m := NewConversationBuffer(WithChatHistory(chatMessageHistory))

	result, err := m.LoadMemoryVariables(context.Background(), map[string]any{})
	require.NoError(t, err)
	expected := map[string]any{"history": "Human: user message test\nAI: ai message test"}
	assert.Equal(t, expected, result)
}

package memory

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func TestTokenBufferMemory(t *testing.T) {
	t.Parallel()

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	llm, err := openai.New()
	require.NoError(t, err)
	m := NewConversationTokenBuffer(llm, 2000)

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

func TestTokenBufferMemoryReturnMessage(t *testing.T) {
	t.Parallel()

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	llm, err := openai.New()
	require.NoError(t, err)
	m := NewConversationTokenBuffer(llm, 2000, WithReturnMessages(true))

	expected1 := map[string]any{"history": []llms.ChatMessage{}}
	result1, err := m.LoadMemoryVariables(context.Background(), map[string]any{})
	require.NoError(t, err)
	assert.Equal(t, expected1, result1)

	err = m.SaveContext(context.Background(), map[string]any{"foo": "bar"}, map[string]any{"bar": "foo"})
	require.NoError(t, err)

	result2, err := m.LoadMemoryVariables(context.Background(), map[string]any{})
	require.NoError(t, err)

	expectedChatHistory := NewChatMessageHistory(
		WithPreviousMessages([]llms.ChatMessage{
			llms.HumanChatMessage{Content: "bar"},
			llms.AIChatMessage{Content: "foo"},
		}),
	)

	messages, err := expectedChatHistory.Messages(context.Background())
	require.NoError(t, err)
	expected2 := map[string]any{"history": messages}
	assert.Equal(t, expected2, result2)
}

func TestTokenBufferMemoryWithPreLoadedHistory(t *testing.T) {
	t.Parallel()

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	llm, err := openai.New()
	require.NoError(t, err)

	m := NewConversationTokenBuffer(llm, 2000, WithChatHistory(NewChatMessageHistory(
		WithPreviousMessages([]llms.ChatMessage{
			llms.HumanChatMessage{Content: "bar"},
			llms.AIChatMessage{Content: "foo"},
		}),
	)))

	result, err := m.LoadMemoryVariables(context.Background(), map[string]any{})
	require.NoError(t, err)
	expected := map[string]any{"history": "Human: bar\nAI: foo"}
	assert.Equal(t, expected, result)
}

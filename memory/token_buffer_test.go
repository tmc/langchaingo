package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
)

func TestTokenBufferMemory(t *testing.T) {
	t.Parallel()

	llm, err := openai.New()
	require.NoError(t, err)
	m := NewTokenBuffer(llm, 2000)

	result1, err := m.LoadMemoryVariables(map[string]any{})
	require.NoError(t, err)
	expected1 := map[string]any{"history": ""}
	assert.Equal(t, expected1, result1)

	err = m.SaveContext(map[string]any{"foo": "bar"}, map[string]any{"bar": "foo"})
	require.NoError(t, err)

	result2, err := m.LoadMemoryVariables(map[string]any{})
	require.NoError(t, err)

	expected2 := map[string]any{"history": "Human: bar\nAI: foo"}
	assert.Equal(t, expected2, result2)
}

func TestTokenBufferMemoryReturnMessage(t *testing.T) {
	t.Parallel()

	llm, err := openai.New()
	require.NoError(t, err)
	m := NewTokenBuffer(llm, 2000, WithReturnMessages(true))

	expected1 := map[string]any{"history": []schema.ChatMessage{}}
	result1, err := m.LoadMemoryVariables(map[string]any{})
	require.NoError(t, err)
	assert.Equal(t, expected1, result1)

	err = m.SaveContext(map[string]any{"foo": "bar"}, map[string]any{"bar": "foo"})
	require.NoError(t, err)

	result2, err := m.LoadMemoryVariables(map[string]any{})
	require.NoError(t, err)

	expectedChatHistory := NewChatMessageHistory(
		WithPreviousMessages([]schema.ChatMessage{
			schema.HumanChatMessage{Text: "bar"},
			schema.AIChatMessage{Text: "foo"},
		}),
	)

	expected2 := map[string]any{"history": expectedChatHistory.Messages()}
	assert.Equal(t, expected2, result2)
}

func TestTokenBufferMemoryWithPreLoadedHistory(t *testing.T) {
	t.Parallel()

	llm, err := openai.New()
	require.NoError(t, err)

	m := NewTokenBuffer(llm, 2000, WithChatHistory(NewChatMessageHistory(
		WithPreviousMessages([]schema.ChatMessage{
			schema.HumanChatMessage{Text: "bar"},
			schema.AIChatMessage{Text: "foo"},
		}),
	)))

	result, err := m.LoadMemoryVariables(map[string]any{})
	require.NoError(t, err)
	expected := map[string]any{"history": "Human: bar\nAI: foo"}
	assert.Equal(t, expected, result)
}

func TestTokenBufferMemoryPrune(t *testing.T) {
	t.Parallel()

	llm, err := openai.New()
	require.NoError(t, err)

	m := NewTokenBuffer(llm, 20, WithChatHistory(NewChatMessageHistory(
		WithPreviousMessages([]schema.ChatMessage{
			schema.HumanChatMessage{Text: "human message test for max token"},
			schema.AIChatMessage{Text: "ai message test for max token"},
		}),
	)))

	buffStringMsg1, err := schema.GetBufferString([]schema.ChatMessage{
		schema.HumanChatMessage{Text: "human message test for max token"},
	}, "Human", "AI")
	require.NoError(t, err)
	tokenNumMsg1 := m.LLM.GetNumTokens(buffStringMsg1)
	assert.Equal(t, 9, tokenNumMsg1)

	buffStringMsg2, err := schema.GetBufferString([]schema.ChatMessage{
		schema.AIChatMessage{Text: "ai message test for max token"},
	}, "Human", "AI")
	require.NoError(t, err)
	tokenNumMsg2 := m.LLM.GetNumTokens(buffStringMsg2)
	assert.Equal(t, 8, tokenNumMsg2)

	assert.Equal(t, tokenNumMsg1+tokenNumMsg2, 17)

	_, err = llm.Call()
	require.NoError(t, err)

}

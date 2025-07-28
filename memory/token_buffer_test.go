package memory

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/0xDezzy/langchaingo/internal/httprr"
	"github.com/0xDezzy/langchaingo/llms"
	"github.com/0xDezzy/langchaingo/llms/openai"
)

// newTestOpenAIClient creates an OpenAI client with httprr support for testing.
func newTestOpenAIClient(t *testing.T) *openai.LLM {
	t.Helper()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	// Only run tests in parallel when not recording (to avoid rate limits)
	if !rr.Recording() {
		t.Parallel()
	}

	openaiOpts := []openai.Option{
		openai.WithHTTPClient(rr.Client()),
	}

	// Only add fake token when NOT recording (i.e., during replay)
	if !rr.Recording() {
		openaiOpts = append(openaiOpts, openai.WithToken("test-api-key"))
	}
	// When recording, openai.New() will read OPENAI_API_KEY from environment

	llm, err := openai.New(openaiOpts...)
	require.NoError(t, err)
	return llm
}

func TestTokenBufferMemory(t *testing.T) {
	ctx := context.Background()

	llm := newTestOpenAIClient(t)
	m := NewConversationTokenBuffer(llm, 2000)

	result1, err := m.LoadMemoryVariables(ctx, map[string]any{})
	require.NoError(t, err)
	expected1 := map[string]any{"history": ""}
	assert.Equal(t, expected1, result1)

	err = m.SaveContext(ctx, map[string]any{"foo": "bar"}, map[string]any{"bar": "foo"})
	require.NoError(t, err)

	result2, err := m.LoadMemoryVariables(ctx, map[string]any{})
	require.NoError(t, err)

	expected2 := map[string]any{"history": "Human: bar\nAI: foo"}
	assert.Equal(t, expected2, result2)
}

func TestTokenBufferMemoryReturnMessage(t *testing.T) {
	ctx := context.Background()

	llm := newTestOpenAIClient(t)
	m := NewConversationTokenBuffer(llm, 2000, WithReturnMessages(true))

	expected1 := map[string]any{"history": []llms.ChatMessage{}}
	result1, err := m.LoadMemoryVariables(ctx, map[string]any{})
	require.NoError(t, err)
	assert.Equal(t, expected1, result1)

	err = m.SaveContext(ctx, map[string]any{"foo": "bar"}, map[string]any{"bar": "foo"})
	require.NoError(t, err)

	result2, err := m.LoadMemoryVariables(ctx, map[string]any{})
	require.NoError(t, err)

	expectedChatHistory := NewChatMessageHistory(
		WithPreviousMessages([]llms.ChatMessage{
			llms.HumanChatMessage{Content: "bar"},
			llms.AIChatMessage{Content: "foo"},
		}),
	)

	messages, err := expectedChatHistory.Messages(ctx)
	require.NoError(t, err)
	expected2 := map[string]any{"history": messages}
	assert.Equal(t, expected2, result2)
}

func TestTokenBufferMemoryWithPreLoadedHistory(t *testing.T) {
	ctx := context.Background()

	llm := newTestOpenAIClient(t)

	m := NewConversationTokenBuffer(llm, 2000, WithChatHistory(NewChatMessageHistory(
		WithPreviousMessages([]llms.ChatMessage{
			llms.HumanChatMessage{Content: "bar"},
			llms.AIChatMessage{Content: "foo"},
		}),
	)))

	result, err := m.LoadMemoryVariables(ctx, map[string]any{})
	require.NoError(t, err)
	expected := map[string]any{"history": "Human: bar\nAI: foo"}
	assert.Equal(t, expected, result)
}

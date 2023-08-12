package chains

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/memory"
)

func TestConversation(t *testing.T) {
	t.Parallel()

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	llm, err := openai.New()
	require.NoError(t, err)

	c := NewConversation(llm, memory.NewConversationBuffer())
	_, err = Run(context.Background(), c, "Hi! I'm Jim")
	require.NoError(t, err)

	res, err := Run(context.Background(), c, "What is my name?")
	require.NoError(t, err)
	require.True(t, strings.Contains(res, "Jim"), `result does not contain the keyword 'Jim'`)
}

func TestConversationWithChatLLM(t *testing.T) {
	t.Parallel()

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	llm, err := openai.NewChat()
	require.NoError(t, err)

	c := NewConversation(llm, memory.NewConversationTokenBuffer(llm, 2000))
	_, err = Run(context.Background(), c, "Hi! I'm Jim")
	require.NoError(t, err)

	res, err := Run(context.Background(), c, "What is my name?")
	require.NoError(t, err)
	require.True(t, strings.Contains(res, "Jim"), `result does contain the keyword 'Jim'`)

	// this message will hit the maxTokenLimit and will initiate the prune of the messages to fit the context
	res, err = Run(context.Background(), c, "Are you sure that my name is Jim?")
	require.NoError(t, err)
	require.True(t, strings.Contains(res, "Jim"), `result does contain the keyword 'Jim'`)
}

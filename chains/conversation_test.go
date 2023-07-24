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
	model, err := openai.New()
	require.NoError(t, err)

	c := NewConversation(model, memory.NewBuffer())
	_, err = Run(context.Background(), c, "Hi! I'm Jim")
	require.NoError(t, err)

	res, err := Run(context.Background(), c, "What is my name?")
	require.NoError(t, err)
	require.True(t, strings.Contains(res, "Jim"), `result does not contain the keyword 'Jim'`)
}

func TestConversationMemoryPrune(t *testing.T) {
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	llm, err := openai.New()
	require.NoError(t, err)

	c := NewConversation(llm, memory.NewTokenBuffer(llm, 100, memory.WithReturnMessages(true)))
	_, err = Run(context.Background(), c, "Hi! I'm Jim")
	require.NoError(t, err)

	res, err := Run(context.Background(), c, "What is my name?")
	require.NoError(t, err)
	require.True(t, strings.Contains(res, "Jim"), `result does not contain the keyword 'Jim'`)

}

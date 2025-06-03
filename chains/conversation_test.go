package chains

import (
	"context"
	"os"
	"strings"
	"testing"

	z "github.com/getzep/zep-go"
	zClient "github.com/getzep/zep-go/client"
	zOption "github.com/getzep/zep-go/option"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/memory/zep"
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

func TestConversationWithZepMemory(t *testing.T) {
	t.Parallel()

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	if zepKey := os.Getenv("ZEP_API_KEY"); zepKey == "" {
		t.Skip("ZEP_API_KEY not set")
	}
	if sessionID := os.Getenv("ZEP_SESSION_ID"); sessionID == "" {
		t.Skip("ZEP_SESSION_ID not set")
	}

	llm, err := openai.New()
	require.NoError(t, err)

	zc := zClient.NewClient(
		zOption.WithAPIKey(os.Getenv("ZEP_API_KEY")),
	)
	sessionID := os.Getenv("ZEP_SESSION_ID")

	c := NewConversation(
		llm,
		zep.NewMemory(
			zc,
			sessionID,
			zep.WithMemoryType(z.MemoryGetRequestMemoryTypePerpetual),
			zep.WithHumanPrefix("Joe"),
			zep.WithAIPrefix("Robot"),
		),
	)
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

	llm, err := openai.New()
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

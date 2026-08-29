package openai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vxcontrol/langchaingo/llms"
)

func TestABudgetClientSurvivesACallWithoutReasoning(t *testing.T) {
	t.Parallel()

	llm, err := New(WithToken("t"), WithModel("gpt-4o-mini"),
		WithBaseURL("http://127.0.0.1:1"), WithUsingReasoningMaxTokens())
	require.NoError(t, err)

	_, err = llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")})

	require.Error(t, err, "the unreachable endpoint must surface as an error, not a panic")
	require.NotContains(t, err.Error(), "nil pointer")
}

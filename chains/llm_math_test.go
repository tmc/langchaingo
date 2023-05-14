package chains

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms/openai"
)

func TestLLMMath(t *testing.T) {
	t.Parallel()
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	llm, err := openai.New()
	require.NoError(t, err)

	chain := NewLLMMathChain(llm)
	q := "what is forty plus three? take that then multiply it by ten thousand divided by 7324.3"
	result, err := Run(context.Background(), chain, q)
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "58.708"), "expected 58.708 in result")
}

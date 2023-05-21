package executor_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/exp/agent/executor"
	"github.com/tmc/langchaingo/exp/tools"
	"github.com/tmc/langchaingo/exp/tools/serpapi"
	"github.com/tmc/langchaingo/llms/openai"
)

func TestMRKL(t *testing.T) {
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	if serpapiKey := os.Getenv("SERPAPI_API_KEY"); serpapiKey == "" {
		t.Skip("SERPAPI_API_KEY not set")
	}

	llm, err := openai.New()
	require.NoError(t, err)

	searchTool, err := serpapi.New()
	require.NoError(t, err)

	calculator := tools.Calculator{}

	a, err := executor.Initialize(
		llm,
		[]tools.Tool{searchTool, calculator},
		executor.ZeroShotReactDescription,
	)
	require.NoError(t, err)

	result, err := chains.Run(context.Background(), a, "If a person lived three times as long as Jacklyn Zeman, how long would they live") //nolint:lll
	require.NoError(t, err)

	require.True(t, strings.Contains(result, "210"), "correct answer 210 not in response")
}

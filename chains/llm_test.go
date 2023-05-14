package chains

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
)

func TestLLMChain(t *testing.T) {
	t.Parallel()
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	model, err := openai.New()
	require.NoError(t, err)

	prompt, err := prompts.NewPromptTemplate(
		"What is the capital of {{.country}}",
		[]string{"country"},
	)
	require.NoError(t, err)

	chain := NewLLMChain(model, prompt)

	resultChainValue, err := Call(context.Background(), chain,
		map[string]any{
			"country": "France",
		},
		llms.WithStopWords([]string{"\nObservation", "\n\tObservation"}),
	)
	require.NoError(t, err)

	resultAny, ok := resultChainValue["text"]
	require.True(t, ok)

	result, _ := resultAny.(string)
	result = strings.TrimSpace(result)
	assert.Equal(t, "Paris.", result)
}

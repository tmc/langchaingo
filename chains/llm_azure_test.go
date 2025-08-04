package chains

import (
	"context"
	"os"
	"strings"
	"testing"
	
	"github.com/stretchr/testify/require"
	"github.com/yincongcyincong/langchaingo/llms/openai"
	"github.com/yincongcyincong/langchaingo/prompts"
)

func TestLLMChainAzure(t *testing.T) {
	t.Parallel()
	// Azure OpenAI Key is used as OPENAI_API_KEY
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	// Azure OpenAI URL is used as OPENAI_BASE_URL
	if openaiBase := os.Getenv("OPENAI_BASE_URL"); openaiBase == "" {
		t.Skip("OPENAI_BASE_URL not set")
	}
	
	model, err := openai.New(
		openai.WithAPIType(openai.APITypeAzure),
		// Azure deployment that uses desired model, the name depends on what we define in the Azure deployment section
		openai.WithModel("model-name"),
		// Azure deployment that uses embeddings model, the name depends on what we define in the Azure deployment section
		openai.WithEmbeddingModel("embeddings-model-name"),
	)
	require.NoError(t, err)
	
	prompt := prompts.NewPromptTemplate(
		"What is the capital of {{.country}}",
		[]string{"country"},
	)
	require.NoError(t, err)
	
	chain := NewLLMChain(model, prompt)
	
	result, err := Predict(context.Background(), chain,
		map[string]any{
			"country": "France",
		},
	)
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "Paris"))
}

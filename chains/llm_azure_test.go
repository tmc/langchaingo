package chains

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
)

func TestLLMChainAzure(t *testing.T) {
	ctx := context.Background()
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	// Only run tests in parallel when not recording (to avoid rate limits)
	if rr.Replaying() {
		t.Parallel()
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
		openai.WithHTTPClient(rr.Client()),
	)
	require.NoError(t, err)

	prompt := prompts.NewPromptTemplate(
		"What is the capital of {{.country}}",
		[]string{"country"},
	)
	require.NoError(t, err)

	chain := NewLLMChain(model, prompt)

	result, err := Predict(ctx, chain,
		map[string]any{
			"country": "France",
		},
	)
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "Paris"))
}

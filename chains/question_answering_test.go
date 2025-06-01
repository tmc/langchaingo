package chains

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/llms/openai"
)

// createOpenAILLMForQA creates an OpenAI LLM with httprr support for testing.
func createOpenAILLMForQA(t *testing.T) *openai.LLM {
	t.Helper()
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	// Only run tests in parallel when not recording
	if !rr.Recording() {
		t.Parallel()
	}

	opts := []openai.Option{
		openai.WithHTTPClient(rr.Client()),
	}

	// Only add fake token when NOT recording (i.e., during replay)
	if !rr.Recording() {
		opts = append(opts, openai.WithToken("test-api-key"))
	}
	// When recording, openai.New() will read OPENAI_API_KEY from environment

	llm, err := openai.New(opts...)
	require.NoError(t, err)
	return llm
}

func TestRefineQA(t *testing.T) {
	ctx := context.Background()

	llm := createOpenAILLMForQA(t)

	docs := loadTestData(t)
	qaChain := LoadRefineQA(llm)

	results, err := Call(
		ctx,
		qaChain,
		map[string]any{
			"input_documents": docs,
			"question":        "What is the name of the lion?",
		},
	)
	require.NoError(t, err)

	_, ok := results["text"].(string)
	require.True(t, ok, "result does not contain text key")
}

func TestMapReduceQA(t *testing.T) {
	ctx := context.Background()

	llm := createOpenAILLMForQA(t)

	docs := loadTestData(t)
	qaChain := LoadMapReduceQA(llm)

	result, err := Predict(
		ctx,
		qaChain,
		map[string]any{
			"input_documents": docs,
			"question":        "What is the name of the lion?",
		},
	)

	require.NoError(t, err)
	require.True(t, strings.Contains(result, "Leo"), "result does not contain correct answer Leo")
}

func TestMapRerankQA(t *testing.T) {
	t.Skip("Test currently fails; see #415")
	t.Parallel()
	ctx := context.Background()

	llm := createOpenAILLMForQA(t)

	docs := loadTestData(t)
	mapRerankChain := LoadMapRerankQA(llm)

	results, err := Call(
		ctx,
		mapRerankChain,
		map[string]any{
			"input_documents": docs,
			"question":        "What is the name of the lion?",
		},
	)

	require.NoError(t, err)

	answer, ok := results["text"].(string)
	require.True(t, ok, "result does not contain text key")
	require.True(t, strings.Contains(answer, "Leo"), "result does not contain correct answer Leo")
}

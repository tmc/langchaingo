package chains

import (
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
	httprr.SkipIfNoCredentialsOrRecording(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	t.Cleanup(func() { rr.Close() })
	llm, err := openai.New(openai.WithHTTPClient(rr.Client()))
	require.NoError(t, err)
	return llm
}

func TestRefineQA(t *testing.T) {
	t.Parallel()

	llm := createOpenAILLMForQA(t)

	docs := loadTestData(t)
	qaChain := LoadRefineQA(llm)

	results, err := Call(
		t.Context(),
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
	t.Parallel()

	llm := createOpenAILLMForQA(t)

	docs := loadTestData(t)
	qaChain := LoadMapReduceQA(llm)

	result, err := Predict(
		t.Context(),
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

	llm := createOpenAILLMForQA(t)

	docs := loadTestData(t)
	mapRerankChain := LoadMapRerankQA(llm)

	results, err := Call(
		t.Context(),
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

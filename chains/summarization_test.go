package chains

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

func loadTestData(t *testing.T) []schema.Document {
	t.Helper()

	file, err := os.Open("./testdata/mouse_story.txt")
	require.NoError(t, err)

	docs, err := documentloaders.NewText(file).LoadAndSplit(
		t.Context(),
		textsplitter.NewRecursiveCharacter(),
	)
	require.NoError(t, err)

	return docs
}

// createOpenAILLMForTest creates an OpenAI LLM with httprr support for testing.
func createOpenAILLMForTest(t *testing.T) *openai.LLM {
	t.Helper()
	httprr.SkipIfNoCredentialsOrRecording(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	t.Cleanup(func() { rr.Close() })
	llm, err := openai.New(openai.WithHTTPClient(rr.Client()))
	require.NoError(t, err)
	return llm
}

func TestStuffSummarization(t *testing.T) {
	t.Parallel()

	llm := createOpenAILLMForTest(t)

	docs := loadTestData(t)

	chain := LoadStuffSummarization(llm)
	_, err := Call(
		t.Context(),
		chain,
		map[string]any{"input_documents": docs},
	)
	require.NoError(t, err)
}

func TestRefineSummarization(t *testing.T) {
	t.Parallel()

	llm := createOpenAILLMForTest(t)

	docs := loadTestData(t)

	chain := LoadRefineSummarization(llm)
	_, err := Call(
		t.Context(),
		chain,
		map[string]any{"input_documents": docs},
	)
	require.NoError(t, err)
}

func TestMapReduceSummarization(t *testing.T) {
	t.Parallel()

	llm := createOpenAILLMForTest(t)

	docs := loadTestData(t)

	chain := LoadMapReduceSummarization(llm)
	_, err := Run(
		t.Context(),
		chain,
		docs,
	)
	require.NoError(t, err)
}

package chains

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

type testRetriever struct{}

var _ schema.Retriever = testRetriever{}

func (r testRetriever) GetRelevantDocuments(_ context.Context, _ string) ([]schema.Document, error) {
	return []schema.Document{
		{PageContent: "foo is 34"},
		{PageContent: "bar is 1"},
	}, nil
}

// createOpenAILLMForRetrieval creates an OpenAI LLM with httprr support for testing.
func createOpenAILLMForRetrieval(t *testing.T) *openai.LLM {
	t.Helper()
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	t.Cleanup(func() { rr.Close() })
	llm, err := openai.New(openai.WithHTTPClient(rr.Client()))
	require.NoError(t, err)
	return llm
}

func TestRetrievalQA(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	llm := createOpenAILLMForRetrieval(t)

	prompt := prompts.NewPromptTemplate(
		"answer this question {{.question}} with this context {{.context}}",
		[]string{"question", "context"},
	)

	combineChain := NewStuffDocuments(NewLLMChain(llm, prompt))
	r := testRetriever{}

	chain := NewRetrievalQA(combineChain, r)

	result, err := Run(ctx, chain, "what is foo? ")
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "34"), "expected 34 in result")
}

func TestRetrievalQAFromLLM(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	r := testRetriever{}
	llm := createOpenAILLMForRetrieval(t)

	chain := NewRetrievalQAFromLLM(llm, r)
	result, err := Run(ctx, chain, "what is foo? ")
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "34"), "expected 34 in result")
}

package chains

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/0xDezzy/langchaingo/internal/httprr"
	"github.com/0xDezzy/langchaingo/llms/openai"
	"github.com/0xDezzy/langchaingo/prompts"
	"github.com/0xDezzy/langchaingo/schema"
	"github.com/stretchr/testify/require"
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

func TestRetrievalQA(t *testing.T) {
	ctx := context.Background()

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

	r := testRetriever{}
	llm := createOpenAILLMForRetrieval(t)

	chain := NewRetrievalQAFromLLM(llm, r)
	result, err := Run(ctx, chain, "what is foo? ")
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "34"), "expected 34 in result")
}

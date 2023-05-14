package chains

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
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

func TestRetrievalQA(t *testing.T) {
	t.Parallel()
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	llm, err := openai.New()
	require.NoError(t, err)

	prompt := prompts.NewPromptTemplate(
		"answer this question {{.question}} with this context {{.context}}",
		[]string{"question", "context"},
	)
	require.NoError(t, err)

	combineChain := NewStuffDocuments(NewLLMChain(llm, prompt))
	r := testRetriever{}

	chain := NewRetrievalQA(combineChain, r)

	result, err := Run(context.Background(), chain, "what is foo? ")
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "34"), "expected 34 in result")
}

func TestRetrievalQAFromLLM(t *testing.T) {
	t.Parallel()
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	r := testRetriever{}
	llm, err := openai.New()
	require.NoError(t, err)

	chain := NewRetrievalQAFromLLM(llm, r)
	result, err := Run(context.Background(), chain, "what is foo? ")
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "34"), "expected 34 in result")
}

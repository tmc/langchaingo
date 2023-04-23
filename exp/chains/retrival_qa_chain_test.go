package chains

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
)

type testRetriever struct{}

func (r testRetriever) GetRelevantDocuments(query string) ([]schema.Document, error) {
	return []schema.Document{
		{PageContent: "foo"},
		{PageContent: "bar"},
	}, nil
}

func TestRetrievalQAChain(t *testing.T) {
	t.Parallel()

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	r := testRetriever{}
	llm, err := openai.New()
	require.NoError(t, err)

	chain := NewRetrievalQAChainFromLLM(llm, r)

	_, err = Call(chain, map[string]any{
		"query": "foz?",
	})
	require.NoError(t, err)
}

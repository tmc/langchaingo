package chains

import (
	"testing"

	"github.com/tmc/langchaingo/exp/schema"
	"github.com/tmc/langchaingo/llms/openai"
)

type testRetriever struct{}

func (r testRetriever) GetRelevantDocuments(query string) ([]schema.Document, error) {
	return []schema.Document{
		{PageContent: "foo"},
		{PageContent: "bar"},
	}, nil
}

func TestRetrievalQAChain(t *testing.T) {
	r := testRetriever{}
	llm, err := openai.New()
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
		return
	}

	chain := NewRetrievalQAChainFromLLM(llm, r)

	_, err = Call(chain, map[string]any{
		"query": "foz?",
	})
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
		return
	}
}

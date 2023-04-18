package chains_test

import (
	"testing"

	"github.com/tmc/langchaingo/chains"
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
	r := testRetriever{}
	llm, err := openai.New()
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
		return
	}

	chain := chains.NewRetrievalQAChainFromLLM(llm, r)

	_, err = chains.Call(chain, map[string]any{
		"query": "foz?",
	}, nil)
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
		return
	}
}

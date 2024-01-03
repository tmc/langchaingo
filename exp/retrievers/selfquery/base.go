package selfquery

import (
	"context"

	"github.com/tmc/langchaingo/schema"
)

var _ schema.Retriever = SelfQueryRetriever{}

type SelfQueryRetriever struct {
}

func (sqr SelfQueryRetriever) GetRelevantDocuments(ctx context.Context, query string) ([]schema.Document, error) {

	return nil, nil
}

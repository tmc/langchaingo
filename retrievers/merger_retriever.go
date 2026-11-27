package retrievers

import (
	"context"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/schema"
)

var _ schema.Retriever = &MergerRetriever{}

// MergerRetriever is a retriever that merges the results of multiple retrievers.
type MergerRetriever struct {
	Retrievers       []schema.Retriever
	CallbacksHandler callbacks.Handler
}

// NewMergerRetriever creates a new MergerRetriever.
func NewMergerRetriever(
	retrievers []schema.Retriever,
) MergerRetriever {
	return MergerRetriever{
		Retrievers:       retrievers,
		CallbacksHandler: nil,
	}
}

// GetRelevantDocuments returns documents from the MergerRetriever's all retrievers.
func (m *MergerRetriever) GetRelevantDocuments(ctx context.Context, query string) ([]schema.Document, error) {
	if m.CallbacksHandler != nil {
		m.CallbacksHandler.HandleRetrieverStart(ctx, query)
	}
	docs := make([]schema.Document, 0)
	for _, r := range m.Retrievers {
		doc, err := r.GetRelevantDocuments(ctx, query)
		if err != nil {
			return nil, err
		}
		docs = append(docs, doc...)
	}
	if m.CallbacksHandler != nil {
		m.CallbacksHandler.HandleRetrieverEnd(ctx, query, docs)
	}
	return docs, nil
}

package selfquery

import (
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

var _ schema.Retriever = SelfQueryRetriever{}

type SelfQueryRetriever struct {
	VectorStore *vectorstores.VectorStore
}

func New() (*SelfQueryRetriever, error) {

	return nil, nil
}

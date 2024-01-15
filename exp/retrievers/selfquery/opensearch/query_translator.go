package selfquery_opensearch

import (
	"github.com/tmc/langchaingo/exp/retrievers/selfquery"
	"github.com/tmc/langchaingo/vectorstores"
)

type SelfQueryOpensearchTranslator struct {
	vectorstore vectorstores.VectorStore
}

var _ selfquery.StoreWithQueryTranslator = SelfQueryOpensearchTranslator{}

func New(vectorstore vectorstores.VectorStore) *SelfQueryOpensearchTranslator {
	return &SelfQueryOpensearchTranslator{
		vectorstore: vectorstore,
	}
}

package selfquery_opensearch

import (
	"github.com/tmc/langchaingo/exp/retrievers/selfquery"
	"github.com/tmc/langchaingo/vectorstores"
)

type SelfQueryOpensearchTranslator struct {
	vectorstore vectorstores.VectorStore
	indexName   string
}

var _ selfquery.StoreWithQueryTranslator = SelfQueryOpensearchTranslator{}

func New(vectorstore vectorstores.VectorStore, indexName string) *SelfQueryOpensearchTranslator {
	return &SelfQueryOpensearchTranslator{
		vectorstore: vectorstore,
		indexName:   indexName,
	}
}

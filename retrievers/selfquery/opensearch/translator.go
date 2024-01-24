package selfqueryopensearch

import (
	"github.com/tmc/langchaingo/retrievers/selfquery"
	"github.com/tmc/langchaingo/tools/queryconstructor"
	"github.com/tmc/langchaingo/vectorstores"
)

// translator for opensearch.
type Translator struct {
	vectorstore   vectorstores.VectorStore
	indexName     string
	comparatorMap map[string]string
	operatorMap   map[string]string
}

var _ selfquery.StoreWithQueryTranslator = Translator{}

// pseudo constructor.
func New(vectorstore vectorstores.VectorStore, indexName string) *Translator {
	return &Translator{
		vectorstore: vectorstore,
		indexName:   indexName,
		comparatorMap: map[queryconstructor.Comparator]string{
			queryconstructor.ComparatorEQ:      "term",
			queryconstructor.ComparatorLT:      "lt",
			queryconstructor.ComparatorLTE:     "lte",
			queryconstructor.ComparatorGT:      "gt",
			queryconstructor.ComparatorGTE:     "gte",
			queryconstructor.ComparatorCONTAIN: "match",
			queryconstructor.ComparatorLIKE:    "fuzzy",
		},
		operatorMap: map[queryconstructor.Operator]string{
			queryconstructor.OperatorAnd: "must",
			queryconstructor.OperatorOr:  "should",
			queryconstructor.OperatorNot: "must_not",
		},
	}
}

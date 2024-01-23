package selfquery

import (
	"github.com/tmc/langchaingo/exp/tools/queryconstructor"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

var _ schema.Retriever = SelfQueryRetriever{}

type SelfQueryRetriever struct {
	Store             StoreWithQueryTranslator
	LLM               llms.LLM
	DocumentContents  string
	MetadataFieldInfo []queryconstructor.AttributeInfo
	EnableLimit       *bool
	DefaultLimit      int
}

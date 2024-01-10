package selfquery

import (
	"github.com/tmc/langchaingo/exp/tools/queryconstrutor"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

var _ schema.Retriever = SelfQueryRetriever{}

type SelfQueryRetriever struct {
	VectorStore       vectorstores.VectorStore
	LLM               llms.LLM
	DocumentContents  string
	MetadataFieldInfo []queryconstrutor.AttributeInfo
	EnableLimit       *bool
}

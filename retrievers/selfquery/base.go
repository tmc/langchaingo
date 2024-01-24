package selfquery

import (
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

var _ schema.Retriever = Retriever{}

// selfquery a database.
type Retriever struct {
	Store             StoreWithQueryTranslator
	LLM               llms.Model
	DocumentContents  string
	MetadataFieldInfo []schema.AttributeInfo
	EnableLimit       *bool
	DefaultLimit      int
}

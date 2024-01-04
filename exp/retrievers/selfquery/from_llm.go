package selfquery

import (
	"context"

	"github.com/tmc/langchaingo/exp/tools/queryconstrutor"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

type FromLLMArgs struct {
	LLM               *llms.LLM
	VectorStore       *vectorstores.VectorStore
	DocumentContents  string
	MetadataFieldInfo []queryconstrutor.AttributeInfo
	EnableLimit       *bool
}

func (sqr SelfQueryRetriever) FromLLM(ctx context.Context, query string) ([]schema.Document, error) {

	return nil, nil
}

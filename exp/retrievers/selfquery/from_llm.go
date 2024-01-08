package selfquery

import (
	"context"

	"github.com/tmc/langchaingo/exp/tools/queryconstrutor"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/vectorstores"
)

type FromLLMArgs struct {
	LLM               *llms.LLM
	VectorStore       *vectorstores.VectorStore
	DocumentContents  string
	MetadataFieldInfo []queryconstrutor.AttributeInfo
	EnableLimit       *bool
}

func FromLLM(ctx context.Context, args FromLLMArgs) (*SelfQueryRetriever, error) {

	return nil, nil
}

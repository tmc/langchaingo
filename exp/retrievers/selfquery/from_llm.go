package selfquery

import (
	"github.com/tmc/langchaingo/exp/tools/queryconstrutor"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/vectorstores"
)

type FromLLMArgs struct {
	LLM               llms.LLM
	VectorStore       vectorstores.VectorStore
	DocumentContents  string
	MetadataFieldInfo []queryconstrutor.AttributeInfo
	EnableLimit       *bool
}

func FromLLM(args FromLLMArgs) *SelfQueryRetriever {
	retriever := SelfQueryRetriever{
		VectorStore:       args.VectorStore,
		LLM:               args.LLM,
		DocumentContents:  args.DocumentContents,
		MetadataFieldInfo: args.MetadataFieldInfo,
		EnableLimit:       args.EnableLimit,
	}
	return &retriever
}

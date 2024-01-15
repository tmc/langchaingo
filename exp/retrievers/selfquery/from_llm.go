package selfquery

import (
	"github.com/tmc/langchaingo/exp/tools/queryconstructor"
	"github.com/tmc/langchaingo/llms"
)

type FromLLMArgs struct {
	LLM               llms.LLM
	Store             StoreWithQueryTranslator
	DocumentContents  string
	MetadataFieldInfo []queryconstructor.AttributeInfo
	EnableLimit       *bool
}

func FromLLM(args FromLLMArgs) *SelfQueryRetriever {
	retriever := SelfQueryRetriever{
		Store:             args.Store,
		LLM:               args.LLM,
		DocumentContents:  args.DocumentContents,
		MetadataFieldInfo: args.MetadataFieldInfo,
		EnableLimit:       args.EnableLimit,
	}
	return &retriever
}

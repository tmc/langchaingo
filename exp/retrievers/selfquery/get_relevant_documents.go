package selfquery

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/exp/tools/queryconstrutor"
	"github.com/tmc/langchaingo/schema"
)

func (sqr SelfQueryRetriever) GetRelevantDocuments(ctx context.Context, query string) ([]schema.Document, error) {

	prompt, err := queryconstrutor.GetQueryConstructorPrompt(queryconstrutor.GetQueryConstructorPromptArgs{
		DocumentContents: sqr.DocumentContents,
		AttributeInfo:    sqr.MetadataFieldInfo,
		EnableLimit:      sqr.EnableLimit,
	})

	if err != nil {
		return nil, fmt.Errorf("error load query constructor %w", err)
	}

	promptChain := *chains.NewLLMChain(sqr.LLM, prompt)
	result, err := promptChain.Call(ctx, map[string]any{
		"query": query,
	})

	if err != nil {
		fmt.Printf("err: %v\n", err)
	}

	fmt.Printf("result: %v\n", result)

	return nil, nil
}

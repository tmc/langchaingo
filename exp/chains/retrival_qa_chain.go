package chains

import (
	"fmt"

	"github.com/tmc/langchaingo/exp/memory"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

type RetrievalQAChain struct {
	retriever             schema.Retriever
	combineDocumentChain  Chain
	OutputKey             string
	InputKey              string
	ReturnSourceDocuments bool
}

func NewRetrievalQAChainFromLLM(llm llms.LLM, retriever schema.Retriever) RetrievalQAChain {
	qaChain := loadQAStuffChain(llm)
	return RetrievalQAChain{
		retriever:             retriever,
		combineDocumentChain:  qaChain,
		OutputKey:             "result",
		InputKey:              "query",
		ReturnSourceDocuments: false,
	}
}

func (c RetrievalQAChain) Call(values map[string]any) (map[string]any, error) {
	queryAny, ok := values[c.InputKey]
	if !ok {
		return nil, fmt.Errorf("Input key %s not found", c.InputKey)
	}

	query, ok := queryAny.(string)
	if !ok {
		return nil, fmt.Errorf("Input value %s not string", c.InputKey)
	}

	docs, err := c.retriever.GetRelevantDocuments(query)
	if err != nil {
		return nil, err
	}

	inputs := map[string]any{
		"question":        query,
		"input_documents": docs,
	}

	result, err := Call(c.combineDocumentChain, inputs)
	if err != nil {
		return nil, err
	}

	if c.ReturnSourceDocuments {
		result["sourceDocuments"] = docs
	}

	return result, nil
}

func (c RetrievalQAChain) GetMemory() schema.Memory {
	return memory.NewEmpty()
}

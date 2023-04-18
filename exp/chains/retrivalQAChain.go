package chains

import (
	"fmt"

	"github.com/tmc/langchaingo/exp/memory"
	"github.com/tmc/langchaingo/exp/schema"
	"github.com/tmc/langchaingo/llms"
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

func (c RetrievalQAChain) Call(values ChainValues, stop []string) (ChainValues, error) {
	queryAny, ok := values[c.InputKey]
	if !ok {
		return map[string]any{}, fmt.Errorf("Input key %s not found", c.InputKey)
	}

	query, ok := queryAny.(string)
	if !ok {
		return map[string]any{}, fmt.Errorf("Input value %s not string", c.InputKey)
	}

	docs, err := c.retriever.GetRelevantDocuments(query)
	if err != nil {
		return map[string]any{}, err
	}

	inputs := map[string]any{
		"question":        query,
		"input_documents": docs,
	}

	result, err := Call(c.combineDocumentChain, inputs, nil)
	if err != nil {
		return map[string]any{}, err
	}

	if c.ReturnSourceDocuments {
		result["sourceDocuments"] = docs
	}

	return result, nil
}

func (c RetrievalQAChain) GetMemory() memory.Memory {
	return memory.NewEmptyMemory()
}

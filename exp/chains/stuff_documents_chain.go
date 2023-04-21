package chains

import (
	"fmt"

	"github.com/tmc/langchaingo/exp/memory"
	"github.com/tmc/langchaingo/exp/prompts"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

type StuffDocumentsChain struct {
	llmChain             LLMChain // LLMChain to use after formatting documents
	InputKey             string
	OutputKey            string
	DocumentVariableName string // Variable name in the llmChain to put the documents in
}

func NewStuffDocumentsChain(llmChain LLMChain) StuffDocumentsChain {
	return StuffDocumentsChain{
		llmChain:             llmChain,
		InputKey:             "input_documents",
		OutputKey:            "output_texts",
		DocumentVariableName: "context",
	}
}

func (c StuffDocumentsChain) Call(values map[string]any) (map[string]any, error) {
	docsAny, ok := values[c.InputKey]
	if !ok {
		return map[string]any{}, fmt.Errorf("Document key %s not found", c.InputKey)
	}

	docs, ok := docsAny.([]schema.Document)
	if !ok {
		return map[string]any{}, fmt.Errorf("Document key %s not of type []Document", c.InputKey)
	}

	text := ""
	for _, doc := range docs {
		text += doc.PageContent + "\n\n"
	}

	inputValues := make(map[string]any)
	for key, value := range values {
		inputValues[key] = value
	}

	inputValues[c.DocumentVariableName] = text
	return Call(c.llmChain, inputValues)
}

func (c StuffDocumentsChain) GetMemory() schema.Memory {
	return memory.NewEmpty()
}

var DefaultQAPrompt, _ = prompts.NewPromptTemplate(
	"Use the following pieces of context to answer the question at the end. If you don't know the answer, just say that you don't know, don't try to make up an answer.\n\n{context}\n\nQuestion: {question}\nHelpful Answer:",
	[]string{"context", "question"},
)

// TODO: add conditional after chat model is added.
var QAPromptSelector = prompts.NewConditionalPromptSelector(DefaultQAPrompt, []prompts.Conditional{})

func loadQAStuffChain(llm llms.LLM) StuffDocumentsChain {
	prompt := QAPromptSelector.GetPrompt(llm)
	llmChain := NewLLMChain(llm, prompt)
	return NewStuffDocumentsChain(llmChain)
}

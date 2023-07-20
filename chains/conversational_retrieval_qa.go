package chains

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/schema"
)

const (
	_conversationalRetrievalQADefaultInputKey             = "question"
	_conversationalRetrievalQADefaultSourceDocumentKey    = "source_documents"
	_conversationalRetrievalQADefaultGeneratedQuestionKey = "generated_question"
)

// ConversationalRetrievalQA chain builds on RetrievalQA to provide a chat history component.
type ConversationalRetrievalQA struct {
	// Retriever used to retrieve the relevant documents.
	Retriever               schema.Retriever
	Memory                  memory.Buffer
	CombineDocumentsChain   Chain
	CondenseQuestionChain   Chain
	OutputKey               string
	RephraseQuestion        bool
	ReturnGeneratedQuestion bool
	InputKey                string
	ReturnSourceDocuments   bool
}

var _ Chain = ConversationalRetrievalQA{}

// NewConversationalRetrievalQA creates a new NewConversationalRetrievalQA
func NewConversationalRetrievalQA(combineDocumentsChain Chain, condenseQuestionChain Chain, retriever schema.Retriever, memory memory.Buffer) ConversationalRetrievalQA {
	return ConversationalRetrievalQA{
		Retriever:               retriever,
		CombineDocumentsChain:   combineDocumentsChain,
		CondenseQuestionChain:   condenseQuestionChain,
		InputKey:                _conversationalRetrievalQADefaultInputKey,
		ReturnSourceDocuments:   false,
		Memory:                  memory,
		OutputKey:               _llmChainDefaultOutputKey,
		RephraseQuestion:        true,
		ReturnGeneratedQuestion: false,
	}
}

func NewConversationalRetrievalQAFromLLM(llm llms.LanguageModel, retriever schema.Retriever, memory memory.Buffer) ConversationalRetrievalQA {
	return NewConversationalRetrievalQA(
		LoadStuffQA(llm),
		LoadQuestionGenerator(llm),
		retriever,
		memory,
	)
}

// Call gets question, and relevant documents by question from the retriever and gives them to the combine
// documents chain.
func (c ConversationalRetrievalQA) Call(ctx context.Context, values map[string]any, options ...ChainCallOption) (map[string]any, error) { //nolint: lll
	query, ok := values[c.InputKey].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %w", ErrInvalidInputValues, ErrInputValuesWrongType)
	}

	newQuestion, err := c.getQuestion(ctx, query)
	if err != nil {
		return nil, err
	}

	docs, err := c.Retriever.GetRelevantDocuments(ctx, newQuestion)
	if err != nil {
		return nil, err
	}
	//docs = append(docs)

	result, err := Predict(ctx, c.CombineDocumentsChain, map[string]any{
		"question":        c.rephraseQuestion(query, newQuestion),
		"input_documents": docs,
	}, options...)
	if err != nil {
		return nil, err
	}

	output := make(map[string]any)

	output[_llmChainDefaultOutputKey] = result
	if c.ReturnSourceDocuments {
		output[_conversationalRetrievalQADefaultSourceDocumentKey] = docs
	}
	if c.ReturnGeneratedQuestion {
		output[_conversationalRetrievalQADefaultGeneratedQuestionKey] = newQuestion
	}

	return output, nil
}

func (c ConversationalRetrievalQA) GetMemory() schema.Memory {
	return &c.Memory
}

func (c ConversationalRetrievalQA) GetInputKeys() []string {
	return []string{c.InputKey}
}

func (c ConversationalRetrievalQA) GetOutputKeys() []string {
	outputKeys := append([]string{}, c.CombineDocumentsChain.GetOutputKeys()...)
	if c.ReturnSourceDocuments {
		outputKeys = append(outputKeys, _conversationalRetrievalQADefaultSourceDocumentKey)
	}

	return outputKeys
}

func (c ConversationalRetrievalQA) getQuestion(ctx context.Context, question string) (string, error) {
	if len(c.Memory.ChatHistory.Messages()) == 0 {
		return question, nil
	}

	chatHistoryStr, err := schema.GetBufferString(c.Memory.ChatHistory.Messages(), "Human", "Assistant")
	if err != nil {
		return "", err
	}

	results, err := Call(
		ctx,
		c.CondenseQuestionChain,
		map[string]any{
			"chat_history": chatHistoryStr,
			"question":     question,
		},
	)
	if err != nil {
		return "", err
	}

	newQuestion, ok := results[c.OutputKey].(string)
	if !ok {
		return "", ErrInvalidOutputValues
	}

	return newQuestion, nil
}

func (c ConversationalRetrievalQA) rephraseQuestion(question string, newQuestion string) string {
	if c.RephraseQuestion {
		return newQuestion
	}
	return question
}

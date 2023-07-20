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
	_conversationalRetrievalQADefaultOutputKey            = "text"
	_conversationalRetrievalQADefaultSourceDocumentKey    = "source_documents"
	_conversationalRetrievalQADefaultGeneratedQuestionKey = "generated_question"
)

// ConversationalRetrievalQA chain builds on RetrievalQA to provide a chat history component.
type ConversationalRetrievalQA struct {
	RetrievalQA
	QuestionGeneratorChain  Chain
	ChatHistory             []schema.ChatMessage
	OutputKey               string
	RephraseQuestion        bool
	ReturnGeneratedQuestion bool
}

var _ Chain = ConversationalRetrievalQA{}

// NewConversationalRetrievalQA creates a new NewConversationalRetrievalQA
func NewConversationalRetrievalQA(combineDocumentsChain Chain, questionGeneratorChain Chain, retriever schema.Retriever, chatHistory []schema.ChatMessage) ConversationalRetrievalQA {
	return ConversationalRetrievalQA{
		RetrievalQA: RetrievalQA{
			Retriever:             retriever,
			CombineDocumentsChain: combineDocumentsChain,
			InputKey:              _conversationalRetrievalQADefaultInputKey,
			ReturnSourceDocuments: false,
		},
		QuestionGeneratorChain:  questionGeneratorChain,
		OutputKey:               _conversationalRetrievalQADefaultOutputKey,
		ChatHistory:             chatHistory,
		RephraseQuestion:        true,
		ReturnGeneratedQuestion: false,
	}
}

func NewConversationalRetrievalQAFromLLM(llm llms.LanguageModel, retriever schema.Retriever, chatHistory []schema.ChatMessage) ConversationalRetrievalQA {
	return NewConversationalRetrievalQA(
		LoadStuffQA(llm),
		LoadQuestionGenerator(llm),
		retriever,
		chatHistory,
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

	result, err := Call(ctx, c.CombineDocumentsChain, map[string]any{
		"question":        c.rephraseQuestion(query, newQuestion),
		"input_documents": docs,
	}, options...)
	if err != nil {
		return nil, err
	}

	if c.ReturnSourceDocuments {
		result[_conversationalRetrievalQADefaultSourceDocumentKey] = docs
	}
	if c.ReturnGeneratedQuestion {
		result[_conversationalRetrievalQADefaultGeneratedQuestionKey] = newQuestion
	}

	return result, nil
}

func (c ConversationalRetrievalQA) GetMemory() schema.Memory { //nolint:ireturn
	return memory.NewSimple()
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
	if len(c.ChatHistory) == 0 {
		return question, nil
	}

	chatHistoryStr, err := schema.GetBufferString(c.ChatHistory, "Human", "AI")
	if err != nil {
		return "", err
	}

	results, err := Call(
		ctx,
		c.QuestionGeneratorChain,
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

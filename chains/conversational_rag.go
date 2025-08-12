package chains

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

const (
	_conversationalRAGDefaultInputKey             = "question"
	_conversationalRAGDefaultSourceDocumentKey    = "source_documents"
	_conversationalRAGDefaultGeneratedQuestionKey = "generated_question"
)

const _defaultConversationalRAGTemplate = `Use the following pieces of context to answer the question at the end. If you don't know the answer, just say that you don't know, don't try to make up an answer.

Context:
{{.context}}

Chat History:
{{.chat_history}}

Question: {{.question}}
Helpful Answer:`

// ConversationalRAG chain combines retrieval-augmented generation with conversational history.
// Unlike ConversationalRetrievalQA, this chain passes chat history to the final LLM call
// for contextually aware responses.
type ConversationalRAG struct {
	// Retriever used to retrieve the relevant documents.
	Retriever schema.Retriever

	// Memory that remembers previous conversational back and forths directly.
	Memory schema.Memory

	// LLMChain The chain used to generate the final answer with context and history.
	LLMChain *LLMChain

	// CondenseQuestionChain The chain used to generate a new question for retrieval.
	// This chain will take in the current question (with variable `question`)
	// and any chat history (with variable `chat_history`) and will produce
	// a new standalone question to be used for retrieval.
	CondenseQuestionChain Chain

	// OutputKey The output key to return the final answer of this chain in.
	OutputKey string

	// RephraseQuestion Whether to pass the new generated question to the LLMChain.
	// If true, will pass the new generated question along.
	// If false, will only use the new generated question for retrieval and pass the
	// original question along to the LLMChain.
	RephraseQuestion bool

	// ReturnGeneratedQuestion Return the generated question as part of the final result.
	ReturnGeneratedQuestion bool

	// InputKey The input key to get the query from, by default "question".
	InputKey string

	// ReturnSourceDocuments Return the retrieved source documents as part of the final result.
	ReturnSourceDocuments bool
}

var _ Chain = ConversationalRAG{}

// NewConversationalRAG creates a new ConversationalRAG chain.
func NewConversationalRAG(
	llmChain *LLMChain,
	condenseQuestionChain Chain,
	retriever schema.Retriever,
	memory schema.Memory,
) ConversationalRAG {
	return ConversationalRAG{
		Memory:                  memory,
		Retriever:               retriever,
		LLMChain:                llmChain,
		CondenseQuestionChain:   condenseQuestionChain,
		InputKey:                _conversationalRAGDefaultInputKey,
		OutputKey:               _llmChainDefaultOutputKey,
		RephraseQuestion:        true,
		ReturnGeneratedQuestion: false,
		ReturnSourceDocuments:   false,
	}
}

// NewConversationalRAGFromLLM creates a ConversationalRAG chain using a single LLM
// for both question condensation and final answer generation.
func NewConversationalRAGFromLLM(
	llm llms.Model,
	retriever schema.Retriever,
	memory schema.Memory,
) ConversationalRAG {
	// if systemPrompt == "" {
	// 	systemPrompt = _defaultConversationalRAGTemplate
	// }
	// Create the conversational RAG prompt template that includes chat history
	ragPromptTemplate := prompts.NewPromptTemplate(
		_defaultConversationalRAGTemplate,
		[]string{"context", "chat_history", "question"},
	)
	ragLLMChain := NewLLMChain(llm, ragPromptTemplate)

	return NewConversationalRAG(
		ragLLMChain,
		LoadCondenseQuestionGenerator(llm),
		retriever,
		memory,
	)
}

// Call executes the conversational RAG chain.
func (c ConversationalRAG) Call(ctx context.Context, values map[string]any, options ...ChainCallOption) (map[string]any, error) {
	query, ok := values[c.InputKey].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %w", ErrInvalidInputValues, ErrInputValuesWrongType)
	}

	// Get chat history from memory
	chatHistoryStr, ok := values[c.Memory.GetMemoryKey(ctx)].(string)
	if !ok {
		chatHistory, ok := values[c.Memory.GetMemoryKey(ctx)].([]llms.ChatMessage)
		if !ok {
			return nil, fmt.Errorf("%w: %w", ErrMissingMemoryKeyValues, ErrMemoryValuesWrongType)
		}

		bufferStr, err := llms.GetBufferString(chatHistory, "Human", "AI")
		if err != nil {
			return nil, err
		}

		chatHistoryStr = bufferStr
	}

	// Generate standalone question using chat history
	question, err := c.getQuestion(ctx, query, chatHistoryStr)
	if err != nil {
		return nil, err
	}

	// Retrieve relevant documents using the standalone question
	docs, err := c.Retriever.GetRelevantDocuments(ctx, question)
	if err != nil {
		return nil, err
	}

	// Prepare context from documents
	context := ""
	for i, doc := range docs {
		if i > 0 {
			context += "\n\n"
		}
		context += doc.PageContent
	}

	// Generate final answer with context, chat history, and question
	result, err := Predict(ctx, c.LLMChain, map[string]any{
		"context":      context,
		"chat_history": chatHistoryStr,
		"question":     c.rephraseQuestion(query, question),
	}, options...)
	if err != nil {
		return nil, err
	}

	output := make(map[string]any)
	output[c.OutputKey] = result

	if c.ReturnSourceDocuments {
		output[_conversationalRAGDefaultSourceDocumentKey] = docs
	}
	if c.ReturnGeneratedQuestion {
		output[_conversationalRAGDefaultGeneratedQuestionKey] = question
	}

	return output, nil
}

func (c ConversationalRAG) GetMemory() schema.Memory {
	return c.Memory
}

func (c ConversationalRAG) GetInputKeys() []string {
	return []string{c.InputKey}
}

func (c ConversationalRAG) GetOutputKeys() []string {
	outputKeys := []string{c.OutputKey}
	if c.ReturnSourceDocuments {
		outputKeys = append(outputKeys, _conversationalRAGDefaultSourceDocumentKey)
	}
	if c.ReturnGeneratedQuestion {
		outputKeys = append(outputKeys, _conversationalRAGDefaultGeneratedQuestionKey)
	}
	return outputKeys
}

func (c ConversationalRAG) getQuestion(
	ctx context.Context,
	question string,
	chatHistoryStr string,
) (string, error) {
	if len(chatHistoryStr) == 0 {
		return question, nil
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

func (c ConversationalRAG) rephraseQuestion(question string, newQuestion string) string {
	if c.RephraseQuestion {
		return newQuestion
	}
	return question
}

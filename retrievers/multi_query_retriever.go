package retrievers

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

const (
	//nolint:lll
	_defaultQueryTemplate = `You are an AI language model assistant. Your task is to generate 3 different versions of the given user question to retrieve relevant documents from a vector database. 
By generating multiple perspectives on the user question, your goal is to help the user overcome some of the limitations of distance-based similarity search. Provide these alternative questions separated by newlines. 
Original question: {{.question}}`
	_defaultInputKey = "question"
)

var _ schema.Retriever = &MultiQueryRetriever{}

type MultiQueryRetriever struct {
	Retriever schema.Retriever
	LLMChain  *chains.LLMChain
	// Whether to include the original query in the list of generated queries.
	IncludeOriginal bool
	InputKey        string
	// Delay time in seconds between each query from base retriever, default to 0.
	DelayTime        int
	CallbacksHandler callbacks.Handler
}

// NewMultiQueryRetriever creates a new MultiQueryRetriever.
func NewMultiQueryRetriever(
	retriever schema.Retriever,
	llmChain *chains.LLMChain,
	includeOriginal bool,
) MultiQueryRetriever {
	return MultiQueryRetriever{
		Retriever:       retriever,
		LLMChain:        llmChain,
		IncludeOriginal: includeOriginal,
		InputKey:        _defaultInputKey,
	}
}

func NewMultiQueryRetrieverFromLLM(
	retriever schema.Retriever,
	llm llms.Model,
	prompt prompts.FormatPrompter,
	includeOriginal bool,
	opts ...chains.ChainCallOption,
) MultiQueryRetriever {
	if prompt == nil {
		prompt = prompts.NewPromptTemplate(_defaultQueryTemplate, []string{"question"})
	}
	return MultiQueryRetriever{
		Retriever:       retriever,
		LLMChain:        chains.NewLLMChain(llm, prompt, opts...),
		IncludeOriginal: includeOriginal,
		InputKey:        _defaultInputKey,
	}
}

func (m *MultiQueryRetriever) GetRelevantDocuments(ctx context.Context, query string) ([]schema.Document, error) {
	if m.CallbacksHandler != nil {
		m.CallbacksHandler.HandleRetrieverStart(ctx, query)
	}
	queries, err := m.GenerateQueries(ctx, query)
	if err != nil {
		return nil, err
	}
	if m.IncludeOriginal {
		queries = append(queries, query)
	}
	docs, err := m.RetrieveDocuments(ctx, queries)
	if err != nil {
		return nil, err
	}
	docs = UniqueDocuments(docs)

	if m.CallbacksHandler != nil {
		m.CallbacksHandler.HandleRetrieverEnd(ctx, query, docs)
	}
	return docs, nil
}

// GenerateQueries Generate queries based upon user input.
func (m *MultiQueryRetriever) GenerateQueries(ctx context.Context, query string) ([]string, error) {
	out, err := m.LLMChain.Call(ctx, map[string]any{
		m.InputKey: query,
	})
	if err != nil {
		return nil, err
	}

	v, ok := out[m.LLMChain.OutputKey]
	if !ok {
		return nil, fmt.Errorf("output key %s not found", m.LLMChain.OutputKey)
	}
	text, ok := v.(string)
	if !ok {
		return nil, fmt.Errorf("output key %s is not a string", m.LLMChain.OutputKey)
	}
	return strings.Split(text, "\n"), nil
}

// RetrieveDocuments Run all LLM generated queries and return the results.
func (m *MultiQueryRetriever) RetrieveDocuments(ctx context.Context, queries []string) ([]schema.Document, error) {
	documents := make([]schema.Document, 0)
	for _, q := range queries {
		docs, err := m.Retriever.GetRelevantDocuments(ctx, q)
		if m.DelayTime != 0 {
			time.Sleep(time.Second * time.Duration(m.DelayTime))
		}
		if err != nil {
			return nil, err
		}
		documents = append(documents, docs...)
	}
	return documents, nil
}

func UniqueDocuments(docs []schema.Document) []schema.Document {
	docsMap := make(map[string]schema.Document, len(docs))
	for i, doc := range docs {
		if has, ok := docsMap[doc.PageContent]; ok && has.Score == doc.Score && reflect.DeepEqual(has.Metadata, doc.Metadata) {
			continue
		}
		docsMap[doc.PageContent] = docs[i]
	}

	uniqueDocs := make([]schema.Document, 0, len(docsMap))
	for _, doc := range docsMap {
		uniqueDocs = append(uniqueDocs, doc)
	}
	return uniqueDocs
}

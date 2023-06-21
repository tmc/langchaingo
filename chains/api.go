package chains

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

const (
	// nolint: lll
	_llmAPIURLPrompt = `You are given the below API Documentation:
{{.api_docs}}
Using this documentation, generate the full API url to call for answering the user question.
You should build the API url in order to get a response that is as short as possible, while still getting the necessary information to answer the question. Pay attention to deliberately exclude any unnecessary pieces of data in the API call.

Question:{{.question}}
API url:`

	// nolint: lll
	_llmAPIResponsePrompt = _llmAPIURLPrompt + `{api_url}

Here is the response from the API:

{{.api_response}}

Summarize this response to answer the original question.

Summary:`
)

// HTTPRequester http requester interface.
type HTTPRequester interface {
	Get(url string) (resp *http.Response, err error)
}

// APIChain is a chain used for request api.
type APIChain struct {
	RequestChain *LLMChain
	AnswerChain  *LLMChain
	Requester    HTTPRequester
}

func NewAPIChain(llm llms.LLM) APIChain {
	reqP := prompts.NewPromptTemplate(_llmAPIURLPrompt, []string{"api_docs", "question"})
	reqC := NewLLMChain(llm, reqP)

	respP := prompts.NewPromptTemplate(_llmAPIResponsePrompt, []string{"api_docs", "question", "api_url", "api_response"})
	respC := NewLLMChain(llm, respP)

	return APIChain{
		RequestChain: reqC,
		AnswerChain:  respC,
		Requester:    http.DefaultClient,
	}
}

// Call call api chain.
// Input: api_docs, question.
// Output: answer.
func (a APIChain) Call(ctx context.Context, values map[string]any, opts ...ChainCallOption) (map[string]any, error) {
	tmpOutput, err := Call(ctx, a.RequestChain, values, opts...)
	if err != nil {
		return nil, err
	}
	apiURL, ok := tmpOutput["text"].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %w", ErrInvalidOutputValues, ErrInputValuesWrongType)
	}
	// this is a hack to get the first line of the output
	apiURL = strings.TrimSpace(strings.Split(apiURL, "\n")[0])
	apiResponse, err := a.get(apiURL)
	if err != nil {
		return nil, err
	}

	tmpOutput["api_docs"] = values["api_docs"]
	tmpOutput["question"] = values["question"]
	tmpOutput["api_response"] = apiResponse
	tmpOutput["api_url"] = apiURL

	tmpOutput, err = Call(ctx, a.AnswerChain, tmpOutput, opts...)
	if err != nil {
		return nil, err
	}

	return map[string]any{"answer": tmpOutput["text"]}, err
}

func (a APIChain) GetMemory() schema.Memory { //nolint:ireturn
	return memory.NewSimple()
}

func (a APIChain) GetInputKeys() []string {
	return []string{"api_docs", "question"}
}

func (a APIChain) GetOutputKeys() []string {
	return []string{"answer"}
}

func (a APIChain) get(url string) (string, error) {
	resp, err := a.Requester.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

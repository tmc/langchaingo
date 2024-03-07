package chains

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

//go:embed prompts/llm_api_url.txt
var _llmAPIURLPrompt string //nolint:gochecknoglobals

//go:embed prompts/llm_api_url_response.txt
var _llmAPIURLResponseTmpPrompt string //nolint:gochecknoglobals

var _llmAPIResponsePrompt = _llmAPIURLPrompt + _llmAPIURLResponseTmpPrompt //nolint:gochecknoglobals

// HTTPRequest http requester interface.
type HTTPRequest interface {
	Do(req *http.Request) (*http.Response, error)
}

type APIChain struct {
	RequestChain *LLMChain
	AnswerChain  *LLMChain
	Request      HTTPRequest
}

// NewAPIChain creates a new APIChain object.
//
// It takes a language model (llm) and an HTTPRequest (request) as parameters.
// It returns an APIChain object.
func NewAPIChain(llm llms.Model, request HTTPRequest) APIChain {
	reqPrompt := prompts.NewPromptTemplate(_llmAPIURLPrompt, []string{"api_docs", "input"})
	reqChain := NewLLMChain(llm, reqPrompt)

	resPrompt := prompts.NewPromptTemplate(_llmAPIResponsePrompt, []string{"input", "api_docs", "api_response"})
	resChain := NewLLMChain(llm, resPrompt)

	return APIChain{
		RequestChain: reqChain,
		AnswerChain:  resChain,
		Request:      request,
	}
}

// Call executes the APIChain and returns the result.
//
// It takes a context.Context object, a map[string]any values, and optional ChainCallOption
// values as input parameters. It returns a map[string]any and an error as output.
func (a APIChain) Call(ctx context.Context, values map[string]any, opts ...ChainCallOption) (map[string]any, error) {
	reqChainTmp := 0.0
	opts = append(opts, WithTemperature(reqChainTmp))

	tmpOutput, err := Call(ctx, a.RequestChain, values, opts...)
	if err != nil {
		return nil, err
	}

	outputText, ok := tmpOutput["text"].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %w", ErrInvalidInputValues, ErrInputValuesWrongType)
	}

	// Extract the json from llm output
	re := regexp.MustCompile(`(?s)\{.*\}`)
	jsonString := re.FindString(outputText)

	// Convert the LLM output into the anonymous struct.
	var output struct {
		Method  string            `json:"method"`
		Headers map[string]string `json:"headers"`
		URL     string            `json:"url"`
		Body    map[string]string `json:"body"`
	}

	err = json.Unmarshal([]byte(jsonString), &output)
	if err != nil {
		return nil, err
	}

	apiResponse, err := a.runRequest(ctx, output.Method, output.URL, output.Headers, output.Body)
	if err != nil {
		return nil, err
	}

	tmpOutput["input"] = values["input"]
	tmpOutput["api_docs"] = values["api_docs"]
	tmpOutput["api_response"] = apiResponse

	answer, err := Call(ctx, a.AnswerChain, tmpOutput, opts...)
	if err != nil {
		return nil, err
	}

	return map[string]any{"answer": answer["text"]}, err
}

// GetMemory returns the memory of the APIChain.
//
// This function does not take any parameters.
// It returns a schema.Memory object.
func (a APIChain) GetMemory() schema.Memory { //nolint:ireturn
	return memory.NewSimple()
}

// GetInputKeys returns the input keys of the APIChain.
//
// No parameters.
// Returns a slice of strings, which contains the output keys.
func (a APIChain) GetInputKeys() []string {
	return []string{"api_docs", "input"}
}

// GetOutputKeys returns the output keys of the APIChain.
//
// It does not take any parameters.
// It returns a slice of strings, which contains the output keys.
func (a APIChain) GetOutputKeys() []string {
	return []string{"answer"}
}

func (a APIChain) runRequest(
	ctx context.Context,
	method string,
	url string,
	headers map[string]string,
	body map[string]string,
) (string, error) {
	var bodyReader io.Reader

	if method == "POST" || method == "PUT" {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return "", err
		}

		bodyReader = bytes.NewBuffer(bodyBytes)
	}

	// Create the new request defined by reqChain
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return "", err
	}

	// set request headers passed from reqChain
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	resp, err := a.Request.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(resBody), nil
}

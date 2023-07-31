package chains

import (
	"bytes"
	"context"
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

const (
	// nolint: lll
	_llmAPIURLPrompt = `
	You are given the API Documentation:

	{{.api_docs}}

	Your task is to construct a full API JSON object based on the provided input. The input could be a question that requires an API call for its answer, or a direct or indirect instruction to consume an API. The input will be unpredictable and could come from a user or an agent.

	Your goal is to create an API call that accurately reflects the intent of the input. Be sure to exclude any unnecessary data in the API call to ensure efficiency.

	Input: {{.input}}

	Respond with a JSON object.

	{
		"method":  [the HTTP method for the API call, such as GET or POST],
		"headers": [object representing the HTTP headers required for the API call, always add a "Content-Type" header],
		"url": 	   [full for the API call],
		"body":    [object containing the data sent with the request, if needed]
	}`

	// nolint: lll
	_llmAPIResponsePrompt = _llmAPIURLPrompt + `
	Here is the response from the API:

	{{.api_response}}

	Now, summarize this response. Your summary should reflect the original input and highlight the key information from the API response that answers or relates to that input. Try to make your summary concise, yet informative.

	Summary:`
)

// HTTPRequest http requester interface.
type HTTPRequest interface {
	Do(*http.Request) (*http.Response, error)
}

type APIChain struct {
	RequestChain *LLMChain
	AnswerChain  *LLMChain
	Request      HTTPRequest
}

// NewAPIChain creates a new APIChain object.
//
// It takes a LanguageModel(llm) and an HTTPRequest(request) as parameters.
// It returns an APIChain object.
func NewAPIChain(llm llms.LanguageModel, request HTTPRequest) APIChain {
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

package llamafileclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
)

type Client struct {
	base       *url.URL
	httpClient *http.Client
}

func checkError(resp *http.Response, body []byte) error {
	if resp.StatusCode < http.StatusBadRequest {
		return nil
	}

	apiError := StatusError{StatusCode: resp.StatusCode}

	err := json.Unmarshal(body, &apiError)
	if err != nil {
		// Use the full body as the message if we fail to decode a response.
		apiError.ErrorMessage = string(body)
	}

	return apiError
}

func NewClient(ourl *url.URL, ohttp *http.Client) (*Client, error) {
	if ourl == nil {
		scheme, hostport, ok := strings.Cut(os.Getenv("LLAMAFILE_HOST"), "://")
		if !ok {
			scheme, hostport = "http", os.Getenv("LLAMAFILE_HOST")
		}

		host, port, err := net.SplitHostPort(hostport)
		if err != nil {
			host, port = "127.0.0.1", "8080"
			if ip := net.ParseIP(strings.Trim(os.Getenv("LLAMAFILE_HOST"), "[]")); ip != nil {
				host = ip.String()
			}
		}

		ourl = &url.URL{
			Scheme: scheme,
			Host:   net.JoinHostPort(host, port),
		}
	}

	if ohttp == nil {
		ohttp = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		}
	}

	client := Client{
		base:       ourl,
		httpClient: ohttp,
	}

	return &client, nil
}

func (c *Client) do(ctx context.Context, method, path string, reqData, respData any) error {
	var reqBody io.Reader
	var data []byte
	var err error
	if reqData != nil {
		data, err = json.Marshal(reqData)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(data)
	}

	requestURL := c.base.JoinPath(path)
	request, err := http.NewRequestWithContext(ctx, method, requestURL.String(), reqBody)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent",
		fmt.Sprintf("langchaingo/ (%s %s) Go/%s", runtime.GOARCH, runtime.GOOS, runtime.Version()))

	respObj, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer respObj.Body.Close()

	respBody, err := io.ReadAll(respObj.Body)
	if err != nil {
		return err
	}

	if err := checkError(respObj, respBody); err != nil {
		return err
	}

	if len(respBody) > 0 && respData != nil {
		if err := json.Unmarshal(respBody, respData); err != nil {
			return err
		}
	}
	return nil
}

const maxBufferSize = 512 * 1000

// func (c *Client) stream(ctx context.Context, method, path string, data any, fn func([]byte) error) error {
// 	var buf *bytes.Buffer
// 	if data != nil {
// 		bts, err := json.Marshal(data)
// 		if err != nil {
// 			return err
// 		}

// 		buf = bytes.NewBuffer(bts)
// 	}

// 	// show payload

// 	requestURL := c.base.JoinPath(path)
// 	request, err := http.NewRequestWithContext(ctx, method, requestURL.String(), buf)
// 	if err != nil {
// 		return err
// 	}

// 	request.Header.Set("Content-Type", "application/json")
// 	request.Header.Set("Accept", "application/x-ndjson")
// 	request.Header.Set("User-Agent",
// 		fmt.Sprintf("langchaingo (%s %s) Go/%s", runtime.GOARCH, runtime.GOOS, runtime.Version()))

// 	response, err := c.httpClient.Do(request)
// 	if err != nil {
// 		return err
// 	}
// 	defer response.Body.Close()

// 	scanner := bufio.NewScanner(response.Body)
// 	// increase the buffer size to avoid running out of space
// 	scanBuf := make([]byte, 0, maxBufferSize)
// 	scanner.Buffer(scanBuf, maxBufferSize)
// 	for scanner.Scan() {
// 		var errorResponse struct {
// 			Error string `json:"error,omitempty"`
// 		}

// 		bts, errBts := ExtractJSONFromBytes(scanner.Bytes())
// 		if errBts != nil {
// 			return errBts
// 		}

// 		// if bts is nil then continue
// 		if bts == nil {
// 			continue
// 		}

// 		if err := json.Unmarshal(bts, &errorResponse); err != nil {
// 			return err
// 		}

// 		if errorResponse.Error != "" {
// 			return errors.New(errorResponse.Error) //nolint
// 		}

// 		if response.StatusCode >= http.StatusBadRequest {
// 			return StatusError{
// 				StatusCode:   response.StatusCode,
// 				Status:       response.Status,
// 				ErrorMessage: errorResponse.Error,
// 			}
// 		}

// 		if err := fn(bts); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// stream manages the streaming and processing of data from an HTTP request.
func (c *Client) stream(ctx context.Context, method, path string, data any, fn func([]byte) error) error {
	buf, err := prepareBuffer(data)
	if err != nil {
		return err
	}

	response, err := c.sendHTTPRequest(ctx, method, path, buf)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	return c.processResponse(response, fn)
}

// prepareBuffer marshals data to JSON if not nil, returning a buffer.
func prepareBuffer(data any) (*bytes.Buffer, error) {
	if data == nil {
		return nil, errors.New("data is nil")
	}
	bts, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(bts), nil
}

// sendHTTPRequest sends an HTTP request and returns the response.
func (c *Client) sendHTTPRequest(ctx context.Context, method, path string, buf *bytes.Buffer) (*http.Response, error) {
	requestURL := c.base.JoinPath(path)
	request, err := http.NewRequestWithContext(ctx, method, requestURL.String(), buf)
	if err != nil {
		return nil, err
	}
	setRequestHeaders(request)

	return c.httpClient.Do(request)
}

// setRequestHeaders sets the necessary headers for the HTTP request.
func setRequestHeaders(request *http.Request) {
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/x-ndjson")
	request.Header.Set("User-Agent", fmt.Sprintf("langchaingo (%s %s) Go/%s", runtime.GOARCH, runtime.GOOS, runtime.Version()))
}

// processResponse handles the HTTP response, parsing and forwarding JSON data.
func (c *Client) processResponse(response *http.Response, fn func([]byte) error) error {
	scanner := bufio.NewScanner(response.Body)
	scanner.Buffer(make([]byte, 0, maxBufferSize), maxBufferSize) // Assume maxBufferSize is defined

	for scanner.Scan() {
		if err := processScan(scanner.Bytes(), response, fn); err != nil {
			return err
		}
	}
	return scanner.Err() // Check for scanning errors
}

// processScan handles the scanned bytes from the response body.
func processScan(bts []byte, response *http.Response, fn func([]byte) error) error {
	bts, err := ExtractJSONFromBytes(bts)
	if err != nil {
		return err
	}
	if bts == nil { // if bts is nil then continue
		return nil
	}

	var errorResponse struct {
		Error string `json:"error,omitempty"`
	}
	if err := json.Unmarshal(bts, &errorResponse); err != nil {
		return err
	}
	if errorResponse.Error != "" {
		return errors.New(errorResponse.Error)
	}
	if response.StatusCode >= http.StatusBadRequest {
		return StatusError{
			StatusCode:   response.StatusCode,
			Status:       response.Status,
			ErrorMessage: errorResponse.Error,
		}
	}

	return fn(bts)
}

type (
	GenerateResponseFunc func(GenerateResponse) error
	ChatResponseFunc     func(ChatResponse) error
)

func (c *Client) Generate(ctx context.Context, req *GenerateRequest, fn GenerateResponseFunc) error {
	return c.stream(ctx, http.MethodPost, "/completion", req, func(bts []byte) error {
		var resp GenerateResponse
		if err := json.Unmarshal(bts, &resp); err != nil {
			return err
		}

		return fn(resp)
	})
}

func (c *Client) GenerateChat(ctx context.Context, req *ChatRequest, fn ChatResponseFunc) error {
	prompt := "<s>[INST]"
	for _, msg := range req.Messages {
		switch msg.Role {
		// "system", "user", "assistant"]
		case "system":
			prompt += fmt.Sprintf("<<SYS>> %s <</SYS>>\n", msg.Content)
		case "user":
			prompt += fmt.Sprintf("USER: %s\n", msg.Content)
		case "assistant":
			prompt += fmt.Sprintf("ASSISTANT: %s\n", msg.Content)
		default:
			prompt += fmt.Sprintf("[UNKNOWN]: %s\n", msg.Content)
		}
	}
	prompt += "[/INST]</s>"
	req.Prompt = &prompt

	if req.Temperature == 0 {
		req.Temperature = 0.7
	}

	if req.Temperature == 0 {
		req.Temperature = 0.7
	}

	return c.stream(ctx, http.MethodPost, "/completion", req, func(bts []byte) error {
		var resp ChatResponse
		if err := json.Unmarshal(bts, &resp); err != nil {
			return err
		}

		return fn(resp)
	})
}

type EmbeddingRequest struct {
	Content []string `json:"content"`
}
type EmbeddingResponse struct {
	Results []EmbeddingData `json:"results"`
}
type EmbeddingData struct {
	Embedding []float64 `json:"embedding"`
}

func (c *Client) CreateEmbedding(ctx context.Context, texts []string) ([][]float32, error) {
	req := &EmbeddingRequest{
		Content: texts,
	}

	var resp EmbeddingResponse

	err := c.do(ctx, http.MethodPost, "/embedding", req, &resp)

	embeddings := convertEmbeddingsToFloat32(resp.Results)

	return embeddings, err
}

func ExtractJSONFromBytes(input []byte) ([]byte, error) {
	// Convert input byte slice to string
	inputStr := string(input)

	if inputStr == "" {
		return nil, errors.New("input is empty")
	}

	// Trim the prefix "data: " from the string
	trimmedStr := strings.TrimPrefix(inputStr, "data: ")

	// The trimmed string is supposed to be a JSON, but it's potentially in escaped format.
	// We'll use json.RawMessage for its ability to be a valid JSON component
	var raw json.RawMessage
	if err := json.Unmarshal([]byte(trimmedStr), &raw); err != nil {
		return nil, errors.New("failed to unmarshal JSON: " + err.Error())
	}

	// Return the cleaned JSON as byte slice
	return raw, nil
}

// convertEmbeddingsToFloat32 takes a slice of EmbeddingData (as found in EmbeddingResponse.Results)
// and converts each embedding into a slice of float32, returning a slice of slices of float32.
func convertEmbeddingsToFloat32(data []EmbeddingData) [][]float32 {
	// Preallocate result with the same number of embeddings (outer slice length)
	result := make([][]float32, 0, len(data))

	// Iterate through each EmbeddingData
	for _, embeddingData := range data {
		// Convert embedding from float64 to float32
		innerSlice := make([]float32, len(embeddingData.Embedding))
		for i, value := range embeddingData.Embedding {
			innerSlice[i] = float32(value)
		}

		// Append the converted embedding to the result
		result = append(result, innerSlice)
	}
	return result
}

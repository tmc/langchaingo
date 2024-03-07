package ollamaclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
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
		scheme, hostport, ok := strings.Cut(os.Getenv("OLLAMA_HOST"), "://")
		if !ok {
			scheme, hostport = "http", os.Getenv("OLLAMA_HOST")
		}

		host, port, err := net.SplitHostPort(hostport)
		if err != nil {
			host, port = "127.0.0.1", "11434"
			if ip := net.ParseIP(strings.Trim(os.Getenv("OLLAMA_HOST"), "[]")); ip != nil {
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

func (c *Client) stream(ctx context.Context, method, path string, data any, fn func([]byte) error) error {
	var buf *bytes.Buffer
	if data != nil {
		bts, err := json.Marshal(data)
		if err != nil {
			return err
		}

		buf = bytes.NewBuffer(bts)
	}

	requestURL := c.base.JoinPath(path)
	request, err := http.NewRequestWithContext(ctx, method, requestURL.String(), buf)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/x-ndjson")
	request.Header.Set("User-Agent",
		fmt.Sprintf("langchaingo (%s %s) Go/%s", runtime.GOARCH, runtime.GOOS, runtime.Version()))

	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	scanner := bufio.NewScanner(response.Body)
	// increase the buffer size to avoid running out of space
	scanBuf := make([]byte, 0, maxBufferSize)
	scanner.Buffer(scanBuf, maxBufferSize)
	for scanner.Scan() {
		var errorResponse struct {
			Error string `json:"error,omitempty"`
		}

		bts := scanner.Bytes()
		if err := json.Unmarshal(bts, &errorResponse); err != nil {
			return err
		}

		if errorResponse.Error != "" {
			return fmt.Errorf(errorResponse.Error) //nolint
		}

		if response.StatusCode >= http.StatusBadRequest {
			return StatusError{
				StatusCode:   response.StatusCode,
				Status:       response.Status,
				ErrorMessage: errorResponse.Error,
			}
		}

		if err := fn(bts); err != nil {
			return err
		}
	}

	return nil
}

type (
	GenerateResponseFunc func(GenerateResponse) error
	ChatResponseFunc     func(ChatResponse) error
)

func (c *Client) Generate(ctx context.Context, req *GenerateRequest, fn GenerateResponseFunc) error {
	return c.stream(ctx, http.MethodPost, "/api/generate", req, func(bts []byte) error {
		var resp GenerateResponse
		if err := json.Unmarshal(bts, &resp); err != nil {
			return err
		}

		return fn(resp)
	})
}

func (c *Client) GenerateChat(ctx context.Context, req *ChatRequest, fn ChatResponseFunc) error {
	return c.stream(ctx, http.MethodPost, "/api/chat", req, func(bts []byte) error {
		var resp ChatResponse
		if err := json.Unmarshal(bts, &resp); err != nil {
			return err
		}

		return fn(resp)
	})
}

func (c *Client) CreateEmbedding(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	resp := &EmbeddingResponse{}
	if err := c.do(ctx, http.MethodPost, "/api/embeddings", req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

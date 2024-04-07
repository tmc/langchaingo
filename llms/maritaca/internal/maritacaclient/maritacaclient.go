package maritacaclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

var baseUrl = "https://chat.maritaca.ai/api"

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}
type Client struct {
	Token      string
	Model      string
	baseURL    string
	httpClient Doer
}

func NewClient(ohttp Doer) (*Client, error) {
	client := Client{
		baseURL:    baseUrl,
		httpClient: ohttp,
	}

	return &client, nil
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

	requestURL := fmt.Sprintf("%s%s", c.baseURL, path)
	request, err := http.NewRequestWithContext(ctx, method, requestURL, buf)
	if err != nil {
		return err
	}

	token := fmt.Sprintf("key %v", c.Token)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", token)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusBadRequest {
		var errorResponse struct {
			Error string `json:"detail,omitempty"`
		}

		if err := json.NewDecoder(response.Body).Decode(&errorResponse); err != nil {
			return err
		}

		return StatusError{
			StatusCode:   response.StatusCode,
			Status:       response.Status,
			ErrorMessage: errorResponse.Error,
		}
	}

	scanner := bufio.NewScanner(response.Body)
	scanBuf := make([]byte, 0, maxBufferSize)
	scanner.Buffer(scanBuf, maxBufferSize)
	for scanner.Scan() {
		bts := scanner.Bytes()
		if err := fn(bts); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

type (
	ChatResponseFunc func(ChatResponse) error
)

func (c *Client) Generate(ctx context.Context, req *ChatRequest, fn ChatResponseFunc) error {
	return c.stream(ctx, http.MethodPost, "/chat/inference", req, func(bts []byte) error {
		var resp ChatResponse
		if req.Options.Stream {
			text, errP := parseData(string(bts))
			resp.Event = "message"
			if string(bts) == "event: end" {
				resp.Event = "end"
			}

			if errP != nil && resp.Event != "end" {
				return nil
			}

			resp.Text = text
			return fn(resp)
		}
		if err := json.Unmarshal(bts, &resp); err != nil {
			return err
		}

		resp.Event = "nostream"

		return fn(resp)
	})
}

func parseData(input string) (string, error) {
	if !strings.Contains(input, "data:") {
		return "", nil
	}

	parts := strings.SplitAfter(input, "data:")
	if len(parts) < 2 {
		return "", nil
	}

	var data map[string]interface{}
	err := json.Unmarshal([]byte(parts[1]), &data)
	if err != nil {
		return "", err
	}

	text, ok := data["text"].(string)
	if !ok {
		return "", nil
	}

	return text, nil
}

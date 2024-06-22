package cloudflareclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// CreateEmbedding creates an embedding from the given texts.
func (c *Client) CreateEmbedding(ctx context.Context, texts *CreateEmbeddingRequest) (*CreateEmbeddingResponse, error) {
	requestBody, err := json.Marshal(texts)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpointURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", c.bearerToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("error: %s", body)
	}

	var createEmbeddingResponse CreateEmbeddingResponse
	err = json.Unmarshal(body, &createEmbeddingResponse)
	if err != nil {
		return nil, err
	}

	return &createEmbeddingResponse, nil
}

const maxBufferSize = 512 * 1000

// GenerateContent generates text based on the given prompts.
func (c *Client) GenerateContent(ctx context.Context, request *GenerateContentRequest) (*GenerateContentResponse, error) { // nolint:funlen,cyclop
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpointURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", c.bearerToken)
	req.Header.Add("Content-Type", "application/json")

	response, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if !request.Stream || request.StreamingFunc == nil {
		var body []byte

		body, err = io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}

		if response.StatusCode >= http.StatusMultipleChoices {
			return nil, fmt.Errorf("error: %s", body)
		}

		var generateResponse GenerateContentResponse
		err = json.Unmarshal(body, &generateResponse)
		if err != nil {
			return nil, err
		}

		return &generateResponse, nil
	}

	scanner := bufio.NewScanner(response.Body)
	// increase the buffer size to avoid running out of space
	scanBuf := make([]byte, 0, maxBufferSize)
	scanner.Buffer(scanBuf, maxBufferSize)
	for scanner.Scan() {
		var streamingResponse StreamingResponse

		stext := scanner.Text()

		if stext == "" {
			continue
		}

		stext = strings.TrimPrefix(stext, "data: ")

		if strings.Contains(stext, "[DONE]") {
			break
		}

		bts := []byte(stext)

		if err = json.Unmarshal(bts, &streamingResponse); err != nil {
			return nil, err
		}

		if response.StatusCode >= http.StatusBadRequest {
			var body []byte

			body, err = io.ReadAll(response.Body)
			if err != nil {
				return nil, err
			}
			return &GenerateContentResponse{
				Errors: []APIError{{Message: string(body)}},
			}, nil
		}

		if err = request.StreamingFunc(ctx, bts); err != nil {
			return nil, err
		}
	}

	return &GenerateContentResponse{}, nil
}

// Summarize summarizes the given input text.
func (c *Client) Summarize(ctx context.Context, inputText string, maxLength int) (*SummarizeResponse, error) {
	requestBody, err := json.Marshal(SummarizeRequest{InputText: inputText, MaxLength: maxLength})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpointURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", c.bearerToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("error: %s", body)
	}

	var summarizeResponse SummarizeResponse
	err = json.Unmarshal(body, &summarizeResponse)
	if err != nil {
		return nil, err
	}

	return &summarizeResponse, nil
}

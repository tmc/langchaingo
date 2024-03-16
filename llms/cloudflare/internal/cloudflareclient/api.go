package cloudflareclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c Client) CreateEmbedding(ctx context.Context, texts *CreateEmbeddingRequest) (*CreateEmbeddingResponse, error) {
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

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("error: %s", body)
	}

	var createEmbeddingResponse CreateEmbeddingResponse
	err = json.Unmarshal(body, &createEmbeddingResponse)
	if err != nil {
		return nil, err
	}

	return &createEmbeddingResponse, nil
}

func (c Client) GenerateContent(ctx context.Context, request *GenerateContentRequest, stream bool) (*GenerateContentResponse, error) {
	request.Stream = stream

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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("error: %s", body)
	}

	var generateResponse GenerateContentResponse
	err = json.Unmarshal(body, &generateResponse)
	if err != nil {
		return nil, err
	}

	return &generateResponse, nil
}

func (c Client) Summarize(ctx context.Context, inputText string, maxLength int) (*SummarizeResponse, error) {
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

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("error: %s", body)
	}

	var summarizeResponse SummarizeResponse
	err = json.Unmarshal(body, &summarizeResponse)
	if err != nil {
		return nil, err
	}

	return &summarizeResponse, nil
}

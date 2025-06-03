package huggingfaceclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/tmc/langchaingo/httputil"
)

type embeddingPayload struct {
	Options map[string]any
	Inputs  []string `json:"inputs"`
}

// nolint:lll
func (c *Client) createEmbedding(ctx context.Context, model string, task string, payload *embeddingPayload) ([][]float32, error) {
	body := map[string]any{
		"inputs": payload.Inputs,
	}
	for key, value := range payload.Options {
		body[key] = value
	}

	payloadBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	// Use /models/ endpoint for embeddings as /pipeline/ is deprecated
	url := fmt.Sprintf("%s/models/%s", c.url, model)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	r, err := httputil.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		var body []byte
		if r.Body != nil {
			body, _ = io.ReadAll(r.Body)
		}
		msg := fmt.Sprintf("API returned unexpected status code: %d for URL: %s", r.StatusCode, url)
		if len(body) > 0 {
			msg = fmt.Sprintf("%s, body: %s", msg, string(body))
		}
		return nil, fmt.Errorf("%s: %s", msg, "unable to create embeddings") // nolint:goerr113
	}

	var response [][]float32
	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return response, nil
}

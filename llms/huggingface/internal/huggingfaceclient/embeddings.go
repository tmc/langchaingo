package huggingfaceclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/pipeline/%s/%s", c.url, task, model), bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-wait-for-model", "true")

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("API returned unexpected status code: %d", r.StatusCode)
		b, _ := io.ReadAll(r.Body)

		return nil, fmt.Errorf("%s: %s: %s",
			msg, "unable to create embeddings", string(b)) // nolint:goerr113
	}

	var response [][]float32
	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return response, nil
}

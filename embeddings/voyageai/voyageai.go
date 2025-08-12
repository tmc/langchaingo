package voyageai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vendasta/langchaingo/embeddings"
)

var _ embeddings.Embedder = &VoyageAI{}

// VoyageAI is the embedder using the VoyageAI api to create embeddings.
type VoyageAI struct {
	baseURL       string
	token         string
	client        *http.Client
	Model         string
	StripNewLines bool
	BatchSize     int
}

// NewVoyageAI returns a new embedder that uses the VoyageAI api.
// The default model is "voyage-2". Use `WithModel` to change the model.
func NewVoyageAI(opts ...Option) (*VoyageAI, error) {
	v, err := applyOptions(opts...)
	if err != nil {
		return nil, err
	}
	return v, nil
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	}
}

type embedDocumentsRequest struct {
	Model     string   `json:"model"`
	Input     []string `json:"input"`
	InputType string   `json:"input_type"`
}

// EmbedDocuments implements the `embeddings.Embedder` and creates an embedding for each of the texts.
func (v *VoyageAI) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	batchedTexts := embeddings.BatchTexts(
		embeddings.MaybeRemoveNewLines(texts, v.StripNewLines),
		v.BatchSize,
	)

	embeddings := make([][]float32, 0, len(texts))
	for _, batch := range batchedTexts {
		req := embedDocumentsRequest{
			Model:     v.Model,
			Input:     batch,
			InputType: "document",
		}

		resp, err := v.request(ctx, "/embeddings", req)
		if err != nil {
			return nil, fmt.Errorf("embed documents request error: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, v.decodeError(resp)
		}

		var embeddingResp embeddingResponse
		if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
			return nil, err
		}

		for _, data := range embeddingResp.Data {
			embeddings = append(embeddings, data.Embedding)
		}
	}
	return embeddings, nil
}

type embedQueryRequest struct {
	Model     string `json:"model"`
	Input     string `json:"input"`
	InputType string `json:"input_type"`
}

// EmbedQuery implements the `embeddings.Embedder` and creates an embedding for the query text.
func (v *VoyageAI) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	req := embedQueryRequest{
		Model:     v.Model,
		Input:     text,
		InputType: "query",
	}
	resp, err := v.request(ctx, "/embeddings", req)
	if err != nil {
		return nil, fmt.Errorf("embed query request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, v.decodeError(resp)
	}

	var embeddingResp embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		return nil, err
	}
	return embeddingResp.Data[0].Embedding, nil
}

func (v *VoyageAI) request(ctx context.Context, path string, body any) (*http.Response, error) {
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, v.baseURL+path, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+v.token)
	httpReq.Header.Set("Content-Type", "application/json")

	return v.client.Do(httpReq)
}

func (v *VoyageAI) decodeError(resp *http.Response) error {
	var errResp struct {
		Detail string `json:"detail"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return fmt.Errorf("unexpected error: %w", err)
	}
	return fmt.Errorf("embedding error: %s", errResp.Detail)
}

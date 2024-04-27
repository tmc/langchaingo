package jina

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/tmc/langchaingo/embeddings"
)

type Jina struct {
	Model         string
	InputText     []string
	StripNewLines bool
	BatchSize     int
	APIBaseURL    string
	APIKey        string
}

type EmbeddingRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

type EmbeddingResponse struct {
	Model  string `json:"model"`
	Object string `json:"object"`
	Usage  struct {
		TotalTokens  int `json:"total_tokens"`
		PromptTokens int `json:"prompt_tokens"`
	} `json:"usage"`
	Data []struct {
		Object    string    `json:"object"`
		Index     int       `json:"index"`
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

var _ embeddings.Embedder = &Jina{}

func NewJina(opts ...Option) (*Jina, error) {
	v, err := applyOptions(opts...)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (j *Jina) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	batchedTexts := embeddings.BatchTexts(
		embeddings.MaybeRemoveNewLines(texts, j.StripNewLines),
		j.BatchSize,
	)

	emb := make([][]float32, 0, len(texts))
	for _, batch := range batchedTexts {
		curBatchEmbeddings, err := j.CreateEmbedding(ctx, batch)
		if err != nil {
			return nil, err
		}
		emb = append(emb, curBatchEmbeddings...)
	}

	return emb, nil
}

func (j *Jina) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	if j.StripNewLines {
		text = strings.ReplaceAll(text, "\n", " ")
	}

	emb, err := j.CreateEmbedding(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	return emb[0], nil
}

// CreateEmbedding sends texts to the Jina API and retrieves their embeddings.
func (j *Jina) CreateEmbedding(ctx context.Context, texts []string) ([][]float32, error) {
	requestBody := EmbeddingRequest{
		Input: texts,
		Model: j.Model,
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", j.APIBaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+j.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("API request failed with status: " + resp.Status)
	}

	var embeddingResponse EmbeddingResponse
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &embeddingResponse)
	if err != nil {
		return nil, err
	}

	embs := make([][]float32, 0, len(embeddingResponse.Data))
	for _, data := range embeddingResponse.Data {
		embs = append(embs, data.Embedding)
	}

	return embs, nil
}

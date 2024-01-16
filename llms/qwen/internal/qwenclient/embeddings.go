package qwen_client

import (
	"context"
)

const (
	embeddingUrl          = "https://dashscope.aliyuncs.com/api/v1/services/embeddings/text-embedding/text-embedding"
	defaultEmbeddingModel = "text-embedding-v1"
)

type EmbeddingRequest struct {
	Model string `json:"model"`
	Input struct {
		Texts []string `json:"texts"`
	} `json:"input"`
	Params struct {
		TextType string `json:"text_type"` // query or document
	} `json:"parameters"`
}

type Embedding struct {
	TextIndex int       `json:"text_index"`
	Embedding []float32 `json:"embedding"`
}

type EmbeddingOutput struct {
	Embeddings []Embedding `json:"embeddings"`
	Usgae      struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
	RequestId string `json:"request_id"`
}

type EmbeddingResponse struct {
	Output EmbeddingOutput `json:"output"`
}

func (q *QwenClient) createEmbedding(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {

	header := map[string]string{
		"Authorization": "Bearer " + q.token,
	}
	resp := EmbeddingResponse{}
	headerOpt := WithHeader(header)

	err := q.httpCli.Post(ctx, embeddingUrl, req, &resp, headerOpt)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

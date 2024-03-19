package cloudflareclient

import "context"

type GenerateContentRequest struct {
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`

	// StreamingFunc is a function to be called for each chunk of a streaming response.
	// Return an error to stop streaming early.
	StreamingFunc func(ctx context.Context, chunk []byte) error `json:"-"`
}

type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

type GenerateContentResponse struct {
	Errors   []APIError `json:"errors"`
	Messages []string   `json:"messages"`
	Result   struct {
		Response string `json:"response"`
	} `json:"result"`
	Success bool `json:"success"`
}

type StreamingResponse struct {
	Response string `json:"response"`
	P        string `json:"p"`
}

type APIError struct {
	Message string `json:"message"`
}

type SummarizeRequest struct {
	InputText string `json:"input_text"`
	MaxLength int    `json:"max_length"`
}

type SummarizeResponse struct {
	Result struct {
		Summary string `json:"summary"`
	} `json:"result"`
	Success bool `json:"success"`
}

type CreateEmbeddingRequest struct {
	Text []string `json:"text"`
}

type CreateEmbeddingResponse struct {
	Result struct {
		Shape []int       `json:"shape"`
		Data  [][]float32 `json:"data"`
	} `json:"result"`
}

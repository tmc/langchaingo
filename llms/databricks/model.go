package databricks

import (
	"context"

	"github.com/tmc/langchaingo/llms"
)

// Model is the interface that wraps the methods to format the payload and response.
type Model interface {
	FormatPayload(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) ([]byte, error)
	FormatResponse(ctx context.Context, response []byte) (*llms.ContentResponse, error)
	FormatStreamResponse(ctx context.Context, response []byte) (*llms.ContentResponse, error)
}

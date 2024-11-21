package databricks

import (
	"context"

	"github.com/tmc/langchaingo/llms"
)

type Model interface {
	FormatPayload(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) ([]byte, error)
	FormatResponse(ctx context.Context, response []byte) (*llms.ContentResponse, error)
	FormatStreamResponse(ctx context.Context, response []byte) (*llms.ContentResponse, error)
}

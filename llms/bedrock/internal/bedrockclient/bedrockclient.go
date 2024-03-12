package bedrockclient

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

type Client struct {
	client *bedrockruntime.Client
}

type Message struct {
	Role    schema.ChatMessageType
	Content string
	// Type may be "text" or "image"
	Type string
	// MimeType is the MIME type
	MimeType string
}

func getProvider(modelID string) string {
	return strings.Split(modelID, ".")[0]
}

func NewClient(client *bedrockruntime.Client) *Client {
	return &Client{
		client: client,
	}
}

func (c *Client) CreateCompletion(ctx context.Context,
	modelID string,
	messages []Message,
	options llms.CallOptions,
) (*llms.ContentResponse, error) {
	provider := getProvider(modelID)
	switch provider {
	case "ai21":
		return createAi21Completion(ctx, c.client, modelID, messages, options)
	case "amazon":
		return createAmazonCompletion(ctx, c.client, modelID, messages, options)
	case "anthropic":
		return createAnthropicCompletion(ctx, c.client, modelID, messages, options)
	case "cohere":
		return createCohereCompletion(ctx, c.client, modelID, messages, options)
	case "meta":
		return createMetaCompletion(ctx, c.client, modelID, messages, options)
	default:
		return nil, errors.New("unsupported provider")
	}
}

func processInputMessagesGeneric(messages []Message) string {
	var sb strings.Builder
	var hasRole bool
	for _, message := range messages {
		if message.Role != "" {
			hasRole = true
			sb.WriteString("\n")
			sb.WriteString(string(message.Role))
			sb.WriteString(": ")
		}
		if message.Type == "text" {
			sb.WriteString(message.Content)
		}
	}
	if hasRole {
		sb.WriteString("\n")
		sb.WriteString("AI: ")
	}
	return sb.String()
}

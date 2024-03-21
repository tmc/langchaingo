package awsclient

import (
	"context"
	"errors"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

type Options interface{}
type AwsRuntimeClient interface {
	InvokeModel(ctx context.Context, params *InvokeModelInput, optFns ...func(Options)) (*InvokeModelOutput, error)
}

// Client is a Bedrock client.
type Client struct {
	Client AwsRuntimeClient
}

// Message is a chunk of text or an data
// that will be sent to the provider.
//
// The provider may then transform the message to its own
// format before sending it to the LLM model API.
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

// NewClient creates a new Bedrock client.
func NewClient(client interface{}) *Client {
	//interface to AwsRuntimeClient
	return &Client{
		Client: client.(AwsRuntimeClient),
	}
}

// CreateCompletion creates a new completion response from the provider
// after sending the messages to the provider.
func (c *Client) CreateCompletion(ctx context.Context,
	modelID string,
	messages []Message,
	options llms.CallOptions,
) (*llms.ContentResponse, error) {
	provider := getProvider(modelID)
	switch provider {
	case "ai21":
		return createAi21Completion(ctx, c.Client, modelID, messages, options)
	case "amazon":
		return createAmazonCompletion(ctx, c.Client, modelID, messages, options)
	case "anthropic":
		return createAnthropicCompletion(ctx, c.Client, modelID, messages, options)
	case "cohere":
		return createCohereCompletion(ctx, c.Client, modelID, messages, options)
	case "meta":
		return createMetaCompletion(ctx, c.Client, modelID, messages, options)
	default:
		return nil, errors.New("unsupported provider")
	}
}

// Helper function to process input text chat
// messages as a single string.
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

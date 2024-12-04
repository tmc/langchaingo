package bedrockclient

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/tmc/langchaingo/llms"
)

// Client is a Bedrock client.
type Client struct {
	client *bedrockruntime.Client
}

// Message is a chunk of text or an data
// that will be sent to the provider.
//
// The provider may then transform the message to its own
// format before sending it to the LLM model API.
type Message struct {
	Role    llms.ChatMessageType
	Content string
	// Type may be "text" or "image"
	Type string
	// MimeType is the MIME type
	MimeType string
}

func getProvider(modelID string) string {
	parts := strings.Split(modelID, ".")
	if len(parts) >= 2 && strings.HasPrefix(parts[1], "nova") {
		return "nova"
	}
	
	// For backward compatibility with the original provider detection
	switch {
	case strings.Contains(modelID, "ai21"):
		return "ai21"
	case strings.Contains(modelID, "amazon"):
		return "amazon"
	case strings.Contains(modelID, "anthropic"):
		return "anthropic"
	case strings.Contains(modelID, "cohere"):
		return "cohere"
	case strings.Contains(modelID, "meta"):
		return "meta"
	}

	// Default to using the first part of the model ID
	if len(parts) > 0 {
		return parts[0]
	}
	
	return ""
}

// NewClient creates a new Bedrock client.
func NewClient(client *bedrockruntime.Client) *Client {
	return &Client{
		client: client,
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
		return createAi21Completion(ctx, c.client, modelID, messages, options)
	case "amazon":
		return createAmazonCompletion(ctx, c.client, modelID, messages, options)
	case "nova":
		return createNovaCompletion(ctx, c.client, modelID, messages, options)
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

func getMaxTokens(maxTokens, defaultValue int) int {
	if maxTokens <= 0 {
		return defaultValue
	}
	return maxTokens
}

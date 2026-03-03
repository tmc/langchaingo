package bedrockclient

import (
	"context"
	"errors"
	"strings"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
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
	Role llms.ChatMessageType
	// Content contains the main message content
	Content string
	// Type may be "text", "image", "tool_use", "tool_result"
	Type string
	// MimeType is the MIME type for image content
	MimeType string
	// Tool calling fields
	ToolCall   *ToolCall   `json:"tool_call,omitempty"`
	ToolResult *ToolResult `json:"tool_result,omitempty"`
	// Reasoning contains thinking content for AI messages
	Reasoning *reasoning.ContentReasoning `json:"reasoning,omitempty"`
	// CacheControl for prompt caching (used by Legacy Anthropic API)
	CacheControl *CacheControl `json:"cache_control,omitempty"`
}

// CacheControl represents cache configuration for prompt caching
type CacheControl struct {
	Type string `json:"type"`
	TTL  string `json:"ttl,omitempty"`
}

// ToolCall represents a function call request from the model
type ToolCall struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// ToolResult represents the result of a function call execution
type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	ToolName   string `json:"tool_name"`
	Content    string `json:"content"`
}

func GetProvider(modelID string) string {
	// Check for Nova models (including inference profiles like us.amazon.nova-*)
	if strings.Contains(modelID, ".nova-") || strings.Contains(modelID, "amazon.nova-") {
		return "nova"
	}

	parts := strings.Split(modelID, ".")

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
	case strings.Contains(modelID, "deepseek"):
		return "deepseek"
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
	provider := GetProvider(modelID)
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
	case "deepseek":
		return createDeepSeekCompletion(ctx, c.client, modelID, messages, options)
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

package bedrockclient

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
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
	return strings.Split(modelID, ".")[0]
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

// Converse converses with the model.
func (c *Client) Converse(ctx context.Context,
	modelID string,
	llmsMessages []llms.MessageContent,
	options *llms.CallOptions,
) (*llms.ContentResponse, error) {
	system, messages, err := processInputMessagesBedrock(llmsMessages)
	if err != nil {
		return nil, err
	}
	temperature := float32(options.Temperature)
	inferenceConfig := &types.InferenceConfiguration{
		StopSequences: options.StopWords,
		Temperature:   &temperature,
	}
	if options.MaxTokens > 0 {
		maxTokens := int32(options.MaxTokens)
		inferenceConfig.MaxTokens = &maxTokens
	}
	if options.TopP > 0 {
		topP := float32(options.TopP)
		inferenceConfig.TopP = &topP
	}
	var toolConfig *types.ToolConfiguration

	if len(options.Tools) > 0 {
		toolConfig = &types.ToolConfiguration{
			Tools: convertTools(options.Tools),
		}
		if options.ToolChoice != nil {
			toolChoice, err := convertToolChoice(options.ToolChoice)
			if err != nil {
				return nil, err
			}
			toolConfig.ToolChoice = toolChoice
		}
	}

	if options.StreamingFunc != nil {
		input := &bedrockruntime.ConverseStreamInput{
			Messages:        messages,
			ModelId:         &modelID,
			System:          system,
			GuardrailConfig: nil,
			InferenceConfig: inferenceConfig,
			ToolConfig:      toolConfig,
		}
		output, err := c.client.ConverseStream(ctx, input)
		if err != nil {
			return nil, err
		}
		resp, err := handleConverseStreamEvents(ctx, output, options)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}

	input := &bedrockruntime.ConverseInput{
		Messages:        messages,
		ModelId:         &modelID,
		GuardrailConfig: nil,
		InferenceConfig: inferenceConfig,
		System:          system,
		ToolConfig:      toolConfig,
	}
	output, err := c.client.Converse(ctx, input)
	if err != nil {
		return nil, err
	}
	return handleConverseOutput(ctx, output)
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

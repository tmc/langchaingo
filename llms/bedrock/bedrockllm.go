package bedrock

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/bedrock/internal/bedrockclient"
)

const defaultModel = ModelAmazonTitanTextLiteV1

// LLM is a Bedrock LLM implementation.
type LLM struct {
	modelID          string
	client           *bedrockclient.Client
	CallbacksHandler callbacks.Handler
}

// New creates a new Bedrock LLM implementation.
func New(opts ...Option) (*LLM, error) {
	return NewWithContext(context.Background(), opts...)
}

// NewWithContext creates a new Bedrock LLM implementation with context.
func NewWithContext(ctx context.Context, opts ...Option) (*LLM, error) {
	o, c, err := newClient(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return &LLM{
		client:           c,
		modelID:          o.modelID,
		CallbacksHandler: o.callbackHandler,
	}, nil
}

func newClient(ctx context.Context, opts ...Option) (*options, *bedrockclient.Client, error) {
	options := &options{
		modelID: defaultModel,
	}

	for _, opt := range opts {
		opt(options)
	}

	if options.client == nil {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return options, nil, err
		}
		options.client = bedrockruntime.NewFromConfig(cfg)
	}

	return options, bedrockclient.NewClient(options.client), nil
}

// Call implements llms.Model.
func (l *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, l, prompt, options...)
}

// GenerateContent implements llms.Model.
func (l *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	if l.CallbacksHandler != nil {
		l.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	opts := llms.CallOptions{
		Model: l.modelID,
	}
	for _, opt := range options {
		opt(&opts)
	}

	m, err := processMessages(messages)
	if err != nil {
		return nil, err
	}

	res, err := l.client.CreateCompletion(ctx, opts.Model, m, opts)
	if err != nil {
		if l.CallbacksHandler != nil {
			l.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}

	if l.CallbacksHandler != nil {
		l.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, res)
	}

	return res, nil
}

func processMessages(messages []llms.MessageContent) ([]bedrockclient.Message, error) {
	bedrockMsgs := make([]bedrockclient.Message, 0, len(messages))

	for _, m := range messages {
		for _, part := range m.Parts {
			switch part := part.(type) {
			case llms.TextContent:
				bedrockMsgs = append(bedrockMsgs, bedrockclient.Message{
					Role:    m.Role,
					Content: part.Text,
					Type:    "text",
				})
			case llms.BinaryContent:
				bedrockMsgs = append(bedrockMsgs, bedrockclient.Message{
					Role:     m.Role,
					Content:  string(part.Data),
					MimeType: part.MIMEType,
					Type:     "image",
				})
			case llms.ToolCall:
				// Serialize tool call as JSON
				toolCallData := map[string]interface{}{
					"id":    part.ID,
					"name":  part.FunctionCall.Name,
					"input": map[string]interface{}{},
				}
				if part.FunctionCall.Arguments != "" {
					var args map[string]interface{}
					if err := json.Unmarshal([]byte(part.FunctionCall.Arguments), &args); err == nil {
						toolCallData["input"] = args
					}
				}
				content, err := json.Marshal(toolCallData)
				if err != nil {
					return nil, err
				}
				bedrockMsgs = append(bedrockMsgs, bedrockclient.Message{
					Role:    m.Role,
					Content: string(content),
					Type:    "tool_use",
				})
			case llms.ToolCallResponse:
				// Serialize tool response as JSON
				toolResultData := map[string]interface{}{
					"tool_use_id": part.ToolCallID,
					"content":     part.Content,
				}
				content, err := json.Marshal(toolResultData)
				if err != nil {
					return nil, err
				}
				bedrockMsgs = append(bedrockMsgs, bedrockclient.Message{
					Role:    m.Role,
					Content: string(content),
					Type:    "tool_result",
				})
			default:
				return nil, errors.New("unsupported message type")
			}
		}
	}
	return bedrockMsgs, nil
}

var _ llms.Model = (*LLM)(nil)

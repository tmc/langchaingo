package bedrock

import (
	"context"
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
	modelProvider    string
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
		modelProvider:    o.modelProvider,
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

	res, err := l.client.CreateCompletion(ctx, l.modelProvider, opts.Model, m, opts)
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
			var cacheControl *llms.CacheControl
			if cached, ok := part.(llms.CachedContent); ok {
				cacheControl = cached.CacheControl
				part = cached.ContentPart
			}

			switch p := part.(type) {
			case llms.TextContent:
				bedrockMsgs = append(bedrockMsgs, bedrockclient.Message{
					Role:         m.Role,
					Content:      p.Text,
					Type:         "text",
					CacheControl: cacheControl,
				})
			case llms.BinaryContent:
				bedrockMsgs = append(bedrockMsgs, bedrockclient.Message{
					Role:         m.Role,
					Content:      string(p.Data),
					MimeType:     p.MIMEType,
					Type:         "image",
					CacheControl: cacheControl,
				})
			case llms.ToolCall:
				bedrockMsgs = append(bedrockMsgs, bedrockclient.Message{
					Role:         m.Role,
					Content:      "", // Content will be empty for tool calls
					Type:         "tool_call",
					ToolCallID:   p.ID,
					ToolName:     p.FunctionCall.Name,
					ToolArgs:     p.FunctionCall.Arguments,
					CacheControl: cacheControl,
				})
			case llms.ToolCallResponse:
				bedrockMsgs = append(bedrockMsgs, bedrockclient.Message{
					Role:         m.Role,
					Content:      p.Content,
					Type:         "tool_result",
					ToolUseID:    p.ToolCallID,
					CacheControl: cacheControl,
				})
			default:
				return nil, errors.New("unsupported message type")
			}
		}
	}
	return bedrockMsgs, nil
}

var _ llms.Model = (*LLM)(nil)

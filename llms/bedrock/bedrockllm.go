package bedrock

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/vxcontrol/langchaingo/callbacks"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/bedrock/internal/bedrockclient"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

const defaultModel = ModelAmazonTitanTextPremierV1

// LLM is a Bedrock LLM implementation.
type LLM struct {
	modelID          string
	client           *bedrockclient.Client
	converseClient   *bedrockclient.ConverseClient
	useConverseAPI   bool
	CallbacksHandler callbacks.Handler
}

// New creates a new Bedrock LLM implementation.
func New(opts ...Option) (*LLM, error) {
	return NewWithContext(context.Background(), opts...)
}

// NewWithContext creates a new Bedrock LLM implementation with context.
func NewWithContext(ctx context.Context, opts ...Option) (*LLM, error) {
	o, c, converseC, err := newClient(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return &LLM{
		client:           c,
		converseClient:   converseC,
		useConverseAPI:   o.useConverseAPI,
		modelID:          o.modelID,
		CallbacksHandler: o.callbackHandler,
	}, nil
}

func newClient(ctx context.Context, opts ...Option) (*options, *bedrockclient.Client, *bedrockclient.ConverseClient, error) {
	options := &options{
		modelID: defaultModel,
	}

	for _, opt := range opts {
		opt(options)
	}

	if options.client == nil {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, nil, nil, err
		}
		options.client = bedrockruntime.NewFromConfig(cfg)
	}

	// Create both clients - legacy and Converse API
	legacyClient := bedrockclient.NewClient(options.client)
	converseClient := bedrockclient.NewConverseClient(options.client)

	return options, legacyClient, converseClient, nil
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

	// Use Converse API if enabled
	if l.useConverseAPI {
		return l.generateContentWithConverseAPI(ctx, messages, opts)
	}

	// Use legacy implementation
	return l.generateContentWithLegacyAPI(ctx, messages, opts)
}

// generateContentWithConverseAPI uses the unified Converse API
func (l *LLM) generateContentWithConverseAPI(ctx context.Context, messages []llms.MessageContent, opts llms.CallOptions) (*llms.ContentResponse, error) {
	m, err := processMessages(messages)
	if err != nil {
		return nil, err
	}

	// Build Converse input
	input := &bedrockclient.ConverseInput{
		ModelID:         opts.Model,
		Messages:        m,
		Tools:           opts.Tools,
		StreamingFunc:   opts.StreamingFunc,
		ReasoningConfig: opts.Reasoning,
	}

	// Set inference parameters
	if opts.MaxTokens > 0 {
		input.MaxTokens = &opts.MaxTokens
	}
	if opts.Temperature > 0 {
		input.Temperature = &opts.Temperature
	}
	if opts.TopP > 0 {
		input.TopP = &opts.TopP
	}
	if len(opts.StopWords) > 0 {
		input.StopSequences = opts.StopWords
	}

	res, err := l.converseClient.CreateCompletionConverse(ctx, input)
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

// generateContentWithLegacyAPI uses the original model-specific implementations
func (l *LLM) generateContentWithLegacyAPI(ctx context.Context, messages []llms.MessageContent, opts llms.CallOptions) (*llms.ContentResponse, error) {
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
				var arguments map[string]any
				if part.FunctionCall != nil {
					if err := json.Unmarshal([]byte(part.FunctionCall.Arguments), &arguments); err != nil {
						return nil, err
					}
				}
				bedrockMsgs = append(bedrockMsgs, bedrockclient.Message{
					Role: m.Role,
					Type: "tool_use",
					ToolCall: &bedrockclient.ToolCall{
						ID:        part.ID,
						Name:      part.FunctionCall.Name,
						Arguments: arguments,
					},
				})
			case llms.ToolCallResponse:
				bedrockMsgs = append(bedrockMsgs, bedrockclient.Message{
					Role: m.Role,
					Type: "tool_result",
					ToolResult: &bedrockclient.ToolResult{
						ToolCallID: part.ToolCallID,
						ToolName:   part.Name,
						Content:    part.Content,
					},
				})
			default:
				return nil, errors.New("unsupported message type")
			}
		}
	}
	return bedrockMsgs, nil
}

var _ llms.Model = (*LLM)(nil)

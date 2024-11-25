package gigachat

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/gigachat/internal/gigachatclient"
)

var (
	ErrEmptyResponse            = errors.New("no response")
	ErrMissingCreds             = errors.New("missing the Gigachat API creds, set it in the GIGACHAT_CLIENT_ID and GIGACHAT_CLIENT_SECRET environment variables")
	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
	ErrInvalidContentType       = errors.New("invalid content type")
	ErrUnsupportedMessageType   = errors.New("unsupported message type")
	ErrUnsupportedContentType   = errors.New("unsupported content type")
)

const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"
)

const (
	ScopePersonal = "GIGACHAT_API_PERS"
	ScopeB2b      = "GIGACHAT_API_B2B"
	ScopeCorp     = "GIGACHAT_API_CORP"
)

type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *gigachatclient.Client
}

var _ llms.Model = (*LLM)(nil)

// New returns a new Gigachat LLM.
func New(opts ...Option) (*LLM, error) {
	c, err := newClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("gigachat: failed to create client: %w", err)
	}
	return &LLM{
		client: c,
	}, nil
}

func newClient(opts ...Option) (*gigachatclient.Client, error) {
	options := &options{
		clientId:     os.Getenv(clientIdEnvVarName),
		clientSecret: os.Getenv(clientSecretEnvVarName),
		scope:        ScopePersonal,
		baseURL:      gigachatclient.DefaultBaseURL,
		httpClient:   http.DefaultClient,
	}

	for _, opt := range opts {
		opt(options)
	}

	if options.clientId == "" || options.clientSecret == "" {
		return nil, ErrMissingCreds
	}

	return gigachatclient.New(
		gigachatclient.WithClientIdAndSecret(options.clientId, options.clientSecret),
		gigachatclient.WithBaseURL(options.baseURL),
		gigachatclient.WithScope(options.scope),
		gigachatclient.WithModel(options.model),
		gigachatclient.WithHTTPClient(options.httpClient),
	)
}

func (o *LLM) CreateEmbedding(ctx context.Context, texts []string) ([][]float32, error) {
	resp, e := o.client.CreateEmbedding(ctx, texts)
	if e != nil {
		return nil, e
	}

	emb := make([][]float32, 0, len(texts))
	for i := range resp.Data {
		emb = append(emb, resp.Data[i].Embedding)
	}

	return emb, nil
}

// Call requests a completion for the given prompt.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, o, prompt, options...)
}

// GenerateContent implements the Model interface.
func (o *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	opts := &llms.CallOptions{}
	for _, opt := range options {
		opt(opts)
	}

	return generateMessagesContent(ctx, o, messages, opts)
}

func generateMessagesContent(ctx context.Context, o *LLM, messages []llms.MessageContent, opts *llms.CallOptions) (*llms.ContentResponse, error) {
	chatMessages, err := processMessages(messages)
	if err != nil {
		return nil, fmt.Errorf("gigachat: failed to process messages: %w", err)
	}
	if opts.StreamingFunc != nil {
		return nil, errors.New("gigachat: streaming is not supported yet")
	}
	if len(opts.Tools) > 0 {
		return nil, errors.New("gigachat: tool use is not supported yet")
	}

	result, err := o.client.CreateCompletion(ctx, &gigachatclient.ChatPayload{
		Model:             opts.Model,
		Messages:          chatMessages,
		MaxTokens:         opts.MaxTokens,
		Temperature:       opts.Temperature,
		TopP:              opts.TopP,
		RepetitionPenalty: opts.RepetitionPenalty,
	})
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, fmt.Errorf("gigachat: failed to create message: %w", err)
	}
	if result == nil {
		return nil, ErrEmptyResponse
	}

	choices := make([]*llms.ContentChoice, len(result.Choices))
	for i, choice := range result.Choices {
		choices[i] = &llms.ContentChoice{
			Content:    choice.Message.Content,
			StopReason: choice.FinishReason,
			GenerationInfo: map[string]any{
				"InputTokens":  result.Usage.PromptTokens,
				"OutputTokens": result.Usage.CompletionTokens,
			},
		}
	}

	resp := &llms.ContentResponse{
		Choices: choices,
	}
	return resp, nil
}

func processMessages(messages []llms.MessageContent) ([]gigachatclient.PayloadMessage, error) {
	chatMessages := make([]gigachatclient.PayloadMessage, 0, len(messages))
	for _, msg := range messages {
		switch msg.Role {
		case llms.ChatMessageTypeSystem:
			chatMessage, err := handleSystemMessage(msg)
			if err != nil {
				return nil, fmt.Errorf("gigachat: failed to handle system message: %w", err)
			}
			chatMessages = append(chatMessages, chatMessage)
		case llms.ChatMessageTypeHuman:
			chatMessage, err := handleHumanMessage(msg)
			if err != nil {
				return nil, fmt.Errorf("gigachat: failed to handle human message: %w", err)
			}
			chatMessages = append(chatMessages, chatMessage)
		case llms.ChatMessageTypeAI:
			chatMessage, err := handleAIMessage(msg)
			if err != nil {
				return nil, fmt.Errorf("gigachat: failed to handle AI message: %w", err)
			}
			chatMessages = append(chatMessages, chatMessage)
		case llms.ChatMessageTypeTool:
			chatMessage, err := handleToolMessage(msg)
			if err != nil {
				return nil, fmt.Errorf("gigachat: failed to handle tool message: %w", err)
			}
			chatMessages = append(chatMessages, chatMessage)
		case llms.ChatMessageTypeGeneric, llms.ChatMessageTypeFunction:
			return nil, fmt.Errorf("gigachat: %w: %v", ErrUnsupportedMessageType, msg.Role)
		default:
			return nil, fmt.Errorf("gigachat: %w: %v", ErrUnsupportedMessageType, msg.Role)
		}
	}
	return chatMessages, nil
}

func handleSystemMessage(msg llms.MessageContent) (gigachatclient.PayloadMessage, error) {
	if textContent, ok := msg.Parts[0].(llms.TextContent); ok {
		return gigachatclient.PayloadMessage{
			Role:    RoleSystem,
			Content: textContent.Text,
		}, nil
	}
	return gigachatclient.PayloadMessage{}, fmt.Errorf("gigachat: %w for system message", ErrInvalidContentType)
}

func handleHumanMessage(msg llms.MessageContent) (gigachatclient.PayloadMessage, error) {
	if textContent, ok := msg.Parts[0].(llms.TextContent); ok {
		return gigachatclient.PayloadMessage{
			Role:    RoleUser,
			Content: textContent.Text,
		}, nil
	}
	return gigachatclient.PayloadMessage{}, fmt.Errorf("gigachat: %w for human message", ErrInvalidContentType)
}

func handleAIMessage(msg llms.MessageContent) (gigachatclient.PayloadMessage, error) {
	if textContent, ok := msg.Parts[0].(llms.TextContent); ok {
		return gigachatclient.PayloadMessage{
			Role:    RoleAssistant,
			Content: textContent.Text,
		}, nil
	}
	return gigachatclient.PayloadMessage{}, fmt.Errorf("gigachat: %w for AI message", ErrInvalidContentType)
}

func handleToolMessage(_ llms.MessageContent) (gigachatclient.PayloadMessage, error) {
	return gigachatclient.PayloadMessage{}, errors.New("gigachat: tool use not supported yet")
}

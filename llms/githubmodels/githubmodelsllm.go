package githubmodels

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/githubmodels/internal/githubmodelsclient"
)

var (
	ErrEmptyResponse = errors.New("no response")
	ErrMissingToken  = errors.New("missing token")
)

// LLM is a GitHub Models LLM implementation.
type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *githubmodelsclient.Client
	options          options
}

var _ llms.Model = (*LLM)(nil)

// New creates a new GitHub Models LLM implementation.
func New(opts ...Option) (*LLM, error) {
	// Initialize options with defaults and environment variables
	o := options{
		token:      os.Getenv(tokenEnvVarName),
		model:      defaultModel,
		httpClient: http.DefaultClient,
	}
	
	// Apply all options
	for _, opt := range opts {
		opt(&o)
	}

	// Validate token
	if o.token == "" {
		return nil, ErrMissingToken
	}

	client, err := githubmodelsclient.NewClient(o.token, o.model, o.httpClient)
	if err != nil {
		return nil, err
	}

	return &LLM{client: client, options: o, CallbacksHandler: o.callbacksHandler}, nil
}

// Call implements the call interface for LLM.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, o, prompt, options...)
}

// GenerateContent implements the Model interface.
func (o *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	// Override LLM model if set as llms.CallOption
	model := o.options.model
	if opts.Model != "" {
		model = opts.Model
	}

	// Convert messages to GitHub Models format
	chatMsgs := make([]githubmodelsclient.Message, 0, len(messages))
	for _, mc := range messages {		var contentText string
		for _, part := range mc.Parts {
			if textContent, ok := part.(llms.TextContent); ok {
				contentText = textContent.Text
				break
			}
		}

		chatMsgs = append(chatMsgs, githubmodelsclient.Message{
			Role:    typeToRole(mc.Role),
			Content: contentText,
		})
	}
	// Create chat request with options
	req := &githubmodelsclient.ChatRequest{
		Model:       model,
		Messages:    chatMsgs,
		Temperature: opts.Temperature,
		MaxTokens:   opts.MaxTokens,
		TopP:        opts.TopP,
	}

	// Call the client
	resp, err := o.client.CreateChat(ctx, req)
	if err != nil {
		return nil, err
	}

	// Process response
	if len(resp.Choices) == 0 {
		return nil, ErrEmptyResponse
	}

	choices := make([]*llms.ContentChoice, len(resp.Choices))
	for i, c := range resp.Choices {
		choices[i] = &llms.ContentChoice{
			Content:    c.Message.Content,
			StopReason: c.FinishReason,
		}
	}
	response := &llms.ContentResponse{Choices: choices}
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, response)
	}
	return response, nil
}

// CreateEmbedding creates embeddings for the given input texts.
// Note: As of this implementation, GitHub Models may not support embeddings.
// This is a placeholder implementation to satisfy the interface.
func (o *LLM) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float32, error) {
	return nil, errors.New("embedding is not supported by GitHub Models at this time")
}

// typeToRole converts llms.ChatMessageType to GitHub Models role string
func typeToRole(typ llms.ChatMessageType) string {
	switch typ {
	case llms.ChatMessageTypeSystem:
		return "system"
	case llms.ChatMessageTypeAI:
		return "assistant"
	case llms.ChatMessageTypeHuman:
		return "user"
	case llms.ChatMessageTypeGeneric:
		return "user"
	case llms.ChatMessageTypeFunction:
		return "function"
	case llms.ChatMessageTypeTool:
		return "tool"
	default:
		return "user"
	}
}

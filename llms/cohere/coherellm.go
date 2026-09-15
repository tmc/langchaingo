package cohere

import (
	"context"
	"errors"
	"os"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/cohere/internal/cohereclient"
)

var (
	ErrEmptyResponse = errors.New("no response")
	ErrMissingToken  = errors.New("missing the COHERE_API_KEY key, set it in the COHERE_API_KEY environment variable")

	ErrUnexpectedResponseLength = errors.New("unexpected length of response")

	errEmptyMessages      = errors.New("messages must not be empty")
	errEmptyMessageParts  = errors.New("message parts must not be empty")
	errNonTextMessagePart = errors.New("first message part must be text")
)

type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *cohereclient.Client
}

var _ llms.Model = (*LLM)(nil)

func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, o, prompt, options...)
}

// GenerateContent implements the Model interface.
func (o *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) { //nolint: lll, cyclop, whitespace

	if len(messages) == 0 {
		return nil, errEmptyMessages
	}
	if len(messages[0].Parts) == 0 {
		return nil, errEmptyMessageParts
	}
	text, ok := messages[0].Parts[0].(llms.TextContent)
	if !ok {
		return nil, errNonTextMessagePart
	}

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	opts := &llms.CallOptions{}
	for _, opt := range options {
		opt(opts)
	}

	result, err := o.client.CreateGeneration(ctx, &cohereclient.GenerationRequest{
		Prompt: text.Text,
	})
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}

	resp := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: result.Text,
			},
		},
	}

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, resp)
	}

	return resp, nil
}

func New(opts ...Option) (*LLM, error) {
	c, err := newClient(opts...)
	return &LLM{
		client: c,
	}, err
}

func newClient(opts ...Option) (*cohereclient.Client, error) {
	options := &options{
		token:   os.Getenv(tokenEnvVarName),
		baseURL: os.Getenv(baseURLEnvVarName),
		model:   os.Getenv(modelEnvVarName),
	}

	for _, opt := range opts {
		opt(options)
	}

	if len(options.token) == 0 {
		return nil, ErrMissingToken
	}

	return cohereclient.New(options.token, options.baseURL, options.model)
}

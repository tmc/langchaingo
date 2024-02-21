package anthropic

import (
	"context"
	"errors"
	"os"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic/internal/anthropicclient"
	"github.com/tmc/langchaingo/schema"
)

var (
	ErrEmptyResponse = errors.New("no response")
	ErrMissingToken  = errors.New("missing the Anthropic API key, set it in the ANTHROPIC_API_KEY environment variable")

	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
)

type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *anthropicclient.Client
}

var _ llms.Model = (*LLM)(nil)

// New returns a new Anthropic LLM.
func New(opts ...Option) (*LLM, error) {
	c, err := newClient(opts...)
	return &LLM{
		client: c,
	}, err
}

func newClient(opts ...Option) (*anthropicclient.Client, error) {
	options := &options{
		token: os.Getenv(tokenEnvVarName),
	}

	for _, opt := range opts {
		opt(options)
	}

	if len(options.token) == 0 {
		return nil, ErrMissingToken
	}

	return anthropicclient.New(options.token, options.model)
}

// Call requests a completion for the given prompt.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, o, prompt, options...)
}

// GenerateContent implements the Model interface.
func (o *LLM) GenerateContent(ctx context.Context, messages []schema.MessageContent, options ...llms.CallOption) (*schema.ContentResponse, error) { //nolint: lll, cyclop, whitespace

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	opts := &llms.CallOptions{}
	for _, opt := range options {
		opt(opts)
	}

	// Assume we get a single text message
	msg0 := messages[0]
	part := msg0.Parts[0]
	result, err := o.client.CreateCompletion(ctx, &anthropicclient.CompletionRequest{
		Model:         opts.Model,
		Prompt:        part.(schema.TextContent).Text,
		MaxTokens:     opts.MaxTokens,
		StopWords:     opts.StopWords,
		Temperature:   opts.Temperature,
		TopP:          opts.TopP,
		StreamingFunc: opts.StreamingFunc,
	})
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}

	resp := &schema.ContentResponse{
		Choices: []*schema.ContentChoice{
			{
				Content: result.Text,
			},
		},
	}
	return resp, nil
}

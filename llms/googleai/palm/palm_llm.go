// package palm implements a langchaingo provider for Google Vertex AI legacy
// PaLM models. Use the newer Gemini models via llms/googleai/vertex if
// possible.
package palm

import (
	"context"
	"errors"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai/internal/palmclient"
)

var (
	ErrMissingProjectID         = errors.New("missing the GCP Project ID, set it in the GOOGLE_CLOUD_PROJECT environment variable") //nolint:lll
	ErrMissingLocation          = errors.New("missing the GCP Location, set it in the GOOGLE_CLOUD_LOCATION environment variable")  //nolint:lll
	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
	ErrNotImplemented           = errors.New("not implemented")
)

type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *palmclient.PaLMClient
}

var _ llms.Model = (*LLM)(nil)

// Call requests a completion for the given prompt.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, o, prompt, options...)
}

// GenerateContent implements the Model interface.
func (o *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) { //nolint: lll, cyclop, whitespace

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	// Assume we get a single text message
	msg0 := messages[0]
	part := msg0.Parts[0]

	results, err := o.client.CreateCompletion(ctx, &palmclient.CompletionRequest{
		Prompts:       []string{part.(llms.TextContent).Text},
		MaxTokens:     opts.MaxTokens,
		Temperature:   opts.Temperature,
		StopSequences: opts.StopWords,
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
				Content: results[0].Text,
			},
		},
	}
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, resp)
	}

	return resp, nil
}

// CreateEmbedding creates embeddings for the given input texts.
func (o *LLM) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float32, error) {
	embeddings, err := o.client.CreateEmbedding(ctx, &palmclient.EmbeddingRequest{
		Input: inputTexts,
	})
	if err != nil {
		return [][]float32{}, err
	}

	if len(embeddings) == 0 {
		return nil, llms.ErrEmptyResponse
	}
	if len(inputTexts) != len(embeddings) {
		return embeddings, ErrUnexpectedResponseLength
	}

	return embeddings, nil
}

// New returns a new palmclient PaLM LLM.
func New(opts ...Option) (*LLM, error) {
	client, err := newClient(opts...)
	return &LLM{client: client}, err
}

func newClient(opts ...Option) (*palmclient.PaLMClient, error) {
	// Ensure options are initialized only once.
	initOptions.Do(initOpts)
	options := &options{}
	*options = *defaultOptions // Copy default options.

	for _, opt := range opts {
		opt(options)
	}
	if len(options.projectID) == 0 {
		return nil, ErrMissingProjectID
	}
	if len(options.location) == 0 {
		return nil, ErrMissingLocation
	}

	return palmclient.New(context.TODO(), options.projectID, options.location, options.clientOptions...)
}

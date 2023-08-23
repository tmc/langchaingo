package vertexai

import (
	"context"
	"errors"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/vertexai/internal/vertexaiclient"
	"github.com/tmc/langchaingo/schema"
)

var (
	ErrEmptyResponse            = errors.New("no response")
	ErrMissingProjectID         = errors.New("missing the GCP Project ID, set it in the GOOGLE_CLOUD_PROJECT environment variable") //nolint:lll
	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
	ErrNotImplemented           = errors.New("not implemented")
)

type LLM struct {
	client *vertexaiclient.PaLMClient
}

var (
	_ llms.LLM           = (*LLM)(nil)
	_ llms.LanguageModel = (*LLM)(nil)
)

// Call requests a completion for the given prompt.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	r, err := o.Generate(ctx, []string{prompt}, options...)
	if err != nil {
		return "", err
	}
	if len(r) == 0 {
		return "", ErrEmptyResponse
	}
	return r[0].Text, nil
}

func (o *LLM) Generate(ctx context.Context, prompts []string, options ...llms.CallOption) ([]*llms.Generation, error) {
	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	results, err := o.client.CreateCompletion(ctx, &vertexaiclient.CompletionRequest{
		Prompts:     prompts,
		MaxTokens:   opts.MaxTokens,
		Temperature: opts.Temperature,
	})
	if err != nil {
		return nil, err
	}

	generations := []*llms.Generation{}
	for _, r := range results {
		generations = append(generations, &llms.Generation{
			Text: r.Text,
		})
	}
	return generations, nil
}

// CreateEmbedding creates embeddings for the given input texts.
func (o *LLM) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float64, error) {
	embeddings, err := o.client.CreateEmbedding(ctx, &vertexaiclient.EmbeddingRequest{
		Input: inputTexts,
	})
	if err != nil {
		return [][]float64{}, err
	}

	if len(embeddings) == 0 {
		return [][]float64{}, ErrEmptyResponse
	}
	if len(inputTexts) != len(embeddings) {
		return embeddings, ErrUnexpectedResponseLength
	}

	return embeddings, nil
}

func (o *LLM) GeneratePrompt(ctx context.Context, promptValues []schema.PromptValue, options ...llms.CallOption) (llms.LLMResult, error) { //nolint:lll
	return llms.GeneratePrompt(ctx, o, promptValues, options...)
}

func (o *LLM) GetNumTokens(text string) int {
	return llms.CountTokens(vertexaiclient.TextModelName, text)
}

// New returns a new VertexAI PaLM LLM.
func New(opts ...Option) (*LLM, error) {
	client, err := newClient(opts...)
	return &LLM{client: client}, err
}

func newClient(opts ...Option) (*vertexaiclient.PaLMClient, error) {
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

	return vertexaiclient.New(options.projectID, options.clientOptions...)
}

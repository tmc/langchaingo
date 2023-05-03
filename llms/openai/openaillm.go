package openai

import (
	"context"
	"errors"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai/internal/openaiclient"
)

var (
	ErrEmptyResponse = errors.New("no response")
	ErrMissingToken  = errors.New("missing the OpenAI API key, set it in the OPENAI_API_KEY environment variable")

	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
)

type LLM struct {
	client *openaiclient.Client
}

var _ llms.LLM = (*LLM)(nil)

// Call requests a completion for the given prompt.
func (o *LLM) Call(ctx context.Context, prompt string, stopWords []string) (string, error) {
	r, err := o.Generate(ctx, []string{prompt}, stopWords)
	if err != nil {
		return "", err
	}
	if len(r) == 0 {
		return "", ErrEmptyResponse
	}
	return r[0].Text, nil
}

func (o *LLM) Generate(ctx context.Context, prompts []string, stopWords []string) ([]*llms.Generation, error) {
	result, err := o.client.CreateCompletion(ctx, &openaiclient.CompletionRequest{
		Prompt:    prompts[0],
		StopWords: stopWords,
	})
	if err != nil {
		return nil, err
	}
	return []*llms.Generation{
		{Text: result.Text},
	}, nil
}

// Chat requests a chat response for the given prompt.
func (o *LLM) Chat(prompt string) (string, error) {
	result, err := o.client.CreateChat(context.TODO(), &openaiclient.ChatRequest{Prompt: prompt})
	if err != nil {
		return "", err
	}

	return result.Text, nil
}

// CreateEmbedding creates embeddings for the given input texts.
func (o *LLM) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float64, error) {
	embeddings, err := o.client.CreateEmbedding(ctx, &openaiclient.EmbeddingRequest{
		Input: inputTexts,
	})

	if len(embeddings) == 0 {
		return [][]float64{}, ErrEmptyResponse
	}

	if err != nil {
		return [][]float64{}, err
	}

	if len(inputTexts) != len(embeddings) {
		return embeddings, ErrUnexpectedResponseLength
	}

	return embeddings, nil
}

// New returns a new OpenAI LLM.
func New(opts ...Option) (*LLM, error) {
	// Ensure options are initialized only once.
	initOptions.Do(initOpts)

	options := &options{}
	*options = *defaultOptions // Copy default options.

	for _, opt := range opts {
		opt(options)
	}

	if len(options.token) == 0 {
		return nil, ErrMissingToken
	}

	client, err := openaiclient.New(options.token, options.model)
	if err != nil {
		return nil, err
	}

	return &LLM{
		client: client,
	}, nil
}

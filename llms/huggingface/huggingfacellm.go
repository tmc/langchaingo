package huggingface

import (
	"context"
	"errors"
	"os"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/huggingface/internal/huggingfaceclient"
)

var (
	ErrEmptyResponse            = errors.New("empty response")
	ErrMissingToken             = errors.New("missing the Hugging Face API token. Set it in the HUGGINGFACEHUB_API_TOKEN environment variable") //nolint:lll
	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
)

type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *huggingfaceclient.Client
}

var _ llms.LLM = (*LLM)(nil)

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
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMStart(ctx, prompts)
	}

	opts := &llms.CallOptions{Model: defaultModel}
	for _, opt := range options {
		opt(opts)
	}
	result, err := o.client.RunInference(ctx, &huggingfaceclient.InferenceRequest{
		Model:             o.client.Model,
		Prompt:            prompts[0],
		Task:              huggingfaceclient.InferenceTaskTextGeneration,
		Temperature:       opts.Temperature,
		TopP:              opts.TopP,
		TopK:              opts.TopK,
		MinLength:         opts.MinLength,
		MaxLength:         opts.MaxLength,
		RepetitionPenalty: opts.RepetitionPenalty,
		Seed:              opts.Seed,
	})
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}

	generations := []*llms.Generation{
		{Text: result.Text},
	}

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMEnd(ctx, llms.LLMResult{Generations: [][]*llms.Generation{generations}})
	}
	return generations, nil
}

func New(opts ...Option) (*LLM, error) {
	options := &options{
		token: os.Getenv(tokenEnvVarName),
		model: defaultModel,
	}

	for _, opt := range opts {
		opt(options)
	}

	if len(options.token) == 0 {
		return nil, ErrMissingToken
	}

	c, err := huggingfaceclient.New(options.token, options.model)
	if err != nil {
		return nil, err
	}

	return &LLM{
		client: c,
	}, nil
}

// CreateEmbedding creates embeddings for the given input texts.
func (o *LLM) CreateEmbedding(
	ctx context.Context,
	inputTexts []string,
	model string,
	task string,
) ([][]float32, error) {
	embeddings, err := o.client.CreateEmbedding(ctx, model, task, &huggingfaceclient.EmbeddingRequest{
		Inputs: inputTexts,
		Options: map[string]any{
			"use_gpu":        false,
			"wait_for_model": true,
		},
	})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, ErrEmptyResponse
	}
	if len(inputTexts) != len(embeddings) {
		return embeddings, ErrUnexpectedResponseLength
	}
	return embeddings, nil
}

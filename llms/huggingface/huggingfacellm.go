package huggingface

import (
	"context"
	"errors"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/huggingface/internal/huggingfaceclient"
)

var (
	ErrEmptyResponse = errors.New("empty response")
	ErrMissingToken  = errors.New("missing the Hugging Face API token. Set it in the HUGGINGFACEHUB_API_TOKEN environment variable") //nolint:lll
)

type LLM struct {
	client *huggingfaceclient.Client
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
		return nil, err
	}
	return []*llms.Generation{
		{Text: result.Text},
	}, nil
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

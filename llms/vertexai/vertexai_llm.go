package vertexai

import (
	"context"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/vertexai/internal/schema"
	lcgschema "github.com/tmc/langchaingo/schema"
)

type LLM struct {
	*baseLLM
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

	o.setDefaultCallOptions(&opts)

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMStart(ctx, prompts)
	}

	results, err := o.client.CreateCompletion(ctx, &schema.CompletionRequest{
		Prompts:       prompts,
		MaxTokens:     opts.MaxTokens,
		Temperature:   opts.Temperature,
		StopSequences: opts.StopWords,
		TopK:          opts.TopK,
		TopP:          int(opts.TopP),
		Model:         opts.Model,
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

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMEnd(ctx, llms.LLMResult{Generations: [][]*llms.Generation{generations}})
	}
	return generations, nil
}

func (o *LLM) GeneratePrompt(ctx context.Context, promptValues []lcgschema.PromptValue, options ...llms.CallOption) (llms.LLMResult, error) { //nolint:lll
	return llms.GeneratePrompt(ctx, o, promptValues, options...)
}

// New returns a new VertexAI LLM.
func New(opts ...Option) (*LLM, error) {
	// The context should be provided by the caller but that would be a big change, so we just do this
	ctx := context.Background()

	// Ensure options are initialized only once.
	initOptions.Do(initOpts)
	options := &options{}
	*options = *defaultOptions // Copy default options.

	for _, opt := range opts {
		opt(options)
	}

	// The LLM struct uses the prediction model provided in the options, so we configure the base with that
	base, err := newBase(ctx, options.model, *options)

	return &LLM{baseLLM: base}, err
}

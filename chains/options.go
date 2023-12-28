package chains

import (
	"context"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
)

// ChainCallOption is a function that can be used to modify the behavior of the Call function.
type ChainCallOption func(*chainCallOption)

type chainCallOption struct {
	// Model is the model to use in an LLM call.
	Model string
	// MaxTokens is the maximum number of tokens to generate to use in an LLM call.
	MaxTokens int
	// Temperature is the temperature for sampling to use in an LLM call, between 0 and 1.
	Temperature float64
	// StopWords is a list of words to stop on to use in an LLM call.
	StopWords []string
	// StreamingFunc is a function to be called for each chunk of a streaming response.
	// Return an error to stop streaming earl.
	StreamingFunc func(ctx context.Context, chunk []byte) error
	// TopK is the number of tokens to consider for top-k sampling in an LLM call.
	TopK int
	// TopP is the cumulative probability for top-p sampling in an LLM call.
	TopP float64
	// Seed is a seed for deterministic sampling in an LLM call.
	Seed int
	// MinLength is the minimum length of the generated text in an LLM call.
	MinLength int
	// MaxLength is the maximum length of the generated text in an LLM call.
	MaxLength int
	// RepetitionPenalty is the repetition penalty for sampling in an LLM call.
	RepetitionPenalty float64
	// CallbackHandler is the callback handler for Chain
	CallbackHandler callbacks.Handler
}

// WithModel is an option for LLM.Call.
func WithModel(model string) ChainCallOption {
	return func(o *chainCallOption) {
		o.Model = model
	}
}

// WithMaxTokens is an option for LLM.Call.
func WithMaxTokens(maxTokens int) ChainCallOption {
	return func(o *chainCallOption) {
		o.MaxTokens = maxTokens
	}
}

// WithTemperature is an option for LLM.Call.
func WithTemperature(temperature float64) ChainCallOption {
	return func(o *chainCallOption) {
		o.Temperature = temperature
	}
}

// WithOptions is an option for LLM.Call.
func WithOptions(options chainCallOption) ChainCallOption {
	return func(o *chainCallOption) {
		*o = options
	}
}

// WithStreamingFunc is an option for LLM.Call that allows streaming responses.
func WithStreamingFunc(streamingFunc func(ctx context.Context, chunk []byte) error) ChainCallOption {
	return func(o *chainCallOption) {
		o.StreamingFunc = streamingFunc
	}
}

// WithTopK will add an option to use top-k sampling for LLM.Call.
func WithTopK(topK int) ChainCallOption {
	return func(o *chainCallOption) {
		o.TopK = topK
	}
}

// WithTopP	will add an option to use top-p sampling for LLM.Call.
func WithTopP(topP float64) ChainCallOption {
	return func(o *chainCallOption) {
		o.TopP = topP
	}
}

// WithSeed will add an option to use deterministic sampling for LLM.Call.
func WithSeed(seed int) ChainCallOption {
	return func(o *chainCallOption) {
		o.Seed = seed
	}
}

// WithMinLength will add an option to set the minimum length of the generated text for LLM.Call.
func WithMinLength(minLength int) ChainCallOption {
	return func(o *chainCallOption) {
		o.MinLength = minLength
	}
}

// WithMaxLength will add an option to set the maximum length of the generated text for LLM.Call.
func WithMaxLength(maxLength int) ChainCallOption {
	return func(o *chainCallOption) {
		o.MaxLength = maxLength
	}
}

// WithRepetitionPenalty will add an option to set the repetition penalty for sampling.
func WithRepetitionPenalty(repetitionPenalty float64) ChainCallOption {
	return func(o *chainCallOption) {
		o.RepetitionPenalty = repetitionPenalty
	}
}

// WithStopWords is an option for setting the stop words for LLM.Call.
func WithStopWords(stopWords []string) ChainCallOption {
	return func(o *chainCallOption) {
		o.StopWords = stopWords
	}
}

// WithCallback allows setting a custom Callback Handler.
func WithCallback(callbackHandler callbacks.Handler) ChainCallOption {
	return func(opts *chainCallOption) {
		opts.CallbackHandler = callbackHandler
	}
}

func getLLMCallOptions(options ...ChainCallOption) []llms.CallOption {
	opts := &chainCallOption{}
	for _, option := range options {
		option(opts)
	}
	if opts.StreamingFunc == nil && opts.CallbackHandler != nil {
		opts.StreamingFunc = func(ctx context.Context, chunk []byte) error {
			opts.CallbackHandler.HandleStreamingFunc(ctx, chunk)
			return nil
		}
	}

	chainCallOption := []llms.CallOption{
		llms.WithModel(opts.Model),
		llms.WithMaxTokens(opts.MaxTokens),
		llms.WithTemperature(opts.Temperature),
		llms.WithStopWords(opts.StopWords),
		llms.WithStreamingFunc(opts.StreamingFunc),
		llms.WithTopK(opts.TopK),
		llms.WithTopP(opts.TopP),
		llms.WithSeed(opts.Seed),
		llms.WithMinLength(opts.MinLength),
		llms.WithMaxLength(opts.MaxLength),
		llms.WithRepetitionPenalty(opts.RepetitionPenalty),
	}

	return chainCallOption
}

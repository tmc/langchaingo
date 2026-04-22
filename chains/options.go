package chains

import (
	"context"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
)

// ChainCallOption is a function that can be used to modify the behavior of the Call function.
type ChainCallOption func(*ChainCallOptions)

// ChainCallOptions holds the resolved values for a chain call. Custom Chain
// implementations outside this package can call [ApplyChainCallOptions] to
// convert a slice of [ChainCallOption] into a ChainCallOptions value they can
// read.
//
// For issue #626, each field here has a boolean "set" flag so we can
// distinguish between the case where the option was actually set explicitly,
// or asked to remain default. The reason we need this is that in translating
// options from ChainCallOption to llms.CallOption, the notion of "default
// value the user didn't explicitly ask to change" is violated.
// These flags are hopefully a temporary backwards-compatible solution, until
// we find a more fundamental solution for #626.
type ChainCallOptions struct {
	// Model is the model to use in an LLM call.
	Model    string
	ModelSet bool

	// MaxTokens is the maximum number of tokens to generate to use in an LLM call.
	MaxTokens    int
	MaxTokensSet bool

	// Temperature is the temperature for sampling to use in an LLM call, between 0 and 1.
	Temperature    float64
	TemperatureSet bool

	// StopWords is a list of words to stop on to use in an LLM call.
	StopWords    []string
	StopWordsSet bool

	// StreamingFunc is a function to be called for each chunk of a streaming response.
	// Return an error to stop streaming early.
	StreamingFunc func(ctx context.Context, chunk []byte) error

	// TopK is the number of tokens to consider for top-k sampling in an LLM call.
	TopK    int
	TopKSet bool

	// TopP is the cumulative probability for top-p sampling in an LLM call.
	TopP    float64
	TopPSet bool

	// Seed is a seed for deterministic sampling in an LLM call.
	Seed    int
	SeedSet bool

	// MinLength is the minimum length of the generated text in an LLM call.
	MinLength    int
	MinLengthSet bool

	// MaxLength is the maximum length of the generated text in an LLM call.
	MaxLength    int
	MaxLengthSet bool

	// RepetitionPenalty is the repetition penalty for sampling in an LLM call.
	RepetitionPenalty    float64
	RepetitionPenaltySet bool

	// CallbackHandler is the callback handler for Chain
	CallbackHandler callbacks.Handler
}

// ApplyChainCallOptions resolves a slice of [ChainCallOption] into a
// [ChainCallOptions] value. Custom [Chain] implementations that accept
// ChainCallOption arguments use this to read the caller's choices.
func ApplyChainCallOptions(options ...ChainCallOption) ChainCallOptions {
	var opts ChainCallOptions
	for _, option := range options {
		option(&opts)
	}
	return opts
}

// WithModel is an option for LLM.Call.
func WithModel(model string) ChainCallOption {
	return func(o *ChainCallOptions) {
		o.Model = model
		o.ModelSet = true
	}
}

// WithMaxTokens is an option for LLM.Call.
func WithMaxTokens(maxTokens int) ChainCallOption {
	return func(o *ChainCallOptions) {
		o.MaxTokens = maxTokens
		o.MaxTokensSet = true
	}
}

// WithTemperature is an option for LLM.Call.
func WithTemperature(temperature float64) ChainCallOption {
	return func(o *ChainCallOptions) {
		o.Temperature = temperature
		o.TemperatureSet = true
	}
}

// WithStreamingFunc is an option for LLM.Call that allows streaming responses.
func WithStreamingFunc(streamingFunc func(ctx context.Context, chunk []byte) error) ChainCallOption {
	return func(o *ChainCallOptions) {
		o.StreamingFunc = streamingFunc
	}
}

// WithTopK will add an option to use top-k sampling for LLM.Call.
func WithTopK(topK int) ChainCallOption {
	return func(o *ChainCallOptions) {
		o.TopK = topK
		o.TopKSet = true
	}
}

// WithTopP	will add an option to use top-p sampling for LLM.Call.
func WithTopP(topP float64) ChainCallOption {
	return func(o *ChainCallOptions) {
		o.TopP = topP
		o.TopPSet = true
	}
}

// WithSeed will add an option to use deterministic sampling for LLM.Call.
func WithSeed(seed int) ChainCallOption {
	return func(o *ChainCallOptions) {
		o.Seed = seed
		o.SeedSet = true
	}
}

// WithMinLength will add an option to set the minimum length of the generated text for LLM.Call.
func WithMinLength(minLength int) ChainCallOption {
	return func(o *ChainCallOptions) {
		o.MinLength = minLength
		o.MinLengthSet = true
	}
}

// WithMaxLength will add an option to set the maximum length of the generated text for LLM.Call.
func WithMaxLength(maxLength int) ChainCallOption {
	return func(o *ChainCallOptions) {
		o.MaxLength = maxLength
		o.MaxLengthSet = true
	}
}

// WithRepetitionPenalty will add an option to set the repetition penalty for sampling.
func WithRepetitionPenalty(repetitionPenalty float64) ChainCallOption {
	return func(o *ChainCallOptions) {
		o.RepetitionPenalty = repetitionPenalty
		o.RepetitionPenaltySet = true
	}
}

// WithStopWords is an option for setting the stop words for LLM.Call.
func WithStopWords(stopWords []string) ChainCallOption {
	return func(o *ChainCallOptions) {
		o.StopWords = stopWords
		o.StopWordsSet = true
	}
}

// WithCallback allows setting a custom Callback Handler.
func WithCallback(callbackHandler callbacks.Handler) ChainCallOption {
	return func(o *ChainCallOptions) {
		o.CallbackHandler = callbackHandler
	}
}

// GetLLMCallOptions converts ChainCallOption slice to llms.CallOption slice.
// This is useful for agents and other code that needs to propagate chain options
// to direct LLM calls.
func GetLLMCallOptions(options ...ChainCallOption) []llms.CallOption { //nolint:cyclop
	opts := ApplyChainCallOptions(options...)
	if opts.StreamingFunc == nil && opts.CallbackHandler != nil {
		opts.StreamingFunc = func(ctx context.Context, chunk []byte) error {
			opts.CallbackHandler.HandleStreamingFunc(ctx, chunk)
			return nil
		}
	}

	var chainCallOption []llms.CallOption

	if opts.ModelSet {
		chainCallOption = append(chainCallOption, llms.WithModel(opts.Model))
	}
	if opts.MaxTokensSet {
		chainCallOption = append(chainCallOption, llms.WithMaxTokens(opts.MaxTokens))
	}
	if opts.TemperatureSet {
		chainCallOption = append(chainCallOption, llms.WithTemperature(opts.Temperature))
	}
	if opts.StopWordsSet {
		chainCallOption = append(chainCallOption, llms.WithStopWords(opts.StopWords))
	}
	if opts.TopKSet {
		chainCallOption = append(chainCallOption, llms.WithTopK(opts.TopK))
	}
	if opts.TopPSet {
		chainCallOption = append(chainCallOption, llms.WithTopP(opts.TopP))
	}
	if opts.SeedSet {
		chainCallOption = append(chainCallOption, llms.WithSeed(opts.Seed))
	}
	if opts.MinLengthSet {
		chainCallOption = append(chainCallOption, llms.WithMinLength(opts.MinLength))
	}
	if opts.MaxLengthSet {
		chainCallOption = append(chainCallOption, llms.WithMaxLength(opts.MaxLength))
	}
	if opts.RepetitionPenaltySet {
		chainCallOption = append(chainCallOption, llms.WithRepetitionPenalty(opts.RepetitionPenalty))
	}
	chainCallOption = append(chainCallOption, llms.WithStreamingFunc(opts.StreamingFunc))

	return chainCallOption
}

// getLLMCallOptions is a backward-compatibility wrapper for GetLLMCallOptions.
func getLLMCallOptions(options ...ChainCallOption) []llms.CallOption {
	return GetLLMCallOptions(options...)
}

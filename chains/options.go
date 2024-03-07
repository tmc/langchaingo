package chains

import (
	"context"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
)

// ChainCallOption is a function that can be used to modify the behavior of the Call function.
type ChainCallOption func(*chainCallOption)

// For issue #626, each field here has a boolean "set" flag so we can
// distinguish between the case where the option was actually set explicitly
// on chainCallOption, or asked to remain default. The reason we need this is
// that in translating options from ChainCallOption to llms.CallOption, the
// notion of "default value the user didn't explicitly ask to change" is
// violated.
// These flags are hopefully a temporary backwards-compatible solution, until
// we find a more fundamental solution for #626.
type chainCallOption struct {
	// Model is the model to use in an LLM call.
	Model    string
	modelSet bool

	// MaxTokens is the maximum number of tokens to generate to use in an LLM call.
	MaxTokens    int
	maxTokensSet bool

	// Temperature is the temperature for sampling to use in an LLM call, between 0 and 1.
	Temperature    float64
	temperatureSet bool

	// StopWords is a list of words to stop on to use in an LLM call.
	StopWords    []string
	stopWordsSet bool

	// StreamingFunc is a function to be called for each chunk of a streaming response.
	// Return an error to stop streaming early.
	StreamingFunc func(ctx context.Context, chunk []byte) error

	// TopK is the number of tokens to consider for top-k sampling in an LLM call.
	TopK    int
	topkSet bool

	// TopP is the cumulative probability for top-p sampling in an LLM call.
	TopP    float64
	toppSet bool

	// Seed is a seed for deterministic sampling in an LLM call.
	Seed    int
	seedSet bool

	// MinLength is the minimum length of the generated text in an LLM call.
	MinLength    int
	minLengthSet bool

	// MaxLength is the maximum length of the generated text in an LLM call.
	MaxLength    int
	maxLengthSet bool

	// RepetitionPenalty is the repetition penalty for sampling in an LLM call.
	RepetitionPenalty    float64
	repetitionPenaltySet bool

	// CallbackHandler is the callback handler for Chain
	CallbackHandler callbacks.Handler
}

// WithModel is an option for LLM.Call.
func WithModel(model string) ChainCallOption {
	return func(o *chainCallOption) {
		o.Model = model
		o.modelSet = true
	}
}

// WithMaxTokens is an option for LLM.Call.
func WithMaxTokens(maxTokens int) ChainCallOption {
	return func(o *chainCallOption) {
		o.MaxTokens = maxTokens
		o.maxTokensSet = true
	}
}

// WithTemperature is an option for LLM.Call.
func WithTemperature(temperature float64) ChainCallOption {
	return func(o *chainCallOption) {
		o.Temperature = temperature
		o.temperatureSet = true
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
		o.topkSet = true
	}
}

// WithTopP	will add an option to use top-p sampling for LLM.Call.
func WithTopP(topP float64) ChainCallOption {
	return func(o *chainCallOption) {
		o.TopP = topP
		o.toppSet = true
	}
}

// WithSeed will add an option to use deterministic sampling for LLM.Call.
func WithSeed(seed int) ChainCallOption {
	return func(o *chainCallOption) {
		o.Seed = seed
		o.seedSet = true
	}
}

// WithMinLength will add an option to set the minimum length of the generated text for LLM.Call.
func WithMinLength(minLength int) ChainCallOption {
	return func(o *chainCallOption) {
		o.MinLength = minLength
		o.minLengthSet = true
	}
}

// WithMaxLength will add an option to set the maximum length of the generated text for LLM.Call.
func WithMaxLength(maxLength int) ChainCallOption {
	return func(o *chainCallOption) {
		o.MaxLength = maxLength
		o.maxLengthSet = true
	}
}

// WithRepetitionPenalty will add an option to set the repetition penalty for sampling.
func WithRepetitionPenalty(repetitionPenalty float64) ChainCallOption {
	return func(o *chainCallOption) {
		o.RepetitionPenalty = repetitionPenalty
		o.repetitionPenaltySet = true
	}
}

// WithStopWords is an option for setting the stop words for LLM.Call.
func WithStopWords(stopWords []string) ChainCallOption {
	return func(o *chainCallOption) {
		o.StopWords = stopWords
		o.stopWordsSet = true
	}
}

// WithCallback allows setting a custom Callback Handler.
func WithCallback(callbackHandler callbacks.Handler) ChainCallOption {
	return func(o *chainCallOption) {
		o.CallbackHandler = callbackHandler
	}
}

func getLLMCallOptions(options ...ChainCallOption) []llms.CallOption { //nolint:cyclop
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

	var chainCallOption []llms.CallOption

	if opts.modelSet {
		chainCallOption = append(chainCallOption, llms.WithModel(opts.Model))
	}
	if opts.maxTokensSet {
		chainCallOption = append(chainCallOption, llms.WithMaxTokens(opts.MaxTokens))
	}
	if opts.temperatureSet {
		chainCallOption = append(chainCallOption, llms.WithTemperature(opts.Temperature))
	}
	if opts.stopWordsSet {
		chainCallOption = append(chainCallOption, llms.WithStopWords(opts.StopWords))
	}
	if opts.topkSet {
		chainCallOption = append(chainCallOption, llms.WithTopK(opts.TopK))
	}
	if opts.toppSet {
		chainCallOption = append(chainCallOption, llms.WithTopP(opts.TopP))
	}
	if opts.seedSet {
		chainCallOption = append(chainCallOption, llms.WithSeed(opts.Seed))
	}
	if opts.minLengthSet {
		chainCallOption = append(chainCallOption, llms.WithMinLength(opts.MinLength))
	}
	if opts.maxLengthSet {
		chainCallOption = append(chainCallOption, llms.WithMaxLength(opts.MaxLength))
	}
	if opts.repetitionPenaltySet {
		chainCallOption = append(chainCallOption, llms.WithRepetitionPenalty(opts.RepetitionPenalty))
	}
	chainCallOption = append(chainCallOption, llms.WithStreamingFunc(opts.StreamingFunc))

	return chainCallOption
}

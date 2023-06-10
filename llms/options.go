package llms

import "context"

// CallOption is a function that configures a CallOptions.
type CallOption func(*CallOptions)

// CallOptions is a set of options for LLM.Call.
type CallOptions struct {
	// Model is the model to use.
	Model string `json:"model"`
	// MaxTokens is the maximum number of tokens to generate.
	MaxTokens int `json:"max_tokens"`
	// Temperature is the temperature for sampling, between 0 and 1.
	Temperature float64 `json:"temperature"`
	// StopWords is a list of words to stop on.
	StopWords []string `json:"stop_words"`
	// StreamingFunc is a function to be called for each chunk of a streaming response.
	// Return an error to stop streaming early.
	StreamingFunc func(ctx context.Context, chunk []byte) error
	// TopK is the number of tokens to consider for top-k sampling.
	TopK int `json:"top_k"`
	// TopP is the cumulative probability for top-p sampling.
	TopP float64 `json:"top_p"`
	// Seed is a seed for deterministic sampling.
	Seed int `json:"seed"`
	// MinLength is the minimum length of the generated text.
	MinLength int `json:"min_length"`
	// MaxLength is the maximum length of the generated text.
	MaxLength int `json:"max_length"`
	// RepetitionPenalty is the repetition penalty for sampling.
	RepetitionPenalty float64 `json:"repetition_penalty"`
}

// WithModel is an option for LLM.Call.
func WithModel(model string) CallOption {
	return func(o *CallOptions) {
		o.Model = model
	}
}

// WithMaxTokens is an option for LLM.Call.
func WithMaxTokens(maxTokens int) CallOption {
	return func(o *CallOptions) {
		o.MaxTokens = maxTokens
	}
}

// WithTemperature is an option for LLM.Call.
func WithTemperature(temperature float64) CallOption {
	return func(o *CallOptions) {
		o.Temperature = temperature
	}
}

// WithStopWords is an option for LLM.Call.
func WithStopWords(stopWords []string) CallOption {
	return func(o *CallOptions) {
		o.StopWords = stopWords
	}
}

// WithOptions is an option for LLM.Call.
func WithOptions(options CallOptions) CallOption {
	return func(o *CallOptions) {
		(*o) = options
	}
}

// WithStreamingFunc is an option for LLM.Call that allows streaming responses.
func WithStreamingFunc(streamingFunc func(ctx context.Context, chunk []byte) error) CallOption {
	return func(o *CallOptions) {
		o.StreamingFunc = streamingFunc
	}
}

// WithTopK will add an option to use top-k sampling.
func WithTopK(topK int) CallOption {
	return func(o *CallOptions) {
		o.TopK = topK
	}
}

// WithTopP	will add an option to use top-p sampling.
func WithTopP(topP float64) CallOption {
	return func(o *CallOptions) {
		o.TopP = topP
	}
}

// WithSeed will add an option to use deterministic sampling.
func WithSeed(seed int) CallOption {
	return func(o *CallOptions) {
		o.Seed = seed
	}
}

// WithMinLength will add an option to set the minimum length of the generated text.
func WithMinLength(minLength int) CallOption {
	return func(o *CallOptions) {
		o.MinLength = minLength
	}
}

// WithMaxLength will add an option to set the maximum length of the generated text.
func WithMaxLength(maxLength int) CallOption {
	return func(o *CallOptions) {
		o.MaxLength = maxLength
	}
}

// WithRepetitionPenalty will add an option to set the repetition penalty for sampling.
func WithRepetitionPenalty(repetitionPenalty float64) CallOption {
	return func(o *CallOptions) {
		o.RepetitionPenalty = repetitionPenalty
	}
}

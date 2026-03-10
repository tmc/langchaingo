package llms

import (
	"github.com/vxcontrol/langchaingo/llms/reasoning"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

// CallOption is a function that configures a CallOptions.
type CallOption func(*CallOptions)

const MaxReasoningTokens = 64000

const (
	DefaultN                 = 1
	DefaultCandidateCount    = 1
	DefaultMaxTokens         = 16384
	DefaultTemperature       = 0.0
	DefaultTopK              = 0
	DefaultTopP              = 0
	DefaultMinP              = 0
	DefaultSeed              = 0
	DefaultMinLength         = 0
	DefaultMaxLength         = 0
	DefaultRepetitionPenalty = 0.0
	DefaultFrequencyPenalty  = 0.0
	DefaultPresencePenalty   = 0.0
	DefaultSpeed             = 0.0
)

type ReasoningEffort string

const (
	ReasoningHigh   ReasoningEffort = "high"
	ReasoningMedium ReasoningEffort = "medium"
	ReasoningLow    ReasoningEffort = "low"
	ReasoningNone   ReasoningEffort = ""
)

// ReasoningConfig is a set of options for reasoning.
type ReasoningConfig struct {
	Effort ReasoningEffort `json:"effort"`
	Tokens int             `json:"tokens"`
}

// IsEnabled returns true if reasoning is enabled based on the effort and tokens.
func (r *ReasoningConfig) IsEnabled() bool {
	if r == nil {
		return false
	}

	if r.Effort == ReasoningNone && r.Tokens == 0 {
		return false
	}

	return true
}

// GetEffort returns enum value of the effort based on kept values inside.
// If maxTokens is less than 0, it will be set to 8192.
// If neither are set, it will return ReasoningNone.
// If effort is set, it will return the set effort.
// If tokens are set, it will return the effort that is the closest to the set tokens.
//   - (0, maxTokens/4) -> ReasoningLow
//   - [maxTokens/4, maxTokens/3) -> ReasoningMedium
//   - [maxTokens/3, inf) -> ReasoningHigh
func (r *ReasoningConfig) GetEffort(maxTokens int) ReasoningEffort {
	if r == nil {
		return ReasoningNone
	}

	if r.Effort != ReasoningNone {
		return r.Effort
	}

	if maxTokens <= 0 {
		maxTokens = 8192
	}

	if r.Tokens > 0 {
		switch {
		case r.Tokens < maxTokens/4:
			return ReasoningLow
		case r.Tokens < maxTokens/3:
			return ReasoningMedium
		default:
			return ReasoningHigh
		}
	}

	return ReasoningNone
}

// GetTokens returns the number of tokens to use for reasoning based on kept values inside.
// Maximum value is maxTokens*2/3 because we need to leave some tokens for the response.
// If maxTokens is less than 0, it will be set to 8192.
// If tokens are set, it will return the minimum of the set tokens and maxTokens*2/3.
// If effort is set, it will return the maximum of the effort and maxTokens*2/3.
// If neither are set, it will return 0 or -1 if effort is set to an invalid value.
// Minimum correct values are:
//   - 1024 for ReasoningLow
//   - 2048 for ReasoningMedium
//   - 4096 for ReasoningHigh
func (r *ReasoningConfig) GetTokens(maxTokens int) int {
	if r == nil {
		return 0
	}

	if maxTokens <= 0 {
		maxTokens = 8192
	}

	var tokens int
	if r.Tokens > 0 {
		tokens = r.Tokens
	} else {
		switch r.Effort {
		case ReasoningLow:
			tokens = max(maxTokens/4, 1024)
		case ReasoningMedium:
			tokens = max(maxTokens/3, 2048)
		case ReasoningHigh:
			tokens = max(maxTokens/2, 4096)
		case ReasoningNone:
			return 0 // disabled
		default:
			return -1 // error value to be handled on the server side
		}
	}

	return min(min(tokens, maxTokens*2/3), MaxReasoningTokens)
}

// CallOptions is a set of options for calling models. Not all models support
// all options.
type CallOptions struct {
	// Model is the model to use.
	Model *string `json:"model,omitempty"`
	// CandidateCount is the number of response candidates to generate.
	CandidateCount *int `json:"candidate_count,omitempty"`
	// MaxTokens is the maximum number of tokens to generate.
	MaxTokens *int `json:"max_tokens,omitempty"`
	// Temperature is the temperature for sampling, between 0 and 1.
	Temperature *float64 `json:"temperature,omitempty"`
	// StopWords is a list of words to stop on.
	StopWords []string `json:"stop_words,omitempty"`
	// StreamingFunc is a function to be called for each chunk of a streaming response.
	// Return an error to stop streaming early.
	StreamingFunc streaming.Callback `json:"-"`
	// TopK is the number of tokens to consider for top-k sampling.
	TopK *int `json:"top_k,omitempty"`
	// TopP is the cumulative probability for top-p sampling.
	TopP *float64 `json:"top_p,omitempty"`
	// MinP is the minimum probability for top-p sampling.
	MinP *float64 `json:"min_p,omitempty"`
	// Seed is a seed for deterministic sampling.
	Seed *int `json:"seed,omitempty"`
	// MinLength is the minimum length of the generated text.
	MinLength *int `json:"min_length,omitempty"`
	// MaxLength is the maximum length of the generated text.
	MaxLength *int `json:"max_length,omitempty"`
	// N is how many chat completion choices to generate for each input message.
	N *int `json:"n,omitempty"`
	// RepetitionPenalty is the repetition penalty for sampling.
	RepetitionPenalty *float64 `json:"repetition_penalty,omitempty"`
	// FrequencyPenalty is the frequency penalty for sampling.
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`
	// PresencePenalty is the presence penalty for sampling.
	PresencePenalty *float64 `json:"presence_penalty,omitempty"`

	// Reasoning is the configuration for thinking of the model.
	Reasoning *ReasoningConfig `json:"reasoning,omitempty"`

	// JSONMode is a flag to enable JSON mode.
	JSONMode bool `json:"json"`

	// Tools is a list of tools to use. Each tool can be a specific tool or a function.
	Tools []Tool `json:"tools,omitempty"`
	// ToolChoice is the choice of tool to use, it can either be "none", "auto" (the default behavior), or a specific tool as described in the ToolChoice type.
	ToolChoice any `json:"tool_choice,omitempty"`

	// Function defitions to include in the request.
	// Deprecated: Use Tools instead.
	Functions []FunctionDefinition `json:"functions,omitempty"`
	// FunctionCallBehavior is the behavior to use when calling functions.
	//
	// If a specific function should be invoked, use the format:
	// `{"name": "my_function"}`
	// Deprecated: Use ToolChoice instead.
	FunctionCallBehavior FunctionCallBehavior `json:"function_call,omitempty"`

	// Metadata is a map of metadata to include in the request.
	// The meaning of this field is specific to the backend in use.
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// ResponseMIMEType MIME type of the generated candidate text.
	// Supported MIME types are: text/plain: (default) Text output.
	// application/json: JSON response in the response candidates.
	ResponseMIMEType *string `json:"response_mime_type,omitempty"`

	// TTS options.
	Voice          *string  `json:"voice,omitempty"`
	Speed          *float64 `json:"speed,omitempty"`
	ResponseFormat *string  `json:"response_format,omitempty"`

	// WebSearchOptions configures web search behavior for models that support it.
	// Currently supported by OpenAI models like gpt-4o-search-preview.
	WebSearchOptions *WebSearchOptions `json:"web_search_options,omitempty"`
}

// GetModel returns the model to use.
func (o *CallOptions) GetModel() string {
	if o.Model == nil {
		return ""
	}
	return *o.Model
}

// GetCandidateCount returns the number of response candidates to generate.
func (o *CallOptions) GetCandidateCount() int {
	if o.CandidateCount == nil {
		return DefaultCandidateCount
	}
	return *o.CandidateCount
}

// GetMaxTokens returns the max number of tokens to generate.
func (o *CallOptions) GetMaxTokens() int {
	if o.MaxTokens == nil {
		return DefaultMaxTokens
	}
	return *o.MaxTokens
}

// GetTemperature returns the model temperature.
func (o *CallOptions) GetTemperature() float64 {
	if o.Temperature == nil {
		return DefaultTemperature
	}
	if reasoning.IsReasoningModel(o.GetModel()) {
		return 1.0
	}
	return *o.Temperature
}

// GetStopWords returns the list of words to stop generation on.
func (o *CallOptions) GetStopWords() []string {
	return o.StopWords
}

// GetTopK returns the number of tokens to consider for top-k sampling.
func (o *CallOptions) GetTopK() int {
	if o.TopK == nil {
		return DefaultTopK
	}
	return *o.TopK
}

// GetTopP returns the cumulative probability for top-p sampling.
func (o *CallOptions) GetTopP() float64 {
	if o.TopP == nil {
		return DefaultTopP
	}
	return *o.TopP
}

// GetMinP returns the minimum probability for top-p sampling.
func (o *CallOptions) GetMinP() float64 {
	if o.MinP == nil {
		return DefaultMinP
	}
	return *o.MinP
}

// GetSeed returns the seed for deterministic sampling.
func (o *CallOptions) GetSeed() int {
	if o.Seed == nil {
		return DefaultSeed
	}
	return *o.Seed
}

// GetMinLength returns the minimum length of the generated text.
func (o *CallOptions) GetMinLength() int {
	if o.MinLength == nil {
		return DefaultMinLength
	}
	return *o.MinLength
}

// GetMaxLength returns the maximum length of the generated text.
func (o *CallOptions) GetMaxLength() int {
	if o.MaxLength == nil {
		return DefaultMaxLength
	}
	return *o.MaxLength
}

// GetN returns how many chat completion choices to generate for each input message.
func (o *CallOptions) GetN() int {
	if o.N == nil {
		return DefaultN
	}
	return *o.N
}

// GetRepetitionPenalty returns the repetition penalty for sampling.
func (o *CallOptions) GetRepetitionPenalty() float64 {
	if o.RepetitionPenalty == nil {
		return DefaultRepetitionPenalty
	}
	return *o.RepetitionPenalty
}

// GetFrequencyPenalty returns the frequency penalty for sampling.
func (o *CallOptions) GetFrequencyPenalty() float64 {
	if o.FrequencyPenalty == nil {
		return DefaultFrequencyPenalty
	}
	return *o.FrequencyPenalty
}

// GetPresencePenalty returns the presence penalty for sampling.
func (o *CallOptions) GetPresencePenalty() float64 {
	if o.PresencePenalty == nil {
		return DefaultPresencePenalty
	}
	return *o.PresencePenalty
}

// GetReasoning returns the reasoning configuration for the model call.
func (o *CallOptions) GetReasoning() *ReasoningConfig {
	return o.Reasoning
}

// GetJSONMode returns the JSON mode flag.
func (o *CallOptions) GetJSONMode() bool {
	return o.JSONMode
}

// GetMetadata returns the metadata to include in the request.
func (o *CallOptions) GetMetadata() map[string]interface{} {
	return o.Metadata
}

// GetResponseMIMEType returns the ResponseMIMEType.
func (o *CallOptions) GetResponseMIMEType() string {
	if o.ResponseMIMEType == nil {
		return ""
	}
	return *o.ResponseMIMEType
}

// GetVoice returns the voice to use.
func (o *CallOptions) GetVoice() string {
	if o.Voice == nil {
		return ""
	}
	return *o.Voice
}

// GetSpeed returns the speed of the voice.
func (o *CallOptions) GetSpeed() float64 {
	if o.Speed == nil {
		return DefaultSpeed
	}
	return *o.Speed
}

// GetResponseFormat returns the response format.
func (o *CallOptions) GetResponseFormat() string {
	if o.ResponseFormat == nil {
		return ""
	}
	return *o.ResponseFormat
}

// GetWebSearchOptions returns the web search options.
func (o *CallOptions) GetWebSearchOptions() *WebSearchOptions {
	return o.WebSearchOptions
}

// GetToolChoice returns the choice of tool to use.
func (o *CallOptions) GetToolChoice() any {
	return o.ToolChoice
}

// GetTools returns the tools to use.
func (o *CallOptions) GetTools() []Tool {
	return o.Tools
}

// GetFunctions returns the functions to include in the request.
func (o *CallOptions) GetFunctions() []FunctionDefinition {
	return o.Functions
}

// GetFunctionCallBehavior returns the behavior to use when calling functions.
func (o *CallOptions) GetFunctionCallBehavior() FunctionCallBehavior {
	return o.FunctionCallBehavior
}

// GetStreamingFunc returns the streaming function to use.
func (o *CallOptions) GetStreamingFunc() streaming.Callback {
	return o.StreamingFunc
}

// Tool is a tool that can be used by the model.
type Tool struct {
	// Type is the type of the tool.
	Type string `json:"type"`
	// Function is the function to call.
	Function *FunctionDefinition `json:"function,omitempty"`
}

// FunctionDefinition is a definition of a function that can be called by the model.
type FunctionDefinition struct {
	// Name is the name of the function.
	Name string `json:"name"`
	// Description is a description of the function.
	Description string `json:"description"`
	// Parameters is a list of parameters for the function.
	Parameters any `json:"parameters,omitempty"`
	// Strict is a flag to indicate if the function should be called strictly.
	// Provider support varies - typically used for structured output guarantees.
	Strict bool `json:"strict,omitempty"`
}

// ToolChoice is a specific tool to use.
type ToolChoice struct {
	// Type is the type of the tool.
	Type string `json:"type"`
	// Function is the function to call (if the tool is a function).
	Function *FunctionReference `json:"function,omitempty"`
}

// FunctionReference is a reference to a function.
type FunctionReference struct {
	// Name is the name of the function.
	Name string `json:"name"`
}

// FunctionCallBehavior is the behavior to use when calling functions.
type FunctionCallBehavior string

// WebSearchOptions configures web search behavior for models that support web search.
// This is currently supported by OpenAI models like gpt-4o-search-preview.
type WebSearchOptions struct {
	// SearchContextSize controls how much context is gathered from web search.
	// Valid values: "low", "medium", "high". Higher values provide more context
	// but increase latency and cost.
	SearchContextSize string `json:"search_context_size,omitempty"`

	// UserLocation provides approximate user location for localized search results.
	UserLocation *UserLocation `json:"user_location,omitempty"`
}

// UserLocation represents the user's approximate location for web search.
type UserLocation struct {
	// Type must be "approximate" for user-provided location.
	Type string `json:"type"`

	// Approximate contains the approximate location details.
	Approximate *ApproximateLocation `json:"approximate,omitempty"`
}

// ApproximateLocation contains approximate location information.
type ApproximateLocation struct {
	// Country is the two-letter ISO country code (e.g., "US", "GB").
	Country string `json:"country,omitempty"`

	// City is the city name (e.g., "San Francisco", "London").
	City string `json:"city,omitempty"`

	// Region is the region or state (e.g., "California", "London").
	Region string `json:"region,omitempty"`
}

const (
	// FunctionCallBehaviorNone will not call any functions.
	FunctionCallBehaviorNone FunctionCallBehavior = "none"
	// FunctionCallBehaviorAuto will call functions automatically.
	FunctionCallBehaviorAuto FunctionCallBehavior = "auto"
)

// WithModel specifies which model name to use.
func WithModel(model string) CallOption {
	return func(o *CallOptions) {
		o.Model = &model
	}
}

// WithMaxTokens specifies the max number of tokens to generate.
func WithMaxTokens(maxTokens int) CallOption {
	return func(o *CallOptions) {
		o.MaxTokens = &maxTokens
	}
}

// WithCandidateCount specifies the number of response candidates to generate.
func WithCandidateCount(c int) CallOption {
	return func(o *CallOptions) {
		o.CandidateCount = &c
	}
}

// WithTemperature specifies the model temperature, a hyperparameter that
// regulates the randomness, or creativity, of the AI's responses.
func WithTemperature(temperature float64) CallOption {
	return func(o *CallOptions) {
		o.Temperature = &temperature
	}
}

// WithStopWords specifies a list of words to stop generation on.
func WithStopWords(stopWords []string) CallOption {
	return func(o *CallOptions) {
		o.StopWords = stopWords
	}
}

// WithOptions specifies options.
func WithOptions(options CallOptions) CallOption {
	return func(o *CallOptions) {
		(*o) = options
	}
}

// WithStreamingFunc specifies the streaming function to use.
func WithStreamingFunc(streamingFunc streaming.Callback) CallOption {
	return func(o *CallOptions) {
		o.StreamingFunc = streamingFunc
	}
}

// WithTopK will add an option to use top-k sampling.
func WithTopK(topK int) CallOption {
	return func(o *CallOptions) {
		o.TopK = &topK
	}
}

// WithTopP	will add an option to use top-p sampling.
func WithTopP(topP float64) CallOption {
	return func(o *CallOptions) {
		o.TopP = &topP
	}
}

// WithMinP will add an option to use min-p sampling.
func WithMinP(minP float64) CallOption {
	return func(o *CallOptions) {
		o.MinP = &minP
	}
}

// WithSeed will add an option to use deterministic sampling.
func WithSeed(seed int) CallOption {
	return func(o *CallOptions) {
		o.Seed = &seed
	}
}

// WithMinLength will add an option to set the minimum length of the generated text.
func WithMinLength(minLength int) CallOption {
	return func(o *CallOptions) {
		o.MinLength = &minLength
	}
}

// WithMaxLength will add an option to set the maximum length of the generated text.
func WithMaxLength(maxLength int) CallOption {
	return func(o *CallOptions) {
		o.MaxLength = &maxLength
	}
}

// WithN will add an option to set how many chat completion choices to generate for each input message.
func WithN(n int) CallOption {
	return func(o *CallOptions) {
		o.N = &n
	}
}

// WithRepetitionPenalty will add an option to set the repetition penalty for sampling.
func WithRepetitionPenalty(repetitionPenalty float64) CallOption {
	return func(o *CallOptions) {
		o.RepetitionPenalty = &repetitionPenalty
	}
}

// WithFrequencyPenalty will add an option to set the frequency penalty for sampling.
func WithFrequencyPenalty(frequencyPenalty float64) CallOption {
	return func(o *CallOptions) {
		o.FrequencyPenalty = &frequencyPenalty
	}
}

// WithPresencePenalty will add an option to set the presence penalty for sampling.
func WithPresencePenalty(presencePenalty float64) CallOption {
	return func(o *CallOptions) {
		o.PresencePenalty = &presencePenalty
	}
}

// WithReasoning sets the reasoning configuration for the model call.
// You can specify either the reasoning effort or the number of tokens to allocate for reasoning.
// If both effort is ReasoningNone and tokens is 0, reasoning will be disabled.
// Note: Most LLM providers expect only one of these options to be set at a time.
// Internally, the options may be converted between each other according to predefined rules.
func WithReasoning(effort ReasoningEffort, tokens int) CallOption {
	return func(o *CallOptions) {
		o.Reasoning = &ReasoningConfig{
			Effort: effort,
			Tokens: tokens,
		}
	}
}

// WithFunctionCallBehavior will add an option to set the behavior to use when calling functions.
// Deprecated: Use WithToolChoice instead.
func WithFunctionCallBehavior(behavior FunctionCallBehavior) CallOption {
	return func(o *CallOptions) {
		o.FunctionCallBehavior = behavior
	}
}

// WithFunctions will add an option to set the functions to include in the request.
// Deprecated: Use WithTools instead.
func WithFunctions(functions []FunctionDefinition) CallOption {
	return func(o *CallOptions) {
		o.Functions = functions
	}
}

// WithToolChoice will add an option to set the choice of tool to use.
// It can either be "none", "auto" (the default behavior), or a specific tool as described in the ToolChoice type.
func WithToolChoice(choice any) CallOption {
	// TODO: Add type validation for choice.
	return func(o *CallOptions) {
		o.ToolChoice = choice
	}
}

// WithTools will add an option to set the tools to use.
func WithTools(tools []Tool) CallOption {
	return func(o *CallOptions) {
		o.Tools = tools
	}
}

// WithJSONMode will add an option to set the response format to JSON.
// This is useful for models that return structured data.
func WithJSONMode() CallOption {
	return func(o *CallOptions) {
		o.JSONMode = true
	}
}

// WithMetadata will add an option to set metadata to include in the request.
// The meaning of this field is specific to the backend in use.
func WithMetadata(metadata map[string]interface{}) CallOption {
	return func(o *CallOptions) {
		o.Metadata = metadata
	}
}

// WithResponseMIMEType will add an option to set the ResponseMIMEType.
// Provider support varies - check your provider's documentation.
func WithResponseMIMEType(responseMIMEType string) CallOption {
	return func(o *CallOptions) {
		o.ResponseMIMEType = &responseMIMEType
	}
}

// WithVoice will add an option to set the voice to use.
func WithVoice(voice string) CallOption {
	return func(o *CallOptions) {
		o.Voice = &voice
	}
}

// WithSpeed will add an option to set the speed of the voice.
func WithSpeed(speed float64) CallOption {
	return func(o *CallOptions) {
		o.Speed = &speed
	}
}

// WithResponseFormat will add an option to set the response format.
func WithResponseFormat(responseFormat string) CallOption {
	return func(o *CallOptions) {
		o.ResponseFormat = &responseFormat
	}
}

// WithWebSearch enables web search for models that support it.
// Use with OpenAI models like gpt-4o-search-preview and gpt-4o-mini-search-preview.
// Pass nil for default web search behavior, or provide WebSearchOptions to customize.
func WithWebSearch(options *WebSearchOptions) CallOption {
	return func(o *CallOptions) {
		if options == nil {
			o.WebSearchOptions = &WebSearchOptions{}
		} else {
			o.WebSearchOptions = options
		}
	}
}

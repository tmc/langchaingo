package databricksclientsmistralv1

// ChatCompletionPayload represents the payload for the chat completion request.
type ChatCompletionPayload struct {
	Model          string         `json:"model"`                 // The model to use for completion
	Messages       []ChatMessage  `json:"messages"`              // The messages to use for completion
	Temperature    float64        `json:"temperature,omitempty"` // The temperature to use for sampling. Higher values like 0.8 will make the output more random, while lower values like 0.2 will make it more focused and deterministic. We generally recommend altering this or TopP but not both.
	MaxTokens      int            `json:"max_tokens,omitempty"`
	TopP           float64        `json:"top_p,omitempty"` // An alternative to sampling with temperature, called nucleus sampling, where the model considers the results of the tokens with top_p probability mass. So 0.1 means only the tokens comprising the top 10% probability mass are considered. We generally recommend altering this or Temperature but not both.
	RandomSeed     int            `json:"random_seed,omitempty"`
	SafePrompt     bool           `json:"safe_prompt,omitempty"` // Adds a Mistral defined safety message to the system prompt to enforce guardrailing
	Tools          []Tool         `json:"tools,omitempty"`
	ToolChoice     string         `json:"tool_choice,omitempty"`
	ResponseFormat ResponseFormat `json:"response_format,omitempty"`
	Stream         bool           `json:"stream,omitempty"` // Enable token-by-token streaming
}

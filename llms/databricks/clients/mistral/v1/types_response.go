package databricksclientsmistralv1

// ChatCompletionResponse represents the response from the chat completion endpoint.
type ChatCompletionResponse struct {
	ID      string                         `json:"id"`
	Object  string                         `json:"object"`
	Created int                            `json:"created"`
	Model   string                         `json:"model"`
	Choices []ChatCompletionResponseChoice `json:"choices"`
	Usage   UsageInfo                      `json:"usage"`
}

// ChatCompletionStreamResponse represents the streamed response from the chat completion endpoint.
type ChatCompletionStreamResponse struct {
	ID      string                               `json:"id"`
	Model   string                               `json:"model"`
	Choices []ChatCompletionResponseChoiceStream `json:"choices"`
	Created int                                  `json:"created,omitempty"`
	Object  string                               `json:"object,omitempty"`
	Usage   UsageInfo                            `json:"usage,omitempty"`
	Error   error                                `json:"error,omitempty"`
}

// ChatCompletionResponseChoice represents a choice in the chat completion response.
type ChatCompletionResponseChoice struct {
	Index        int          `json:"index"`
	Message      ChatMessage  `json:"message"`
	FinishReason FinishReason `json:"finish_reason,omitempty"`
}

// ChatCompletionResponseChoice represents a choice in the chat completion response.
type ChatCompletionResponseChoiceStream struct {
	Index        int          `json:"index"`
	Delta        ChatMessage  `json:"delta"`
	FinishReason FinishReason `json:"finish_reason,omitempty"`
}

// UsageInfo represents the usage information of a response.
type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	TotalTokens      int `json:"total_tokens"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
}

package databricksclientsllama31

type Role string

const (
	RoleSystem    Role = "system"    // The system role provides instructions or context for the model
	RoleUser      Role = "user"      // The user role represents inputs from the user
	RoleAssistant Role = "assistant" // The assistant role represents responses from the model
	RoleIPython   Role = "ipython"   // The ipython role represents responses from the model
)

type LlamaMessage struct {
	Role    Role   `json:"role"`    // Role of the message sender (e.g., "system", "user", "assistant")
	Content string `json:"content"` // The content of the message
}

type LlamaPayload struct {
	Model            string         `json:"model"`                       // Model to use (e.g., "llama-3.1")
	Messages         []LlamaMessage `json:"messages"`                    // List of structured messages
	Temperature      float64        `json:"temperature,omitempty"`       // Sampling temperature (0 to 1)
	MaxTokens        int            `json:"max_tokens,omitempty"`        // Maximum number of tokens to generate
	TopP             float64        `json:"top_p,omitempty"`             // Top-p (nucleus) sampling
	FrequencyPenalty float64        `json:"frequency_penalty,omitempty"` // Penalizes new tokens based on frequency
	PresencePenalty  float64        `json:"presence_penalty,omitempty"`  // Penalizes tokens based on presence
	Stop             []string       `json:"stop,omitempty"`              // List of stop sequences to end generation
}

type LlamaResponse struct {
	ID      string        `json:"id"`      // Unique ID of the response
	Object  string        `json:"object"`  // Type of response (e.g., "chat.completion")
	Created int64         `json:"created"` // Timestamp of creation
	Model   string        `json:"model"`   // Model used (e.g., "llama-3.1")
	Choices []LlamaChoice `json:"choices"` // List of response choices
	Usage   LlamaUsage    `json:"usage"`   // Token usage details
}

type LlamaChoice struct {
	Index        int          `json:"index"`         // Index of the choice
	Message      LlamaMessage `json:"message"`       // The message content
	FinishReason string       `json:"finish_reason"` // Why the response stopped (e.g., "stop")
}

type LlamaUsage struct {
	PromptTokens     int `json:"prompt_tokens"`     // Tokens used for the prompt
	CompletionTokens int `json:"completion_tokens"` // Tokens used for the completion
	TotalTokens      int `json:"total_tokens"`      // Total tokens used
}

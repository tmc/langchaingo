package databricksclientsllama31

type Role string

const (
	RoleSystem    Role = "system"    // The system role provides instructions or context for the model
	RoleUser      Role = "user"      // The user role represents inputs from the user
	RoleAssistant Role = "assistant" // The assistant role represents responses from the model
	RoleIPython   Role = "ipython"   // The ipython role represents responses from the model
)

// LlamaMessage represents a message in the LLM.
type LlamaMessage struct {
	Role    Role   `json:"role"`    // Role of the message sender (e.g., "system", "user", "assistant")
	Content string `json:"content"` // The content of the message
}

// LlamaMessageDelta represents a message streamed by the LLM.
type LlamaMessageDelta struct {
	Content string `json:"content"` // The content of the message
}

// LlamaPayload represents the payload structure for the Llama model.
type LlamaPayload struct {
	Model            string         `json:"model"`                       // Model to use (e.g., "llama-3.1")
	Messages         []LlamaMessage `json:"messages"`                    // List of structured messages
	Temperature      float64        `json:"temperature,omitempty"`       // Sampling temperature (0 to 1)
	MaxTokens        int            `json:"max_tokens,omitempty"`        // Maximum number of tokens to generate
	TopP             float64        `json:"top_p,omitempty"`             // Top-p (nucleus) sampling
	FrequencyPenalty float64        `json:"frequency_penalty,omitempty"` // Penalizes new tokens based on frequency
	PresencePenalty  float64        `json:"presence_penalty,omitempty"`  // Penalizes tokens based on presence
	Stop             []string       `json:"stop,omitempty"`              // List of stop sequences to end generation
	Stream           bool           `json:"stream,omitempty"`            // Enable token-by-token streaming
}

// LlamaResponse represents the response structure for the Llama model. (full answer or streamed one)
type LlamaResponse[T LlamaChoice | LlamaChoiceDelta] struct {
	ID      string     `json:"id"`      // Unique ID of the response
	Object  string     `json:"object"`  // Type of response (e.g., "chat.completion")
	Created int64      `json:"created"` // Timestamp of creation
	Model   string     `json:"model"`   // Model used (e.g., "llama-3.1")
	Choices []T        `json:"choices"` // List of response choices
	Usage   LlamaUsage `json:"usage"`   // Token usage details
}

// LlamaChoice represents a choice in the Llama response.
type LlamaChoice struct {
	Index        int          `json:"index"`         // Index of the choice
	Message      LlamaMessage `json:"message"`       // The message content
	FinishReason string       `json:"finish_reason"` // Why the response stopped (e.g., "stop")
}

// LlamaChoiceDelta represents a choice in the Llama response.
type LlamaChoiceDelta struct {
	Index        int               `json:"index"`         // Index of the choice
	Delta        LlamaMessageDelta `json:"delta"`         // The message content
	FinishReason string            `json:"finish_reason"` // Why the response stopped (e.g., "stop")
}

// LlamaUsage represents the token usage details of a response.
type LlamaUsage struct {
	PromptTokens     int `json:"prompt_tokens"`     // Tokens used for the prompt
	CompletionTokens int `json:"completion_tokens"` // Tokens used for the completion
	TotalTokens      int `json:"total_tokens"`      // Total tokens used
}

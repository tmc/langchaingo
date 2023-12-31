package schema

import (
	"errors"
	"github.com/tmc/langchaingo/schema"
)

var (
	// ErrMissingValue is returned when a value is missing.
	ErrMissingValue = errors.New("missing value")
	// ErrInvalidValue is returned when a value is invalid.
	ErrInvalidValue = errors.New("invalid value")
	// ErrInvalidReturnType is returned when the model generates the wrong type for the return value
	ErrInvalidReturnType = errors.New("unsupported model")
)

var DefaultParameters = map[string]interface{}{ //nolint:gochecknoglobals
	"temperature":     0.2, //nolint:gomnd
	"maxOutputTokens": 256, //nolint:gomnd
	"topP":            0.8, //nolint:gomnd
	"topK":            40,  //nolint:gomnd
}

// ErrEmptyResponse is returned when the OpenAI API returns an empty response.
var ErrEmptyResponse = errors.New("empty response")

// CompletionRequest is a request to create a completion.
type CompletionRequest struct {
	Prompts       []string `json:"prompts"`
	MaxTokens     int      `json:"max_tokens"`
	Temperature   float64  `json:"temperature,omitempty"`
	TopP          int      `json:"top_p,omitempty"`
	TopK          int      `json:"top_k,omitempty"`
	StopSequences []string `json:"stop_sequences"`

	Model string `json:"model"`
}

// Completion is a completion.
type Completion struct {
	Text string `json:"text"`
}

type ChatRequest struct {
	Context        string         `json:"context"`
	Messages       []*ChatMessage `json:"messages"`
	Temperature    float64        `json:"temperature,omitempty"`
	TopP           int            `json:"top_p,omitempty"`
	TopK           int            `json:"top_k,omitempty"`
	CandidateCount int            `json:"candidate_count,omitempty"`
}

// ChatMessage is a message in a chat.
type ChatMessage struct {
	// The content of the message.
	Content string `json:"content"`
	// The name of the author of this message. user or bot
	Author string `json:"author,omitempty"`
}

// ChatRequest is a request to complete a chat.
// Statically assert that the types implement the interface.
var _ schema.ChatMessage = ChatMessage{}

// GetType returns the type of the message.
func (m ChatMessage) GetType() schema.ChatMessageType {
	switch m.Author {
	case "user":
		return schema.ChatMessageTypeHuman
	default:
		return schema.ChatMessageTypeAI
	}
}

// GetText returns the text of the message.
func (m ChatMessage) GetContent() string {
	return m.Content
}

// ChatResponse is a response to a chat request.
type ChatResponse struct {
	Candidates []ChatMessage
}

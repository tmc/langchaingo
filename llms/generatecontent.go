package llms

import (
	"encoding/json"

	"github.com/tmc/langchaingo/schema"
)

// MessageContent is the content of a message sent to a LLM. It has a role and a
// sequence of parts. For example, it can represent one message in a chat
// session sent by the user, in which case Role will be
// schema.ChatMessageTypeHuman and Parts will be the sequence of items sent in
// this specific message.
type MessageContent struct {
	Role  schema.ChatMessageType
	Parts []ContentPart
}

// TextPart creates TextContent from a given string.
func TextPart(s string) TextContent {
	return TextContent{Text: s}
}

// BinaryPart creates a new BinaryContent from the given MIME type (e.g.
// "image/png" and binary data).
func BinaryPart(mime string, data []byte) BinaryContent {
	return BinaryContent{
		MIMEType: mime,
		Data:     data,
	}
}

// ImageURLPart creates a new ImageURLContent from the given URL.
func ImageURLPart(url string) ImageURLContent {
	return ImageURLContent{
		URL: url,
	}
}

// ContentPart is an interface all parts of content have to implement.
type ContentPart interface {
	isPart()
}

// TextContent is content with some text.
type TextContent struct {
	Text string
}

func (tc TextContent) MarshalJSON() ([]byte, error) {
	m := map[string]string{
		"type": "text",
		"text": tc.Text,
	}
	return json.Marshal(m)
}

func (TextContent) isPart() {}

// ImageURLContent is content with an URL pointing to an image.
type ImageURLContent struct {
	URL string
}

func (iuc ImageURLContent) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"type": "image_url",
		"image_url": map[string]string{
			"url": iuc.URL,
		},
	}
	return json.Marshal(m)
}

func (ImageURLContent) isPart() {}

// BinaryContent is content holding some binary data with a MIME type.
type BinaryContent struct {
	MIMEType string
	Data     []byte
}

func (BinaryContent) isPart() {}

// ToolCall is a call to a tool (as requested by the model) that should be executed.
type ToolCall struct {
	// ID is the unique identifier of the tool call.
	ID string `json:"id"`
	// Type is the type of the tool call.
	Type string `json:"type"`
	// FunctionCall is the function call to be executed.
	FunctionCall *schema.FunctionCall `json:"function,omitempty"`
}

func (ToolCall) isPart() {}

// ToolCallResponse is the response returned by a tool call.
type ToolCallResponse struct {
	// ToolCallID is the ID of the tool call this response is for.
	ToolCallID string `json:"tool_call_id"`
	// Name is the name of the tool that was called.
	Name string `json:"name"`
	// Content is the textual content of the response.
	Content string `json:"content"`
}

func (ToolCallResponse) isPart() {}

// ContentResponse is the response returned by a GenerateContent call.
// It can potentially return multiple content choices.
type ContentResponse struct {
	Choices []*ContentChoice
}

// ContentChoice is one of the response choices returned by GenerateContent
// calls.
type ContentChoice struct {
	// Content is the textual content of a response
	Content string

	// StopReason is the reason the model stopped generating output.
	StopReason string

	// GenerationInfo is arbitrary information the model adds to the response.
	GenerationInfo map[string]any

	// FuncCall is non-nil when the model asks to invoke a function/tool.
	// If a model invokes more than one function/tool, this field will only
	// contain the first one.
	FuncCall *schema.FunctionCall

	// ToolCalls is a list of tool calls the model asks to invoke.
	ToolCalls []schema.ToolCall
}

// TextParts is a helper function to create a MessageContent with a role and a
// list of text parts.
func TextParts(role schema.ChatMessageType, parts ...string) MessageContent {
	result := MessageContent{
		Role:  role,
		Parts: []ContentPart{},
	}
	for _, part := range parts {
		result.Parts = append(result.Parts, TextPart(part))
	}
	return result
}

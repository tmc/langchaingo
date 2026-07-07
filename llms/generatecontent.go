package llms

import (
	"encoding/base64"
	"fmt"
	"io"
)

// MessageContent is the content of a message sent to a LLM. It has a role and a
// sequence of parts. For example, it can represent one message in a chat
// session sent by the user, in which case Role will be
// ChatMessageTypeHuman and Parts will be the sequence of items sent in
// this specific message.
type MessageContent struct {
	Role  ChatMessageType
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

// ImageURLWithDetailPart creates a new ImageURLContent from the given URL and detail.
func ImageURLWithDetailPart(url string, detail string) ImageURLContent {
	return ImageURLContent{
		URL:    url,
		Detail: detail,
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

func (tc TextContent) String() string {
	return tc.Text
}

func (TextContent) isPart() {}

// ImageURLContent is content with an URL pointing to an image.
type ImageURLContent struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"` // Detail is the detail of the image, e.g. "low", "high".
}

func (iuc ImageURLContent) String() string {
	return iuc.URL
}

func (ImageURLContent) isPart() {}

// BinaryContent is content holding some binary data with a MIME type.
type BinaryContent struct {
	MIMEType string
	Data     []byte
}

func (bc BinaryContent) String() string {
	base64Encoded := base64.StdEncoding.EncodeToString(bc.Data)
	return "data:" + bc.MIMEType + ";base64," + base64Encoded
}

func (BinaryContent) isPart() {}

// FunctionCall is the name and arguments of a function call.
type FunctionCall struct {
	// The name of the function to call.
	Name string `json:"name"`
	// The arguments to pass to the function, as a JSON string.
	Arguments string `json:"arguments"`
}

// ToolCall is a call to a tool (as requested by the model) that should be executed.
type ToolCall struct {
	// ID is the unique identifier of the tool call.
	ID string `json:"id"`
	// Type is the type of the tool call. Typically, this would be "function".
	Type string `json:"type"`
	// FunctionCall is the function call to be executed.
	FunctionCall *FunctionCall `json:"function,omitempty"`
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

// ThinkingContent is a thinking/reasoning block produced by a model.
// Signature is an opaque token some providers attach to thinking
// blocks; it must be sent back unmodified when the message is replayed
// to the model, for example in a tool-calling loop.
type ThinkingContent struct {
	Thinking  string `json:"thinking"`
	Signature string `json:"signature,omitempty"`
}

func (tc ThinkingContent) String() string {
	return tc.Thinking
}

func (ThinkingContent) isPart() {}

// RedactedThinkingContent is an encrypted thinking block produced by a
// model. Data is opaque; it must be sent back unmodified when the
// message is replayed to the model.
type RedactedThinkingContent struct {
	Data string `json:"data"`
}

func (RedactedThinkingContent) isPart() {}

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
	FuncCall *FunctionCall

	// ToolCalls is a list of tool calls the model asks to invoke.
	ToolCalls []ToolCall

	// This field is only used with the deepseek-reasoner model and represents the reasoning contents of the assistant message before the final answer.
	ReasoningContent string

	// Parts, when populated by a provider, carries the typed content
	// parts of this choice (TextContent, ToolCall, ThinkingContent,
	// RedactedThinkingContent, ...). Concatenating the Parts of a
	// response's choices yields an AI MessageContent that replays the
	// response to the model, preserving blocks such as signed thinking
	// that must round-trip in tool-calling loops.
	Parts []ContentPart
}

// AssistantMessage returns the response as an AI MessageContent for
// appending to a conversation, for example between the calls of a
// tool-calling loop.
//
// The parts of every choice are concatenated in order, which preserves
// blocks such as signed thinking that a model requires replayed
// verbatim. That imposes a requirement on providers: a provider may
// populate ContentChoice.Parts only when its choices are the blocks of
// a single response. A provider whose choices are alternative
// completions must leave Parts empty, or those alternatives would be
// merged into one message here.
//
// When no choice carries parts, the message is assembled from the
// first choice's text and tool calls. A nil response, or a nil first
// choice, yields an empty AI message.
func (cr *ContentResponse) AssistantMessage() MessageContent {
	msg := MessageContent{Role: ChatMessageTypeAI}
	if cr == nil {
		return msg
	}
	for _, c := range cr.Choices {
		if c != nil {
			msg.Parts = append(msg.Parts, c.Parts...)
		}
	}
	if len(msg.Parts) > 0 || len(cr.Choices) == 0 || cr.Choices[0] == nil {
		return msg
	}
	c := cr.Choices[0]
	if c.Content != "" {
		msg.Parts = append(msg.Parts, TextContent{Text: c.Content})
	}
	for _, tc := range c.ToolCalls {
		msg.Parts = append(msg.Parts, tc)
	}
	return msg
}

// TextParts is a helper function to create a MessageContent with a role and a
// list of text parts.
func TextParts(role ChatMessageType, parts ...string) MessageContent {
	result := MessageContent{
		Role:  role,
		Parts: []ContentPart{},
	}
	for _, part := range parts {
		result.Parts = append(result.Parts, TextPart(part))
	}
	return result
}

// ShowMessageContents is a debugging helper for MessageContent.
func ShowMessageContents(w io.Writer, msgs []MessageContent) {
	fmt.Fprintf(w, "MessageContent (len=%v)\n", len(msgs))
	for i, mc := range msgs {
		fmt.Fprintf(w, "[%d]: Role=%s\n", i, mc.Role)
		for j, p := range mc.Parts {
			fmt.Fprintf(w, "  Parts[%v]: ", j)
			switch pp := p.(type) {
			case TextContent:
				fmt.Fprintf(w, "TextContent %q\n", pp.Text)
			case ImageURLContent:
				fmt.Fprintf(w, "ImageURLPart %q\n", pp.URL)
			case BinaryContent:
				fmt.Fprintf(w, "BinaryContent MIME=%q, size=%d\n", pp.MIMEType, len(pp.Data))
			case ToolCall:
				fmt.Fprintf(w, "ToolCall ID=%v, Type=%v, Func=%v(%v)\n", pp.ID, pp.Type, pp.FunctionCall.Name, pp.FunctionCall.Arguments)
			case ToolCallResponse:
				fmt.Fprintf(w, "ToolCallResponse ID=%v, Name=%v, Content=%v\n", pp.ToolCallID, pp.Name, pp.Content)
			case ThinkingContent:
				fmt.Fprintf(w, "ThinkingContent len=%d, signed=%v\n", len(pp.Thinking), pp.Signature != "")
			case RedactedThinkingContent:
				fmt.Fprintf(w, "RedactedThinkingContent len=%d\n", len(pp.Data))
			default:
				fmt.Fprintf(w, "unknown type %T\n", pp)
			}
		}
	}
}

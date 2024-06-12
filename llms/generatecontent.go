package llms

import (
	"encoding/base64"
	"encoding/json"
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

func (iuc *ImageURLContent) UnmarshalJSON(data []byte) error {
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	_, ok := m["type"].(string)
	if !ok {
		return fmt.Errorf(`missing "type" field in ImageURLContent`)
	}
	imageURL, ok := m["image_url"].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid image_url field in ImageURLContent")
	}
	url, ok := imageURL["url"].(string)
	if !ok {
		return fmt.Errorf("invalid url field in ImageURLContent")
	}
	iuc.URL = url
	return nil
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

func (bc BinaryContent) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"type": "binary",
		"binary": map[string]string{
			"mime_type": bc.MIMEType,
			"data":      base64.StdEncoding.EncodeToString(bc.Data),
		},
	}
	return json.Marshal(m)
}

func (bc *BinaryContent) UnmarshalJSON(data []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	if m["type"] != "binary" {
		return fmt.Errorf("invalid type for BinaryContent: %v", m["type"])
	}
	binary, ok := m["binary"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid binary field in BinaryContent")
	}
	mimeType, ok := binary["mime_type"].(string)
	if !ok {
		return fmt.Errorf("invalid mime_type field in BinaryContent")
	}
	encodedData, ok := binary["data"].(string)
	if !ok {
		return fmt.Errorf("invalid data field in BinaryContent")
	}
	enc, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return fmt.Errorf("error decoding data: %v", err)
	}
	bc.MIMEType = mimeType
	bc.Data = enc
	return nil
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

func (tc ToolCall) MarshalJSON() ([]byte, error) {
	fc, err := json.Marshal(tc.FunctionCall)
	if err != nil {
		return nil, err
	}

	m := map[string]any{
		"type": "tool_call",
		"tool_call": map[string]any{
			"id":       tc.ID,
			"type":     tc.Type,
			"function": json.RawMessage(fc),
		},
	}
	return json.Marshal(m)
}

func (tc *ToolCall) UnmarshalJSON(data []byte) error {
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	_, ok := m["type"].(string)
	if !ok {
		return fmt.Errorf(`missing "type" field in ToolCall`)
	}
	toolCall, ok := m["tool_call"].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid tool_call field in ToolCall")
	}
	fmt.Println("Dec tC:", toolCall)
	id, ok := toolCall["id"].(string)
	if !ok {
		return fmt.Errorf("invalid id field in ToolCall")
	}
	typ, ok := toolCall["type"].(string)
	if !ok {
		return fmt.Errorf("invalid type field in ToolCall")
	}
	var fc FunctionCall
	fcData, ok := toolCall["function"].(json.RawMessage)
	if ok {
		if err := json.Unmarshal(fcData, &fc); err != nil {
			return fmt.Errorf("error unmarshalling function call: %v", err)
		}
	}
	tc.ID = id
	tc.Type = typ
	tc.FunctionCall = &fc
	return nil
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

func (tc ToolCallResponse) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"type": "tool_response",
		"tool_response": map[string]string{
			"tool_call_id": tc.ToolCallID,
			"name":         tc.Name,
			"content":      tc.Content,
		},
	}
	return json.Marshal(m)
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
	FuncCall *FunctionCall

	// ToolCalls is a list of tool calls the model asks to invoke.
	ToolCalls []ToolCall
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
			default:
				fmt.Fprintf(w, "unknown type %T\n", pp)
			}
		}
	}
}

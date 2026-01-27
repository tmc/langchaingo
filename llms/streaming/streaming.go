package streaming

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

type ChunkType string

const (
	ChunkTypeNone      ChunkType = ""
	ChunkTypeText      ChunkType = "text"
	ChunkTypeReasoning ChunkType = "reasoning"
	ChunkTypeToolCall  ChunkType = "tool_call"
	ChunkTypeDone      ChunkType = "done"
)

type ToolCall struct {
	ID        string                      `json:"id"`
	Name      string                      `json:"name"`
	Arguments string                      `json:"arguments"`
	Reasoning *reasoning.ContentReasoning `json:"reasoning,omitempty"`
}

func (t *ToolCall) String() string {
	return fmt.Sprintf("ToolCall{ID: %s, Name: %s, Arguments: %s}", t.ID, t.Name, t.Arguments)
}

func (t *ToolCall) Parse() (map[string]any, error) {
	var result map[string]any
	if t.Arguments == "" { // it's a special case for tool call without arguments
		return map[string]any{}, nil
	}
	if err := json.Unmarshal([]byte(t.Arguments), &result); err != nil {
		return nil, err
	}
	return result, nil
}

type Chunk struct {
	Type      ChunkType                   `json:"type"`
	Content   string                      `json:"content"`
	Reasoning *reasoning.ContentReasoning `json:"reasoning,omitempty"`
	ToolCall  ToolCall                    `json:"tool_call"`
}

func (c *Chunk) String() string {
	switch c.Type {
	case ChunkTypeNone:
		return "None"
	case ChunkTypeText:
		return fmt.Sprintf("Text: %s", c.Content)
	case ChunkTypeReasoning:
		return c.Reasoning.String()
	case ChunkTypeToolCall:
		return fmt.Sprintf("ToolCall: %s", c.ToolCall.String())
	case ChunkTypeDone:
		return "Done"
	default:
		return fmt.Sprintf("unexpected chunk type: %s", c.Type)
	}
}

type Callback func(ctx context.Context, chunk Chunk) error

var (
	ErrToolCallIDRequired   = errors.New("tool call id is required")
	ErrToolCallNameRequired = errors.New("tool call name is required")
)

func NewTextChunk(text string) Chunk {
	return Chunk{
		Type:    ChunkTypeText,
		Content: text,
	}
}

func NewReasoningChunk(reasoning *reasoning.ContentReasoning) Chunk {
	return Chunk{
		Type:      ChunkTypeReasoning,
		Reasoning: reasoning,
	}
}

func NewReasoningChunkWithContent(content string) Chunk {
	return Chunk{
		Type:      ChunkTypeReasoning,
		Reasoning: &reasoning.ContentReasoning{Content: content},
	}
}

func NewToolCallChunk(toolCall ToolCall) Chunk {
	return Chunk{
		Type:     ChunkTypeToolCall,
		ToolCall: toolCall,
	}
}

func NewToolCall(id, name, arguments string) ToolCall {
	return ToolCall{
		ID:        id,
		Name:      name,
		Arguments: arguments,
	}
}

func NewToolCallWithReasoning(id, name, arguments string, reasoning *reasoning.ContentReasoning) ToolCall {
	return ToolCall{
		ID:        id,
		Name:      name,
		Arguments: arguments,
		Reasoning: reasoning,
	}
}

func NewDoneChunk() Chunk {
	return Chunk{
		Type: ChunkTypeDone,
	}
}

func CallWithText(ctx context.Context, cb Callback, text string) error {
	if cb == nil {
		return nil
	}
	if text == "" {
		return nil
	}
	return cb(ctx, NewTextChunk(text))
}

func CallWithReasoning(ctx context.Context, cb Callback, reasoning *reasoning.ContentReasoning) error {
	if cb == nil {
		return nil
	}
	if reasoning == nil || reasoning.IsEmpty() {
		return nil
	}
	return cb(ctx, NewReasoningChunk(reasoning))
}

func CallWithReasoningContent(ctx context.Context, cb Callback, content string) error {
	if cb == nil {
		return nil
	}
	if content == "" {
		return nil
	}
	return cb(ctx, NewReasoningChunkWithContent(content))
}

func CallWithToolCall(ctx context.Context, cb Callback, toolCall ToolCall) error {
	if cb == nil {
		return nil
	}
	if toolCall.ID == "" {
		return ErrToolCallIDRequired
	}
	if toolCall.Name == "" {
		return ErrToolCallNameRequired
	}
	return cb(ctx, NewToolCallChunk(toolCall))
}

func CallWithDone(ctx context.Context, cb Callback) error {
	if cb == nil {
		return nil
	}
	return cb(ctx, NewDoneChunk())
}

func AppendToolCall(src ToolCall, dst *ToolCall) {
	dst.ID = src.ID
	dst.Name = src.Name
	dst.Arguments += src.Arguments
}

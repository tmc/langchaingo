package llms

import (
	"reflect"
	"testing"
)

func TestTextParts(t *testing.T) {
	t.Parallel()
	type args struct {
		role  ChatMessageType
		parts []string
	}
	tests := []struct {
		name string
		args args
		want MessageContent
	}{
		{"basics", args{ChatMessageTypeHuman, []string{"a", "b", "c"}}, MessageContent{
			Role: ChatMessageTypeHuman,
			Parts: []ContentPart{
				TextContent{Text: "a"},
				TextContent{Text: "b"},
				TextContent{Text: "c"},
			},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := TextParts(tt.args.role, tt.args.parts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TextParts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssistantMessage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		resp *ContentResponse
		want MessageContent
	}{
		{
			name: "concatenates parts across block-per-choice response",
			resp: &ContentResponse{Choices: []*ContentChoice{
				{Parts: []ContentPart{ThinkingContent{Thinking: "hm", Signature: "sig"}}},
				{Content: "hi", Parts: []ContentPart{TextContent{Text: "hi"}}},
				{Parts: []ContentPart{ToolCall{ID: "t1", FunctionCall: &FunctionCall{Name: "f", Arguments: "{}"}}}},
			}},
			want: MessageContent{Role: ChatMessageTypeAI, Parts: []ContentPart{
				ThinkingContent{Thinking: "hm", Signature: "sig"},
				TextContent{Text: "hi"},
				ToolCall{ID: "t1", FunctionCall: &FunctionCall{Name: "f", Arguments: "{}"}},
			}},
		},
		{
			name: "falls back to first choice content and tool calls",
			resp: &ContentResponse{Choices: []*ContentChoice{
				{
					Content:   "answer",
					ToolCalls: []ToolCall{{ID: "t2", FunctionCall: &FunctionCall{Name: "g", Arguments: "{}"}}},
				},
				{Content: "alternative"},
			}},
			want: MessageContent{Role: ChatMessageTypeAI, Parts: []ContentPart{
				TextContent{Text: "answer"},
				ToolCall{ID: "t2", FunctionCall: &FunctionCall{Name: "g", Arguments: "{}"}},
			}},
		},
		{
			name: "empty response",
			resp: &ContentResponse{},
			want: MessageContent{Role: ChatMessageTypeAI},
		},
		{
			// A nil response reaches here when a caller uses the
			// result of a failed call without checking the error.
			name: "nil response",
			resp: nil,
			want: MessageContent{Role: ChatMessageTypeAI},
		},
		{
			// A provider that reports a malformed choice must not
			// take the caller down with it.
			name: "nil choice among parts",
			resp: &ContentResponse{Choices: []*ContentChoice{
				nil,
				{Parts: []ContentPart{TextContent{Text: "hi"}}},
			}},
			want: MessageContent{Role: ChatMessageTypeAI, Parts: []ContentPart{
				TextContent{Text: "hi"},
			}},
		},
		{
			name: "nil first choice with no parts",
			resp: &ContentResponse{Choices: []*ContentChoice{nil}},
			want: MessageContent{Role: ChatMessageTypeAI},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.resp.AssistantMessage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AssistantMessage() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

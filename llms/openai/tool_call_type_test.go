package openai

import (
	"strings"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
)

func TestToolCallTypeIsNamedOnTheWire(t *testing.T) {
	t.Parallel()

	history := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "weather?"),
		{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{
			llms.ToolCall{
				ID:           "call_1",
				FunctionCall: &llms.FunctionCall{Name: "get_weather", Arguments: `{"city":"Minsk"}`},
			},
		}},
		{Role: llms.ChatMessageTypeTool, Parts: []llms.ContentPart{
			llms.ToolCallResponse{ToolCallID: "call_1", Name: "get_weather", Content: "sunny"},
		}},
	}

	body := sendMessagesForWire(t, "gpt-4o", history)

	if strings.Contains(body, `"type":""`) {
		t.Errorf("a tool call reassembled from a provider that omits the type must not"+
			" echo an empty type, got body: %s", body)
	}
	if !strings.Contains(body, `"type":"function"`) {
		t.Errorf("want the function tool type on the wire, got body: %s", body)
	}
}

package bedrock_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/bedrock"
)

func converseAgentTurnOnWire(t *testing.T, choice any) string {
	t.Helper()

	const answer = `{"output":{"message":{"role":"assistant","content":[{"text":"ok"}]}},` +
		`"stopReason":"end_turn","usage":{"inputTokens":1,"outputTokens":1,"totalTokens":2}}`

	tool := llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "get_weather",
			Description: "weather",
			Parameters:  map[string]any{"type": "object", "properties": map[string]any{}},
		},
	}

	history := []llms.MessageContent{
		{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{
			llms.TextContent{Text: "what is the weather in Paris?"},
		}},
		{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{
			llms.ToolCall{ID: "call_1", Type: "function", FunctionCall: &llms.FunctionCall{
				Name:      "get_weather",
				Arguments: `{"city":"Paris"}`,
			}},
		}},
		{Role: llms.ChatMessageTypeTool, Parts: []llms.ContentPart{
			llms.ToolCallResponse{ToolCallID: "call_1", Name: "get_weather", Content: "20C"},
		}},
	}

	var body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, answer)
	}))
	t.Cleanup(srv.Close)

	llm := bedrockLLMAgainst(t, srv,
		bedrock.WithModel("amazon.nova-lite-v1:0"), bedrock.WithConverseAPI())

	opts := []llms.CallOption{llms.WithTools([]llms.Tool{tool})}
	if choice != nil {
		opts = append(opts, llms.WithToolChoice(choice))
	}
	if _, err := llm.GenerateContent(context.Background(), history, opts...); err != nil {
		t.Fatalf("GenerateContent() error: %v", err)
	}
	return body
}

func TestConverseKeepsToolsWhenTheHistoryCarriesToolBlocks(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		choice any
	}{
		{"none as a bare string", "none"},
		{"none as a struct", llms.ToolChoice{Type: "none"}},
		{"none as a raw map", map[string]any{"type": "none"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			body := converseAgentTurnOnWire(t, tc.choice)
			if !strings.Contains(body, `"toolConfig"`) {
				t.Fatalf("a turn carrying tool blocks needs toolConfig, got body: %s", body)
			}
			if !strings.Contains(body, `"get_weather"`) {
				t.Fatalf("toolConfig must carry the caller's tools, got body: %s", body)
			}
			if !strings.Contains(body, `"auto":{}`) {
				t.Fatalf("converse has no none, so the choice degrades to auto, got body: %s", body)
			}
		})
	}
}

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

func TestConverseCarriesTheToolChoice(t *testing.T) {
	t.Parallel()

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

	capture := func(t *testing.T, choice any) string {
		t.Helper()
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
		if _, err := llm.GenerateContent(context.Background(),
			[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}, opts...); err != nil {
			t.Fatalf("GenerateContent() error: %v", err)
		}
		return body
	}

	t.Run("a named tool reaches the wire", func(t *testing.T) {
		body := capture(t, llms.ToolChoice{
			Type:     "tool",
			Function: &llms.FunctionReference{Name: "get_weather"},
		})
		if !strings.Contains(body, `"tool":{"name":"get_weather"}`) {
			t.Fatalf("named choice must reach the wire, got body: %s", body)
		}
	})

	t.Run("any tool reaches the wire", func(t *testing.T) {
		body := capture(t, llms.ToolChoice{Type: "any"})
		if !strings.Contains(body, `"any":{}`) {
			t.Fatalf("forced choice must reach the wire, got body: %s", body)
		}
	})

	t.Run("an unset choice still leaves it to the model", func(t *testing.T) {
		body := capture(t, nil)
		if !strings.Contains(body, `"auto":{}`) {
			t.Fatalf("an unset choice must stay auto, got body: %s", body)
		}
	})
}

func TestForcedToolName(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		choice any
		want   string
		forced bool
	}{
		{"named tool", llms.ToolChoice{Type: "tool", Function: &llms.FunctionReference{Name: "f"}}, "f", true},
		{"pointer", &llms.ToolChoice{Type: "tool", Function: &llms.FunctionReference{Name: "g"}}, "g", true},
		{"any", llms.ToolChoice{Type: "any"}, "", true},
		{"map with function", map[string]any{"type": "tool", "function": map[string]any{"name": "h"}}, "h", true},
		{"auto", llms.ToolChoice{Type: "auto"}, "", false},
		{"nil", nil, "", false},
	} {
		name, forced := llms.ForcedToolName(tc.choice)
		if name != tc.want || forced != tc.forced {
			t.Errorf("%s: ForcedToolName() = (%q, %v), want (%q, %v)", tc.name, name, forced, tc.want, tc.forced)
		}
	}
}

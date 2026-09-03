package mistral

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
)

func captureMistralRequest(t *testing.T, opts ...llms.CallOption) map[string]any {
	t.Helper()

	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"x","object":"chat.completion","created":1,` +
			`"model":"mistral-small-latest","choices":[{"index":0,` +
			`"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],` +
			`"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`))
	}))
	t.Cleanup(srv.Close)

	model, err := New(WithAPIKey("unit-test-token"), WithEndpoint(srv.URL),
		WithModel("mistral-small-latest"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := model.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}, opts...); err != nil {
		t.Fatalf("GenerateContent: %v", err)
	}
	return body
}

func TestTopPReachesTheMistralWire(t *testing.T) {
	t.Parallel()

	if got := captureMistralRequest(t, llms.WithTopP(0.5))["top_p"]; got != 0.5 {
		t.Fatalf("the caller's top_p must reach the wire, got %v", got)
	}
}

func TestJSONModeReachesTheMistralWire(t *testing.T) {
	t.Parallel()

	body := captureMistralRequest(t, llms.WithJSONMode())
	format, _ := body["response_format"].(map[string]any)
	if format == nil || format["type"] != "json_object" {
		t.Fatalf("JSON mode must reach the wire, got body: %v", body)
	}

	if _, ok := captureMistralRequest(t)["response_format"]; ok {
		t.Fatal("an unset JSON mode must leave the field off the wire")
	}
}

func TestToolChoiceReachesTheMistralWire(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		choice any
		want   any
	}{
		{"any", "any", "any"},
		{"auto", "auto", "auto"},
		{"none", "none", "none"},
		{"required", "required", "any"},
		{"function without a name", "function", "any"},
		{"struct asking for any", llms.ToolChoice{Type: "any"}, "any"},
		{"struct asking for auto", llms.ToolChoice{Type: "auto"}, "auto"},
		{"map asking for required", map[string]any{"type": "required"}, "any"},
		{"named tool", llms.ToolChoice{
			Type: "function", Function: &llms.FunctionReference{Name: "get_weather"},
		}, nil},
		{"a spelling nobody uses", "something-else", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, sent := captureMistralRequest(t, llms.WithToolChoice(tc.choice))["tool_choice"]
			if tc.want == nil {
				if sent {
					t.Errorf("must stay off the wire, got %v", got)
				}
				return
			}
			if got != tc.want {
				t.Errorf("want %v on the wire, got %v", tc.want, got)
			}
		})
	}

	if _, ok := captureMistralRequest(t)["tool_choice"]; ok {
		t.Error("an unset choice must leave the field off the wire")
	}
}

func TestMistralDoesNotPinTheSeedWhenTheCallerDidNotAskForOne(t *testing.T) {
	t.Parallel()

	first, sentFirst := captureMistralRequest(t)["random_seed"]
	second, sentSecond := captureMistralRequest(t)["random_seed"]
	if sentFirst && sentSecond && first == second {
		t.Fatalf("two seedless calls must not share a seed, both got %v", first)
	}
}

func TestMistralKeepsTheSeedTheCallerAsksFor(t *testing.T) {
	t.Parallel()

	if got := captureMistralRequest(t, llms.WithSeed(7))["random_seed"]; got != float64(7) {
		t.Fatalf("the caller's seed must reach the wire, got %v", got)
	}
}

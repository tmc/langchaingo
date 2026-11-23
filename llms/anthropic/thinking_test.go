package anthropic_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"unicode/utf8"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
)

func TestGenerateContentThinkingText(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, text, thinking, output string
	}{
		{"ASCII output", "<thinking>plan</thinking>Done", "plan", "Done"},
		{"UTF-8 output", "<thinking>plan</thinking>完成", "plan", "完成"},
		{"single byte output", "<thinking>plan</thinking>X", "plan", "X"},
		{"single rune output", "<thinking>plan</thinking>好", "plan", "好"},
		{"no tags", "  Done  ", "", "  Done  "},
		{"empty output", "<thinking>plan</thinking>", "plan", ""},
		{"whitespace after tags", "<thinking>plan</thinking> \nDone \t", "plan", "Done"},
		{"Unicode whitespace after tags", "<thinking>plan</thinking>\u2003Done", "plan", "Done"},
		{"prefix and suffix", "Before<thinking>plan</thinking>Done", "plan", "Before\n\nDone"},
		{"prefix only", "Before<thinking>plan</thinking>", "plan", "Before"},
		{"whitespace output", "<thinking>plan</thinking> \n\t", "plan", ""},
		{"trim thinking", "<thinking> \nplan \t</thinking>Done", "plan", "Done"},
		{"unclosed tag", "<thinking>plan", "", "<thinking>plan"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/v1/messages" {
					t.Errorf("request = %s %s, want POST /v1/messages", r.Method, r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(map[string]any{
					"content":     []map[string]string{{"type": "text", "text": tt.text}},
					"stop_reason": "end_turn",
				}); err != nil {
					t.Errorf("encode response: %v", err)
				}
			}))
			defer server.Close()
			model, err := anthropic.New(
				anthropic.WithToken("test-api-key"),
				anthropic.WithBaseURL(server.URL+"/v1"),
				anthropic.WithHTTPClient(server.Client()),
			)
			if err != nil {
				t.Fatal(err)
			}
			response, err := model.GenerateContent(context.Background(), []llms.MessageContent{
				llms.TextParts(llms.ChatMessageTypeHuman, "Hello"),
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(response.Choices) != 1 {
				t.Fatalf("got %d choices, want 1", len(response.Choices))
			}
			choice := response.Choices[0]
			if choice.Content != tt.text {
				t.Errorf("Content = %q, want original text %q", choice.Content, tt.text)
			}
			if got := choice.GenerationInfo["ThinkingContent"]; got != tt.thinking {
				t.Errorf("ThinkingContent = %q, want %q", got, tt.thinking)
			}
			output, ok := choice.GenerationInfo["OutputContent"].(string)
			if !ok {
				t.Fatalf("OutputContent type = %T, want string", choice.GenerationInfo["OutputContent"])
			}
			if output != tt.output {
				t.Errorf("OutputContent = %q, want %q", output, tt.output)
			}
			if !utf8.ValidString(output) {
				t.Errorf("OutputContent contains invalid UTF-8: % x", output)
			}
		})
	}
}

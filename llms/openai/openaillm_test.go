package openai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

func TestToolFromToolFunctionParameters(t *testing.T) {
	tests := []struct {
		name       string
		parameters any
		want       string
	}{
		{
			name:       "zero arguments",
			parameters: map[string]any{"type": "object"},
			want:       `{"type":"object","properties":{}}`,
		},
		{
			name: "one argument",
			parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{"type": "string"},
				},
			},
			want: `{"type":"object","properties":{"query":{"type":"string"}}}`,
		},
		{
			name: "nested argument",
			parameters: json.RawMessage(
				`{"type":"object","properties":{"filter":{"type":"object"}}}`,
			),
			want: `{"type":"object","properties":{"filter":{"type":"object"}}}`,
		},
		{
			name:       "non-object schema",
			parameters: json.RawMessage(`{"type":"string"}`),
			want:       `{"type":"string"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool, err := toolFromTool(llms.Tool{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:       "test_tool",
					Parameters: tt.parameters,
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			got, err := json.Marshal(tool.Function.Parameters)
			if err != nil {
				t.Fatal(err)
			}
			var gotSchema, wantSchema any
			if err := json.Unmarshal(got, &gotSchema); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(tt.want), &wantSchema); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(gotSchema, wantSchema) {
				t.Errorf("parameters = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestGenerateContentFunctionParameters(t *testing.T) {
	var body []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"c","object":"chat.completion","model":"gpt-4o","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}]}`)
	}))
	defer server.Close()

	llm, err := New(WithToken("test-token"), WithBaseURL(server.URL), WithModel("gpt-4o"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "what time is it")},
		llms.WithFunctions([]llms.FunctionDefinition{{Name: "now", Parameters: map[string]any{"type": "object"}}}),
	)
	if err != nil {
		t.Fatal(err)
	}

	var req struct {
		Tools []struct {
			Function struct {
				Parameters any `json:"parameters"`
			} `json:"function"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatalf("decode request: %v: %s", err, body)
	}
	if len(req.Tools) != 1 {
		t.Fatalf("request has %d tools, want 1: %s", len(req.Tools), body)
	}
	want := map[string]any{"type": "object", "properties": map[string]any{}}
	if !reflect.DeepEqual(req.Tools[0].Function.Parameters, want) {
		t.Errorf("parameters = %v, want %v", req.Tools[0].Function.Parameters, want)
	}
}

func TestGetModelCapabilities(t *testing.T) {
	tests := []struct {
		model        string
		wantThinking bool
		wantSystem   bool
	}{
		{"o1-mini", true, false},
		{"o3", true, false},
		{"o4-mini", true, true},
		{"o3-2025-04-16", true, true},
		{"gpt-5", true, true},
		{"gpt-5-mini", true, true},
		{"gpt-4o", false, true},
		{"gpt-4", false, true},
		{"gpt-3.5-turbo", false, true},
	}
	for _, tt := range tests {
		caps := getModelCapabilities(tt.model)
		if caps.SupportsThinking != tt.wantThinking {
			t.Errorf("getModelCapabilities(%q).SupportsThinking = %v, want %v", tt.model, caps.SupportsThinking, tt.wantThinking)
		}
		if caps.SupportsSystem != tt.wantSystem {
			t.Errorf("getModelCapabilities(%q).SupportsSystem = %v, want %v", tt.model, caps.SupportsSystem, tt.wantSystem)
		}
	}
}

func TestClampReasoningEffort(t *testing.T) {
	tests := []struct {
		effort llms.ThinkingEffort
		want   string
	}{
		{"low", "low"},
		{"medium", "medium"},
		{"high", "high"},
		// OpenAI accepts xhigh, so it must not be downgraded.
		{"xhigh", "xhigh"},
		// max is Anthropic-only; it clamps to OpenAI's highest.
		{"max", "xhigh"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := clampReasoningEffort(tt.effort); got != tt.want {
			t.Errorf("clampReasoningEffort(%q) = %q, want %q", tt.effort, got, tt.want)
		}
	}
}

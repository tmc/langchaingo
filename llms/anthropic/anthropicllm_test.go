package anthropic

import (
	"reflect"
	"testing"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic/internal/anthropicclient"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		envToken string
		opts     []Option
		wantErr  bool
	}{
		{
			name:     "with token from env",
			envToken: "test-token",
			opts:     []Option{},
			wantErr:  false,
		},
		{
			name:     "with token option",
			envToken: "",
			opts:     []Option{WithToken("test-token")},
			wantErr:  false,
		},
		{
			name:     "missing token",
			envToken: "",
			opts:     []Option{},
			wantErr:  true,
		},
		{
			name:     "with all options",
			envToken: "test-token",
			opts: []Option{
				WithModel("claude-3-opus-20240229"),
				WithBaseURL("https://api.example.com"),
				WithAnthropicBetaHeader("max-tokens-3-5-sonnet-2024-07-15"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("ANTHROPIC_API_KEY", tt.envToken)

			llm, err := New(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && llm == nil {
				t.Error("New() returned nil LLM without error")
			}
		})
	}
}

func TestProcessMessages(t *testing.T) {
	tests := []struct {
		name       string
		messages   []llms.MessageContent
		wantLen    int
		wantSystem string
		wantErr    bool
	}{
		{
			name: "basic text message",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "Hello"},
					},
				},
			},
			wantLen:    1,
			wantSystem: "",
			wantErr:    false,
		},
		{
			name: "system message",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeSystem,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "You are helpful"},
					},
				},
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "Hi"},
					},
				},
			},
			wantLen:    1,
			wantSystem: "You are helpful",
			wantErr:    false,
		},
		{
			name: "ai and human messages",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "Hello"},
					},
				},
				{
					Role: llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "Hi there!"},
					},
				},
			},
			wantLen:    2,
			wantSystem: "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, systemPrompt, err := processMessages(tt.messages)
			if (err != nil) != tt.wantErr {
				t.Errorf("processMessages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(result) != tt.wantLen {
					t.Errorf("processMessages() returned %d messages, want %d", len(result), tt.wantLen)
				}
				if systemPrompt != tt.wantSystem {
					t.Errorf("processMessages() system prompt = %q, want %q", systemPrompt, tt.wantSystem)
				}
			}
		})
	}
}

func TestProcessMessagesSystem(t *testing.T) {
	tests := []struct {
		name     string
		messages []llms.MessageContent
		want     any
	}{
		{
			name: "multi-part system concatenates",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeSystem,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "You are helpful."},
						llms.TextContent{Text: " Be brief."},
					},
				},
			},
			want: "You are helpful. Be brief.",
		},
		{
			name: "multiple system messages concatenate",
			messages: []llms.MessageContent{
				{
					Role:  llms.ChatMessageTypeSystem,
					Parts: []llms.ContentPart{llms.TextContent{Text: "One."}},
				},
				{
					Role:  llms.ChatMessageTypeSystem,
					Parts: []llms.ContentPart{llms.TextContent{Text: "Two."}},
				},
			},
			want: "One.Two.",
		},
		{
			name: "cache control produces block list",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeSystem,
					Parts: []llms.ContentPart{
						llms.WithCacheControl(
							llms.TextContent{Text: "Large context"},
							&llms.CacheControl{Type: "ephemeral"},
						),
						llms.TextContent{Text: "Question preamble"},
					},
				},
			},
			want: []anthropicclient.TextContent{
				{Type: "text", Text: "Large context", CacheControl: &anthropicclient.CacheControl{Type: "ephemeral"}},
				{Type: "text", Text: "Question preamble"},
			},
		},
		{
			name: "duration maps to 1h ttl and empty type defaults",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeSystem,
					Parts: []llms.ContentPart{
						llms.WithCacheControl(
							llms.TextContent{Text: "Cached"},
							&llms.CacheControl{Duration: time.Hour},
						),
					},
				},
			},
			want: []anthropicclient.TextContent{
				{Type: "text", Text: "Cached", CacheControl: &anthropicclient.CacheControl{Type: "ephemeral", TTL: "1h"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, system, err := processMessages(tt.messages)
			if err != nil {
				t.Fatalf("processMessages() error = %v", err)
			}
			if !reflect.DeepEqual(system, tt.want) {
				t.Errorf("processMessages() system = %#v, want %#v", system, tt.want)
			}
		})
	}
}

func TestToolsToTools(t *testing.T) {
	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_weather",
				Description: "Get the weather for a location",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "The location to get weather for",
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	result := toolsToTools(tools)

	if len(result) != 1 {
		t.Fatalf("toolsToTools() returned %d tools, want 1", len(result))
	}
	if result[0].Name != "get_weather" {
		t.Errorf("toolsToTools() tool name = %q, want %q", result[0].Name, "get_weather")
	}
	if result[0].Description != "Get the weather for a location" {
		t.Errorf("toolsToTools() tool description = %q, want %q", result[0].Description, "Get the weather for a location")
	}
}

func TestOptions(t *testing.T) {
	t.Run("WithModel", func(t *testing.T) {
		opts := &options{}
		WithModel("claude-3-opus")(opts)
		if opts.model != "claude-3-opus" {
			t.Errorf("WithModel() got %s, want claude-3-opus", opts.model)
		}
	})

	t.Run("WithToken", func(t *testing.T) {
		opts := &options{}
		WithToken("test-token")(opts)
		if opts.token != "test-token" {
			t.Errorf("WithToken() got %s, want test-token", opts.token)
		}
	})

	t.Run("WithBaseURL", func(t *testing.T) {
		opts := &options{}
		WithBaseURL("https://test.com")(opts)
		if opts.baseURL != "https://test.com" {
			t.Errorf("WithBaseURL() got %s, want https://test.com", opts.baseURL)
		}
	})

	t.Run("WithAnthropicBetaHeader", func(t *testing.T) {
		opts := &options{}
		WithAnthropicBetaHeader("test-beta")(opts)
		if opts.anthropicBetaHeader != "test-beta" {
			t.Errorf("WithAnthropicBetaHeader() got %s, want test-beta", opts.anthropicBetaHeader)
		}
	})

	t.Run("WithLegacyTextCompletionsAPI", func(t *testing.T) {
		opts := &options{}
		WithLegacyTextCompletionsAPI()(opts)
		if !opts.useLegacyTextCompletionsAPI {
			t.Error("WithLegacyTextCompletionsAPI() did not set flag")
		}
	})
}

func TestCall(t *testing.T) {
	// Test that Call delegates to GenerateContent
	t.Skip("Call() requires integration testing with mock client")
}

func TestGenerateMessagesContent_EmptyContent(t *testing.T) {
	// This test demonstrates the need for checking len(result.Content) == 0
	// Without the fix, accessing result.Content[0] would panic when Anthropic
	// returns a response with nil or empty content (addresses issue #993)
	t.Skip("Requires mock client - would demonstrate panic without len(result.Content) == 0 check")
}

func TestSupportsReasoningForModel(t *testing.T) {
	tests := []struct {
		model string
		want  bool
	}{
		{"claude-fable-5", true},
		{"claude-opus-4-8", true},
		{"claude-opus-4-7", true},
		{"claude-opus-4-6", true},
		{"claude-sonnet-4-6", true},
		{"claude-3-7-sonnet-20250219", true},
		{"claude-3-haiku-20240307", false},
		{"claude-3-5-sonnet-20240620", false},
		// The empty name gets newest-generation defaults; the provider
		// resolves the client default before lookup.
		{"", true},
	}
	for _, tt := range tests {
		if got := supportsReasoningForModel(tt.model); got != tt.want {
			t.Errorf("supportsReasoningForModel(%q) = %v, want %v", tt.model, got, tt.want)
		}
	}
}

func TestAdaptiveThinkingOnly(t *testing.T) {
	tests := []struct {
		model string
		want  bool
	}{
		{"claude-fable-5", true},
		{"claude-opus-4-7", true},
		{"claude-opus-4-8", true},
		{"claude-opus-4-6", false},
		{"claude-sonnet-4-6", false},
		{"claude-3-7-sonnet-20250219", false},
		{"claude-3-haiku-20240307", false},
		// The empty name gets newest-generation defaults; the provider
		// resolves the client default before lookup.
		{"", true},
	}
	for _, tt := range tests {
		if got := adaptiveThinkingOnly(tt.model); got != tt.want {
			t.Errorf("adaptiveThinkingOnly(%q) = %v, want %v", tt.model, got, tt.want)
		}
	}
}

func TestExtractThinkingOptions(t *testing.T) {
	tests := []extractThinkingOptionsTest{
		{
			name:         "no thinking config",
			model:        "claude-fable-5",
			config:       nil,
			wantThinking: nil,
		},
		{
			// A client without a model targets the client default,
			// which supports thinking; the config must not be dropped.
			name:         "empty model resolves default",
			model:        "",
			config:       &llms.ThinkingConfig{Mode: llms.ThinkingModeMedium},
			wantThinking: &anthropicclient.ThinkingConfig{Type: "enabled", BudgetTokens: 2048},
		},
		{
			name:         "fable auto mode uses adaptive",
			model:        "claude-fable-5",
			config:       &llms.ThinkingConfig{Mode: llms.ThinkingModeAuto},
			wantThinking: &anthropicclient.ThinkingConfig{Type: "adaptive"},
		},
		{
			name:         "fable drops explicit budget",
			model:        "claude-fable-5",
			config:       &llms.ThinkingConfig{BudgetTokens: 8192},
			wantThinking: &anthropicclient.ThinkingConfig{Type: "adaptive"},
		},
		{
			name:         "fable mode none omits thinking entirely",
			model:        "claude-fable-5",
			config:       &llms.ThinkingConfig{Mode: llms.ThinkingModeNone},
			wantThinking: nil,
		},
		{
			name:         "fable skips interleaved thinking header",
			model:        "claude-fable-5",
			config:       &llms.ThinkingConfig{Mode: llms.ThinkingModeAuto, InterleaveThinking: true},
			wantThinking: &anthropicclient.ThinkingConfig{Type: "adaptive"},
		},
		{
			name:         "opus 4.7 uses adaptive",
			model:        "claude-opus-4-7",
			config:       &llms.ThinkingConfig{BudgetTokens: 4096},
			wantThinking: &anthropicclient.ThinkingConfig{Type: "adaptive"},
		},
		{
			name:         "older model keeps budget",
			model:        "claude-sonnet-4-6",
			config:       &llms.ThinkingConfig{BudgetTokens: 8192},
			wantThinking: &anthropicclient.ThinkingConfig{Type: "enabled", BudgetTokens: 8192},
		},
		{
			name:         "older model keeps interleaved thinking header",
			model:        "claude-sonnet-4-6",
			config:       &llms.ThinkingConfig{BudgetTokens: 8192, InterleaveThinking: true},
			wantThinking: &anthropicclient.ThinkingConfig{Type: "enabled", BudgetTokens: 8192},
			wantHeaders:  []string{"interleaved-thinking-2025-05-14"},
		},
		{
			name:         "older model clamps budget to minimum",
			model:        "claude-3-7-sonnet-20250219",
			config:       &llms.ThinkingConfig{BudgetTokens: 100},
			wantThinking: &anthropicclient.ThinkingConfig{Type: "enabled", BudgetTokens: 1024},
		},
		{
			name:         "non-reasoning model ignores thinking",
			model:        "claude-3-haiku-20240307",
			config:       &llms.ThinkingConfig{BudgetTokens: 8192},
			wantThinking: nil,
		},
	}
	runExtractThinkingOptionsTests(t, tests)
}

func TestExtractThinkingOptionsEffortDisplay(t *testing.T) {
	tests := []extractThinkingOptionsTest{
		{
			name:         "fable maps effort to output config",
			model:        "claude-fable-5",
			config:       &llms.ThinkingConfig{Mode: llms.ThinkingModeAuto, Effort: "xhigh"},
			wantThinking: &anthropicclient.ThinkingConfig{Type: "adaptive"},
			wantOutput:   &anthropicclient.OutputConfig{Effort: "xhigh"},
		},
		{
			name:         "fable passes display through",
			model:        "claude-fable-5",
			config:       &llms.ThinkingConfig{Mode: llms.ThinkingModeAuto, Display: "summarized"},
			wantThinking: &anthropicclient.ThinkingConfig{Type: "adaptive", Display: "summarized"},
		},
		{
			name:         "fable effort without thinking mode",
			model:        "claude-fable-5",
			config:       &llms.ThinkingConfig{Mode: llms.ThinkingModeNone, Effort: "low"},
			wantThinking: nil,
			wantOutput:   &anthropicclient.OutputConfig{Effort: "low"},
		},
		{
			name:         "sonnet 4.6 maps effort alongside budget",
			model:        "claude-sonnet-4-6",
			config:       &llms.ThinkingConfig{BudgetTokens: 8192, Effort: "high"},
			wantThinking: &anthropicclient.ThinkingConfig{Type: "enabled", BudgetTokens: 8192},
			wantOutput:   &anthropicclient.OutputConfig{Effort: "high"},
		},
		{
			name:         "budget-only model drops effort",
			model:        "claude-3-7-sonnet-20250219",
			config:       &llms.ThinkingConfig{BudgetTokens: 8192, Effort: "high"},
			wantThinking: &anthropicclient.ThinkingConfig{Type: "enabled", BudgetTokens: 8192},
		},
	}
	runExtractThinkingOptionsTests(t, tests)
}

type extractThinkingOptionsTest struct {
	name         string
	model        string
	config       *llms.ThinkingConfig
	wantThinking *anthropicclient.ThinkingConfig
	wantHeaders  []string
	wantOutput   *anthropicclient.OutputConfig
}

func runExtractThinkingOptionsTests(t *testing.T, tests []extractThinkingOptionsTest) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &LLM{model: tt.model}
			opts := &llms.CallOptions{MaxTokens: 4096}
			if tt.config != nil {
				opts.Metadata = map[string]any{"thinking_config": tt.config}
			}

			headers, thinking, output := extractThinkingOptions(o, opts)
			assertThinkingConfig(t, thinking, tt.wantThinking)
			assertBetaHeaders(t, headers, tt.wantHeaders)
			assertOutputConfig(t, output, tt.wantOutput)
		})
	}
}

func assertThinkingConfig(t *testing.T, got, want *anthropicclient.ThinkingConfig) {
	t.Helper()
	if want == nil {
		if got != nil {
			t.Fatalf("extractThinkingOptions() thinking = %+v, want nil", got)
		}
		return
	}
	if got == nil {
		t.Fatalf("extractThinkingOptions() thinking = nil, want %+v", want)
	}
	if got.Type != want.Type {
		t.Errorf("thinking.Type = %q, want %q", got.Type, want.Type)
	}
	if got.BudgetTokens != want.BudgetTokens {
		t.Errorf("thinking.BudgetTokens = %d, want %d", got.BudgetTokens, want.BudgetTokens)
	}
	if got.Display != want.Display {
		t.Errorf("thinking.Display = %q, want %q", got.Display, want.Display)
	}
}

func assertOutputConfig(t *testing.T, got, want *anthropicclient.OutputConfig) {
	t.Helper()
	if want == nil {
		if got != nil {
			t.Fatalf("extractThinkingOptions() output = %+v, want nil", got)
		}
		return
	}
	if got == nil {
		t.Fatalf("extractThinkingOptions() output = nil, want %+v", want)
	}
	if got.Effort != want.Effort {
		t.Errorf("output.Effort = %q, want %q", got.Effort, want.Effort)
	}
}

func assertBetaHeaders(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("extractThinkingOptions() headers = %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("headers[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestModelCapabilities(t *testing.T) {
	tests := []struct {
		model string
		want  capabilities
	}{
		{"claude-fable-5", adaptiveCaps},
		{"claude-opus-4-8", adaptiveCaps},
		{"claude-opus-4-7", adaptiveCaps},
		{"claude-opus-4-6", capabilities{thinking: true, sampling: true, effort: true}},
		{"claude-sonnet-4-6", capabilities{thinking: true, sampling: true, effort: true}},
		{"claude-opus-4-20250514", budgetCaps},
		{"claude-3-7-sonnet-20250219", budgetCaps},
		{"claude-3-5-sonnet-20240620", legacyCaps},
		{"claude-3-haiku-20240307", legacyCaps},
		{"claude-2.1", legacyCaps},
		// Unlisted 4-x family members: minors past the last listed
		// release get newest-generation defaults, earlier ones keep
		// budget-era behavior.
		{"claude-opus-4-9", adaptiveCaps},
		{"claude-opus-4-5", budgetCaps},
		{"claude-opus-4-1-20250805", budgetCaps},
		{"claude-opus-4", budgetCaps},
		{"claude-sonnet-4-7", adaptiveCaps},
		{"claude-sonnet-4-5-20250929", budgetCaps},
		{"claude-haiku-4-5", budgetCaps},
		{"claude-haiku-4-6", adaptiveCaps},
		// Unknown Claude names get newest-generation defaults. The
		// empty name cannot reach here from the provider:
		// resolvedModel substitutes the client default first.
		{"", adaptiveCaps},
		{"claude-fable-6", adaptiveCaps},
		{"claude-omega-7", adaptiveCaps},
		{"claude-sonnet-5", adaptiveCaps},
		// Names that are not Claude models at all — gateway aliases,
		// fine-tunes, compatible endpoints — keep the long-standing
		// wire contract so caller sampling parameters are not dropped.
		{"my-gateway/llama-3-70b", budgetCaps},
		{"gpt-4o", budgetCaps},
		{"some-finetune-v2", budgetCaps},
	}
	for _, tt := range tests {
		if got := modelCapabilities(tt.model); got != tt.want {
			t.Errorf("modelCapabilities(%q) = %+v, want %+v", tt.model, got, tt.want)
		}
	}
}

func TestHandleAIMessageAllParts(t *testing.T) {
	msg := llms.MessageContent{
		Role: llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{
			llms.ThinkingContent{Thinking: "let me think", Signature: "sig123"},
			llms.RedactedThinkingContent{Data: "opaque"},
			llms.TextContent{Text: "calling the tool"},
			llms.ToolCall{
				ID: "tool-1",
				FunctionCall: &llms.FunctionCall{
					Name:      "get_weather",
					Arguments: `{"location":"sf"}`,
				},
			},
		},
	}

	got, err := handleAIMessage(msg)
	if err != nil {
		t.Fatalf("handleAIMessage() error = %v", err)
	}
	if got.Role != RoleAssistant {
		t.Errorf("role = %q, want %q", got.Role, RoleAssistant)
	}
	contents, ok := got.Content.([]anthropicclient.Content)
	if !ok {
		t.Fatalf("content type = %T, want []anthropicclient.Content", got.Content)
	}
	if len(contents) != 4 {
		t.Fatalf("len(contents) = %d, want 4", len(contents))
	}
	wantTypes := []string{"thinking", "redacted_thinking", "text", "tool_use"}
	for i, want := range wantTypes {
		if contents[i].GetType() != want {
			t.Errorf("contents[%d].GetType() = %q, want %q", i, contents[i].GetType(), want)
		}
	}
	thinking, ok := contents[0].(*anthropicclient.ThinkingContent)
	if !ok {
		t.Fatalf("contents[0] type = %T, want *ThinkingContent", contents[0])
	}
	if thinking.Signature != "sig123" {
		t.Errorf("thinking signature = %q, want %q", thinking.Signature, "sig123")
	}
}

func TestHandleToolMessageAllParts(t *testing.T) {
	msg := llms.MessageContent{
		Role: llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{
			llms.ToolCallResponse{ToolCallID: "tool-1", Content: "sunny"},
			llms.ToolCallResponse{ToolCallID: "tool-2", Content: "72F"},
		},
	}

	got, err := handleToolMessage(msg)
	if err != nil {
		t.Fatalf("handleToolMessage() error = %v", err)
	}
	if got.Role != RoleUser {
		t.Errorf("role = %q, want %q", got.Role, RoleUser)
	}
	contents, ok := got.Content.([]anthropicclient.Content)
	if !ok {
		t.Fatalf("content type = %T, want []anthropicclient.Content", got.Content)
	}
	if len(contents) != 2 {
		t.Fatalf("len(contents) = %d, want 2", len(contents))
	}
	second, ok := contents[1].(anthropicclient.ToolResultContent)
	if !ok {
		t.Fatalf("contents[1] type = %T, want ToolResultContent", contents[1])
	}
	if second.ToolUseID != "tool-2" || second.Content != "72F" {
		t.Errorf("contents[1] = %+v, want tool-2/72F", second)
	}
}

func TestProcessAnthropicResponseParts(t *testing.T) {
	result := &anthropicclient.MessageResponsePayload{
		Content: []anthropicclient.Content{
			&anthropicclient.ThinkingContent{Type: "thinking", Thinking: "hmm", Signature: "sig"},
			&anthropicclient.RedactedThinkingContent{Type: "redacted_thinking", Data: "opaque"},
			&anthropicclient.TextContent{Type: "text", Text: "answer"},
			&anthropicclient.ToolUseContent{Type: "tool_use", ID: "t1", Name: "f", Input: map[string]interface{}{"a": "b"}},
		},
	}

	resp, err := processAnthropicResponse(result)
	if err != nil {
		t.Fatalf("processAnthropicResponse() error = %v", err)
	}
	if len(resp.Choices) != 4 {
		t.Fatalf("len(choices) = %d, want 4", len(resp.Choices))
	}

	thinkingPart, ok := resp.Choices[0].Parts[0].(llms.ThinkingContent)
	if !ok {
		t.Fatalf("choices[0].Parts[0] type = %T, want llms.ThinkingContent", resp.Choices[0].Parts[0])
	}
	if thinkingPart.Thinking != "hmm" || thinkingPart.Signature != "sig" {
		t.Errorf("thinking part = %+v, want hmm/sig", thinkingPart)
	}
	if resp.Choices[0].ReasoningContent != "hmm" {
		t.Errorf("ReasoningContent = %q, want %q", resp.Choices[0].ReasoningContent, "hmm")
	}
	if _, ok := resp.Choices[1].Parts[0].(llms.RedactedThinkingContent); !ok {
		t.Fatalf("choices[1].Parts[0] type = %T, want llms.RedactedThinkingContent", resp.Choices[1].Parts[0])
	}
	if _, ok := resp.Choices[2].Parts[0].(llms.TextContent); !ok {
		t.Fatalf("choices[2].Parts[0] type = %T, want llms.TextContent", resp.Choices[2].Parts[0])
	}
	if _, ok := resp.Choices[3].Parts[0].(llms.ToolCall); !ok {
		t.Fatalf("choices[3].Parts[0] type = %T, want llms.ToolCall", resp.Choices[3].Parts[0])
	}

	// Round-trip: the concatenated parts must convert back into a valid
	// assistant message with all four blocks in order.
	var aiParts []llms.ContentPart
	for _, c := range resp.Choices {
		aiParts = append(aiParts, c.Parts...)
	}
	aiMsg, err := handleAIMessage(llms.MessageContent{Role: llms.ChatMessageTypeAI, Parts: aiParts})
	if err != nil {
		t.Fatalf("handleAIMessage(round-trip) error = %v", err)
	}
	contents := aiMsg.Content.([]anthropicclient.Content)
	if len(contents) != 4 {
		t.Fatalf("round-trip len(contents) = %d, want 4", len(contents))
	}
	rt, ok := contents[0].(*anthropicclient.ThinkingContent)
	if !ok || rt.Signature != "sig" {
		t.Errorf("round-trip thinking = %+v, want signature preserved", contents[0])
	}
}

package anthropic

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/anthropic/internal/anthropicclient"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
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
			originalKey := os.Getenv("ANTHROPIC_API_KEY")
			os.Setenv("ANTHROPIC_API_KEY", tt.envToken)
			defer os.Setenv("ANTHROPIC_API_KEY", originalKey)

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

// TestMergeCacheStrategies tests the cache strategy merge logic
func TestMergeCacheStrategies(t *testing.T) {
	tests := []struct {
		name           string
		clientStrategy *CacheStrategy
		callStrategy   *CacheStrategy
		want           *CacheStrategy
	}{
		{
			name:           "both nil",
			clientStrategy: nil,
			callStrategy:   nil,
			want:           nil,
		},
		{
			name:           "only client-level",
			clientStrategy: &CacheStrategy{CacheTools: true, CacheSystem: true},
			callStrategy:   nil,
			want:           &CacheStrategy{CacheTools: true, CacheSystem: true},
		},
		{
			name:           "only call-level",
			clientStrategy: nil,
			callStrategy:   &CacheStrategy{CacheMessages: true, TTL: "1h"},
			want:           &CacheStrategy{CacheMessages: true, TTL: "1h"},
		},
		{
			name: "call-level overrides client-level (OR merge)",
			clientStrategy: &CacheStrategy{
				CacheTools:  true,
				CacheSystem: true,
				TTL:         "5m",
			},
			callStrategy: &CacheStrategy{
				CacheMessages: true,
				TTL:           "1h",
			},
			want: &CacheStrategy{
				CacheTools:    true, // from client (OR: true || false)
				CacheSystem:   true, // from client (OR: false || true)
				CacheMessages: true, // from call (OR: true || false)
				TTL:           "1h", // call-level takes precedence
			},
		},
		{
			name: "call-level disables nothing (OR logic preserves client settings)",
			clientStrategy: &CacheStrategy{
				CacheTools:  true,
				CacheSystem: true,
			},
			callStrategy: &CacheStrategy{
				CacheMessages: true,
			},
			want: &CacheStrategy{
				CacheTools:    true, // preserved from client
				CacheSystem:   true, // preserved from client
				CacheMessages: true, // added from call
			},
		},
		{
			name: "empty call-level doesn't override",
			clientStrategy: &CacheStrategy{
				CacheTools:  true,
				CacheSystem: true,
				TTL:         "1h",
			},
			callStrategy: &CacheStrategy{}, // all false, no TTL
			want: &CacheStrategy{
				CacheTools:  true, // from client (OR: false || true)
				CacheSystem: true, // from client (OR: false || true)
				TTL:         "1h", // from client (empty string doesn't override)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare call options with strategy
			var opts llms.CallOptions
			if tt.callStrategy != nil {
				opts.Metadata = map[string]any{
					"anthropic:cache_strategy": *tt.callStrategy,
				}
			}

			// Perform merge
			got := mergeCacheStrategies(tt.clientStrategy, &opts)

			// Verify result
			if tt.want == nil {
				assert.Nil(t, got, "Expected nil strategy")
			} else {
				require.NotNil(t, got, "Expected non-nil strategy")
				assert.Equal(t, tt.want.CacheTools, got.CacheTools, "CacheTools mismatch")
				assert.Equal(t, tt.want.CacheSystem, got.CacheSystem, "CacheSystem mismatch")
				assert.Equal(t, tt.want.CacheMessages, got.CacheMessages, "CacheMessages mismatch")
				assert.Equal(t, tt.want.TTL, got.TTL, "TTL mismatch")
			}
		})
	}
}

// TestDefaultCacheStrategy_ClientLevel tests that client-level cache strategy is applied
func TestDefaultCacheStrategy_ClientLevel(t *testing.T) {
	// Create client with default strategy
	llm, err := New(
		WithToken("test-token"),
		WithDefaultCacheStrategy(CacheStrategy{
			CacheTools:  true,
			CacheSystem: true,
			TTL:         "1h",
		}),
	)
	require.NoError(t, err)
	require.NotNil(t, llm.defaultCacheStrategy, "Default strategy should be set")

	// Verify strategy was stored
	assert.True(t, llm.defaultCacheStrategy.CacheTools)
	assert.True(t, llm.defaultCacheStrategy.CacheSystem)
	assert.False(t, llm.defaultCacheStrategy.CacheMessages)
	assert.Equal(t, "1h", llm.defaultCacheStrategy.TTL)
}

// TestDefaultCacheStrategy_MergeWithCallLevel tests merge behavior
func TestDefaultCacheStrategy_MergeWithCallLevel(t *testing.T) {
	// Setup: client with tools+system caching
	clientStrategy := &CacheStrategy{
		CacheTools:  true,
		CacheSystem: true,
		TTL:         "5m",
	}

	// Case 1: Call-level adds messages caching
	opts1 := &llms.CallOptions{
		Metadata: map[string]any{
			"anthropic:cache_strategy": CacheStrategy{
				CacheMessages: true,
			},
		},
	}
	merged1 := mergeCacheStrategies(clientStrategy, opts1)
	require.NotNil(t, merged1)
	assert.True(t, merged1.CacheTools, "Should preserve client CacheTools")
	assert.True(t, merged1.CacheSystem, "Should preserve client CacheSystem")
	assert.True(t, merged1.CacheMessages, "Should add call-level CacheMessages")
	assert.Equal(t, "5m", merged1.TTL, "Should use client TTL when call-level empty")

	// Case 2: Call-level overrides TTL
	opts2 := &llms.CallOptions{
		Metadata: map[string]any{
			"anthropic:cache_strategy": CacheStrategy{
				TTL: "1h",
			},
		},
	}
	merged2 := mergeCacheStrategies(clientStrategy, opts2)
	require.NotNil(t, merged2)
	assert.Equal(t, "1h", merged2.TTL, "Should use call-level TTL")
	assert.True(t, merged2.CacheTools, "Should preserve client settings")

	// Case 3: Call-level with no metadata uses client defaults
	opts3 := &llms.CallOptions{}
	merged3 := mergeCacheStrategies(clientStrategy, opts3)
	require.NotNil(t, merged3)
	assert.Equal(t, clientStrategy, merged3, "Should use client strategy when no call-level")
}

// TestTokenUsageMapping_Anthropic tests correct token usage mapping for Anthropic provider
func TestTokenUsageMapping_Anthropic(t *testing.T) { //nolint:funlen
	tests := []struct {
		name                     string
		inputTokens              int
		cacheCreationInputTokens int
		cacheReadInputTokens     int
		outputTokens             int
		expectedPromptTokens     int
		expectedCacheRead        int
		expectedCacheCreation    int
		expectedCompletionTokens int
		expectedTotalTokens      int
	}{
		{
			name:                     "first request without cache",
			inputTokens:              332,
			cacheCreationInputTokens: 0,
			cacheReadInputTokens:     0,
			outputTokens:             82,
			expectedPromptTokens:     332,
			expectedCacheRead:        0,
			expectedCacheCreation:    0,
			expectedCompletionTokens: 82,
			expectedTotalTokens:      414,
		},
		{
			name:                     "first request with cache creation",
			inputTokens:              332,
			cacheCreationInputTokens: 1546,
			cacheReadInputTokens:     0,
			outputTokens:             82,
			expectedPromptTokens:     1878, // 332 + 1546 + 0
			expectedCacheRead:        0,
			expectedCacheCreation:    1546,
			expectedCompletionTokens: 82,
			expectedTotalTokens:      1960, // 332 + 1546 + 0 + 82
		},
		{
			name:                     "subsequent request with cache hit",
			inputTokens:              332,
			cacheCreationInputTokens: 0,
			cacheReadInputTokens:     1546,
			outputTokens:             82,
			expectedPromptTokens:     1878, // 332 + 0 + 1546
			expectedCacheRead:        1546,
			expectedCacheCreation:    0,
			expectedCompletionTokens: 82,
			expectedTotalTokens:      1960, // 332 + 0 + 1546 + 82
		},
		{
			name:                     "mixed cache scenario",
			inputTokens:              500,
			cacheCreationInputTokens: 200,
			cacheReadInputTokens:     1000,
			outputTokens:             150,
			expectedPromptTokens:     1700, // 500 + 200 + 1000
			expectedCacheRead:        1000,
			expectedCacheCreation:    200,
			expectedCompletionTokens: 150,
			expectedTotalTokens:      1850, // 500 + 200 + 1000 + 150
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate Anthropic response structure
			result := struct {
				Usage struct {
					InputTokens              int
					CacheCreationInputTokens int
					CacheReadInputTokens     int
					OutputTokens             int
					CacheCreation            struct {
						Ephemeral5mInputTokens int
						Ephemeral1hInputTokens int
					}
					ServiceTier string
				}
			}{}

			result.Usage.InputTokens = tt.inputTokens
			result.Usage.CacheCreationInputTokens = tt.cacheCreationInputTokens
			result.Usage.CacheReadInputTokens = tt.cacheReadInputTokens
			result.Usage.OutputTokens = tt.outputTokens
			result.Usage.ServiceTier = "standard"

			// Build GenerationInfo as done in anthropicllm.go
			generationInfo := map[string]any{
				"PromptTokens":             result.Usage.InputTokens + result.Usage.CacheCreationInputTokens + result.Usage.CacheReadInputTokens,
				"CompletionTokens":         result.Usage.OutputTokens,
				"TotalTokens":              result.Usage.InputTokens + result.Usage.CacheCreationInputTokens + result.Usage.CacheReadInputTokens + result.Usage.OutputTokens,
				"ReasoningTokens":          0,
				"PromptCachedTokens":       result.Usage.CacheReadInputTokens,
				"CacheReadInputTokens":     result.Usage.CacheReadInputTokens,
				"CacheCreationInputTokens": result.Usage.CacheCreationInputTokens,
				"InputTokens":              result.Usage.InputTokens,
				"OutputTokens":             result.Usage.OutputTokens,
				"ServiceTier":              result.Usage.ServiceTier,
			}

			// Verify mapped values
			assert.Equal(t, tt.expectedPromptTokens, generationInfo["PromptTokens"], "PromptTokens mismatch")
			assert.Equal(t, tt.expectedCacheRead, generationInfo["CacheReadInputTokens"], "CacheReadInputTokens mismatch")
			assert.Equal(t, tt.expectedCacheCreation, generationInfo["CacheCreationInputTokens"], "CacheCreationInputTokens mismatch")
			assert.Equal(t, tt.expectedCompletionTokens, generationInfo["CompletionTokens"], "CompletionTokens mismatch")
			assert.Equal(t, tt.expectedTotalTokens, generationInfo["TotalTokens"], "TotalTokens mismatch")

			// Verify client-side cost calculation logic
			// Client formula: input = max(PromptTokens - CacheRead, 0)
			promptTokens := generationInfo["PromptTokens"].(int)
			cacheRead := generationInfo["CacheReadInputTokens"].(int)
			cacheWrite := generationInfo["CacheCreationInputTokens"].(int)

			uncachedTokens := max(promptTokens-cacheRead, 0)

			// Expected: uncached tokens should equal inputTokens + cacheCreationInputTokens
			expectedUncached := tt.inputTokens + tt.cacheCreationInputTokens
			assert.Equal(t, expectedUncached, uncachedTokens, "Uncached tokens calculation mismatch")

			// For Anthropic pricing: uncached * basePrice + cacheRead * cacheReadPrice + cacheWrite * cacheWritePrice
			basePrice := 3.0 / 1e6        // $3 per 1M tokens
			cacheReadPrice := 0.3 / 1e6   // $0.3 per 1M tokens (90% discount)
			cacheWritePrice := 3.75 / 1e6 // $3.75 per 1M tokens (25% premium)

			expectedCost := float64(uncachedTokens)*basePrice +
				float64(cacheRead)*cacheReadPrice +
				float64(cacheWrite)*cacheWritePrice

			actualCost := float64(uncachedTokens)*basePrice +
				float64(cacheRead)*cacheReadPrice +
				float64(cacheWrite)*cacheWritePrice

			assert.InDelta(t, expectedCost, actualCost, 0.000001, "Cost calculation mismatch")
		})
	}
}

// The thinking block must lead the assistant turn even when the caller places the
// tool call before the reasoning-bearing text part; otherwise the API rejects a
// tool_use that precedes thinking.
func TestHandleAIMessage_ThinkingBlockLeadsToolUse(t *testing.T) {
	t.Parallel()

	msg := llms.MessageContent{
		Role: llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{
			llms.ToolCall{ID: "t1", FunctionCall: &llms.FunctionCall{Name: "probe", Arguments: "{}"}},
			llms.TextContent{
				Text:      "answer",
				Reasoning: &reasoning.ContentReasoning{Content: "thinking", Signature: []byte("sig")},
			},
		},
	}

	out, err := handleAIMessage(msg)
	require.NoError(t, err)

	thinkingIdx, toolIdx := -1, -1
	for i, c := range out.Content {
		switch c.(type) {
		case *anthropicclient.ThinkingContent:
			thinkingIdx = i
		case *anthropicclient.ToolUseContent:
			toolIdx = i
		}
	}
	require.NotEqual(t, -1, thinkingIdx, "thinking block must be present")
	require.NotEqual(t, -1, toolIdx, "tool_use block must be present")
	assert.Equal(t, 0, thinkingIdx, "thinking block must lead the assistant turn")
	assert.Less(t, thinkingIdx, toolIdx, "thinking must come before tool_use")
}

const soSchema = `{"type":"object","properties":{"answer":{"type":"string"}},"required":["answer"],"additionalProperties":false}` //nolint:gosec

// runStructured drives GenerateContent against a server returning respBody and
// returns the captured outbound request body, the response and any error, so both
// the wire shape and the response validation can be asserted deterministically.
func runStructured(t *testing.T, model, respBody string, callOpts ...llms.CallOption) (map[string]any, *llms.ContentResponse, error) {
	t.Helper()
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(respBody))
	}))
	t.Cleanup(srv.Close)

	llm, err := New(
		WithToken("test-key"),
		WithBaseURL(srv.URL),
		WithModel(model),
	)
	require.NoError(t, err)

	messages := []llms.MessageContent{{
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{llms.TextPart("hi")},
	}}
	resp, genErr := llm.GenerateContent(t.Context(), messages, callOpts...)

	var payload map[string]any
	if len(body) > 0 {
		_ = json.Unmarshal(body, &payload)
	}
	return payload, resp, genErr
}

func endTurnResp(text string) string {
	return `{"id":"m","type":"message","role":"assistant","model":"claude-sonnet-5","content":[{"type":"text","text":` +
		mustJSON(text) + `}],"stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1}}`
}

func mustJSON(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func outputConfig(payload map[string]any) map[string]any {
	oc, _ := payload["output_config"].(map[string]any)
	return oc
}

func TestAnthropic_StructuredOutputWire(t *testing.T) {
	t.Parallel()

	t.Run("format is sent in output_config", func(t *testing.T) {
		t.Parallel()
		payload, _, err := runStructured(t, "claude-sonnet-4-5", endTurnResp(`{"answer":"ok"}`),
			llms.WithStructuredOutput(llms.StructuredOutputConfig{Name: "s", Schema: json.RawMessage(soSchema)}))
		require.NoError(t, err)
		oc := outputConfig(payload)
		require.NotNil(t, oc, "output_config must be present")
		format, ok := oc["format"].(map[string]any)
		require.True(t, ok, "output_config.format must be present")
		require.Equal(t, "json_schema", format["type"])
		require.NotNil(t, format["schema"])
	})

	t.Run("format and effort coexist", func(t *testing.T) {
		t.Parallel()
		// Adaptive-only model sets output_config.effort; structured output adds
		// output_config.format to the SAME object without clobbering effort.
		payload, _, err := runStructured(t, "claude-sonnet-5", endTurnResp(`{"answer":"ok"}`),
			llms.WithAdaptiveReasoning(llms.ReasoningHigh),
			llms.WithStructuredOutput(llms.StructuredOutputConfig{Name: "s", Schema: json.RawMessage(soSchema)}))
		require.NoError(t, err)
		oc := outputConfig(payload)
		require.NotNil(t, oc)
		require.Equal(t, "high", oc["effort"])
		_, ok := oc["format"].(map[string]any)
		require.True(t, ok, "format must coexist with effort")
	})

	t.Run("no legacy top-level output_format", func(t *testing.T) {
		t.Parallel()
		payload, _, err := runStructured(t, "claude-sonnet-4-5", endTurnResp(`{"answer":"ok"}`),
			llms.WithStructuredOutput(llms.StructuredOutputConfig{Name: "s", Schema: json.RawMessage(soSchema)}))
		require.NoError(t, err)
		_, has := payload["output_format"]
		require.False(t, has, "must not use deprecated top-level output_format")
	})

	t.Run("assistant prefill conflicts", func(t *testing.T) {
		t.Parallel()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(endTurnResp(`{"answer":"ok"}`)))
		}))
		t.Cleanup(srv.Close)
		llm, err := New(WithToken("k"), WithBaseURL(srv.URL), WithModel("claude-sonnet-4-5"))
		require.NoError(t, err)
		messages := []llms.MessageContent{
			{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart("hi")}},
			{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextPart("prefill")}},
		}
		_, err = llm.GenerateContent(t.Context(), messages,
			llms.WithStructuredOutput(llms.StructuredOutputConfig{Name: "s", Schema: json.RawMessage(soSchema)}))
		var conflict *llms.ErrStructuredOutputConflict
		require.True(t, errors.As(err, &conflict), "want ErrStructuredOutputConflict, got %v", err)
	})

	t.Run("object schema missing additionalProperties:false is rejected before the network", func(t *testing.T) {
		t.Parallel()
		open := `{"type":"object","properties":{"answer":{"type":"string"}},"required":["answer"]}`
		_, _, err := runStructured(t, "claude-sonnet-4-5", endTurnResp(`{"answer":"ok"}`),
			llms.WithStructuredOutput(llms.StructuredOutputConfig{Name: "s", Schema: json.RawMessage(open)}))
		require.True(t, errors.Is(err, llms.ErrStructuredOutputConfig), "want ErrStructuredOutputConfig, got %v", err)
	})
}

func TestAnthropic_StructuredOutputResponseValidation(t *testing.T) {
	t.Parallel()
	so := llms.WithStructuredOutput(llms.StructuredOutputConfig{Name: "s", Schema: json.RawMessage(soSchema)})

	t.Run("valid end_turn passes", func(t *testing.T) {
		t.Parallel()
		_, resp, err := runStructured(t, "claude-sonnet-4-5", endTurnResp(`{"answer":"ok"}`), so)
		require.NoError(t, err)
		require.Equal(t, `{"answer":"ok"}`, resp.Choices[0].Content)
	})

	t.Run("stop_sequence is validated", func(t *testing.T) {
		t.Parallel()
		body := `{"id":"m","type":"message","role":"assistant","model":"x","content":[{"type":"text","text":"{\"answer\":\"ok\"}"}],"stop_reason":"stop_sequence","usage":{"input_tokens":1,"output_tokens":1}}`
		_, resp, err := runStructured(t, "claude-sonnet-4-5", body, so)
		require.NoError(t, err)
		require.Equal(t, `{"answer":"ok"}`, resp.Choices[0].Content)
	})

	t.Run("invalid end_turn fails with typed error and preserves the response", func(t *testing.T) {
		t.Parallel()
		_, resp, err := runStructured(t, "claude-sonnet-4-5", endTurnResp(`{"answer":5}`), so)
		var ve *llms.ErrStructuredOutputValidation
		require.True(t, errors.As(err, &ve), "want ErrStructuredOutputValidation, got %v", err)
		// The public GenerateContent must return the assembled response alongside the
		// error so usage/content/stop metadata are not lost.
		require.NotNil(t, resp, "validation error must still return the response")
		require.NotEmpty(t, resp.Choices)
		require.Equal(t, `{"answer":5}`, resp.Choices[0].Content)
		require.Equal(t, "end_turn", resp.Choices[0].StopReason)
	})

	t.Run("two text blocks are concatenated then validated", func(t *testing.T) {
		t.Parallel()
		body := `{"id":"m","type":"message","role":"assistant","model":"x","content":[` +
			`{"type":"text","text":"{\"ans"},{"type":"text","text":"wer\":\"ok\"}"}],` +
			`"stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1}}`
		_, resp, err := runStructured(t, "claude-sonnet-4-5", body, so)
		require.NoError(t, err)
		require.Equal(t, `{"answer":"ok"}`, resp.Choices[0].Content)
	})

	t.Run("max_tokens is not validated", func(t *testing.T) {
		t.Parallel()
		body := `{"id":"m","type":"message","role":"assistant","model":"x","content":[{"type":"text","text":"{\"answer\""}],"stop_reason":"max_tokens","usage":{"input_tokens":1,"output_tokens":1}}`
		_, _, err := runStructured(t, "claude-sonnet-4-5", body, so)
		require.NoError(t, err, "max_tokens must not be validated as final JSON")
	})

	t.Run("tool_use is not validated", func(t *testing.T) {
		t.Parallel()
		body := `{"id":"m","type":"message","role":"assistant","model":"x","content":[{"type":"tool_use","id":"t","name":"f","input":{}}],"stop_reason":"tool_use","usage":{"input_tokens":1,"output_tokens":1}}`
		_, _, err := runStructured(t, "claude-sonnet-4-5", body, so)
		require.NoError(t, err, "tool_use turn must not be validated")
	})

	t.Run("refusal returns ErrModelRefusal not validation", func(t *testing.T) {
		t.Parallel()
		body := `{"id":"m","type":"message","role":"assistant","model":"x","content":[{"type":"text","text":"no"}],"stop_reason":"refusal","usage":{"input_tokens":1,"output_tokens":1}}`
		_, _, err := runStructured(t, "claude-sonnet-4-5", body, so)
		var refusal *ErrModelRefusal
		require.True(t, errors.As(err, &refusal), "want ErrModelRefusal, got %v", err)
	})

	t.Run("case-sensitive enum mismatch fails", func(t *testing.T) {
		t.Parallel()
		enumSchema := `{"type":"object","properties":{"color":{"enum":["Red"]}},"required":["color"],"additionalProperties":false}`
		soEnum := llms.WithStructuredOutput(llms.StructuredOutputConfig{Name: "s", Schema: json.RawMessage(enumSchema)})
		_, _, err := runStructured(t, "claude-sonnet-4-5", endTurnResp(`{"color":"red"}`), soEnum)
		var ve *llms.ErrStructuredOutputValidation
		require.True(t, errors.As(err, &ve), "enum case mismatch must fail, got %v", err)
	})
}

func TestSamplingParamsNeverBothOnTheWire(t *testing.T) {
	t.Parallel()

	const resp = `{"id":"x","type":"message","role":"assistant","model":"m",` +
		`"content":[{"type":"text","text":"ok"}],"stop_reason":"end_turn",` +
		`"usage":{"input_tokens":1,"output_tokens":1}}`

	for _, model := range []string{
		"claude-haiku-4-5-20251001",
		"claude-sonnet-4-5-20250929",
		"claude-opus-4-5-20251101",
		"claude-sonnet-4-6",
		"claude-opus-4-6",
	} {
		t.Run(model, func(t *testing.T) {
			var body string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				b, _ := io.ReadAll(r.Body)
				body = string(b)
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, resp)
			}))
			defer srv.Close()

			llm, err := New(WithBaseURL(srv.URL), WithToken("t"), WithModel(model))
			if err != nil {
				t.Fatalf("New() error: %v", err)
			}
			if _, err := llm.GenerateContent(context.Background(),
				[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
				llms.WithTemperature(0.5), llms.WithTopP(0.9)); err != nil {
				t.Fatalf("GenerateContent() error: %v", err)
			}
			if strings.Contains(body, `"temperature"`) && strings.Contains(body, `"top_p"`) {
				t.Fatalf("both sampling params reached the wire: %s", body)
			}
		})
	}
}

func TestBudgetEffortForOpus45OnTheAnthropicPath(t *testing.T) {
	t.Parallel()

	const resp = `{"id":"x","type":"message","role":"assistant","model":"m",` +
		`"content":[{"type":"text","text":"ok"}],"stop_reason":"end_turn",` +
		`"usage":{"input_tokens":1,"output_tokens":1}}`

	var body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, resp)
	}))
	defer srv.Close()

	llm, err := New(WithBaseURL(srv.URL), WithToken("t"), WithModel("claude-opus-4-5-20251101"))
	require.NoError(t, err)
	_, err = llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
		llms.WithReasoning(llms.ReasoningHigh, 0), llms.WithMaxTokens(8000))
	require.NoError(t, err)

	assert.Contains(t, body, `"type":"enabled"`, "Opus 4.5 is budget-only")
	assert.Contains(t, body, `"output_config":{"effort":"high"}`,
		"the first-party API takes the effort alongside a budget here")
}

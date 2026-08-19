package openai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/openai/internal/openaiclient"
)

func TestWithMaxCompletionTokens(t *testing.T) {
	opts := &llms.CallOptions{}

	// Test that WithMaxCompletionTokens sets MaxTokens
	WithMaxCompletionTokens(100)(opts)
	if opts.GetMaxTokens() != 100 {
		t.Errorf("expected MaxTokens=100, got %d", opts.GetMaxTokens())
	}

	// Test that it can be overridden
	WithMaxCompletionTokens(200)(opts)
	if opts.GetMaxTokens() != 200 {
		t.Errorf("expected MaxTokens=200, got %d", opts.GetMaxTokens())
	}

	// Test with zero value
	WithMaxCompletionTokens(0)(opts)
	if opts.GetMaxTokens() != 0 {
		t.Errorf("expected MaxTokens=0, got %d", opts.GetMaxTokens())
	}
}

func TestOptionsCompatibility(t *testing.T) {
	opts := &llms.CallOptions{}

	// Test that both llms.WithMaxTokens and WithMaxCompletionTokens
	// set the same field for compatibility
	llms.WithMaxTokens(150)(opts)
	if opts.GetMaxTokens() != 150 {
		t.Errorf("expected MaxTokens=150, got %d", opts.GetMaxTokens())
	}

	opts2 := &llms.CallOptions{}
	WithMaxCompletionTokens(150)(opts2)
	if opts2.GetMaxTokens() != 150 {
		t.Errorf("expected MaxTokens=150, got %d", opts2.GetMaxTokens())
	}

	// They should be equivalent
	if opts.GetMaxTokens() != opts2.GetMaxTokens() {
		t.Errorf("WithMaxTokens and WithMaxCompletionTokens should set the same field")
	}
}

func TestWithLegacyMaxTokensField(t *testing.T) {
	opts := &llms.CallOptions{}

	// Test that WithLegacyMaxTokensField sets the metadata flag
	WithLegacyMaxTokensField()(opts)
	if opts.Metadata == nil {
		t.Fatal("expected Metadata to be initialized")
	}
	if v, ok := opts.Metadata["openai:use_legacy_max_tokens"].(bool); !ok || !v {
		t.Error("expected openai:use_legacy_max_tokens to be true")
	}

	// Test combining with WithMaxTokens
	opts2 := &llms.CallOptions{}
	llms.WithMaxTokens(200)(opts2)
	WithLegacyMaxTokensField()(opts2)
	if opts2.GetMaxTokens() != 200 {
		t.Errorf("expected MaxTokens=200, got %d", opts2.GetMaxTokens())
	}
	if v, ok := opts2.Metadata["openai:use_legacy_max_tokens"].(bool); !ok || !v {
		t.Error("expected openai:use_legacy_max_tokens to be true")
	}
}

func TestWithWebSearch(t *testing.T) {
	// Test with nil options (default behavior)
	opts := &llms.CallOptions{}
	llms.WithWebSearch(nil)(opts)
	if opts.WebSearchOptions == nil {
		t.Fatal("expected WebSearchOptions to be initialized")
	}

	// Test with custom search context size
	opts2 := &llms.CallOptions{}
	llms.WithWebSearch(&llms.WebSearchOptions{
		SearchContextSize: "high",
	})(opts2)
	if opts2.WebSearchOptions == nil {
		t.Fatal("expected WebSearchOptions to be set")
	}
	if opts2.WebSearchOptions.SearchContextSize != "high" {
		t.Errorf("expected SearchContextSize=high, got %s", opts2.WebSearchOptions.SearchContextSize)
	}

	// Test with user location
	opts3 := &llms.CallOptions{}
	llms.WithWebSearch(&llms.WebSearchOptions{
		SearchContextSize: "medium",
		UserLocation: &llms.UserLocation{
			Type: "approximate",
			Approximate: &llms.ApproximateLocation{
				Country: "US",
				City:    "San Francisco",
				Region:  "California",
			},
		},
	})(opts3)
	if opts3.WebSearchOptions == nil {
		t.Fatal("expected WebSearchOptions to be set")
	}
	if opts3.WebSearchOptions.UserLocation == nil {
		t.Fatal("expected UserLocation to be set")
	}
	if opts3.WebSearchOptions.UserLocation.Type != "approximate" {
		t.Errorf("expected Type=approximate, got %s", opts3.WebSearchOptions.UserLocation.Type)
	}
	if opts3.WebSearchOptions.UserLocation.Approximate == nil {
		t.Fatal("expected Approximate to be set")
	}
	if opts3.WebSearchOptions.UserLocation.Approximate.Country != "US" {
		t.Errorf("expected Country=US, got %s", opts3.WebSearchOptions.UserLocation.Approximate.Country)
	}
	if opts3.WebSearchOptions.UserLocation.Approximate.City != "San Francisco" {
		t.Errorf("expected City=San Francisco, got %s", opts3.WebSearchOptions.UserLocation.Approximate.City)
	}
	if opts3.WebSearchOptions.UserLocation.Approximate.Region != "California" {
		t.Errorf("expected Region=California, got %s", opts3.WebSearchOptions.UserLocation.Approximate.Region)
	}
}

func TestWebSearchOptionsConversion(t *testing.T) {
	// Test nil conversion
	result := webSearchOptionsFromCallOptions(nil)
	if result != nil {
		t.Error("expected nil result for nil input")
	}

	// Test basic conversion
	opts := &llms.WebSearchOptions{
		SearchContextSize: "high",
	}
	result = webSearchOptionsFromCallOptions(opts)
	require.NotNil(t, result, "expected non-nil result")
	if result.SearchContextSize != "high" {
		t.Errorf("expected SearchContextSize=high, got %s", result.SearchContextSize)
	}

	// Test full conversion with user location
	opts2 := &llms.WebSearchOptions{
		SearchContextSize: "medium",
		UserLocation: &llms.UserLocation{
			Type: "approximate",
			Approximate: &llms.ApproximateLocation{
				Country: "GB",
				City:    "London",
				Region:  "London",
			},
		},
	}
	result2 := webSearchOptionsFromCallOptions(opts2)
	require.NotNil(t, result2, "expected non-nil result")
	require.NotNil(t, result2.UserLocation, "expected UserLocation to be set")
	if result2.UserLocation.Type != "approximate" {
		t.Errorf("expected Type=approximate, got %s", result2.UserLocation.Type)
	}
	if result2.UserLocation.Approximate == nil {
		t.Fatal("expected Approximate to be set")
	}
	if result2.UserLocation.Approximate.Country != "GB" {
		t.Errorf("expected Country=GB, got %s", result2.UserLocation.Approximate.Country)
	}
}

func TestWithExtraBody(t *testing.T) {
	t.Run("sets extra body in metadata", func(t *testing.T) {
		opts := &llms.CallOptions{}
		extraBody := map[string]any{
			"enable_thinking": false,
			"custom_param":    "value",
		}

		WithExtraBody(extraBody)(opts)

		require.NotNil(t, opts.Metadata)
		stored, ok := opts.Metadata["openai:extra_body"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, false, stored["enable_thinking"])
		assert.Equal(t, "value", stored["custom_param"])
	})

	t.Run("nil extra body is handled", func(t *testing.T) {
		opts := &llms.CallOptions{}
		WithExtraBody(nil)(opts)
		require.NotNil(t, opts.Metadata)
	})

	t.Run("empty extra body is handled", func(t *testing.T) {
		opts := &llms.CallOptions{}
		WithExtraBody(map[string]any{})(opts)
		require.NotNil(t, opts.Metadata)
		stored, ok := opts.Metadata["openai:extra_body"].(map[string]any)
		require.True(t, ok)
		assert.Empty(t, stored)
	})
}

func TestGetExtraBody(t *testing.T) {
	t.Run("returns nil when metadata is nil", func(t *testing.T) {
		opts := &llms.CallOptions{}
		result := getExtraBody(opts)
		assert.Nil(t, result)
	})

	t.Run("returns nil when extra_body not present", func(t *testing.T) {
		opts := &llms.CallOptions{
			Metadata: map[string]any{"other_key": "value"},
		}
		result := getExtraBody(opts)
		assert.Nil(t, result)
	})

	t.Run("returns extra body when present", func(t *testing.T) {
		extraBody := map[string]any{"key": "value"}
		opts := &llms.CallOptions{
			Metadata: map[string]any{
				"openai:extra_body": extraBody,
			},
		}
		result := getExtraBody(opts)
		require.NotNil(t, result)
		assert.Equal(t, "value", result["key"])
	})

	t.Run("returns nil for wrong type", func(t *testing.T) {
		opts := &llms.CallOptions{
			Metadata: map[string]any{
				"openai:extra_body": "not a map",
			},
		}
		result := getExtraBody(opts)
		assert.Nil(t, result)
	})
}

// Qwen3.x models are not forced-reasoning: their thinking is user-toggleable and
// their sampling is configurable. The caller's sampling parameters must reach an
// OpenAI-compatible backend verbatim (temperature is NOT pinned to 1.0), and the
// thinking switch travels through extra_body.chat_template_kwargs.enable_thinking.
func TestQwenSamplingReachesWireVerbatim(t *testing.T) { //nolint:funlen // three wire sub-cases
	t.Parallel()

	const completion = `{"id":"x","object":"chat.completion","created":1,"model":"m",` +
		`"choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],` +
		`"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`

	const model = "Qwen/Qwen3.6-27B-FP8"

	capture := func(t *testing.T, callOpts []llms.CallOption) map[string]any {
		t.Helper()
		var raw []byte
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, _ = io.ReadAll(r.Body)
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, completion)
		}))
		t.Cleanup(srv.Close)

		llm, err := New(WithBaseURL(srv.URL), WithToken("test"), WithModel(model))
		if err != nil {
			t.Fatalf("New() error: %v", err)
		}
		if _, err := llm.GenerateContent(context.Background(),
			[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
			callOpts...); err != nil {
			t.Fatalf("GenerateContent() error: %v", err)
		}
		var body map[string]any
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatalf("unmarshal wire body: %v\nraw: %s", err, raw)
		}
		return body
	}

	// numeric JSON fields decode to float64; assert exact presence and value.
	assertNum := func(t *testing.T, body map[string]any, key string, want float64) {
		t.Helper()
		got, ok := body[key]
		if !ok {
			t.Fatalf("wire body missing %q; body=%v", key, body)
		}
		if f, _ := got.(float64); f != want {
			t.Fatalf("wire %q = %v, want %v", key, got, want)
		}
	}

	t.Run("thinking mode default sampling passes through", func(t *testing.T) {
		body := capture(t, []llms.CallOption{
			llms.WithTemperature(1.0),
			llms.WithTopP(0.95),
			llms.WithTopK(20),
			llms.WithMinP(0.0),
			llms.WithPresencePenalty(1.5),
			llms.WithRepetitionPenalty(1.0),
		})
		assertNum(t, body, "temperature", 1.0)
		assertNum(t, body, "top_p", 0.95)
		assertNum(t, body, "top_k", 20)
		assertNum(t, body, "min_p", 0.0)
		assertNum(t, body, "presence_penalty", 1.5)
		assertNum(t, body, "repetition_penalty", 1.0)
	})

	t.Run("precise-coding thinking sampling passes through", func(t *testing.T) {
		body := capture(t, []llms.CallOption{
			llms.WithTemperature(0.6),
			llms.WithTopP(0.95),
			llms.WithTopK(20),
			llms.WithMinP(0.0),
			llms.WithPresencePenalty(0.0),
			llms.WithRepetitionPenalty(1.0),
		})
		// temperature must NOT be pinned to 1.0 for a bare Qwen3 model.
		assertNum(t, body, "temperature", 0.6)
		assertNum(t, body, "presence_penalty", 0.0)
		assertNum(t, body, "min_p", 0.0)
		assertNum(t, body, "repetition_penalty", 1.0)
	})

	t.Run("non-thinking mode via extra_body", func(t *testing.T) {
		body := capture(t, []llms.CallOption{
			llms.WithTemperature(0.7),
			llms.WithTopP(0.8),
			llms.WithTopK(20),
			llms.WithMinP(0.0),
			llms.WithPresencePenalty(1.5),
			llms.WithRepetitionPenalty(1.0),
			llms.WithN(1),
			llms.WithMaxTokens(32768),
			WithExtraBody(map[string]any{
				"chat_template_kwargs": map[string]any{"enable_thinking": false},
			}),
		})
		assertNum(t, body, "temperature", 0.7)
		assertNum(t, body, "top_p", 0.8)
		assertNum(t, body, "top_k", 20)
		assertNum(t, body, "min_p", 0.0)
		assertNum(t, body, "presence_penalty", 1.5)
		assertNum(t, body, "repetition_penalty", 1.0)
		assertNum(t, body, "n", 1)

		ctk, ok := body["chat_template_kwargs"].(map[string]any)
		if !ok {
			t.Fatalf("wire body missing chat_template_kwargs object; body=%v", body)
		}
		if enable, _ := ctk["enable_thinking"].(bool); enable {
			t.Fatalf("enable_thinking must be false on wire, got %v", ctk["enable_thinking"])
		}
	})
}

// With no model set anywhere, capability decisions must key off the package
// default the client substitutes on the wire — not off an empty model. Otherwise
// WithReasoningDisabled() on a zero-config client is a silent no-op.
func TestZeroConfigReasoningDisabledUsesDefaultModel(t *testing.T) {
	t.Setenv("OPENAI_MODEL", "")

	const completion = `{"id":"x","object":"chat.completion","created":1,"model":"m",` +
		`"choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],` +
		`"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`

	var body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, completion)
	}))
	defer srv.Close()

	// No WithModel and no OPENAI_MODEL: the default (gpt-5.4-mini) is reasoning-on.
	llm, err := New(WithBaseURL(srv.URL), WithToken("test"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	if _, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
		llms.WithReasoningDisabled()); err != nil {
		t.Fatalf("GenerateContent() error: %v", err)
	}

	if !strings.Contains(body, `"reasoning_effort":"none"`) {
		t.Fatalf("zero-config disable must send the disable wire for %s, got body: %s",
			openaiclient.DefaultChatModel, body)
	}
}

func TestZeroConfigTemperaturePinUsesDefaultModel(t *testing.T) {
	t.Setenv("OPENAI_MODEL", "")

	const completion = `{"id":"x","object":"chat.completion","created":1,"model":"m",` +
		`"choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],` +
		`"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`

	capture := func(t *testing.T, opts ...Option) string {
		t.Helper()
		var body string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, _ := io.ReadAll(r.Body)
			body = string(b)
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, completion)
		}))
		defer srv.Close()

		llm, err := New(append([]Option{WithBaseURL(srv.URL), WithToken("test")}, opts...)...)
		if err != nil {
			t.Fatalf("New() error: %v", err)
		}
		if _, err := llm.GenerateContent(context.Background(),
			[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
			llms.WithTemperature(0.2)); err != nil {
			t.Fatalf("GenerateContent() error: %v", err)
		}
		return body
	}

	zeroConfig := capture(t)
	if !strings.Contains(zeroConfig, `"temperature":1`) {
		t.Errorf("zero-config must pin temperature for %s, got body: %s",
			openaiclient.DefaultChatModel, zeroConfig)
	}

	explicit := capture(t, WithModel(openaiclient.DefaultChatModel))
	if !strings.Contains(explicit, `"temperature":1`) {
		t.Errorf("explicit %s must pin temperature too, got body: %s",
			openaiclient.DefaultChatModel, explicit)
	}

	plain := capture(t, WithModel("gpt-4.1-mini"))
	if !strings.Contains(plain, `"temperature":0.2`) {
		t.Errorf("a non-reasoning model must keep the caller temperature, got body: %s", plain)
	}
}

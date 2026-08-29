package openai

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vxcontrol/langchaingo/llms"
)

func TestARefusalReachesTheCallerAsATypedError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"x","object":"chat.completion","created":1,"model":"gpt-4o",
			"choices":[{"index":0,"message":{"role":"assistant","content":"",
			"refusal":"I can't help with that."},"finish_reason":"stop"}],
			"usage":{"prompt_tokens":11,"completion_tokens":7,"total_tokens":18}}`)
	}))
	t.Cleanup(srv.Close)

	llm := newUnitLLM(t, WithBaseURL(srv.URL), WithModel("gpt-4o"))
	resp, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")})

	var refusal *llms.ErrModelRefusal
	require.ErrorAs(t, err, &refusal)
	assert.Equal(t, "I can't help with that.", refusal.Message)
	assert.Equal(t, 11, refusal.InputTokens)
	assert.Equal(t, 7, refusal.OutputTokens)

	require.NotNil(t, resp, "the response travels with the error so usage survives")
	assert.Equal(t, "I can't help with that.", resp.Choices[0].GenerationInfo["Refusal"])
}

func TestAStructuredOutputRefusalKeepsBothItsTypes(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"x","object":"chat.completion","created":1,"model":"gpt-4o",
			"choices":[{"index":0,"message":{"role":"assistant","content":"",
			"refusal":"I can't help with that."},"finish_reason":"stop"}],
			"usage":{"prompt_tokens":11,"completion_tokens":7,"total_tokens":18}}`)
	}))
	t.Cleanup(srv.Close)

	llm := newUnitLLM(t, WithBaseURL(srv.URL), WithModel("gpt-4o"))
	resp, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
		llms.WithStructuredOutput(llms.StructuredOutputConfig{
			Name:   "answer",
			Schema: []byte(`{"type":"object","additionalProperties":false,"properties":{"a":{"type":"string"}},"required":["a"]}`),
		}))

	var structured *ErrStructuredOutputRefusal
	require.ErrorAs(t, err, &structured,
		"a refusal to a structured-output request must reach the caller as the exported type")
	assert.Equal(t, "gpt-4o", structured.Model)
	assert.Equal(t, 0, structured.Choice)
	assert.Equal(t, "I can't help with that.", structured.Refusal)

	var refusal *llms.ErrModelRefusal
	require.ErrorAs(t, err, &refusal, "the cross-door type must still match, or door parity breaks")
	assert.Equal(t, 11, refusal.InputTokens)

	require.NotNil(t, resp)
}

func TestARefusalReportsTheUncachedInputSeparately(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"x","object":"chat.completion","created":1,"model":"gpt-4o",
			"choices":[{"index":0,"message":{"role":"assistant","content":"",
			"refusal":"I can't help with that."},"finish_reason":"stop"}],
			"usage":{"prompt_tokens":100,"completion_tokens":7,"total_tokens":107,
			"prompt_tokens_details":{"cached_tokens":80}}}`)
	}))
	t.Cleanup(srv.Close)

	llm := newUnitLLM(t, WithBaseURL(srv.URL), WithModel("gpt-4o"))
	_, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")})

	var refusal *llms.ErrModelRefusal
	require.ErrorAs(t, err, &refusal)
	assert.Equal(t, 20, refusal.InputTokens, "InputTokens counts only the uncached input")
	assert.Equal(t, 80, refusal.CacheReadInputTokens)
	assert.Equal(t, 100, refusal.InputTokens+refusal.CacheCreationInputTokens+refusal.CacheReadInputTokens,
		"the three input counts must add up to the billed prompt")
}

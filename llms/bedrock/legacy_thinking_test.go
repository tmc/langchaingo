package bedrock_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/bedrock"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

func legacyLLMCapturing(t *testing.T, body string, opts ...bedrock.Option) (*bedrock.LLM, *string) {
	t.Helper()

	sent := new(string)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		*sent = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)

	return bedrockLLMAgainst(t, srv, opts...), sent
}

func bedrockLLMAgainst(t *testing.T, srv *httptest.Server, opts ...bedrock.Option) *bedrock.LLM {
	t.Helper()

	client := bedrockruntime.NewFromConfig(aws.Config{
		Region:      "us-east-1",
		Credentials: credentials.NewStaticCredentialsProvider("unit", "test", ""),
	}, func(o *bedrockruntime.Options) {
		o.BaseEndpoint = aws.String(srv.URL)
	})

	llm, err := bedrock.New(append([]bedrock.Option{bedrock.WithClient(client)}, opts...)...)
	require.NoError(t, err)
	return llm
}

func TestLegacyBudgetStaysBelowMaxTokens(t *testing.T) {
	t.Parallel()

	const resp = `{"id":"x","type":"message","role":"assistant","model":"m",` +
		`"content":[{"type":"text","text":"ok"}],"stop_reason":"end_turn",` +
		`"usage":{"input_tokens":1,"output_tokens":1}}`

	llm, sent := legacyLLMCapturing(t, resp,
		bedrock.WithModel("anthropic.claude-sonnet-4-5-20250929-v1:0"))

	_, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
		llms.WithMaxTokens(0), llms.WithReasoning(llms.ReasoningLow, 0))
	require.NoError(t, err)

	var got struct {
		MaxTokens int `json:"max_tokens"`
		Thinking  struct {
			BudgetTokens int `json:"budget_tokens"`
		} `json:"thinking"`
	}
	require.NoError(t, json.Unmarshal([]byte(*sent), &got))

	assert.Positive(t, got.Thinking.BudgetTokens, "budget thinking must carry a budget")
	assert.Less(t, got.Thinking.BudgetTokens, got.MaxTokens,
		"a missing max-tokens must be substituted once, not twice with different values")
}

func TestLegacyBudgetThinkingReachesEveryBudgetGeneration(t *testing.T) {
	t.Parallel()

	const resp = `{"id":"x","type":"message","role":"assistant","model":"m",` +
		`"content":[{"type":"text","text":"ok"}],"stop_reason":"end_turn",` +
		`"usage":{"input_tokens":1,"output_tokens":1}}`

	for _, model := range []string{
		"anthropic.claude-3-7-sonnet-20250219-v1:0",
		"us.anthropic.claude-3-7-sonnet-20250219-v1:0",
		"anthropic.claude-opus-4-20250514-v1:0",
		"anthropic.claude-sonnet-4-20250514-v1:0",
		"anthropic.claude-sonnet-4-5-20250929-v1:0",
	} {
		t.Run(model, func(t *testing.T) {
			t.Parallel()

			llm, sent := legacyLLMCapturing(t, resp, bedrock.WithModel(model))

			_, err := llm.GenerateContent(context.Background(),
				[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
				llms.WithReasoning(llms.ReasoningLow, 0))
			require.NoError(t, err)

			var got struct {
				Temperature float64 `json:"temperature"`
				Thinking    *struct {
					Type         string `json:"type"`
					BudgetTokens int    `json:"budget_tokens"`
				} `json:"thinking"`
				OutputConfig *struct {
					Effort string `json:"effort"`
				} `json:"output_config"`
			}
			require.NoError(t, json.Unmarshal([]byte(*sent), &got))

			require.NotNil(t, got.Thinking, "budget generation must receive a thinking block")
			assert.Equal(t, "enabled", got.Thinking.Type)
			assert.Positive(t, got.Thinking.BudgetTokens)
			assert.InDelta(t, 1.0, got.Temperature, 1e-9,
				"budget thinking runs at temperature 1")
			assert.Nil(t, got.OutputConfig,
				"a budget-only generation takes no effort alongside the budget")
		})
	}
}

func writeBedrockChunk(t *testing.T, w io.Writer, enc *eventstream.Encoder, chunk string) {
	t.Helper()

	envelope, err := json.Marshal(map[string]string{
		"bytes": base64.StdEncoding.EncodeToString([]byte(chunk)),
	})
	require.NoError(t, err)

	require.NoError(t, enc.Encode(w, eventstream.Message{
		Headers: eventstream.Headers{
			{Name: ":message-type", Value: eventstream.StringValue("event")},
			{Name: ":event-type", Value: eventstream.StringValue("chunk")},
			{Name: ":content-type", Value: eventstream.StringValue("application/json")},
		},
		Payload: envelope,
	}))
}

func TestLegacyStreamKeepsASignatureOnlyThinkingBlock(t *testing.T) {
	t.Parallel()

	chunks := []string{
		`{"type":"message_start","message":{"usage":{"input_tokens":1,"output_tokens":0}}}`,
		`{"type":"content_block_start","content_block":{"type":"thinking"}}`,
		`{"type":"content_block_delta","delta":{"type":"signature_delta","signature":"sig-only"}}`,
		`{"type":"content_block_stop"}`,
		`{"type":"content_block_delta","delta":{"type":"text_delta","text":"ok"}}`,
		`{"type":"message_delta","delta":{"stop_reason":"end_turn"},"message":{"usage":{"input_tokens":1,"output_tokens":1}}}`,
		`{"type":"message_stop"}`,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/vnd.amazon.eventstream")
		enc := eventstream.NewEncoder()
		for _, chunk := range chunks {
			writeBedrockChunk(t, w, enc, chunk)
		}
	}))
	t.Cleanup(srv.Close)

	llm := bedrockLLMAgainst(t, srv, bedrock.WithModel("anthropic.claude-sonnet-4-5-20250929-v1:0"))

	got, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
		llms.WithStreamingFunc(func(_ context.Context, _ streaming.Chunk) error { return nil }))
	require.NoError(t, err)

	choice := got.Choices[0]
	require.NotNil(t, choice.Reasoning,
		"a thinking block that arrived as a signature alone must survive, as it does off the stream")
	assert.Equal(t, "sig-only", string(choice.Reasoning.Signature))
}

func TestLegacyPromptTokensCountTheCachedInput(t *testing.T) {
	t.Parallel()

	const resp = `{"id":"x","type":"message","role":"assistant","model":"m",` +
		`"content":[{"type":"text","text":"ok"}],"stop_reason":"end_turn",` +
		`"usage":{"input_tokens":13,"output_tokens":4,` +
		`"cache_creation_input_tokens":300,"cache_read_input_tokens":3602}}`

	llm, _ := legacyLLMCapturing(t, resp,
		bedrock.WithModel("anthropic.claude-sonnet-4-5-20250929-v1:0"))

	got, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")})
	require.NoError(t, err)
	require.Len(t, got.Choices, 1)

	info := got.Choices[0].GenerationInfo
	assert.Equal(t, int32(13), info["input_tokens"],
		"the vendor's own field reports the uncached input alone")
	assert.Equal(t, int32(3915), info["PromptTokens"],
		"the standardized field must count the cached input the request was billed for")
	assert.Equal(t, int32(3919), info["TotalTokens"])
}

package bedrock_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/bedrock"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

type familyCase struct {
	name        string
	model       string
	whole       string
	chunks      []string
	wantText    string
	wantReason  string
	hasCounters bool
}

func nonClaudeFamilies() []familyCase {
	return []familyCase{
		{
			name:        "amazon",
			hasCounters: true,
			model:       "amazon.titan-text-express-v1",
			whole: `{"inputTextTokenCount":5,"results":[{"tokenCount":3,"outputText":"sixty rooms",` +
				`"completionReason":"LENGTH"}]}`,
			chunks: []string{
				`{"outputText":"sixty ","index":0}`,
				`{"outputText":"rooms","index":1,"completionReason":"LENGTH"}`,
			},
			wantText:   "sixty rooms",
			wantReason: "LENGTH",
		},
		{
			name:        "meta",
			hasCounters: true,
			model:       "meta.llama3-70b-instruct-v1:0",
			whole:       `{"generation":"sixty rooms","prompt_token_count":5,"generation_token_count":3,"stop_reason":"length"}`,
			chunks:      []string{`{"generation":"sixty "}`, `{"generation":"rooms","stop_reason":"length"}`},
			wantText:    "sixty rooms",
			wantReason:  "length",
		},
		{
			name:  "cohere",
			model: "cohere.command-text-v14",
			whole: `{"id":"x","generations":[{"id":"g","text":"sixty rooms","finish_reason":"MAX_TOKENS"}],` +
				`"text":"","finish_reason":""}`,
			chunks: []string{
				`{"text":"sixty ","is_finished":false}`,
				`{"text":"rooms","finish_reason":"MAX_TOKENS","is_finished":true}`,
			},
			wantText:   "sixty rooms",
			wantReason: "MAX_TOKENS",
		},
		{
			name:        "ai21",
			hasCounters: true,
			model:       "ai21.jamba-1-5-large-v1:0",
			whole: `{"id":"x","choices":[{"index":0,"message":{"role":"assistant","content":"sixty rooms"},` +
				`"finish_reason":"length"}],"usage":{"prompt_tokens":5,"completion_tokens":3,"total_tokens":8}}`,
			chunks: []string{
				`{"text":"sixty ","index":0}`,
				`{"text":"rooms","index":1,"finish_reason":"length"}`,
			},
			wantText:   "sixty rooms",
			wantReason: "length",
		},
		{
			name:       "deepseek",
			model:      "us.deepseek.r1-v1:0",
			whole:      `{"choices":[{"text":"sixty rooms","stop_reason":"length"}]}`,
			chunks:     []string{`{"choices":[{"text":"sixty "}]}`, `{"choices":[{"text":"rooms","stop_reason":"length"}]}`},
			wantText:   "sixty rooms",
			wantReason: "length",
		},
		{
			name:        "nova",
			hasCounters: true,
			model:       "us.amazon.nova-pro-v1:0",
			whole: `{"output":{"message":{"content":[{"text":"sixty rooms"}],"role":"assistant"}},` +
				`"stopReason":"max_tokens","usage":{"inputTokens":5,"outputTokens":3,"totalTokens":8}}`,
			chunks: []string{
				`{"contentBlockDelta":{"delta":{"text":"sixty "}}}`,
				`{"contentBlockDelta":{"delta":{"text":"rooms"}}}`,
				`{"messageDelta":{"stopReason":"max_tokens","usage":{"outputTokens":3}}}`,
			},
			wantText:   "sixty rooms",
			wantReason: "max_tokens",
		},
	}
}

func TestANonClaudeFamilyReportsTruncationAndKeepsItsWholeAnswer(t *testing.T) {
	t.Parallel()

	for _, tc := range nonClaudeFamilies() {
		t.Run(tc.name+"/whole answer", func(t *testing.T) {
			t.Parallel()

			llm := truncationLLMWithBody(t, tc.whole, bedrock.WithModel(tc.model))
			resp, err := llm.GenerateContent(context.Background(),
				[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "how many rooms are free?")})
			require.NoError(t, err)
			require.Len(t, resp.Choices, 1)

			assert.Equal(t, tc.wantText, resp.Choices[0].Content)
			assert.Equal(t, tc.wantReason, resp.Choices[0].StopReason)
			assert.True(t, resp.Choices[0].Truncated,
				"a stop reason meaning the output limit must set Truncated")

			if tc.hasCounters {
				for _, key := range []string{"PromptTokens", "CompletionTokens"} {
					_, ok := resp.Choices[0].GenerationInfo[key].(int)
					assert.True(t, ok, "%s must be an int, as the Claude doors report it, got %#v",
						key, resp.Choices[0].GenerationInfo[key])
				}
			}
		})

		t.Run(tc.name+"/streamed answer", func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.Copy(io.Discard, r.Body)
				w.Header().Set("Content-Type", "application/vnd.amazon.eventstream")
				enc := eventstream.NewEncoder()
				for _, chunk := range tc.chunks {
					writeLegacyChunk(t, w, enc, chunk)
				}
			}))
			t.Cleanup(srv.Close)

			llm := bedrockLLMAgainst(t, srv, bedrock.WithModel(tc.model))
			resp, err := llm.GenerateContent(context.Background(),
				[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "how many rooms are free?")},
				llms.WithStreamingFunc(func(context.Context, streaming.Chunk) error { return nil }))
			require.NoError(t, err)
			require.Len(t, resp.Choices, 1)

			assert.Equal(t, tc.wantText, resp.Choices[0].Content,
				"every delta belongs in the assembled answer")
			assert.True(t, resp.Choices[0].Truncated,
				"the streamed stop reason must set Truncated too")
		})
	}
}

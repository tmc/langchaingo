package bedrockclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

func newRespondingBedrockClient(t *testing.T, responseBody string) *bedrockruntime.Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(responseBody))
	}))
	t.Cleanup(server.Close)

	return bedrockruntime.New(bedrockruntime.Options{
		Region:       "us-east-1",
		BaseEndpoint: aws.String(server.URL),
		Credentials: aws.CredentialsProviderFunc(func(context.Context) (aws.Credentials, error) {
			return aws.Credentials{AccessKeyID: "test", SecretAccessKey: "test"}, nil
		}),
	})
}

func requireRefusalError(t *testing.T, err error, wantCategory, wantExplanation any) {
	t.Helper()
	var llmErr *llms.Error
	require.ErrorAs(t, err, &llmErr)
	assert.Equal(t, llms.ErrCodeContentFilter, llmErr.Code)
	assert.Equal(t, map[string]any{
		DetailType:        BlockedByRefusal,
		DetailCategory:    wantCategory,
		DetailExplanation: wantExplanation,
	}, llmErr.Details)
}

func TestCreateAnthropicCompletionReturnsRefusals(t *testing.T) {
	tests := []struct {
		name            string
		responseBody    string
		wantCategory    any
		wantExplanation any
	}{
		{
			name:            "refusal before any output",
			responseBody:    `{"type":"message","role":"assistant","content":[],"stop_reason":"refusal","stop_details":{"type":"refusal","category":"cyber","explanation":"Declined."}}`,
			wantCategory:    "cyber",
			wantExplanation: "Declined.",
		},
		{
			name:            "refusal after partial output",
			responseBody:    `{"type":"message","role":"assistant","content":[{"type":"text","text":"Sure, first"}],"stop_reason":"refusal","stop_details":{"type":"refusal","category":"general_harms","explanation":"Declined."}}`,
			wantCategory:    "general_harms",
			wantExplanation: "Declined.",
		},
		{
			name:         "refusal without a named category",
			responseBody: `{"type":"message","role":"assistant","content":[],"stop_reason":"refusal","stop_details":{"type":"refusal","category":null,"explanation":null}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newRespondingBedrockClient(t, tt.responseBody)
			messages := []Message{{Role: llms.ChatMessageTypeHuman, Type: "text", Content: "Hi"}}

			resp, err := createAnthropicCompletion(context.Background(), client, "anthropic.claude-sonnet-5-5", messages, llms.CallOptions{})

			assert.Nil(t, resp)
			requireRefusalError(t, err, tt.wantCategory, tt.wantExplanation)
		})
	}
}

func TestHandleAnthropicStreamingChunkReturnsRefusals(t *testing.T) {
	tests := []struct {
		name            string
		event           string
		wantCategory    any
		wantExplanation any
	}{
		{
			name:            "message_delta with stop_details",
			event:           `{"type":"message_delta","delta":{"stop_reason":"refusal","stop_details":{"type":"refusal","category":"bio","explanation":"Declined."}},"usage":{"output_tokens":3}}`,
			wantCategory:    "bio",
			wantExplanation: "Declined.",
		},
		{
			name:  "message_delta without stop_details",
			event: `{"type":"message_delta","delta":{"stop_reason":"refusal"},"usage":{"output_tokens":3}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp streamingCompletionResponseChunk
			require.NoError(t, json.Unmarshal([]byte(tt.event), &resp))
			choice := &llms.ContentChoice{GenerationInfo: map[string]any{}}

			err := handleAnthropicStreamingChunk(context.Background(), resp, choice, llms.CallOptions{})

			requireRefusalError(t, err, tt.wantCategory, tt.wantExplanation)
		})
	}
}

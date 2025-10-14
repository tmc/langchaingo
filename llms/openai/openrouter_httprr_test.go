package openai

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vendasta/langchaingo/internal/httprr"
	"github.com/vendasta/langchaingo/llms"
)

// TestOpenRouterWithHTTPRR tests OpenRouter integration with recorded HTTP responses
func TestOpenRouterWithHTTPRR(t *testing.T) {
	// This test uses httprr to record/replay OpenRouter API responses
	// Run with -httprecord=. and OPENROUTER_API_KEY set to record new responses

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENROUTER_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	// Only run tests in parallel when not recording
	if !rr.Recording() {
		t.Parallel()
	}

	apiKey := "test-api-key"
	if key := os.Getenv("OPENROUTER_API_KEY"); key != "" && rr.Recording() {
		apiKey = key
	}

	// Create OpenAI client configured for OpenRouter
	llm, err := New(
		WithToken(apiKey),
		WithBaseURL("https://openrouter.ai/api/v1"),
		WithModel("meta-llama/llama-3.2-3b-instruct:free"),
		WithHTTPClient(rr.Client()),
	)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("streaming", func(t *testing.T) {
		var chunks []string

		_, err := llm.Call(ctx, "Reply with exactly 'OK' and nothing else",
			llms.WithTemperature(0.0),
			llms.WithMaxTokens(10),
			llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
				chunks = append(chunks, string(chunk))
				return nil
			}),
		)

		// Skip test if rate limited during recording
		if err != nil && strings.Contains(err.Error(), "429") && rr.Recording() {
			t.Skip("Rate limited during recording - this is expected with free tier")
		}

		require.NoError(t, err, "streaming should work with OpenRouter")
		assert.NotEmpty(t, chunks, "should receive streamed chunks")

		// The response should contain OK
		fullResponse := strings.Join(chunks, "")
		assert.Contains(t, strings.ToUpper(fullResponse), "OK", "response should contain 'OK'")
	})
}

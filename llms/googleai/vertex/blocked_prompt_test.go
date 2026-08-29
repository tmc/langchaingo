package vertex

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/googleai"
)

type blockedPrompt struct{}

func (blockedPrompt) RoundTrip(r *http.Request) (*http.Response, error) {
	const body = `{"promptFeedback":{"blockReason":"SAFETY",` +
		`"blockReasonMessage":"the prompt was rejected by the safety filter"}}`
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader([]byte(body))),
		Request:    r,
	}, nil
}

func TestABlockedPromptReachesTheCallerClassified(t *testing.T) {
	t.Setenv("GOOGLE_API_KEY", "unit-test-key")

	v, err := New(context.Background(),
		googleai.WithRest(), googleai.WithAPIKey("unit-test-key"),
		googleai.WithCloudProject("unit"), googleai.WithCloudLocation("us-central1"),
		googleai.WithDefaultModel("gemini-2.5-flash"),
		googleai.WithHTTPClient(&http.Client{Transport: blockedPrompt{}}))
	require.NoError(t, err)
	t.Cleanup(func() { _ = v.Close() })

	_, err = v.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")})

	var apiErr *llms.Error
	require.ErrorAs(t, err, &apiErr, "a blocked prompt must reach the caller in this project's terms")
	assert.Equal(t, llms.ErrCodeContentFilter, apiErr.Code)
}
